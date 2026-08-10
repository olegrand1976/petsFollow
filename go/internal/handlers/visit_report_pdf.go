package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/consultationpdf"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// visitReportAPI enriches VisitReport with optional last completed improve-run citations (Phase 4).
type visitReportAPI struct {
	store.VisitReport
	LastImproveCitations json.RawMessage `json:"lastImproveCitations,omitempty"`
}

func (a *API) writeVisitReportData(w http.ResponseWriter, r *http.Request, report store.VisitReport) {
	out := visitReportAPI{VisitReport: redactVisitReportAudio(report)}
	if a.cfg.AiCrAdvancedEnabled && report.ID != "" {
		if run, err := a.store.LatestCompletedImproveRunForReport(r.Context(), report.ID); err == nil && len(run.Citations) > 0 && string(run.Citations) != "[]" {
			out.LastImproveCitations = run.Citations
		}
	}
	httpx.WriteData(w, http.StatusOK, out)
}

// getVisitReportPDF streams a branded PDF of the caller's CR for the visit (attachment).
func (a *API) getVisitReportPDF(w http.ResponseWriter, r *http.Request) {
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

	report, err := a.store.GetVisitReport(r.Context(), visitID, id.UserID)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "report_not_found")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	body := strings.TrimSpace(report.BodyText)
	if body == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "report_empty")
		return
	}
	if q := strings.TrimSpace(r.URL.Query().Get("stripCitations")); q == "1" || strings.EqualFold(q, "true") {
		body = consultationpdf.StripCitations(body)
		if strings.TrimSpace(body) == "" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "report_empty")
			return
		}
	}

	pet, err := a.store.GetPet(r.Context(), visit.PetID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	ownerName := ""
	if owner, err := a.store.GetUserByID(r.Context(), pet.OwnerUserID); err == nil {
		ownerName = owner.FullName
	}
	authorName := ""
	locale := i18n.NormalizeLocale(r.Header.Get("Accept-Language"))
	if author, err := a.store.GetUserByID(r.Context(), id.UserID); err == nil {
		authorName = author.FullName
		if author.PreferredLocale != "" {
			locale = i18n.NormalizeLocale(author.PreferredLocale)
		}
	}
	practiceName := ""
	if visit.PracticeID != "" {
		if name, err := a.store.GetPracticeName(r.Context(), visit.PracticeID); err == nil {
			practiceName = name
		}
	}

	visitWhen := visit.CreatedAt.UTC().Format("2006-01-02 15:04")
	if visit.ScheduledAt != nil {
		visitWhen = visit.ScheduledAt.UTC().Format("2006-01-02 15:04")
	}
	finalizedAt := ""
	if report.FinalizedAt != nil {
		finalizedAt = report.FinalizedAt.UTC().Format("2006-01-02 15:04")
	}

	site := strings.TrimRight(a.cfg.ProPublicSiteURL, "/")
	pdfBytes, err := consultationpdf.BuildPDF(consultationpdf.PDFInput{
		Marketing: consultationpdf.Marketing{
			SiteURL:     site,
			RegisterURL: site + "/register",
			ProductsURL: site + "/produits",
		},
		PetName:      pet.Name,
		Species:      pet.Species,
		Breed:        pet.Breed,
		OwnerName:    ownerName,
		PracticeName: practiceName,
		VisitWhen:    visitWhen,
		Reports: []consultationpdf.ReportSection{{
			AuthorName:  authorName,
			FinalizedAt: finalizedAt,
			BodyText:    body,
		}},
		GeneratedAt: time.Now().UTC(),
		Locale:      locale,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "pdf_failed")
		return
	}

	filename := "cr-" + sanitizeFilename(pet.Name) + ".pdf"
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}
