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
	r.Post("/auth/login", a.login)
	r.Post("/auth/refresh", a.refresh)
	a.registerBillingRoutes(r)
	a.registerAdminRoutes(r)

	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Get("/me", a.me)
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
	})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	u, err := a.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return
	}
	pair, err := a.tokens.Issue(u.ID, u.Email, u.Role, u.PracticeID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, pair)
}

func (a *API) refresh(w http.ResponseWriter, r *http.Request) {
	httpx.WriteError(w, http.StatusNotImplemented, "not_implemented", "use login in MVP")
}

func (a *API) me(w http.ResponseWriter, r *http.Request) {
	id, _ := authx.FromContext(r.Context())
	httpx.WriteData(w, http.StatusOK, id)
}

func (a *API) listClients(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "vet only")
		return
	}
	clients, err := a.store.ListClientsByPractice(r.Context(), id.PracticeID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, clients)
}

func (a *API) getClient(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "vet only")
		return
	}
	client, err := a.store.GetClientByPractice(r.Context(), id.PracticeID, chi.URLParam(r, "clientID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "not_found", "client not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, client)
}

func (a *API) listClientPets(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "vet only")
		return
	}
	pets, err := a.store.ListPetsByClientForVet(r.Context(), id.PracticeID, chi.URLParam(r, "clientID"))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, pets)
}

func (a *API) listMyPets(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}
	pets, err := a.store.ListPetsByOwner(r.Context(), id.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
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
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "client only")
		return
	}
	var req petReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "invalid json")
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
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	a.startPetBillingCheckout(w, r, created, id, createPetBilling{
		Plan: req.Plan, BillingMode: req.BillingMode, SuccessURL: req.SuccessURL, CancelURL: req.CancelURL,
	})
}

func (a *API) updatePet(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "client only")
		return
	}
	var req petReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	p := store.Pet{ID: chi.URLParam(r, "petID"), Name: req.Name, Species: req.Species, Breed: req.Breed, WeightKg: req.WeightKg, PhotoURL: req.PhotoURL, OwnerUserID: id.UserID}
	if req.BirthDate != nil {
		if t, err := time.Parse("2006-01-02", *req.BirthDate); err == nil {
			p.BirthDate = &t
		}
	}
	if err := a.store.UpdatePet(r.Context(), p); err != nil {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "pet not found")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *API) getPet(w http.ResponseWriter, r *http.Request) {
	pet, err := a.store.GetPet(r.Context(), chi.URLParam(r, "petID"))
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "pet not found")
		return
	}
	id, _ := authx.FromContext(r.Context())
	if id.Role == kernel.RoleClient && pet.OwnerUserID != id.UserID {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "not your pet")
		return
	}
	if id.Role == kernel.RoleVet && pet.PracticeID != id.PracticeID {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "wrong practice")
		return
	}
	httpx.WriteData(w, http.StatusOK, pet)
}

func (a *API) petTimeline(w http.ResponseWriter, r *http.Request) {
	id, _ := authx.FromContext(r.Context())
	vetView := id.Role == kernel.RoleVet
	items, err := a.store.PetTimeline(r.Context(), chi.URLParam(r, "petID"), vetView)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

func (a *API) startHeartRate(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "client only")
		return
	}
	pet, err := a.store.GetPet(r.Context(), chi.URLParam(r, "petID"))
	if err != nil || pet.OwnerUserID != id.UserID {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "not your pet")
		return
	}
	if !a.requirePremiumAccess(w, r, pet.ID) {
		return
	}
	sess, err := a.store.StartHeartRateSession(r.Context(), pet.ID, id.UserID, pet.PracticeID, a.cfg.HeartRateSeconds)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
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
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "client only")
		return
	}
	var req completeHRReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	bpm := kernel.CalculateBPM(req.TapCount, a.cfg.HeartRateSeconds)
	alert := kernel.IsHeartRateAlert(bpm, a.cfg.HeartRateMinBPM, a.cfg.HeartRateMaxBPM)
	sess, err := a.store.CompleteHeartRateSession(r.Context(), chi.URLParam(r, "sessionID"), id.UserID, req.TapCount, bpm, alert)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "session not found")
		return
	}
	httpx.WriteData(w, http.StatusOK, sess)
}

func (a *API) validateHeartRate(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "client only")
		return
	}
	sess, err := a.store.ValidateHeartRateSession(r.Context(), chi.URLParam(r, "sessionID"), id.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "session not found")
		return
	}
	vetID, _ := a.store.GetVetForClient(r.Context(), id.UserID, id.PracticeID)
	onMsg, onHR, _ := a.store.EmailPrefs(r.Context(), vetID)
	if onHR {
		vet, _ := a.store.GetUserByID(r.Context(), vetID)
		subject := "petsFollow Pro — Relevé cardiaque validé"
		body := fmt.Sprintf("<p>Relevé cardiaque validé pour un patient.</p><p>BPM: %d</p>", *sess.BPM)
		_ = a.notifier.SendVetAlert(vet.Email, subject, body)
		_ = a.store.LogNotification(r.Context(), vetID, "heartrate_validated", map[string]any{"sessionId": sess.ID, "bpm": sess.BPM})
		_ = onMsg
	}
	httpx.WriteData(w, http.StatusOK, sess)
}

func (a *API) cancelHeartRate(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "client only")
		return
	}
	if err := a.store.CancelHeartRateSession(r.Context(), chi.URLParam(r, "sessionID"), id.UserID); err != nil {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "session not found")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (a *API) listHeartRate(w http.ResponseWriter, r *http.Request) {
	id, _ := authx.FromContext(r.Context())
	vetView := id.Role == kernel.RoleVet
	sessions, err := a.store.ListHeartRateSessions(r.Context(), chi.URLParam(r, "petID"), vetView)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, sessions)
}

func (a *API) listThreads(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}
	if id.Role == kernel.RoleVet {
		threads, err := a.store.ListThreadSummariesForVet(r.Context(), id.UserID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, threads)
		return
	}
	vetID, err := a.store.GetVetForClient(r.Context(), id.UserID, id.PracticeID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	t, err := a.store.GetOrCreateThread(r.Context(), id.PracticeID, id.UserID, vetID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, []store.Thread{t})
}

func (a *API) listMessages(w http.ResponseWriter, r *http.Request) {
	thread, err := a.store.GetThreadByID(r.Context(), chi.URLParam(r, "threadID"))
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "not_found", "thread not found")
		return
	}
	id, _ := authx.FromContext(r.Context())
	if id.UserID != thread.ClientUserID && id.UserID != thread.VetUserID {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "not participant")
		return
	}
	msgs, err := a.store.ListMessages(r.Context(), thread.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
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
		httpx.WriteError(w, http.StatusNotFound, "not_found", "thread not found")
		return
	}
	id, _ := authx.FromContext(r.Context())
	if id.UserID != thread.ClientUserID && id.UserID != thread.VetUserID {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "not participant")
		return
	}
	var req msgReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.Body == "" {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "body required")
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
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	if id.Role == kernel.RoleClient {
		onMsg, _, _ := a.store.EmailPrefs(r.Context(), thread.VetUserID)
		if onMsg {
			vet, _ := a.store.GetUserByID(r.Context(), thread.VetUserID)
			_ = a.notifier.SendVetAlert(vet.Email, "petsFollow Pro — Nouveau message", "<p>Nouveau message client.</p><p>"+req.Body+"</p>")
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
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "vet only")
		return
	}
	var req availReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	if err := a.store.SetVetAvailability(r.Context(), id.UserID, id.PracticeID, req.Status, req.AutoReply); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": string(req.Status)})
}

func (a *API) vetOverview(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "vet only")
		return
	}
	overview, err := a.store.VetOverview(r.Context(), id.PracticeID, id.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, overview)
}

func (a *API) getAvailability(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "vet only")
		return
	}
	status, autoReply, err := a.store.GetVetAvailability(r.Context(), id.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"status": status, "autoReply": autoReply})
}
