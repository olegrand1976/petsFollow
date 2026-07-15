package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/billing"
	"github.com/olegrand1976/petsFollow/go/internal/notifications/email"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
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
}

func NewAPI(st *store.Store, tokens *authx.TokenIssuer, cfg config.Config, notifier *email.Notifier, bill *billing.Service) *API {
	return &API{store: st, tokens: tokens, cfg: cfg, notifier: notifier, billing: bill}
}

func (a *API) Routes(r chi.Router) {
	r.Use(httpx.LocaleMiddleware)
	r.Post("/auth/login", a.login)
	r.Post("/auth/register", a.register)
	r.Post("/auth/confirm-email", a.confirmEmail)
	r.Post("/auth/refresh", a.refresh)
	a.registerAuthRoutes(r)
	a.registerBillingRoutes(r)
	a.registerAdminRoutes(r)

	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/me", a.me)
		pr.Patch("/me", a.updateMe)
		pr.Patch("/me/password", a.changeMePassword)
		pr.Delete("/me", a.deleteMe)
		pr.Patch("/me/locale", a.updateMeLocale)
		pr.Get("/clients", a.listClients)
		pr.Get("/clients/{clientID}", a.getClient)
		pr.Get("/clients/{clientID}/pets", a.listClientPets)
		pr.Get("/pets", a.listMyPets)
		pr.Post("/pets", a.createPet)
		pr.Put("/pets/{petID}", a.updatePet)
		pr.Get("/pets/{petID}", a.getPet)
		pr.Get("/pets/{petID}/timeline", a.petTimeline)
		pr.Post("/pets/{petID}/heartrate/sessions", a.startHeartRate)
		pr.Get("/pets/{petID}/heartrate/sessions", a.listHeartRate)
		pr.Patch("/heartrate/sessions/{sessionID}", a.completeHeartRate)
		pr.Post("/heartrate/sessions/{sessionID}/validate", a.validateHeartRate)
		pr.Post("/heartrate/sessions/{sessionID}/cancel", a.cancelHeartRate)
		pr.Get("/messaging/threads", a.listThreads)
		pr.Get("/messaging/threads/{threadID}/messages", a.listMessages)
		pr.Post("/messaging/threads/{threadID}/messages", a.sendMessage)
		pr.Put("/vet/availability", a.setAvailability)
		pr.Get("/vet/availability", a.getAvailability)
		pr.Get("/vet/overview", a.vetOverview)
		pr.Get("/vet/profile", a.getVetProfile)
		pr.Put("/vet/profile", a.updateVetProfile)
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

func (a *API) refresh(w http.ResponseWriter, r *http.Request) {
	writeErr(w, r, http.StatusNotImplemented, "not_implemented", "not_implemented")
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
	id, _ := authx.FromContext(r.Context())
	vetView := id.Role == kernel.RoleVet
	items, err := a.store.PetTimeline(r.Context(), chi.URLParam(r, "petID"), vetView)
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
	durationSec := req.DurationSec
	if durationSec == 0 {
		durationSec = a.cfg.HeartRateSeconds
	}
	allowed, err := a.store.GetPracticeHeartRateDurations(r.Context(), pet.PracticeID)
	if err != nil {
		allowed = []int{a.cfg.HeartRateSeconds}
	}
	normalized := kernel.NormalizeHeartRateDurations(allowed)
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
	out := map[string]any{
		"message": t(r, "success.email_confirmed", nil),
		"email":   u.Email,
	}
	// Session ouverte directement après confirmation (évite un re-login inutile).
	if pair, err := a.tokens.Issue(u.ID, u.Email, u.Role, u.PracticeID); err == nil {
		out["accessToken"] = pair.AccessToken
		out["refreshToken"] = pair.RefreshToken
		out["expiresIn"] = pair.ExpiresIn
	}
	httpx.WriteData(w, http.StatusOK, out)
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
