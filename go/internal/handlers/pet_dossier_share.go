package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/platform/petdossier"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

const (
	// Le ZIP est assemblé et servi entièrement en mémoire depuis une route publique :
	// sans plafond, un animal très documenté suffit à faire tomber l'instance.
	maxDossierAttachments     = 25
	maxDossierAttachmentBytes = 40 << 20
)

type createDossierShareReq struct {
	Email string `json:"email"`
}

func (a *API) registerDossierSharePublicRoutes(r chi.Router, rateLimit func(http.Handler) http.Handler) {
	r.Group(func(pr chi.Router) {
		if rateLimit != nil {
			pr.Use(rateLimit)
		}
		pr.Get("/public/pet-dossier/{token}", a.getPublicPetDossier)
		pr.Get("/public/pet-dossier/{token}/download", a.downloadPublicPetDossier)
	})
}

func (a *API) createPetDossierShare(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	petID := chi.URLParam(r, "petID")
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
	var req createDossierShareReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	// ParseAddress plutôt qu'un test sur "@" : rejette les CR/LF et les adresses
	// malformées avant de créer un token et de tenter un envoi SMTP voué à l'échec.
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

	// Le job de rétention n'a pas de scheduler : on purge aussi les partages périmés
	// du propriétaire sur son propre chemin de création, seul passage garanti.
	a.purgeExpiredDossierShares(r.Context(), id.UserID)

	tok, err := a.store.CreateDossierShareToken(r.Context(), store.CreateDossierShareInput{
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
			writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "dossier_share_limit")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	downloadURL := site + "/dossier/" + tok.Token
	if a.notifier != nil {
		if err := a.notifier.SendPetDossierShare(
			email, locale, pet.Name, owner.FullName,
			downloadURL, commercialName, commercialPhone, commercialEmail, registerURL, site,
		); err != nil {
			// Ne pas laisser un token orphelin (quota + lien fantôme).
			if key, delErr := a.store.DeleteDossierShareByID(r.Context(), tok.ID); delErr != nil {
				fmt.Printf("dossier share: rollback after SMTP failure id=%s: %v\n", tok.ID, delErr)
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
		out["downloadPath"] = "/dossier/" + tok.Token
		out["token"] = tok.Token
	}
	httpx.WriteData(w, http.StatusCreated, out)
}

func (a *API) resolveDossierCommercial(ctx context.Context, clientUserID, practiceID, site string) (commercialID, name, phone, emailAddr, registerURL string) {
	_, commercialID, err := a.store.ResolveVetCommercial(ctx, clientUserID, practiceID)
	if err == nil && commercialID != "" {
		if c, err2 := a.store.GetUserByID(ctx, commercialID); err2 == nil {
			name = c.FullName
			phone = strings.TrimSpace(c.ContactPhone)
			emailAddr = c.Email
			if inv, err3 := a.store.EnsureAppInviteCode(ctx, commercialID); err3 == nil && inv.Code != "" {
				registerURL = site + "/register?invite=" + inv.Code
			}
		}
	}
	if phone == "" {
		phone = strings.TrimSpace(a.cfg.CommercialContactPhone)
	}
	if registerURL == "" {
		registerURL = site + "/register"
	}
	if name == "" {
		name = "petsFollow"
	}
	return commercialID, name, phone, emailAddr, registerURL
}

func (a *API) getPublicPetDossier(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	tok, err := a.store.GetDossierShareByToken(r.Context(), token)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "dossier_not_found")
		return
	}
	if time.Now().UTC().After(tok.ExpiresAt) {
		// Row + email tiers + ZIP cache — pas seulement object_key.
		a.purgeExpiredDossierShares(r.Context(), tok.OwnerUserID)
		writeErr(w, r, http.StatusGone, "gone", "dossier_expired")
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
		"downloadUrl":     "/api/v1/public/pet-dossier/" + tok.Token + "/download",
	})
}

func (a *API) downloadPublicPetDossier(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	ctx := r.Context()

	tx, err := a.store.Pool().Begin(ctx)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	defer tx.Rollback(ctx)

	tok, err := a.store.LockDossierShareForDownload(ctx, tx, token)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "dossier_not_found")
		return
	}
	if time.Now().UTC().After(tok.ExpiresAt) {
		ownerID := tok.OwnerUserID
		_ = tx.Rollback(ctx)
		a.purgeExpiredDossierShares(ctx, ownerID)
		writeErr(w, r, http.StatusGone, "gone", "dossier_expired")
		return
	}

	var zipBytes []byte
	if key := strings.TrimSpace(tok.ObjectKey); key != "" {
		// Le cache a été écrit par nous : même plafond que la construction, plus la
		// marge du PDF de synthèse.
		zipBytes = a.readDossierAttachment(ctx, key, maxDossierAttachmentBytes+(8<<20))
	}
	cacheKey := ""
	if len(zipBytes) == 0 {
		built, buildErr := a.buildDossierZip(ctx, tok)
		if buildErr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "dossier_build_failed")
			return
		}
		zipBytes = built
		if a.media != nil {
			cacheKey = media.ObjectKey("dossier-shares", tok.ID, ".zip")
			if _, upErr := a.media.Upload(ctx, cacheKey, bytes.NewReader(zipBytes), int64(len(zipBytes)), "application/zip"); upErr != nil {
				cacheKey = ""
			}
		}
	}
	if err := a.store.MarkDossierShareDownloadedTx(ctx, tx, tok.ID, cacheKey); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if err := tx.Commit(ctx); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}

	filename := "dossier-" + sanitizeFilename(tok.PetName) + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(zipBytes)
}

func (a *API) buildDossierZip(ctx context.Context, tok store.DossierShareToken) ([]byte, error) {
	pet, err := a.store.GetPet(ctx, tok.PetID)
	if err != nil {
		return nil, err
	}
	owner, _ := a.store.GetUserByID(ctx, tok.OwnerUserID)
	site := strings.TrimRight(a.cfg.ProPublicSiteURL, "/")
	registerURL := tok.RegisterURL
	if registerURL == "" {
		registerURL = site + "/register"
	}

	weights, _ := a.store.ListWeightReadings(ctx, pet.ID)
	hrs, _ := a.store.ListHeartRateSessions(ctx, pet.ID, false)
	reminders, _ := a.store.ListCareReminders(ctx, pet.ID)
	visits, _ := a.store.ListVisitsForDossier(ctx, pet.ID)
	timeline, _ := a.store.PetTimelineFiltered(ctx, pet.ID, false, false, false)
	// Dossier PDF only uses title/body/date — drop meta (hasReport etc.) so no
	// accidental consultation CTA can be wired from this payload later.
	for i := range timeline {
		timeline[i].Meta = nil
	}
	docs, _ := a.store.ListPetDocuments(ctx, pet.ID)

	commercialEmail := ""
	if tok.CommercialUserID != "" {
		if c, err := a.store.GetUserByID(ctx, tok.CommercialUserID); err == nil {
			commercialEmail = c.Email
		}
	}
	pdfIn := petdossier.PDFInput{
		Marketing: petdossier.Marketing{
			SiteURL:         site,
			RegisterURL:     registerURL,
			ProductsURL:     site + "/produits",
			CommercialName:  tok.CommercialName,
			CommercialPhone: tok.CommercialPhone,
			CommercialEmail: commercialEmail,
		},
		Pet: petdossier.PetInfo{
			Name:             pet.Name,
			Species:          pet.Species,
			Breed:            pet.Breed,
			OwnerName:        owner.FullName,
			OwnerEmail:       owner.Email,
			MicrochipNumber:  pet.MicrochipNumber,
			HealthBookNumber: pet.HealthBookNumber,
		},
		HasHealthBook: strings.TrimSpace(pet.HealthBookPDFObjectKey) != "",
		GeneratedAt:   time.Now().UTC(),
		Locale:        i18n.NormalizeLocale(tok.Locale),
	}
	if pet.BirthDate != nil {
		pdfIn.Pet.BirthDate = pet.BirthDate.Format("2006-01-02")
	}
	if pet.WeightKg != nil {
		pdfIn.Pet.WeightKg = fmt.Sprintf("%.2f kg", *pet.WeightKg)
	}
	for i, w := range weights {
		if i >= 20 {
			break
		}
		line := petdossier.WeightLine{When: w.RecordedAt.Format("2006-01-02"), WeightKg: fmt.Sprintf("%.2f", w.WeightKg)}
		if w.Comment != nil {
			line.Comment = *w.Comment
		}
		pdfIn.Weights = append(pdfIn.Weights, line)
	}
	for i, h := range hrs {
		if i >= 20 {
			break
		}
		bpm := "—"
		if h.BPM != nil {
			bpm = fmt.Sprintf("%d", *h.BPM)
		}
		pdfIn.HeartRates = append(pdfIn.HeartRates, petdossier.HRLine{
			When: h.StartedAt.Format("2006-01-02 15:04"), BPM: bpm, Alert: h.IsAlert,
		})
	}
	for _, rem := range reminders {
		if rem.Status == "done" || rem.Status == "cancelled" {
			continue
		}
		pdfIn.Reminders = append(pdfIn.Reminders, petdossier.ReminderLine{
			Title: rem.Title, Due: rem.DueAt.Format("2006-01-02"),
		})
	}
	for _, v := range visits {
		when := v.CreatedAt.Format("2006-01-02")
		if v.ScheduledAt != nil {
			when = v.ScheduledAt.Format("2006-01-02 15:04")
		}
		notes := v.Notes
		if v.ReportExcerpt != "" {
			if notes != "" {
				notes += " — "
			}
			notes += v.ReportExcerpt
		}
		pro := v.ProConsulted
		if pro == "" && v.PracticeName != "" {
			pro = "Cabinet " + v.PracticeName
		}
		pdfIn.Visits = append(pdfIn.Visits, petdossier.VisitLine{
			When: when, Status: v.Status, PracticeName: v.PracticeName, ProConsulted: pro, Notes: notes,
		})
	}
	for i, t := range timeline {
		if i >= 40 {
			break
		}
		pdfIn.Timeline = append(pdfIn.Timeline, petdossier.TimelineLine{
			When: t.CreatedAt.Format("2006-01-02 15:04"), Title: t.Title, Body: t.Body,
		})
	}
	for _, d := range docs {
		pdfIn.Documents = append(pdfIn.Documents, petdossier.DocLine{Title: d.Title, FileName: d.FileName})
	}

	pdfBytes, err := petdossier.BuildPDF(pdfIn)
	if err != nil {
		return nil, err
	}
	files := []petdossier.ZipFile{{Name: "dossier.pdf", Data: pdfBytes}}
	budget := int64(maxDossierAttachmentBytes)

	if key := strings.TrimSpace(pet.HealthBookPDFObjectKey); key != "" && a.media != nil {
		if data := a.readDossierAttachment(ctx, key, budget); data != nil {
			budget -= int64(len(data))
			files = append(files, petdossier.ZipFile{Name: "carnet-sante.pdf", Data: data})
		}
	}
	attached := 0
	for _, d := range docs {
		if attached >= maxDossierAttachments || budget <= 0 || a.media == nil {
			break
		}
		key := strings.TrimSpace(d.ObjectKey)
		if key == "" {
			continue
		}
		data := a.readDossierAttachment(ctx, key, budget)
		if data == nil {
			continue
		}
		budget -= int64(len(data))
		attached++
		files = append(files, petdossier.ZipFile{
			Name: petdossier.DocumentZipPath(d.Title, d.FileName, key),
			Data: data,
		})
	}
	return petdossier.BuildZip(files)
}

// readDossierAttachment returns nil when the object is missing, empty, or larger
// than the remaining budget — a single oversized file is skipped, never truncated.
func (a *API) readDossierAttachment(ctx context.Context, objectKey string, budget int64) []byte {
	if a.media == nil || budget <= 0 {
		return nil
	}
	rc, _, err := a.media.Open(ctx, objectKey)
	if err != nil {
		return nil
	}
	data, err := io.ReadAll(io.LimitReader(rc, budget+1))
	_ = rc.Close()
	if err != nil || len(data) == 0 || int64(len(data)) > budget {
		return nil
	}
	return data
}

func sanitizeFilename(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "animal"
	}
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if r == ' ' {
			b.WriteRune('-')
		}
	}
	out := b.String()
	if out == "" {
		return "animal"
	}
	return out
}
