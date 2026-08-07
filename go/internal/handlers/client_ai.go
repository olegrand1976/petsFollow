package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

const (
	clientAITriageMaxMsgsPerSession = 20
	clientAITriageMaxBodyRunes      = 2000
	clientAIExplainMaxBodyRunes     = 50000
)

// testClientExplain overrides Gemini explain in integration tests.
var testClientExplain func(ctx context.Context, in gemini.ClientExplainInput) (*gemini.ClientExplainResult, error)

// testClientTriage overrides Gemini triage in integration tests.
var testClientTriage func(ctx context.Context, in gemini.ClientTriageInput) (*gemini.ClientTriageTurn, error)

// TestSetClientExplain installs a mock explain generator (integration tests only).
func TestSetClientExplain(fn func(ctx context.Context, in gemini.ClientExplainInput) (*gemini.ClientExplainResult, error)) {
	testClientExplain = fn
}

// TestClearClientExplain clears the mock explain generator.
func TestClearClientExplain() { testClientExplain = nil }

// TestSetClientTriage installs a mock triage generator (integration tests only).
func TestSetClientTriage(fn func(ctx context.Context, in gemini.ClientTriageInput) (*gemini.ClientTriageTurn, error)) {
	testClientTriage = fn
}

// TestClearClientTriage clears the mock triage generator.
func TestClearClientTriage() { testClientTriage = nil }

func (a *API) requireClientAIEnabled(w http.ResponseWriter, r *http.Request) bool {
	if !a.cfg.ClientAIEnabled {
		writeErr(w, r, http.StatusNotFound, "client_ai_disabled", "client_ai_disabled")
		return false
	}
	return true
}

func (a *API) requireClientAIClient(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	if !a.requireClientAIEnabled(w, r) {
		return authx.Identity{}, false
	}
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return authx.Identity{}, false
	}
	return id, true
}

func hashExplainSource(bodies []string) string {
	h := sha256.New()
	for _, b := range bodies {
		h.Write([]byte(b))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (a *API) getClientConsultationExplain(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireClientAIClient(w, r)
	if !ok {
		return
	}
	visitID := chi.URLParam(r, "visitID")
	refresh := r.URL.Query().Get("refresh") == "1"

	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	pet, err := a.store.GetPet(r.Context(), visit.PetID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	if pet.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	active, err := a.store.HasActiveEntitlement(r.Context(), pet.ID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !active {
		writeErr(w, r, http.StatusPaymentRequired, "payment_required", "pet_inactive")
		return
	}

	reports, err := a.store.ListFinalVisitReportsForClient(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if len(reports) == 0 {
		writeErr(w, r, http.StatusNotFound, "not_found", "consultation_not_found")
		return
	}

	bodies := make([]string, 0, len(reports))
	totalRunes := 0
	for _, rep := range reports {
		b := strings.TrimSpace(rep.BodyText)
		if b == "" {
			continue
		}
		totalRunes += utf8.RuneCountInString(b)
		bodies = append(bodies, b)
	}
	if len(bodies) == 0 {
		writeErr(w, r, http.StatusNotFound, "not_found", "consultation_not_found")
		return
	}
	if totalRunes > clientAIExplainMaxBodyRunes {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "body_too_long")
		return
	}

	locale := "fr"
	if u, uerr := a.store.GetUserByID(r.Context(), id.UserID); uerr == nil {
		locale = i18n.NormalizeLocale(u.PreferredLocale)
	}
	sourceHash := hashExplainSource(bodies)

	cached, cerr := a.store.GetVisitReportExplanation(r.Context(), visitID, locale)
	if cerr == nil && cached.SourceHash == sourceHash && !refresh {
		var payload any
		_ = json.Unmarshal(cached.PayloadJSON, &payload)
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"visitId":    visitID,
			"locale":     locale,
			"cached":     true,
			"sourceHash": sourceHash,
			"disclaimer": payloadField(payload, "disclaimer"),
			"cards":      payloadField(payload, "cards"),
		})
		return
	}

	if refresh && cerr == nil && cached.RefreshedAt != nil {
		if time.Since(*cached.RefreshedAt) < 24*time.Hour {
			writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "explain_refresh_quota")
			return
		}
	}

	// Rate-limit Gemini generations only (cache hits already returned above).
	if a.clientAiExplainRL != nil && !a.clientAiExplainRL.Allow(id.UserID) {
		writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "explain_rate_limited")
		return
	}

	result, genErr := a.generateClientExplain(r.Context(), gemini.ClientExplainInput{
		Locale:    locale,
		PetName:   pet.Name,
		Species:   pet.Species,
		BodyTexts: bodies,
	})
	if genErr != nil {
		a.store.InsertClientAIUsageEvent(r.Context(), id.UserID, "error", map[string]any{"kind": "explain"})
		if errors.Is(genErr, errGeminiNotConfigured) || strings.Contains(genErr.Error(), "gemini_not_configured") {
			writeErr(w, r, http.StatusServiceUnavailable, "gemini_not_configured", "gemini_not_configured")
			return
		}
		writeErr(w, r, http.StatusBadGateway, "gemini_error", "gemini_error")
		return
	}

	payloadBytes, _ := json.Marshal(result)
	model := a.cfg.GeminiLiteModel
	if model == "" {
		model = "gemini-lite"
	}
	saved, uerr := a.store.UpsertVisitReportExplanation(r.Context(), visitID, locale, sourceHash, model, payloadBytes, refresh)
	if uerr != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	a.store.InsertClientAIUsageEvent(r.Context(), id.UserID, "explain", map[string]any{
		"visitId": visitID,
		"cached":  false,
		"refresh": refresh,
	})

	httpx.WriteData(w, http.StatusOK, map[string]any{
		"visitId":    visitID,
		"locale":     locale,
		"cached":     false,
		"sourceHash": saved.SourceHash,
		"disclaimer": result.Disclaimer,
		"cards":      result.Cards,
	})
}

var errGeminiNotConfigured = errors.New("gemini_not_configured")

func (a *API) generateClientExplain(ctx context.Context, in gemini.ClientExplainInput) (*gemini.ClientExplainResult, error) {
	if testClientExplain != nil {
		return testClientExplain(ctx, in)
	}
	if a.gemini == nil || !a.gemini.Configured() {
		return nil, errGeminiNotConfigured
	}
	return a.gemini.ExplainVisitReportForClient(ctx, in)
}

func (a *API) generateClientTriage(ctx context.Context, in gemini.ClientTriageInput) (*gemini.ClientTriageTurn, error) {
	if testClientTriage != nil {
		return testClientTriage(ctx, in)
	}
	if a.gemini == nil || !a.gemini.Configured() {
		return nil, errGeminiNotConfigured
	}
	return a.gemini.TriageClientMessage(ctx, in)
}

func payloadField(payload any, key string) any {
	m, ok := payload.(map[string]any)
	if !ok {
		return nil
	}
	return m[key]
}

type createTriageSessionReq struct {
	PetID string `json:"petId"`
}

func (a *API) createClientAITriageSession(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireClientAIClient(w, r)
	if !ok {
		return
	}
	var req createTriageSessionReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	petID := strings.TrimSpace(req.PetID)
	practiceID := ""
	petName := ""
	species := ""

	if petID != "" {
		pet, err := a.store.GetPet(r.Context(), petID)
		if err != nil {
			writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
			return
		}
		if pet.OwnerUserID != id.UserID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
			return
		}
		active, err := a.store.HasActiveEntitlement(r.Context(), pet.ID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if !active {
			writeErr(w, r, http.StatusPaymentRequired, "payment_required", "pet_inactive")
			return
		}
		practiceID = pet.PracticeID
		petName = pet.Name
		species = pet.Species
	}

	sess, err := a.store.CreateClientAITriageSession(r.Context(), id.UserID, petID, practiceID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	escalation := a.buildTriageEscalation(r.Context(), petID, practiceID)
	httpx.WriteData(w, http.StatusCreated, map[string]any{
		"session":    sess,
		"petName":    petName,
		"species":    species,
		"escalation": escalation,
		"messages":   []any{},
	})
}

func (a *API) getClientAITriageSession(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireClientAIClient(w, r)
	if !ok {
		return
	}
	sessionID := chi.URLParam(r, "sessionID")
	sess, err := a.store.GetClientAITriageSession(r.Context(), sessionID, id.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	msgs, err := a.store.ListClientAITriageMessages(r.Context(), sessionID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	petName, species := "", ""
	if sess.PetID != "" {
		if pet, err := a.store.GetPet(r.Context(), sess.PetID); err == nil {
			petName, species = pet.Name, pet.Species
		}
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"session":    sess,
		"petName":    petName,
		"species":    species,
		"escalation": a.buildTriageEscalation(r.Context(), sess.PetID, sess.PracticeID),
		"messages":   msgs,
	})
}

type triageMessageReq struct {
	Body string `json:"body"`
}

func (a *API) postClientAITriageMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireClientAIClient(w, r)
	if !ok {
		return
	}
	if a.clientAiTriageRL != nil && !a.clientAiTriageRL.Allow(id.UserID) {
		writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "triage_rate_limited")
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	sess, err := a.store.GetClientAITriageSession(r.Context(), sessionID, id.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	count, err := a.store.CountClientAITriageMessages(r.Context(), sessionID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// User + assistant are inserted together — reserve 2 slots.
	if count+2 > clientAITriageMaxMsgsPerSession {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "triage_session_full")
		return
	}

	var req triageMessageReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	body := strings.TrimSpace(req.Body)
	if body == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "body_required")
		return
	}
	if utf8.RuneCountInString(body) > clientAITriageMaxBodyRunes {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "body_too_long")
		return
	}
	if strings.ContainsAny(body, "<>") {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "html_not_allowed")
		return
	}

	history, err := a.store.ListClientAITriageMessages(r.Context(), sessionID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	hist := make([]gemini.ClientTriageMessage, 0, len(history))
	for _, m := range history {
		hist = append(hist, gemini.ClientTriageMessage{Role: m.Role, Body: m.Body})
	}

	locale := "fr"
	if u, uerr := a.store.GetUserByID(r.Context(), id.UserID); uerr == nil {
		locale = i18n.NormalizeLocale(u.PreferredLocale)
	}
	petName, species := "", ""
	if sess.PetID != "" {
		if pet, err := a.store.GetPet(r.Context(), sess.PetID); err == nil {
			petName, species = pet.Name, pet.Species
		}
	}

	// Generate before persist so a Gemini failure never leaves an orphan user message.
	turn, genErr := a.generateClientTriage(r.Context(), gemini.ClientTriageInput{
		Locale:   locale,
		PetName:  petName,
		Species:  species,
		History:  hist,
		UserText: body,
	})
	if genErr != nil {
		a.store.InsertClientAIUsageEvent(r.Context(), id.UserID, "error", map[string]any{"kind": "triage_message"})
		if errors.Is(genErr, errGeminiNotConfigured) || strings.Contains(genErr.Error(), "gemini_not_configured") {
			writeErr(w, r, http.StatusServiceUnavailable, "gemini_not_configured", "gemini_not_configured")
			return
		}
		writeErr(w, r, http.StatusBadGateway, "gemini_error", "gemini_error")
		return
	}

	userMsg, asst, err := a.store.InsertClientAITriageTurn(
		r.Context(), sessionID, body, turn.Reply, string(turn.Level), turn.WatchSigns, turn.RecommendedAction,
	)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	a.store.InsertClientAIUsageEvent(r.Context(), id.UserID, "triage_message", map[string]any{
		"level": string(turn.Level),
	})

	httpx.WriteData(w, http.StatusOK, map[string]any{
		"userMessage":       userMsg,
		"assistantMessage":  asst,
		"level":             turn.Level,
		"watchSigns":        turn.WatchSigns,
		"recommendedAction": turn.RecommendedAction,
		"escalation":        a.buildTriageEscalation(r.Context(), sess.PetID, sess.PracticeID),
	})
}

func (a *API) buildTriageEscalation(ctx context.Context, petID, practiceID string) map[string]any {
	phone := ""
	practiceName := ""
	if practiceID != "" {
		if contact, err := a.store.GetPracticeContact(ctx, practiceID); err == nil {
			phone = strings.TrimSpace(contact.Phone)
			practiceName = strings.TrimSpace(contact.PracticeName)
		}
	}
	return map[string]any{
		"petId":            petID,
		"practiceId":       practiceID,
		"practicePhone":    phone,
		"practiceName":     practiceName,
		"canMessage":       true, // always offer messaging (threads list if no pet)
		"canBookVisit":     petID != "",
		"hasPracticePhone": phone != "",
	}
}
