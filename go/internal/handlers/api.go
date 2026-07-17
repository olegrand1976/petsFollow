package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/billing"
	"github.com/olegrand1976/petsFollow/go/internal/notifications/email"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
	"golang.org/x/crypto/bcrypt"
)

type API struct {
	store    *store.Store
	tokens   *authx.TokenIssuer
	cfg      config.Config
	notifier *email.Notifier
	billing  *billing.Service
	media    media.Store
}

func NewAPI(st *store.Store, tokens *authx.TokenIssuer, cfg config.Config, notifier *email.Notifier, bill *billing.Service, mediaStore media.Store) *API {
	return &API{store: st, tokens: tokens, cfg: cfg, notifier: notifier, billing: bill, media: mediaStore}
}

func (a *API) Routes(r chi.Router) {
	r.Use(httpx.LocaleMiddleware)
	r.Post("/auth/login", a.login)
	r.Post("/auth/register", a.register)
	r.Post("/auth/confirm-email", a.confirmEmail)
	r.Post("/auth/forgot-password", a.forgotPassword)
	r.Post("/auth/reset-password", a.resetPassword)
	r.Post("/auth/refresh", a.refresh)
	a.registerAuthRoutes(r)
	a.registerBillingRoutes(r)
	a.registerAdminRoutes(r)
	a.registerCommissionRoutes(r)
	a.registerCommercialRoutes(r)

	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/me", a.me)
		pr.Patch("/me", a.updateMe)
		pr.Post("/me/avatar", a.uploadMyAvatar)
		pr.Patch("/me/password", a.changeMePassword)
		pr.Delete("/me", a.deleteMe)
		pr.Patch("/me/locale", a.updateMeLocale)
		pr.Get("/me/vets", a.listMyVets)
		pr.Post("/me/vets/invite", a.inviteVet)
		pr.Get("/vet/link-requests", a.listVetLinkRequests)
		pr.Post("/vet/link-requests/{id}/accept", a.acceptVetLinkRequest)
		pr.Post("/vet/link-requests/{id}/reject", a.rejectVetLinkRequest)
		pr.Get("/vet/visits", a.listVetVisits)
		pr.Get("/vet/care-reminders", a.listVetOverdueCare)
		pr.Get("/me/discovery", a.getDiscovery)
		pr.Post("/me/discovery/complete", a.completeDiscovery)
		pr.Put("/me/device-tokens", a.putDeviceToken)
		pr.Get("/me/notification-preferences", a.getClientNotificationPrefs)
		pr.Patch("/me/notification-preferences", a.updateClientNotificationPrefs)
		pr.Get("/me/household", a.getHousehold)
		pr.Get("/clients", a.listClients)
		pr.Get("/clients/{clientID}", a.getClient)
		pr.Post("/clients/{clientID}/send-app-link", a.sendClientAppLink)
		pr.Get("/clients/{clientID}/pets", a.listClientPets)
		pr.Get("/pets", a.listMyPets)
		pr.Post("/pets", a.createPet)
		pr.Patch("/pets/{petID}/primary-practice", a.setPetPrimaryPractice)
		pr.Get("/pets/{petID}/care-reminders", a.listCareReminders)
		pr.Post("/pets/{petID}/care-reminders", a.createCareReminder)
		pr.Get("/pets/{petID}/horse-contacts", a.listHorseContacts)
		pr.Post("/pets/{petID}/horse-contacts", a.createHorseContact)
		pr.Delete("/horse-contacts/{id}", a.deleteHorseContact)
		pr.Get("/pets/{petID}/horse-competitions", a.listHorseCompetitions)
		pr.Post("/pets/{petID}/horse-competitions", a.createHorseCompetition)
		pr.Delete("/horse-competitions/{id}", a.deleteHorseCompetition)
		pr.Get("/pets/{petID}/visits", a.listVisits)
		pr.Post("/pets/{petID}/visits", a.createVisit)
		pr.Put("/pets/{petID}", a.updatePet)
		pr.Post("/pets/{petID}/photo", a.uploadPetPhoto)
		pr.Get("/pets/{petID}", a.getPet)
		pr.Get("/pets/{petID}/timeline", a.petTimeline)
		pr.Post("/pets/{petID}/heartrate/sessions", a.startHeartRate)
		pr.Get("/pets/{petID}/heartrate/sessions", a.listHeartRate)
		pr.Patch("/heartrate/sessions/{sessionID}", a.completeHeartRate)
		pr.Post("/heartrate/sessions/{sessionID}/validate", a.validateHeartRate)
		pr.Post("/heartrate/sessions/{sessionID}/cancel", a.cancelHeartRate)
		pr.Post("/care-reminders/{id}/done", a.markCareReminderDone)
		pr.Post("/care-reminders/{id}/postpone", a.postponeCareReminder)
		pr.Patch("/visits/{id}", a.updateVisit)
		pr.Get("/messaging/threads", a.listThreads)
		pr.Post("/messaging/threads/read-all", a.markAllThreadsRead)
		pr.Get("/messaging/threads/{threadID}/messages", a.listMessages)
		pr.Post("/messaging/threads/{threadID}/messages", a.sendMessage)
		pr.Post("/messaging/threads/{threadID}/messages/media", a.sendMessageMedia)
		pr.Post("/messaging/threads/{threadID}/read", a.markThreadRead)
		pr.Put("/vet/availability", a.setAvailability)
		pr.Get("/vet/availability", a.getAvailability)
		pr.Get("/vet/overview", a.vetOverview)
		pr.Get("/vet/profile", a.getVetProfile)
		pr.Put("/vet/profile", a.updateVetProfile)
		pr.Post("/vet/prospects", a.vetCreateProspect)
		pr.Get("/vet/notification-preferences", a.getVetEmailPrefs)
		pr.Put("/vet/notification-preferences", a.updateVetEmailPrefs)
	})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	u, err := a.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if u.PasswordHash == "" {
		writeErr(w, r, http.StatusUnauthorized, "use_google_sign_in", "use_google_sign_in")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if u.Role == kernel.RoleVet && u.EmailVerifiedAt == nil {
		writeErr(w, r, http.StatusForbidden, "email_not_verified", "email_not_verified")
		return
	}
	a.issueLoginResponse(w, r, u)
}

type refreshReq struct {
	RefreshToken string `json:"refreshToken"`
}

func (a *API) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.RefreshToken == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "refresh_token_required")
		return
	}
	id, err := a.tokens.ParseRefresh(req.RefreshToken)
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_token")
		return
	}
	u, err := a.store.GetUserByID(r.Context(), id.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_token")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	pair, err := a.tokens.Issue(u.ID, u.Email, u.Role, u.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, pair)
}

func (a *API) me(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	data, err := a.store.GetUserMe(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, data)
}

func (a *API) listClients(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	clients, err := a.store.ListClientsByPractice(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, clients)
}

func (a *API) getClient(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	client, err := a.store.GetClientByPractice(r.Context(), id.PracticeID, chi.URLParam(r, "clientID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "client_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, client)
}

func (a *API) sendClientAppLink(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	downloadURL := strings.TrimSpace(a.cfg.PetsAppDownloadURL)
	if downloadURL == "" {
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "app_download_url_missing")
		return
	}
	client, err := a.store.GetClientByPractice(r.Context(), id.PracticeID, chi.URLParam(r, "clientID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "client_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	clientUser, err := a.store.GetUserByID(r.Context(), client.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	vet, err := a.store.GetUserByID(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	practiceName := ""
	if profile, err := a.store.GetPracticeProfile(r.Context(), id.PracticeID, id.UserID); err == nil {
		practiceName = profile.PracticeName
	}
	if practiceName == "" {
		practiceName = "petsFollow"
	}
	locale := clientUser.PreferredLocale
	if locale == "" {
		locale = localeOf(r)
	}
	clientName := client.FullName
	if clientName == "" {
		clientName = client.Email
	}
	vetName := vet.FullName
	if vetName == "" {
		vetName = vet.Email
	}
	if err := a.notifier.SendAppDownloadInvite(client.Email, locale, clientName, vetName, practiceName, downloadURL); err != nil {
		writeErr(w, r, http.StatusBadGateway, "email_send_failed", "email_send_failed")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{
		"status":  "sent",
		"email":   client.Email,
		"message": t(r, "success.app_link_sent", nil),
	})
}

func (a *API) listClientPets(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	pets, err := a.store.ListPetsByClientForVet(r.Context(), id.PracticeID, chi.URLParam(r, "clientID"))
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, pets)
}

func (a *API) listMyPets(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	pets, err := a.store.ListPetsByOwner(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, pets)
}

type petReq struct {
	Name        string   `json:"name"`
	Species     string   `json:"species"`
	Breed       string   `json:"breed"`
	BirthDate   *string  `json:"birthDate"`
	WeightKg    *float64 `json:"weightKg"`
	PhotoURL    string   `json:"photoUrl"`
	Plan        string   `json:"plan"`
	BillingMode string   `json:"billingMode"`
	SuccessURL  string   `json:"successUrl"`
	CancelURL   string   `json:"cancelUrl"`
}

func (a *API) createPet(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var req petReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err := a.store.AssertFamilyCanAddPet(r.Context(), id.UserID); err != nil {
		if errors.Is(err, store.ErrFamilyPetLimit) {
			writeErr(w, r, http.StatusConflict, "family_limit", "family_pet_limit")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	p := store.Pet{Name: req.Name, Species: req.Species, Breed: req.Breed, WeightKg: req.WeightKg, PhotoURL: req.PhotoURL, OwnerUserID: id.UserID, PracticeID: id.PracticeID, PaymentStatus: "pending_payment"}
	if req.BirthDate != nil {
		if t, err := time.Parse("2006-01-02", *req.BirthDate); err == nil {
			p.BirthDate = &t
		}
	}
	created, err := a.store.CreatePet(r.Context(), p)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if err := a.store.SeedDefaultCareReminders(r.Context(), created.ID, created.PracticeID, created.Species); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if created.Species == "horse" {
		if has, err := a.store.HasActiveAddon(r.Context(), id.UserID, string(billing.AddonHorse)); err == nil && has {
			_ = a.store.SeedHorsePackReminders(r.Context(), id.UserID)
		}
	}
	a.startPetBillingCheckout(w, r, created, id, createPetBilling{
		Plan: req.Plan, BillingMode: req.BillingMode, SuccessURL: req.SuccessURL, CancelURL: req.CancelURL,
	})
}

func (a *API) updatePet(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var req petReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	p := store.Pet{ID: chi.URLParam(r, "petID"), Name: req.Name, Species: req.Species, Breed: req.Breed, WeightKg: req.WeightKg, PhotoURL: req.PhotoURL, OwnerUserID: id.UserID}
	if req.BirthDate != nil {
		if t, err := time.Parse("2006-01-02", *req.BirthDate); err == nil {
			p.BirthDate = &t
		}
	}
	if err := a.store.UpdatePet(r.Context(), p); err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *API) getPet(w http.ResponseWriter, r *http.Request) {
	pet, err := a.store.GetPet(r.Context(), chi.URLParam(r, "petID"))
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	id, _ := authx.FromContext(r.Context())
	if id.Role == kernel.RoleClient && pet.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	if id.Role == kernel.RoleVet && pet.PracticeID != id.PracticeID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
		return
	}
	httpx.WriteData(w, http.StatusOK, pet)
}

func (a *API) petTimeline(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	switch id.Role {
	case kernel.RoleClient:
		if pet.OwnerUserID != id.UserID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
			return
		}
	case kernel.RoleVet:
		if pet.PracticeID != id.PracticeID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
			return
		}
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	vetView := id.Role == kernel.RoleVet
	items, err := a.store.PetTimeline(r.Context(), petID, vetView)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type startHRReq struct {
	DurationSec int `json:"durationSec"`
}

func (a *API) startHeartRate(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	pet, err := a.store.GetPet(r.Context(), chi.URLParam(r, "petID"))
	if err != nil || pet.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	if !a.requirePremiumAccess(w, r, pet.ID) {
		return
	}
	var req startHRReq
	_ = httpx.DecodeJSON(r, &req)
	allowed, err := a.store.GetPracticeHeartRateDurations(r.Context(), pet.PracticeID)
	if err != nil {
		allowed = nil
	}
	normalized := kernel.NormalizeHeartRateDurations(allowed)
	durationSec := req.DurationSec
	if durationSec == 0 {
		// Default to the longest duration enabled by the vet (practice settings).
		durationSec = normalized[len(normalized)-1]
	}
	ok := false
	for _, d := range normalized {
		if d == durationSec {
			ok = true
			break
		}
	}
	if !ok {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_duration")
		return
	}
	sess, err := a.store.StartHeartRateSession(r.Context(), pet.ID, id.UserID, pet.PracticeID, durationSec)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, sess)
}

type completeHRReq struct {
	TapCount int `json:"tapCount"`
}

func (a *API) completeHeartRate(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var req completeHRReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	sess, err := a.store.GetHeartRateSession(r.Context(), chi.URLParam(r, "sessionID"), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "session_not_found")
		return
	}
	bpm := kernel.CalculateBPM(req.TapCount, sess.DurationSec)
	alert := kernel.IsHeartRateAlert(bpm, a.cfg.HeartRateMinBPM, a.cfg.HeartRateMaxBPM)
	sess, err = a.store.CompleteHeartRateSession(r.Context(), chi.URLParam(r, "sessionID"), id.UserID, req.TapCount, bpm, alert)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "session_not_found")
		return
	}
	httpx.WriteData(w, http.StatusOK, sess)
}

func (a *API) validateHeartRate(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	sess, err := a.store.ValidateHeartRateSession(r.Context(), chi.URLParam(r, "sessionID"), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "session_not_found")
		return
	}
	vetID, _ := a.store.GetVetForClient(r.Context(), id.UserID, id.PracticeID)
	onMsg, onHR, _ := a.store.EmailPrefs(r.Context(), vetID)
	if onHR {
		vet, _ := a.store.GetUserByID(r.Context(), vetID)
		locale := vet.PreferredLocale
		if locale == "" {
			locale = localeOf(r)
		}
		_ = a.notifier.SendHeartrateValidated(vet.Email, locale, *sess.BPM)
		_ = a.store.LogNotification(r.Context(), vetID, "heartrate_validated", map[string]any{"sessionId": sess.ID, "bpm": sess.BPM})
		_ = onMsg
	}
	httpx.WriteData(w, http.StatusOK, sess)
}

func (a *API) cancelHeartRate(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	if err := a.store.CancelHeartRateSession(r.Context(), chi.URLParam(r, "sessionID"), id.UserID); err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "session_not_found")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (a *API) listHeartRate(w http.ResponseWriter, r *http.Request) {
	id, _ := authx.FromContext(r.Context())
	vetView := id.Role == kernel.RoleVet
	sessions, err := a.store.ListHeartRateSessions(r.Context(), chi.URLParam(r, "petID"), vetView)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, sessions)
}

func (a *API) listThreads(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if id.Role == kernel.RoleVet {
		threads, err := a.store.ListThreadSummariesForVet(r.Context(), id.UserID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		httpx.WriteData(w, http.StatusOK, threads)
		return
	}
	vetID, err := a.store.GetVetForClient(r.Context(), id.UserID, id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	t, err := a.store.GetOrCreateThread(r.Context(), id.PracticeID, id.UserID, vetID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, []store.Thread{t})
}

func (a *API) listMessages(w http.ResponseWriter, r *http.Request) {
	thread, err := a.store.GetThreadByID(r.Context(), chi.URLParam(r, "threadID"))
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "thread_not_found")
		return
	}
	id, _ := authx.FromContext(r.Context())
	if id.UserID != thread.ClientUserID && id.UserID != thread.VetUserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_participant")
		return
	}
	msgs, err := a.store.ListMessages(r.Context(), thread.ID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, msgs)
}

func (a *API) markThreadRead(w http.ResponseWriter, r *http.Request) {
	thread, err := a.store.GetThreadByID(r.Context(), chi.URLParam(r, "threadID"))
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "thread_not_found")
		return
	}
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if id.UserID != thread.ClientUserID && id.UserID != thread.VetUserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_participant")
		return
	}
	if err := a.store.MarkThreadRead(r.Context(), thread.ID, id.UserID); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) markAllThreadsRead(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	if err := a.store.MarkAllUnreadForUser(r.Context(), id.UserID); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]bool{"ok": true})
}

type msgReq struct {
	Body string `json:"body"`
}

func (a *API) sendMessage(w http.ResponseWriter, r *http.Request) {
	thread, err := a.store.GetThreadByID(r.Context(), chi.URLParam(r, "threadID"))
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "thread_not_found")
		return
	}
	id, _ := authx.FromContext(r.Context())
	if id.UserID != thread.ClientUserID && id.UserID != thread.VetUserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_participant")
		return
	}
	var req msgReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.Body == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "body_required")
		return
	}
	if id.Role == kernel.RoleClient {
		if thread.PetID != "" && !a.requirePremiumAccess(w, r, thread.PetID) {
			return
		}
		status, autoReply, _ := a.store.GetVetAvailability(r.Context(), thread.VetUserID)
		if status == kernel.AvailabilityUnavailable && autoReply != "" {
			_, _ = a.store.AddMessage(r.Context(), thread.ID, thread.VetUserID, autoReply)
		}
	}
	msg, err := a.store.AddMessage(r.Context(), thread.ID, id.UserID, req.Body)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if id.Role == kernel.RoleClient {
		onMsg, _, _ := a.store.EmailPrefs(r.Context(), thread.VetUserID)
		if onMsg {
			vet, _ := a.store.GetUserByID(r.Context(), thread.VetUserID)
			locale := vet.PreferredLocale
			if locale == "" {
				locale = localeOf(r)
			}
			_ = a.notifier.SendNewMessage(vet.Email, locale, req.Body)
		}
	}
	httpx.WriteData(w, http.StatusCreated, msg)
}

func (a *API) sendMessageMedia(w http.ResponseWriter, r *http.Request) {
	thread, err := a.store.GetThreadByID(r.Context(), chi.URLParam(r, "threadID"))
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "thread_not_found")
		return
	}
	id, _ := authx.FromContext(r.Context())
	if id.UserID != thread.ClientUserID && id.UserID != thread.VetUserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_participant")
		return
	}
	if id.Role == kernel.RoleClient {
		if thread.PetID != "" && !a.requirePremiumAccess(w, r, thread.PetID) {
			return
		}
	}
	url, ct, err := a.uploadMessageMedia(r, "messages", thread.ID)
	if err != nil {
		a.writeUploadErr(w, r, err)
		return
	}
	body := strings.TrimSpace(r.FormValue("body"))
	kind := media.MediaKind(ct)
	if id.Role == kernel.RoleClient {
		status, autoReply, _ := a.store.GetVetAvailability(r.Context(), thread.VetUserID)
		if status == kernel.AvailabilityUnavailable && autoReply != "" {
			_, _ = a.store.AddMessage(r.Context(), thread.ID, thread.VetUserID, autoReply)
		}
	}
	msg, err := a.store.AddMessageMedia(r.Context(), thread.ID, id.UserID, body, url, kind)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if id.Role == kernel.RoleClient {
		onMsg, _, _ := a.store.EmailPrefs(r.Context(), thread.VetUserID)
		if onMsg {
			vet, _ := a.store.GetUserByID(r.Context(), thread.VetUserID)
			locale := vet.PreferredLocale
			if locale == "" {
				locale = localeOf(r)
			}
			preview := body
			if preview == "" {
				if kind == "video" {
					preview = "[video]"
				} else {
					preview = "[image]"
				}
			}
			_ = a.notifier.SendNewMessage(vet.Email, locale, preview)
		}
	}
	httpx.WriteData(w, http.StatusCreated, msg)
}

type availReq struct {
	Status    kernel.AvailabilityStatus `json:"status"`
	AutoReply string                    `json:"autoReply"`
}

func (a *API) setAvailability(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	var req availReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err := a.store.SetVetAvailability(r.Context(), id.UserID, id.PracticeID, req.Status, req.AutoReply); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": string(req.Status)})
}

func (a *API) vetOverview(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	overview, err := a.store.VetOverview(r.Context(), id.PracticeID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, overview)
}

func (a *API) getAvailability(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	status, autoReply, err := a.store.GetVetAvailability(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"status": status, "autoReply": autoReply})
}

type registerReq struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	FullName     string `json:"fullName"`
	PracticeName string `json:"practiceName"`
}

func (a *API) register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" || req.FullName == "" || req.PracticeName == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	if len(req.Password) < 8 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "password_too_short")
		return
	}
	if _, err := a.store.GetUserByEmail(r.Context(), req.Email); err == nil {
		writeErr(w, r, http.StatusConflict, "conflict", "email_already_exists")
		return
	} else if !errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	locale := localeOf(r)
	result, err := a.store.RegisterVet(r.Context(), store.RegisterVetInput{
		Email: req.Email, Password: req.Password, FullName: req.FullName, PracticeName: req.PracticeName,
		PreferredLocale: locale, AutoReplyDefault: t(r, "defaults.auto_reply_unavailable", nil),
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	confirmURL := fmt.Sprintf("%s/confirm-email?token=%s", a.cfg.ProPublicSiteURL, result.Token)
	_ = a.notifier.SendConfirmRegistration(req.Email, locale, req.FullName, confirmURL)
	httpx.WriteData(w, http.StatusCreated, map[string]any{
		"message":     t(r, "success.confirm_email_sent", nil),
		"confirmPath": "/confirm-email?token=" + result.Token,
	})
}

type confirmEmailReq struct {
	Token string `json:"token"`
}

func (a *API) confirmEmail(w http.ResponseWriter, r *http.Request) {
	var req confirmEmailReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if req.Token == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "token_required")
		return
	}
	u, err := a.store.ConfirmEmail(r.Context(), req.Token)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "invalid_confirm_link")
			return
		}
		writeErr(w, r, http.StatusBadRequest, "bad_request", "internal")
		return
	}
	pair, err := a.tokens.Issue(u.ID, u.Email, u.Role, u.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"message":      t(r, "success.email_confirmed", nil),
		"email":        u.Email,
		"accessToken":  pair.AccessToken,
		"refreshToken": pair.RefreshToken,
		"expiresIn":    pair.ExpiresIn,
	})
}

type forgotPasswordReq struct {
	Email string `json:"email"`
}

func (a *API) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}

	result, err := a.store.RequestPasswordReset(r.Context(), req.Email)
	out := map[string]any{
		"message": t(r, "success.password_reset_sent", nil),
	}
	if err == nil {
		resetURL := fmt.Sprintf("%s/reset-password?token=%s", a.cfg.ProPublicSiteURL, result.Token)
		_ = a.notifier.SendPasswordReset(result.Email, result.Locale, result.FullName, resetURL)
		out["resetPath"] = "/reset-password?token=" + result.Token
	}
	// Always 200 — do not reveal whether the email exists.
	httpx.WriteData(w, http.StatusOK, out)
}

type resetPasswordReq struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (a *API) resetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if req.Token == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "token_required")
		return
	}
	if len(req.Password) < 8 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "password_too_short")
		return
	}
	if err := a.store.ResetPassword(r.Context(), req.Token, req.Password); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "invalid_reset_link")
			return
		}
		if err.Error() == "password_too_short" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "password_too_short")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"message": t(r, "success.password_reset", nil),
	})
}

func (a *API) getVetProfile(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	profile, err := a.store.GetPracticeProfile(r.Context(), id.PracticeID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, profile)
}

func (a *API) updateVetProfile(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	var req store.PracticeProfile
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if req.PracticeName == "" || req.Phone == "" || req.ContactEmail == "" ||
		req.AddressLine1 == "" || req.City == "" || req.PostalCode == "" || req.VetFullName == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "profile_fields_required")
		return
	}
	req.HeartRateDurationsSec = kernel.NormalizeHeartRateDurations(req.HeartRateDurationsSec)
	markComplete := r.URL.Query().Get("complete") == "true"
	if err := a.store.UpdatePracticeProfile(r.Context(), id.PracticeID, id.UserID, req, markComplete); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	profile, err := a.store.GetPracticeProfile(r.Context(), id.PracticeID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, profile)
}
