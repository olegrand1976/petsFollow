package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerPitchTrainingRoutes(r chi.Router) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)

		pr.Get("/commercial/pitch-scripts", a.commercialListPitchScripts)
		pr.Post("/commercial/pitch-scripts/{id}/personalize", a.commercialPersonalizePitchScript)
		pr.Patch("/commercial/pitch-scripts/{id}", a.commercialPatchPitchScript)
		pr.Delete("/commercial/pitch-scripts/{id}", a.commercialDeletePitchScript)

		pr.Get("/commercial/pitch-sims/skip-quota", a.commercialPitchSkipQuota)
		pr.Get("/commercial/pitch-sims", a.commercialListPitchSims)
		pr.Post("/commercial/pitch-sims", a.commercialStartPitchSim)
		pr.Post("/commercial/pitch-sims/{id}/turn", a.commercialPitchSimTurn)
		pr.Post("/commercial/pitch-sims/{id}/finalize", a.commercialFinalizePitchSim)
		pr.Post("/commercial/pitch-sims/{id}/audio", a.commercialUploadPitchAudio)
		pr.Patch("/commercial/pitch-sims/{id}/rating", a.commercialRatePitchSim)
		pr.Post("/commercial/pitch-sims/{id}/feedback", a.commercialPitchSimFeedback)

		pr.Get("/commercial-manager/pitch-sims", a.managerListPitchSims)

		pr.Get("/admin/pitch-scripts", a.adminListPitchScripts)
		pr.Post("/admin/pitch-scripts", a.adminCreatePitchScript)
		pr.Patch("/admin/pitch-scripts/{id}", a.adminPatchPitchScript)
		pr.Get("/admin/agent-prompts/{kind}/versions", a.adminListAgentPromptVersions)
		pr.Post("/admin/agent-prompts/{kind}/versions", a.adminCreateAgentPromptVersion)
		pr.Post("/admin/agent-prompts/versions/{id}/restore", a.adminRestoreAgentPromptVersion)
		pr.Get("/admin/pitch-analyzer/runs", a.adminListAnalyzerRuns)
		pr.Get("/admin/pitch-feedback", a.adminListPitchFeedback)
	})

	r.Post("/internal/pitch-analyzer/run", a.internalRunPitchAnalyzer)
}

func normalizeInterest(level string) string {
	level = strings.TrimSpace(strings.ToLower(level))
	// Strip combining accents for interesse / intéressé
	level = strings.Map(func(r rune) rune {
		switch r {
		case 'é', 'è', 'ê', 'ë':
			return 'e'
		default:
			return r
		}
	}, level)
	switch level {
	case "hostile", "sceptique", "neutre", "interesse", "chaud":
		return level
	default:
		return "neutre"
	}
}

func (a *API) enrichPitchSim(sim *store.PitchSimulation) {
	if sim == nil {
		return
	}
	sim.AudioURL = media.PublicURL(a.cfg, sim.AudioObjectKey)
}

func (a *API) commercialListPitchScripts(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialOrManager(w, r)
	if !ok {
		return
	}
	list, err := a.store.ListPitchScriptsForUser(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, list)
}

func (a *API) commercialPersonalizePitchScript(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialOrManager(w, r)
	if !ok {
		return
	}
	sc, err := a.store.PersonalizePitchScript(r.Context(), chi.URLParam(r, "id"), id.UserID)
	if err != nil {
		if errors.Is(err, store.ErrPitchScriptNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, sc)
}

func (a *API) commercialPatchPitchScript(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialOrManager(w, r)
	if !ok {
		return
	}
	existing, err := a.store.GetPitchScript(r.Context(), chi.URLParam(r, "id"))
	if err != nil || existing.OwnerUserID == nil || *existing.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	var body struct {
		Title               string          `json:"title"`
		Steps               json.RawMessage `json:"steps"`
		ExampleDialogue     json.RawMessage `json:"exampleDialogue"`
		CoachHints          string          `json:"coachHints"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "bad_json")
		return
	}
	if body.Title != "" {
		existing.Title = body.Title
	}
	if len(body.Steps) > 0 {
		existing.StepsJSON = body.Steps
	}
	if len(body.ExampleDialogue) > 0 {
		existing.ExampleDialogueJSON = body.ExampleDialogue
	}
	if body.CoachHints != "" {
		existing.CoachHints = body.CoachHints
	}
	if err := a.store.UpdatePitchScript(r.Context(), existing); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, existing)
}

func (a *API) commercialDeletePitchScript(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialOrManager(w, r)
	if !ok {
		return
	}
	if err := a.store.DeletePitchScript(r.Context(), chi.URLParam(r, "id"), id.UserID); err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) commercialListPitchSims(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialOrManager(w, r)
	if !ok {
		return
	}
	list, err := a.store.ListPitchSimulations(r.Context(), id.UserID, false)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	for i := range list {
		a.enrichPitchSim(&list[i])
	}
	httpx.WriteData(w, http.StatusOK, list)
}

func (a *API) commercialStartPitchSim(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialOrManager(w, r)
	if !ok {
		return
	}
	var body struct {
		ScriptID      string `json:"scriptId"`
		InterestLevel string `json:"interestLevel"`
		VoiceName     string `json:"voiceName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "bad_json")
		return
	}
	if body.ScriptID == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "script_required")
		return
	}
	if err := a.store.CanAccessPitchScript(r.Context(), body.ScriptID, id.UserID); err != nil {
		if errors.Is(err, store.ErrPitchScriptForbidden) {
			writeErr(w, r, http.StatusForbidden, "forbidden", "script_forbidden")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	vetPrompt, err := a.store.GetCurrentAgentPrompt(r.Context(), "vet_live")
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "vet_prompt_missing")
		return
	}
	voice := strings.TrimSpace(body.VoiceName)
	if voice == "" {
		voice = "Charon"
	}
	interest := normalizeInterest(body.InterestLevel)
	sim := store.PitchSimulation{
		UserID:             id.UserID,
		ScriptID:           body.ScriptID,
		InterestLevel:      interest,
		VoiceName:          voice,
		VetPromptVersionID: &vetPrompt.ID,
		Outcome:            "in_progress",
		TranscriptJSON:     json.RawMessage(`[{"role":"vet","text":"Allo ?"}]`),
	}
	sim, err = a.store.CreatePitchSimulation(r.Context(), sim)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, map[string]any{
		"simulation": sim,
		"vetOpening": "Allo ?",
		"mode":       "turn_based",
		"maxSeconds": int(store.PitchMaxCallDuration.Seconds()),
	})
}

func (a *API) commercialPitchSimTurn(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialOrManager(w, r)
	if !ok {
		return
	}
	simID := chi.URLParam(r, "id")
	sim, err := a.store.GetPitchSimulation(r.Context(), simID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if sim.Outcome != "in_progress" {
		writeErr(w, r, http.StatusConflict, "conflict", "already_ended")
		return
	}
	if a.store.PitchCallTimedOut(sim) {
		now := time.Now().UTC()
		sim.Outcome = "timeout"
		sim.EndedAt = &now
		sim.DurationSec = int(store.PitchMaxCallDuration.Seconds())
		if err := a.store.FinalizePitchSimulation(r.Context(), sim); err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		var history []map[string]string
		_ = json.Unmarshal(sim.TranscriptJSON, &history)
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"reply":      "",
			"action":     "timeout",
			"ended":      true,
			"outcome":    "timeout",
			"transcript": history,
		})
		return
	}
	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Text) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "text_required")
		return
	}
	if a.gemini == nil || !a.gemini.Configured() {
		writeErr(w, r, http.StatusServiceUnavailable, "gemini_not_configured", "gemini_not_configured")
		return
	}
	vetPrompt, err := a.store.GetAgentPromptVersion(r.Context(), *sim.VetPromptVersionID)
	if err != nil {
		vetPrompt, err = a.store.GetCurrentAgentPrompt(r.Context(), "vet_live")
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "vet_prompt_missing")
			return
		}
	}
	var history []map[string]string
	_ = json.Unmarshal(sim.TranscriptJSON, &history)
	history = append(history, map[string]string{"role": "commercial", "text": strings.TrimSpace(body.Text)})

	system := gemini.BuildVetSystemPrompt(vetPrompt.ContentJSON, sim.InterestLevel)
	turn, err := a.gemini.VetTurn(r.Context(), system, history, body.Text)
	if err != nil {
		writeErr(w, r, http.StatusBadGateway, "gemini_failed", "gemini_failed")
		return
	}
	history = append(history, map[string]string{"role": "vet", "text": turn.Reply})
	tr, _ := json.Marshal(history)
	sim.TranscriptJSON = tr

	ended := false
	outcome := "in_progress"
	slot := ""
	switch turn.Action {
	case "book_appointment":
		ended = true
		outcome = "appointment"
		slot = turn.AppointmentSlot
		if slot == "" {
			slot = "Créneau démo proposé"
		}
	case "hang_up_not_interested":
		ended = true
		outcome = "hangup"
	}
	if ended {
		now := time.Now().UTC()
		sim.Outcome = outcome
		sim.AppointmentSlot = slot
		sim.EndedAt = &now
		sim.DurationSec = int(now.Sub(sim.CreatedAt).Seconds())
		sim.TranscriptJSON = tr
		if err := a.store.FinalizePitchSimulation(r.Context(), sim); err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
	} else {
		if err := a.store.UpdatePitchSimulationTranscript(r.Context(), sim.ID, sim.UserID, tr); err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
	}

	httpx.WriteData(w, http.StatusOK, map[string]any{
		"reply":           turn.Reply,
		"action":          turn.Action,
		"appointmentSlot": slot,
		"reason":          turn.Reason,
		"ended":           ended,
		"outcome":         outcome,
		"transcript":      history,
	})
}

func (a *API) commercialFinalizePitchSim(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialOrManager(w, r)
	if !ok {
		return
	}
	simID := chi.URLParam(r, "id")
	sim, err := a.store.GetPitchSimulation(r.Context(), simID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	var body struct {
		Outcome     string          `json:"outcome"`
		DurationSec int             `json:"durationSec"`
		Transcript  json.RawMessage `json:"transcript"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	// Idempotent: already coached → return as-is.
	if sim.Outcome != "in_progress" && len(sim.CoachFeedbackJSON) > 0 {
		a.enrichPitchSim(&sim)
		httpx.WriteData(w, http.StatusOK, sim)
		return
	}

	if sim.Outcome == "in_progress" {
		outcome := body.Outcome
		if a.store.PitchCallTimedOut(sim) {
			outcome = "timeout"
		}
		if outcome == "" || outcome == "in_progress" {
			outcome = "manual"
		}
		now := time.Now().UTC()
		sim.Outcome = outcome
		sim.EndedAt = &now
		if body.DurationSec > 0 {
			sim.DurationSec = body.DurationSec
		} else {
			sim.DurationSec = int(now.Sub(sim.CreatedAt).Seconds())
		}
		if sim.DurationSec > int(store.PitchMaxCallDuration.Seconds()) {
			sim.DurationSec = int(store.PitchMaxCallDuration.Seconds())
		}
		if len(body.Transcript) > 0 {
			sim.TranscriptJSON = body.Transcript
		}
	}

	if len(sim.CoachFeedbackJSON) == 0 {
		coachPrompt, err := a.store.GetCurrentAgentPrompt(r.Context(), "coach")
		if err == nil {
			sim.CoachPromptVersionID = &coachPrompt.ID
		}
		script, _ := a.store.GetPitchScript(r.Context(), sim.ScriptID)

		if a.gemini != nil && a.gemini.Configured() && coachPrompt.ID != "" {
			res, err := a.gemini.CoachCall(r.Context(), coachPrompt.ContentJSON, script.CoachHints, sim.InterestLevel, sim.Outcome, sim.TranscriptJSON)
			if err == nil && res != nil {
				fb, _ := json.Marshal(res)
				sim.CoachFeedbackJSON = fb
				score := res.Score
				sim.AIScore = &score
			}
		}
		if sim.CoachFeedbackJSON == nil {
			fb := map[string]any{
				"score":          5.0,
				"dimensions":     map[string]float64{"opener": 5, "listening": 5, "objections": 5, "offerClarity": 5, "cta": 5},
				"strengths":      []string{"Appel terminé — activez GEMINI_API_KEY pour un coaching détaillé."},
				"improvements":   []string{"Rejouer l’appel avec le script étape par étape."},
				"coachingTips":   []string{"Toujours terminer par un créneau concret."},
				"scriptCoverage": []string{},
			}
			raw, _ := json.Marshal(fb)
			sim.CoachFeedbackJSON = raw
			s := 5.0
			sim.AIScore = &s
		}
	}

	if err := a.store.FinalizePitchSimulation(r.Context(), sim); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	_ = a.store.RecalcPitchTop5(r.Context(), id.UserID)
	sim, _ = a.store.GetPitchSimulation(r.Context(), simID, id.UserID)
	a.enrichPitchSim(&sim)
	httpx.WriteData(w, http.StatusOK, sim)
}

func (a *API) commercialUploadPitchAudio(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialOrManager(w, r)
	if !ok {
		return
	}
	simID := chi.URLParam(r, "id")
	if _, err := a.store.GetPitchSimulation(r.Context(), simID, id.UserID); err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if err := r.ParseMultipartForm(media.MaxPitchAudioBytes + (1 << 20)); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "multipart")
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_required")
		return
	}
	defer file.Close()
	ct, err := media.NormalizePitchAudioType(hdr.Header.Get("Content-Type"), hdr.Filename)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_type")
		return
	}
	if err := media.ValidateSizeLimit(hdr.Size, media.MaxPitchAudioBytes); err != nil {
		writeErr(w, r, http.StatusRequestEntityTooLarge, "too_large", "too_large")
		return
	}
	ext, _ := media.ExtForPitchAudio(ct)
	key := media.ObjectKey("pitch-sims", simID, ext)
	url, err := a.media.Upload(r.Context(), key, file, hdr.Size, ct)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "upload_failed")
		return
	}
	if err := a.store.SetPitchSimulationAudio(r.Context(), simID, id.UserID, key); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"url": url, "objectKey": key})
}

func (a *API) commercialRatePitchSim(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialOrManager(w, r)
	if !ok {
		return
	}
	var body struct {
		Score float64 `json:"score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Score < 0 || body.Score > 10 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "score_0_10")
		return
	}
	if err := a.store.SetPitchSimulationUserScore(r.Context(), chi.URLParam(r, "id"), id.UserID, body.Score); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) commercialPitchSimFeedback(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialOrManager(w, r)
	if !ok {
		return
	}
	simID := chi.URLParam(r, "id")
	if _, err := a.store.GetPitchSimulation(r.Context(), simID, id.UserID); err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	var body struct {
		VetRealism      int             `json:"vetRealism"`
		CoachUsefulness int             `json:"coachUsefulness"`
		DifficultyFelt  string          `json:"difficultyFelt"`
		Comment         string          `json:"comment"`
		Flags           json.RawMessage `json:"flags"`
		Skip            bool            `json:"skip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "bad_json")
		return
	}
	if body.Skip {
		n, err := a.store.CountFeedbackSkipsToday(r.Context(), id.UserID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if n >= 1 {
			writeErr(w, r, http.StatusConflict, "conflict", "skip_quota_exceeded")
			return
		}
		if err := a.store.MarkPitchFeedbackSkipped(r.Context(), simID, id.UserID); err != nil {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{"skipped": true})
		return
	}
	if body.VetRealism < 1 || body.VetRealism > 5 || body.CoachUsefulness < 1 || body.CoachUsefulness > 5 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "ratings_1_5")
		return
	}
	felt := body.DifficultyFelt
	if felt == "" {
		felt = "ok"
	}
	fb, err := a.store.UpsertPitchSimFeedback(r.Context(), store.PitchSimFeedback{
		SimulationID:    simID,
		UserID:          id.UserID,
		VetRealism:      body.VetRealism,
		CoachUsefulness: body.CoachUsefulness,
		DifficultyFelt:  felt,
		Comment:         body.Comment,
		Flags:           body.Flags,
	})
	if errors.Is(err, store.ErrPitchFeedbackLocked) {
		writeErr(w, r, http.StatusConflict, "conflict", "feedback_locked")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, fb)
}

func (a *API) commercialPitchSkipQuota(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialOrManager(w, r)
	if !ok {
		return
	}
	n, err := a.store.CountFeedbackSkipsToday(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"skipsToday": n, "maxSkips": 1, "canSkip": n < 1})
}

func (a *API) managerListPitchSims(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercialManager(w, r)
	if !ok {
		return
	}
	list, err := a.store.ListTeamPitchSimulations(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	for i := range list {
		a.enrichPitchSim(&list[i])
	}
	httpx.WriteData(w, http.StatusOK, list)
}

func (a *API) adminListPitchScripts(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	list, err := a.store.ListAdminPitchScripts(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, list)
}

func (a *API) adminCreatePitchScript(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	var body store.PitchScript
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "bad_json")
		return
	}
	body.OwnerUserID = nil
	body.IsActive = true
	if body.Audience == "" {
		body.Audience = "vet"
	}
	if body.Locale == "" {
		body.Locale = "fr"
	}
	sc, err := a.store.CreatePitchScript(r.Context(), body)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, sc)
}

func (a *API) adminPatchPitchScript(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	existing, err := a.store.GetPitchScript(r.Context(), chi.URLParam(r, "id"))
	if err != nil || existing.OwnerUserID != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	var body struct {
		Title           string          `json:"title"`
		Steps           json.RawMessage `json:"steps"`
		ExampleDialogue json.RawMessage `json:"exampleDialogue"`
		CoachHints      string          `json:"coachHints"`
		IsActive        *bool           `json:"isActive"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "bad_json")
		return
	}
	if body.Title != "" {
		existing.Title = body.Title
	}
	if len(body.Steps) > 0 {
		existing.StepsJSON = body.Steps
	}
	if len(body.ExampleDialogue) > 0 {
		existing.ExampleDialogueJSON = body.ExampleDialogue
	}
	if body.CoachHints != "" {
		existing.CoachHints = body.CoachHints
	}
	if body.IsActive != nil {
		existing.IsActive = *body.IsActive
	}
	if err := a.store.UpdatePitchScript(r.Context(), existing); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, existing)
}

func (a *API) adminListAgentPromptVersions(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	kind := chi.URLParam(r, "kind")
	if kind != "vet_live" && kind != "coach" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "kind")
		return
	}
	list, err := a.store.ListAgentPromptVersions(r.Context(), kind)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, list)
}

func (a *API) adminCreateAgentPromptVersion(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	kind := chi.URLParam(r, "kind")
	var body struct {
		Content   json.RawMessage `json:"content"`
		Changelog string          `json:"changelog"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Content) == 0 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "bad_json")
		return
	}
	by := id.UserID
	v, err := a.store.CreateAgentPromptVersion(r.Context(), kind, body.Content, body.Changelog, "admin", &by, nil, true)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, v)
}

func (a *API) adminRestoreAgentPromptVersion(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	v, err := a.store.RestoreAgentPromptVersion(r.Context(), chi.URLParam(r, "id"), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	httpx.WriteData(w, http.StatusOK, v)
}

func (a *API) adminListAnalyzerRuns(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	list, err := a.store.ListAnalyzerRuns(r.Context(), 50)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, list)
}

func (a *API) adminListPitchFeedback(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	list, err := a.store.ListRecentPitchFeedback(r.Context(), 100)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, list)
}

func (a *API) internalRunPitchAnalyzer(w http.ResponseWriter, r *http.Request) {
	secret := a.cfg.PitchAnalyzerSecret
	if secret == "" || r.Header.Get("X-Pitch-Analyzer-Secret") != secret {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	run, err := a.store.CreateAnalyzerRun(r.Context(), store.PitchAnalyzerRun{Status: "noop"})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	feedbacks, err := a.store.ListUnprocessedPitchFeedback(r.Context(), 50)
	if err != nil {
		_ = a.store.FinishAnalyzerRun(r.Context(), store.PitchAnalyzerRun{ID: run.ID, Status: "failed", FeedbackCount: 0})
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	summary := map[string]any{"count": len(feedbacks), "items": feedbacks}
	sumJSON, _ := json.Marshal(summary)
	run.InputSummaryJSON = sumJSON
	run.FeedbackCount = len(feedbacks)

	if len(feedbacks) < 3 {
		run.Status = "noop"
		out, _ := json.Marshal(map[string]string{"noOpReason": "insufficient_feedback"})
		run.OutputJSON = out
		_ = a.store.FinishAnalyzerRun(r.Context(), run)
		httpx.WriteData(w, http.StatusOK, run)
		return
	}
	if a.gemini == nil || !a.gemini.Configured() {
		run.Status = "failed"
		out, _ := json.Marshal(map[string]string{"error": "gemini_not_configured"})
		run.OutputJSON = out
		_ = a.store.FinishAnalyzerRun(r.Context(), run)
		httpx.WriteData(w, http.StatusOK, run)
		return
	}
	vetP, _ := a.store.GetCurrentAgentPrompt(r.Context(), "vet_live")
	coachP, _ := a.store.GetCurrentAgentPrompt(r.Context(), "coach")
	res, err := a.gemini.AnalyzeFeedback(r.Context(), vetP.ContentJSON, coachP.ContentJSON, string(sumJSON))
	if err != nil {
		run.Status = "failed"
		out, _ := json.Marshal(map[string]string{"error": err.Error()})
		run.OutputJSON = out
		_ = a.store.FinishAnalyzerRun(r.Context(), run)
		httpx.WriteData(w, http.StatusOK, run)
		return
	}
	out, _ := json.Marshal(res)
	run.OutputJSON = out
	runID := run.ID
	applied := false
	if res.VetChanges.Apply && len(res.VetChanges.ContentJSON) > 2 {
		v, err := a.store.CreateAgentPromptVersion(r.Context(), "vet_live", res.VetChanges.ContentJSON, res.VetChanges.Changelog, "analyzer", nil, &runID, true)
		if err == nil {
			run.VetVersionID = &v.ID
			applied = true
		}
	}
	if res.CoachChanges.Apply && len(res.CoachChanges.ContentJSON) > 2 {
		v, err := a.store.CreateAgentPromptVersion(r.Context(), "coach", res.CoachChanges.ContentJSON, res.CoachChanges.Changelog, "analyzer", nil, &runID, true)
		if err == nil {
			run.CoachVersionID = &v.ID
			applied = true
		}
	}
	if applied {
		run.Status = "applied"
	} else if res.NoOpReason != "" {
		run.Status = "noop"
	} else {
		run.Status = "needs_review"
	}
	ids := make([]string, 0, len(feedbacks))
	for _, f := range feedbacks {
		ids = append(ids, f.ID)
	}
	_ = a.store.MarkPitchFeedbackProcessed(r.Context(), ids)
	_ = a.store.FinishAnalyzerRun(r.Context(), run)
	httpx.WriteData(w, http.StatusOK, run)
}
