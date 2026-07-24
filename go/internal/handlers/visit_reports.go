package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type visitLocationReq struct {
	AddressText string   `json:"addressText"`
	Lat         *float64 `json:"lat"`
	Lng         *float64 `json:"lng"`
}

type visitReportBodyReq struct {
	BodyText string `json:"bodyText"`
}

func (a *API) updateVisitLocation(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	visitID := chi.URLParam(r, "visitID")
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if !a.canManageVisit(w, r, id, visit) {
		return
	}
	var req visitLocationReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	updated, err := a.store.UpdateVisitLocation(r.Context(), visitID, strings.TrimSpace(req.AddressText), req.Lat, req.Lng)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, updated)
}

func (a *API) getVisitReport(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if id.Role != kernel.RoleVet && id.Role != kernel.RoleCarePro {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	visitID := chi.URLParam(r, "visitID")
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
	ident := store.IdentityOf(id.UserID, id.Role, id.PracticeID)
	canRead, err := a.store.CanAccessPet(r.Context(), ident, pet, store.PermRead)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !canRead {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	canWrite, err := a.store.CanAccessPet(r.Context(), ident, pet, store.PermWriteNotes)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if canWrite {
		report, err := a.store.EnsureVisitReport(r.Context(), visitID, id.UserID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		httpx.WriteData(w, http.StatusOK, redactVisitReportAudio(report))
		return
	}
	report, err := a.store.GetVisitReport(r.Context(), visitID, id.UserID)
	if errors.Is(err, store.ErrNotFound) {
		httpx.WriteData(w, http.StatusOK, store.VisitReport{
			VisitID: visitID, AuthorUserID: id.UserID, Status: "none",
		})
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, redactVisitReportAudio(report))
}

func (a *API) putVisitReport(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	visitID := chi.URLParam(r, "visitID")
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if !a.canManageVisit(w, r, id, visit) {
		return
	}
	var req visitReportBodyReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	report, err := a.store.UpsertVisitReport(r.Context(), visitID, id.UserID, req.BodyText)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "conflict", "report_finalized")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, redactVisitReportAudio(report))
}

func (a *API) getVisitReportAudio(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	visitID := chi.URLParam(r, "visitID")
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if !a.canAccessVisitReport(w, r, id, visit, store.PermWriteNotes, true) {
		return
	}
	report, err := a.store.GetVisitReport(r.Context(), visitID, id.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if report.Status == "final" {
		writeErr(w, r, http.StatusGone, "gone", "report_finalized")
		return
	}
	key := strings.TrimSpace(report.AudioObjectKey)
	if key == "" || a.media == nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	rc, ct, err := a.media.Open(r.Context(), key)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	defer rc.Close()
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = io.Copy(w, rc)
}

func (a *API) finalizeVisitReport(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	visitID := chi.URLParam(r, "visitID")
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if !a.canManageVisit(w, r, id, visit) {
		return
	}
	report, err := a.store.EnsureVisitReport(r.Context(), visitID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if report.Status == "final" {
		writeErr(w, r, http.StatusConflict, "conflict", "report_finalized")
		return
	}
	// RGPD: purge audio before finalize so a failed Delete can still be retried via finalize.
	if err := a.purgeVisitReportAudio(r.Context(), report); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "audio_purge_failed")
		return
	}
	report, err = a.store.FinalizeVisitReport(r.Context(), report.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusConflict, "conflict", "report_finalized")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	report.AudioURL = ""
	report.AudioObjectKey = ""
	httpx.WriteData(w, http.StatusOK, redactVisitReportAudio(report))
}

func (a *API) improveVisitReport(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if a.gemini == nil || !a.gemini.Configured() {
		writeErr(w, r, http.StatusServiceUnavailable, "not_configured", "gemini_not_configured")
		return
	}
	visitID := chi.URLParam(r, "visitID")
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if !a.canManageVisit(w, r, id, visit) {
		return
	}
	report, err := a.store.EnsureVisitReport(r.Context(), visitID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if report.Status == "final" {
		writeErr(w, r, http.StatusConflict, "conflict", "report_finalized")
		return
	}
	source := report.BodyText
	if strings.TrimSpace(source) == "" {
		source = report.TranscriptText
	}
	if strings.TrimSpace(source) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	system := `Tu es un assistant vétérinaire. Reformule le compte-rendu de visite en français clair,
structuré en sections SOAP (Subjectif, Objectif, Analyse, Plan). Garde les faits médicaux; n'invente pas.`
	if id.Role == kernel.RoleCarePro {
		u, uerr := a.store.GetUserByID(r.Context(), id.UserID)
		if uerr == nil {
			switch kernel.ProfessionalSpecialty(u.ProfessionalSpecialty) {
			case kernel.SpecialtyFarrier:
				system = `Tu es un assistant pour maréchal-ferrant. Reformule le compte-rendu de ferrage/intervention
en français clair (état des pieds, fer/type, observations, recommandations). N'invente pas.`
			case kernel.SpecialtyPhysio:
				system = `Tu es un assistant en physiothérapie animale. Reformule le CR de séance
(motif, examen, techniques, exercices, plan). N'invente pas.`
			case kernel.SpecialtyBehaviorist:
				system = `Tu es un assistant comportementaliste animalier. Reformule le CR
(contexte, comportements observés, analyse, plan d'accompagnement). N'invente pas.`
			}
		}
	}
	improved, err := a.gemini.GenerateText(r.Context(), system, source, 0.3)
	if err != nil {
		writeErr(w, r, http.StatusBadGateway, "gemini_error", "internal")
		return
	}
	report, err = a.store.UpdateVisitReportImproved(r.Context(), report.ID, strings.TrimSpace(improved))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusConflict, "conflict", "report_finalized")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, redactVisitReportAudio(report))
}

func (a *API) transcribeVisitReport(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if a.gemini == nil || !a.gemini.Configured() {
		writeErr(w, r, http.StatusServiceUnavailable, "not_configured", "gemini_not_configured")
		return
	}
	visitID := chi.URLParam(r, "visitID")
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if !a.canManageVisit(w, r, id, visit) {
		return
	}
	report, err := a.store.EnsureVisitReport(r.Context(), visitID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if report.Status == "final" {
		writeErr(w, r, http.StatusConflict, "conflict", "report_finalized")
		return
	}
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	file, header, err := r.FormFile("audio")
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil || len(data) == 0 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	safeName := filepath.Base(strings.TrimSpace(header.Filename))
	safeName = strings.ReplaceAll(safeName, "..", "")
	if safeName == "" || safeName == "." || safeName == string(filepath.Separator) {
		safeName = "audio.bin"
	}
	objectKey := fmt.Sprintf("visit-reports/%s/%s-%s", visitID, uuid.NewString(), safeName)
	ct := header.Header.Get("Content-Type")
	if ct == "" {
		ct = mimeFromFilename(safeName)
	}
	ct = normalizeAudioMIME(ct)
	if ct == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_audio_type")
		return
	}
	if a.media == nil {
		writeErr(w, r, http.StatusServiceUnavailable, "media_unavailable", "media_unavailable")
		return
	}
	url, err := a.media.Upload(r.Context(), objectKey, bytes.NewReader(data), int64(len(data)), ct)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// Do not expose public /media URL for PHI audio — store key only; stream via authenticated GET.
	_ = url
	report, err = a.store.UpdateVisitReportAudio(r.Context(), report.ID, "", objectKey)
	if err != nil {
		_ = a.media.Delete(r.Context(), objectKey)
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusConflict, "conflict", "report_finalized")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	hint := strings.TrimSpace(r.FormValue("hint"))
	transcript := hint
	if transcript == "" {
		system := `Tu transcris un compte-rendu vocal vétérinaire ou de soin animalier.
Retourne uniquement le texte transcrit, clair, en français (ou la langue parlée). N'invente pas de faits médicaux absents de l'audio.`
		userPrompt := "Transcris cet enregistrement de visite."
		out, gerr := a.gemini.GenerateTextWithMedia(r.Context(), system, userPrompt, ct, data, 0.2)
		if gerr != nil {
			_ = a.purgeVisitReportAudio(r.Context(), report)
			writeErr(w, r, http.StatusBadGateway, "gemini_error", "transcription_failed")
			return
		}
		transcript = strings.TrimSpace(out)
		if transcript == "" {
			_ = a.purgeVisitReportAudio(r.Context(), report)
			writeErr(w, r, http.StatusBadGateway, "gemini_error", "transcription_failed")
			return
		}
	}
	report, err = a.store.UpdateVisitReportTranscript(r.Context(), report.ID, transcript)
	if err != nil {
		_ = a.purgeVisitReportAudio(r.Context(), report)
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusConflict, "conflict", "report_finalized")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, redactVisitReportAudio(report))
}

func (a *API) canManageVisit(w http.ResponseWriter, r *http.Request, id authx.Identity, visit store.Visit) bool {
	return a.canAccessVisitReport(w, r, id, visit, store.PermWriteNotes, true)
}

// canAccessVisitReport checks role (vet|care_pro) and pet ACL. If writeErrOnFail is false, returns false silently.
func (a *API) canAccessVisitReport(w http.ResponseWriter, r *http.Request, id authx.Identity, visit store.Visit, need store.AccessPermission, writeErrOnFail bool) bool {
	if id.Role != kernel.RoleVet && id.Role != kernel.RoleCarePro {
		if writeErrOnFail {
			writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		}
		return false
	}
	pet, err := a.store.GetPet(r.Context(), visit.PetID)
	if err != nil {
		if writeErrOnFail {
			writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		}
		return false
	}
	ok, err := a.store.CanAccessPet(r.Context(), store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, need)
	if err != nil {
		if writeErrOnFail {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		}
		return false
	}
	if !ok {
		if writeErrOnFail {
			writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		}
		return false
	}
	return true
}

func mimeFromFilename(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".mp3"):
		return "audio/mpeg"
	case strings.HasSuffix(lower, ".m4a"), strings.HasSuffix(lower, ".mp4"):
		return "audio/mp4"
	case strings.HasSuffix(lower, ".wav"):
		return "audio/wav"
	case strings.HasSuffix(lower, ".webm"):
		return "audio/webm"
	case strings.HasSuffix(lower, ".ogg"), strings.HasSuffix(lower, ".oga"):
		return "audio/ogg"
	default:
		return ""
	}
}

func normalizeAudioMIME(ct string) string {
	ct = strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
	switch ct {
	case "audio/mpeg", "audio/mp3", "audio/mp4", "audio/m4a", "audio/x-m4a",
		"audio/wav", "audio/wave", "audio/x-wav", "audio/webm", "audio/ogg":
		if ct == "audio/mp3" {
			return "audio/mpeg"
		}
		if ct == "audio/m4a" || ct == "audio/x-m4a" {
			return "audio/mp4"
		}
		if ct == "audio/wave" || ct == "audio/x-wav" {
			return "audio/wav"
		}
		return ct
	default:
		return ""
	}
}

// redactVisitReportAudio hides public audio URLs from API clients (PHI).
func redactVisitReportAudio(r store.VisitReport) store.VisitReport {
	r.AudioURL = ""
	r.AudioObjectKey = ""
	return r
}

func (a *API) purgeVisitReportAudio(ctx context.Context, report store.VisitReport) error {
	key := strings.TrimSpace(report.AudioObjectKey)
	if key != "" && a.media != nil {
		if err := a.media.Delete(ctx, key); err != nil {
			return err
		}
	}
	if report.ID != "" {
		if err := a.store.ClearVisitReportAudio(ctx, report.ID); err != nil && !errors.Is(err, store.ErrNotFound) {
			return err
		}
	}
	return nil
}
