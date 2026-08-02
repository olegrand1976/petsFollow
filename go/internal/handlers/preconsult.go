package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) getVisitPreconsult(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	visit, pet, ok := a.loadVisitPetForPreconsult(w, r, id, store.PermRead)
	if !ok {
		return
	}
	in, err := a.store.GetPreconsultByVisit(r.Context(), visit.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "preconsult_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	in.PetID = pet.ID
	in.PetName = pet.Name
	httpx.WriteData(w, http.StatusOK, in)
}

type putPreconsultReq struct {
	Answers store.PreconsultAnswers `json:"answers"`
}

func (a *API) putVisitPreconsult(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	visit, pet, ok := a.loadVisitPetForPreconsult(w, r, id, store.PermWriteNotes)
	if !ok {
		return
	}
	if pet.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	if visit.Status != "confirmed" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "visit_not_confirmed")
		return
	}
	var req putPreconsultReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	in, err := a.store.SubmitPreconsult(r.Context(), visit.ID, req.Answers)
	if err != nil {
		if code, ok := store.IsPreconsultValidation(err); ok {
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "conflict", "preconsult_already_submitted")
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "preconsult_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	in.PetID = pet.ID
	in.PetName = pet.Name
	a.afterPreconsultSubmitted(pet, visit, in.Answers)
	httpx.WriteData(w, http.StatusOK, in)
}

func (a *API) loadVisitPetForPreconsult(w http.ResponseWriter, r *http.Request, id authx.Identity, minPerm store.AccessPermission) (store.Visit, store.Pet, bool) {
	visit, err := a.store.GetVisit(r.Context(), chi.URLParam(r, "visitID"))
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "visit_not_found")
		return store.Visit{}, store.Pet{}, false
	}
	pet, err := a.store.GetPet(r.Context(), visit.PetID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return store.Visit{}, store.Pet{}, false
	}
	switch id.Role {
	case kernel.RoleClient, kernel.RoleCarePro:
		ok, aerr := a.store.CanAccessPet(r.Context(), store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, minPerm)
		if aerr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return store.Visit{}, store.Pet{}, false
		}
		if !ok {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
			return store.Visit{}, store.Pet{}, false
		}
	default:
		if !a.checkPracticePerm(w, r, id, "calendar.manage") {
			return store.Visit{}, store.Pet{}, false
		}
		if pet.PracticeID != id.PracticeID {
			ok, aerr := a.store.CanAccessPet(r.Context(), store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, store.PermFull)
			if aerr != nil {
				writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
				return store.Visit{}, store.Pet{}, false
			}
			if !ok {
				writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
				return store.Visit{}, store.Pet{}, false
			}
		}
	}
	return visit, pet, true
}

func (a *API) registerPreconsultPublicRoutes(r chi.Router, rateLimit func(http.Handler) http.Handler) {
	r.Group(func(pr chi.Router) {
		if rateLimit != nil {
			pr.Use(rateLimit)
		}
		pr.Get("/public/preconsult/{token}", a.getPublicPreconsult)
		pr.Post("/public/preconsult/{token}", a.postPublicPreconsult)
		pr.Get("/public/brand-assets", a.getPublicBrandAssets)
	})
}

func (a *API) getPublicBrandAssets(w http.ResponseWriter, r *http.Request) {
	android, ios, err := a.store.StoreQRAssets(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"qrAndroid": brandAssetPublicDTO(android),
		"qrIos":     brandAssetPublicDTO(ios),
		"downloadUrl": strings.TrimSpace(a.cfg.PetsAppDownloadURL),
	})
}

func brandAssetPublicDTO(a store.BrandAsset) map[string]any {
	if a.Key == "" || a.PublicURL == "" {
		return nil
	}
	return map[string]any{
		"publicUrl": a.PublicURL,
		"storeUrl":  a.StoreURL,
	}
}

func (a *API) getPublicPreconsult(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	visitID, err := a.store.ResolvePreconsultToken(r.Context(), token)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "preconsult_token_invalid")
		return
	}
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "visit_not_found")
		return
	}
	if visit.Status == "cancelled" {
		writeErr(w, r, http.StatusGone, "gone", "visit_cancelled")
		return
	}
	ctxData, err := a.store.GetPublicPreconsultContext(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "preconsult_not_found")
		return
	}
	a.enrichPublicPreconsultLinks(r.Context(), &ctxData)
	android, ios, _ := a.store.StoreQRAssets(r.Context())
	// Never echo PHI answers on the public GET — status alone drives the thank-you UI.
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"visitId":      ctxData.VisitID,
		"petName":      ctxData.PetName,
		"practiceName": ctxData.PracticeName,
		"scheduledAt":  ctxData.ScheduledAt,
		"status":       ctxData.Status,
		"inviteUrl":    ctxData.InviteURL,
		"downloadUrl":  ctxData.DownloadURL,
		"qrAndroid":    brandAssetPublicDTO(android),
		"qrIos":        brandAssetPublicDTO(ios),
	})
}

type postPublicPreconsultReq struct {
	Answers store.PreconsultAnswers `json:"answers"`
}

func (a *API) postPublicPreconsult(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	visitID, err := a.store.ResolvePreconsultToken(r.Context(), token)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "preconsult_token_invalid")
		return
	}
	visit, err := a.store.GetVisit(r.Context(), visitID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "visit_not_found")
		return
	}
	if visit.Status != "confirmed" && visit.Status != "done" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "visit_not_confirmed")
		return
	}
	var req postPublicPreconsultReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	in, err := a.store.SubmitPreconsult(r.Context(), visitID, req.Answers)
	if err != nil {
		if code, ok := store.IsPreconsultValidation(err); ok {
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "conflict", "preconsult_already_submitted")
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "preconsult_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	_ = a.store.MarkPreconsultTokenUsed(r.Context(), token)
	if pet, perr := a.store.GetPet(r.Context(), visit.PetID); perr == nil {
		a.afterPreconsultSubmitted(pet, visit, in.Answers)
	}
	ctxData, _ := a.store.GetPublicPreconsultContext(r.Context(), visitID)
	a.enrichPublicPreconsultLinks(r.Context(), &ctxData)
	android, ios, _ := a.store.StoreQRAssets(r.Context())
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"intake":        in,
		"petName":      ctxData.PetName,
		"practiceName": ctxData.PracticeName,
		"inviteUrl":    ctxData.InviteURL,
		"downloadUrl":  ctxData.DownloadURL,
		"qrAndroid":    brandAssetPublicDTO(android),
		"qrIos":        brandAssetPublicDTO(ios),
	})
}

// afterPreconsultSubmitted alerts on declared high immediately, then Gemini (soft-fail), then AI-red if needed.
func (a *API) afterPreconsultSubmitted(pet store.Pet, visit store.Visit, answers store.PreconsultAnswers) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		notified := false
		if answers.Urgency == "high" {
			a.notifyVetsPreconsultUrgent(ctx, pet, visit, answers, "")
			notified = true
		}

		aiUrgency := ""
		aiSummary := ""
		locale := a.preconsultAILocale(ctx, pet)
		if a.gemini != nil && a.gemini.Configured() {
			out, err := a.gemini.AssessPreconsultUrgency(ctx, gemini.PreconsultUrgencyInput{
				Locale:         locale,
				PetName:        pet.Name,
				Species:        pet.Species,
				ChiefComplaint: answers.ChiefComplaint,
				Duration:       answers.Duration,
				Behavior:       answers.Behavior,
				Appetite:       answers.Appetite,
				Thirst:         answers.Thirst,
				Elimination:    answers.Elimination,
				Urgency:        answers.Urgency,
				Comment:        answers.Comment,
			})
			if err != nil {
				log.Printf("preconsult: gemini assess visit %s: %v", visit.ID, err)
			} else if out != nil {
				aiUrgency = string(out.Urgency)
				aiSummary = out.Summary
				if serr := a.store.SavePreconsultAIAssessment(ctx, visit.ID, aiUrgency, aiSummary); serr != nil {
					log.Printf("preconsult: save ai assess visit %s: %v", visit.ID, serr)
				}
			}
		}
		if aiUrgency == "red" && !notified {
			a.notifyVetsPreconsultUrgent(ctx, pet, visit, answers, aiSummary)
		}
	}()
}

// preconsultAILocale picks practice/vet locale for the AI summary (not the client's).
func (a *API) preconsultAILocale(ctx context.Context, pet store.Pet) string {
	if refID, err := a.store.PracticeReferenceVetUserID(ctx, pet.PracticeID); err == nil && refID != "" {
		if loc, err := a.store.GetUserPreferredLocale(ctx, refID); err == nil && strings.TrimSpace(loc) != "" {
			return loc
		}
	}
	if vets, err := a.store.ListVetsForVisitAlert(ctx, pet.PracticeID, pet.OwnerUserID); err == nil {
		for _, vet := range vets {
			if strings.TrimSpace(vet.PreferredLocale) != "" {
				return vet.PreferredLocale
			}
		}
	}
	return "fr"
}

func (a *API) enrichPublicPreconsultLinks(ctx context.Context, out *store.PublicPreconsultContext) {
	out.DownloadURL = strings.TrimSpace(a.cfg.PetsAppDownloadURL)
	visit, err := a.store.GetVisit(ctx, out.VisitID)
	if err != nil {
		return
	}
	pet, err := a.store.GetPet(ctx, visit.PetID)
	if err != nil {
		return
	}
	if owner := a.resolvePreconsultInviteOwner(ctx, pet.PracticeID, pet.OwnerUserID); owner != "" {
		if invite, err := a.store.EnsureAppInviteCode(ctx, owner); err == nil {
			out.InviteURL = a.appInviteWebURL(invite.Code)
		}
	}
}

// errPreconsultAlreadySent is returned when a public invite was already issued.
var errPreconsultAlreadySent = errors.New("preconsult_already_sent")

// issuePreconsultInvite ensures pending intake + public token + email (once per visit).
func (a *API) issuePreconsultInvite(ctx context.Context, pet store.Pet, visit store.Visit) error {
	already, err := a.store.HasPreconsultInviteIssued(ctx, visit.ID)
	if err != nil {
		return err
	}
	if already {
		return errPreconsultAlreadySent
	}
	if err := a.store.SetVisitRequestPreconsult(ctx, visit.ID, true); err != nil {
		return err
	}
	if _, _, err = a.store.EnsurePreconsultPending(ctx, visit.ID); err != nil {
		return err
	}
	token, err := a.store.IssuePreconsultToken(ctx, visit.ID)
	if err != nil {
		return err
	}
	a.emailVisitPreconsult(ctx, pet, visit, token)
	return nil
}

// onVisitConfirmed pushes FCM; emails affiliation or public preconsult when opted-in.
// Token/email are issued at most once per visit (idempotent across confirm retries).
func (a *API) onVisitConfirmed(pet store.Pet, visit store.Visit) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		// Refresh flag in case confirm path set it just before.
		if full, err := a.store.GetVisit(ctx, visit.ID); err == nil {
			visit = full
		}
		a.pushVisitConfirmed(pet.OwnerUserID, visit.ID, pet.ID, pet.Name)
		if visit.RequestPreconsult {
			if err := a.issuePreconsultInvite(ctx, pet, visit); err != nil && !errors.Is(err, errPreconsultAlreadySent) {
				log.Printf("preconsult: issue invite visit %s: %v", visit.ID, err)
			}
			return
		}
		a.emailVisitConfirmedAffiliate(ctx, pet, visit)
	}()
}

// onVisitRescheduleAccepted notifies via push only — never re-issues preconsult mail/token.
func (a *API) onVisitRescheduleAccepted(pet store.Pet, visit store.Visit) {
	go a.pushVisitConfirmed(pet.OwnerUserID, visit.ID, pet.ID, pet.Name)
}

func (a *API) emailVisitPreconsult(ctx context.Context, pet store.Pet, visit store.Visit, token string) {
	if a.notifier == nil {
		return
	}
	prefs, err := a.store.GetClientNotificationPrefs(ctx, pet.OwnerUserID)
	if err != nil || !prefs.Visits {
		return
	}
	client, err := a.store.GetUserByID(ctx, pet.OwnerUserID)
	if err != nil || strings.TrimSpace(client.Email) == "" {
		return
	}
	locale := client.PreferredLocale
	if locale == "" {
		locale = "fr"
	}
	clientName := client.FullName
	if clientName == "" {
		clientName = client.Email
	}
	when := formatVisitWhen(visit)
	practiceName := ""
	if contact, err := a.store.GetPracticeContact(ctx, pet.PracticeID); err == nil {
		practiceName = contact.PracticeName
	}
	ctaURL := strings.TrimRight(a.cfg.ProPublicSiteURL, "/") + "/preconsult/" + token
	if err := a.notifier.SendVisitPreconsult(client.Email, locale, clientName, pet.Name, when, practiceName, ctaURL); err != nil {
		log.Printf("preconsult: email visit %s: %v", visit.ID, err)
	}
}

func (a *API) emailVisitConfirmedAffiliate(ctx context.Context, pet store.Pet, visit store.Visit) {
	if a.notifier == nil {
		return
	}
	prefs, err := a.store.GetClientNotificationPrefs(ctx, pet.OwnerUserID)
	if err != nil || !prefs.Visits {
		return
	}
	client, err := a.store.GetUserByID(ctx, pet.OwnerUserID)
	if err != nil || strings.TrimSpace(client.Email) == "" {
		return
	}
	locale := client.PreferredLocale
	if locale == "" {
		locale = "fr"
	}
	clientName := client.FullName
	if clientName == "" {
		clientName = client.Email
	}
	when := formatVisitWhen(visit)
	practiceName := ""
	ctaURL := strings.TrimSpace(a.cfg.PetsAppDownloadURL)
	if inviteOwner := a.resolvePreconsultInviteOwner(ctx, pet.PracticeID, pet.OwnerUserID); inviteOwner != "" {
		if invite, err := a.store.EnsureAppInviteCode(ctx, inviteOwner); err == nil {
			ctaURL = a.appInviteWebURL(invite.Code)
			if invite.PracticeName != "" {
				practiceName = invite.PracticeName
			}
		}
	}
	if practiceName == "" {
		if contact, err := a.store.GetPracticeContact(ctx, pet.PracticeID); err == nil {
			practiceName = contact.PracticeName
		}
	}
	if ctaURL == "" {
		ctaURL = strings.TrimRight(a.cfg.ProPublicSiteURL, "/")
	}
	if err := a.notifier.SendVisitConfirmedAffiliate(client.Email, locale, clientName, pet.Name, when, practiceName, ctaURL); err != nil {
		log.Printf("visit affiliate email visit %s: %v", visit.ID, err)
	}
}

func formatVisitWhen(visit store.Visit) string {
	loc, _ := time.LoadLocation("Europe/Brussels")
	if loc == nil {
		loc = time.Local
	}
	if visit.ScheduledAt != nil {
		return visit.ScheduledAt.In(loc).Format("02/01/2006 15:04")
	}
	return ""
}

// resolvePreconsultInviteOwner picks a practice user who can issue an app-invite code.
func (a *API) resolvePreconsultInviteOwner(ctx context.Context, practiceID, clientUserID string) string {
	if ref, err := a.store.PracticeReferenceVetUserID(ctx, practiceID); err == nil && ref != "" {
		return ref
	}
	vets, err := a.store.ListVetsForVisitAlert(ctx, practiceID, clientUserID)
	if err == nil {
		for _, v := range vets {
			if v.ID != "" {
				return v.ID
			}
		}
	}
	return ""
}
