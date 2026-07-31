package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/consultationpdf"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type createConsultationShareReq struct {
	Email string `json:"email"`
}

func (a *API) registerConsultationSharePublicRoutes(r chi.Router, rateLimit func(http.Handler) http.Handler) {
	r.Group(func(pr chi.Router) {
		if rateLimit != nil {
			pr.Use(rateLimit)
		}
		pr.Get("/public/consultation/{token}", a.getPublicConsultation)
		pr.Get("/public/consultation/{token}/download", a.downloadPublicConsultation)
	})
}

func (a *API) createConsultationShare(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
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
	hasFinal, err := a.store.VisitHasFinalReport(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !hasFinal {
		writeErr(w, r, http.StatusNotFound, "not_found", "consultation_not_found")
		return
	}

	var req createConsultationShareReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	parsed, err := mail.ParseAddress(strings.TrimSpace(req.Email))
	if err != nil || utf8.RuneCountInString(parsed.Address) > 254 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_email")
		return
	}
	email := strings.ToLower(parsed.Address)

	owner, err := a.store.GetUserByID(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	locale := i18n.NormalizeLocale(owner.PreferredLocale)
	site := strings.TrimRight(a.cfg.ProPublicSiteURL, "/")

	commercialID, commercialName, commercialPhone, commercialEmail, registerURL := a.resolveDossierCommercial(r.Context(), id.UserID, pet.PracticeID, site)

	a.purgeExpiredConsultationShares(r.Context(), id.UserID)

	tok, err := a.store.CreateConsultationShareToken(r.Context(), store.CreateConsultationShareInput{
		VisitID:          visit.ID,
		PetID:            pet.ID,
		OwnerUserID:      id.UserID,
		RecipientEmail:   email,
		Locale:           locale,
		CommercialUserID: commercialID,
		CommercialName:   commercialName,
		CommercialPhone:  commercialPhone,
		RegisterURL:      registerURL,
	})
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_email")
			return
		}
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "consultation_share_limit")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	downloadURL := site + "/consultation/" + tok.Token
	if a.notifier != nil {
		if err := a.notifier.SendConsultationShare(
			email, locale, pet.Name, owner.FullName,
			downloadURL, commercialName, commercialPhone, commercialEmail, registerURL, site,
		); err != nil {
			if key, delErr := a.store.DeleteConsultationShareByID(r.Context(), tok.ID); delErr != nil {
				fmt.Printf("consultation share: rollback after SMTP failure id=%s: %v\n", tok.ID, delErr)
			} else if key != "" {
				a.purgeMediaObjects(r.Context(), []string{key})
			}
			writeErr(w, r, http.StatusBadGateway, "email_failed", "email_send_failed")
			return
		}
	}

	out := map[string]any{
		"ok":        true,
		"expiresAt": tok.ExpiresAt.UTC().Format(time.RFC3339),
		"petName":   pet.Name,
	}
	if a.cfg.DevSeedEnabled {
		out["downloadPath"] = "/consultation/" + tok.Token
		out["token"] = tok.Token
	}
	httpx.WriteData(w, http.StatusCreated, out)
}

func (a *API) getPublicConsultation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	tok, err := a.store.GetConsultationShareByToken(r.Context(), token)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "consultation_not_found")
		return
	}
	if time.Now().UTC().After(tok.ExpiresAt) {
		a.purgeExpiredConsultationShares(r.Context(), tok.OwnerUserID)
		writeErr(w, r, http.StatusGone, "gone", "consultation_expired")
		return
	}
	site := strings.TrimRight(a.cfg.ProPublicSiteURL, "/")
	registerURL := tok.RegisterURL
	if registerURL == "" {
		registerURL = site + "/register"
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"petName":         tok.PetName,
		"expiresAt":       tok.ExpiresAt.UTC().Format(time.RFC3339),
		"commercialName":  tok.CommercialName,
		"commercialPhone": tok.CommercialPhone,
		"registerUrl":     registerURL,
		"siteUrl":         site,
		"productsUrl":     site + "/produits",
		"downloadUrl":     "/api/v1/public/consultation/" + tok.Token + "/download",
	})
}

func (a *API) downloadPublicConsultation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	ctx := r.Context()

	tx, err := a.store.Pool().Begin(ctx)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	defer tx.Rollback(ctx)

	tok, err := a.store.LockConsultationShareForDownload(ctx, tx, token)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "consultation_not_found")
		return
	}
	if time.Now().UTC().After(tok.ExpiresAt) {
		ownerID := tok.OwnerUserID
		_ = tx.Rollback(ctx)
		a.purgeExpiredConsultationShares(ctx, ownerID)
		writeErr(w, r, http.StatusGone, "gone", "consultation_expired")
		return
	}

	var pdfBytes []byte
	staleObjectKey := ""
	if key := strings.TrimSpace(tok.ObjectKey); key != "" {
		// Bust pre-brand / attachment-layout caches (v1 keys under consultation-shares/).
		legacyCache := !strings.Contains(key, "consultation-shares-v2")
		stale, staleErr := a.store.ConsultationShareCacheStale(ctx, tok.VisitID, tok.CachedAt)
		if staleErr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if stale || legacyCache {
			staleObjectKey = key
			if clearErr := a.store.ClearConsultationShareObjectKeyTx(ctx, tx, tok.ID); clearErr != nil {
				writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
				return
			}
		} else {
			pdfBytes = a.readDossierAttachment(ctx, key, 8<<20)
		}
	}
	cacheKey := ""
	if len(pdfBytes) == 0 {
		built, buildErr := a.buildConsultationPDF(ctx, tok)
		if buildErr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "consultation_build_failed")
			return
		}
		pdfBytes = built
		if a.media != nil {
			// v2 = branded PDF layout; new key busts stale mobile-hostile caches.
			cacheKey = media.ObjectKey("consultation-shares-v2", tok.ID, ".pdf")
			if _, upErr := a.media.Upload(ctx, cacheKey, bytes.NewReader(pdfBytes), int64(len(pdfBytes)), "application/pdf"); upErr != nil {
				cacheKey = ""
			}
		}
	}
	if err := a.store.MarkConsultationShareDownloadedTx(ctx, tx, tok.ID, cacheKey); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if err := tx.Commit(ctx); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if staleObjectKey != "" && staleObjectKey != cacheKey {
		a.purgeMediaObjects(ctx, []string{staleObjectKey})
	}

	filename := "consultation-" + sanitizeFilename(tok.PetName) + ".pdf"
	w.Header().Set("Content-Type", "application/pdf")
	// inline: mobile browsers / Safari open the PDF in-viewer instead of a blank blob tab.
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}

func (a *API) buildConsultationPDF(ctx context.Context, tok store.ConsultationShareToken) ([]byte, error) {
	pet, err := a.store.GetPet(ctx, tok.PetID)
	if err != nil {
		return nil, err
	}
	visit, err := a.store.GetVisit(ctx, tok.VisitID)
	if err != nil {
		return nil, err
	}
	owner, _ := a.store.GetUserByID(ctx, tok.OwnerUserID)
	reports, err := a.store.ListFinalVisitReportsForClient(ctx, tok.VisitID)
	if err != nil {
		return nil, err
	}
	if len(reports) == 0 {
		return nil, fmt.Errorf("no final reports")
	}

	practiceName := ""
	if visit.PracticeID != "" {
		if name, err := a.store.GetPracticeName(ctx, visit.PracticeID); err == nil {
			practiceName = name
		}
	}

	site := strings.TrimRight(a.cfg.ProPublicSiteURL, "/")
	registerURL := tok.RegisterURL
	if registerURL == "" {
		registerURL = site + "/register"
	}
	commercialEmail := ""
	if tok.CommercialUserID != "" {
		if c, err := a.store.GetUserByID(ctx, tok.CommercialUserID); err == nil {
			commercialEmail = c.Email
		}
	}

	visitWhen := visit.CreatedAt.UTC().Format("2006-01-02 15:04")
	if visit.ScheduledAt != nil {
		visitWhen = visit.ScheduledAt.UTC().Format("2006-01-02 15:04")
	}

	sections := make([]consultationpdf.ReportSection, 0, len(reports))
	for _, r := range reports {
		fa := ""
		if r.FinalizedAt != nil {
			fa = r.FinalizedAt.UTC().Format("2006-01-02 15:04")
		}
		sections = append(sections, consultationpdf.ReportSection{
			AuthorName:  r.AuthorName,
			FinalizedAt: fa,
			BodyText:    r.BodyText,
		})
	}

	return consultationpdf.BuildPDF(consultationpdf.PDFInput{
		Marketing: consultationpdf.Marketing{
			SiteURL:         site,
			RegisterURL:     registerURL,
			ProductsURL:     site + "/produits",
			CommercialName:  tok.CommercialName,
			CommercialPhone: tok.CommercialPhone,
			CommercialEmail: commercialEmail,
		},
		PetName:      pet.Name,
		Species:      pet.Species,
		Breed:        pet.Breed,
		OwnerName:    owner.FullName,
		PracticeName: practiceName,
		VisitWhen:    visitWhen,
		Reports:      sections,
		GeneratedAt:  time.Now().UTC(),
		Locale:       i18n.NormalizeLocale(tok.Locale),
	})
}

func (a *API) purgeExpiredConsultationShares(ctx context.Context, ownerUserID string) int {
	keys, deleted, err := a.store.PurgeExpiredConsultationShares(ctx, ownerUserID, 200)
	if err != nil {
		fmt.Printf("consultation share purge: %v\n", err)
		return 0
	}
	a.purgeMediaObjects(ctx, keys)
	return deleted
}
