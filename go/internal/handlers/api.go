package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/billing"
	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
	"github.com/olegrand1976/petsFollow/go/internal/invoicing/billit"
	invoicingmock "github.com/olegrand1976/petsFollow/go/internal/invoicing/mock"
	"github.com/olegrand1976/petsFollow/go/internal/notifications/email"
	"github.com/olegrand1976/petsFollow/go/internal/notifications/fcm"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
	"golang.org/x/crypto/bcrypt"
)

type API struct {
	store      *store.Store
	tokens     *authx.TokenIssuer
	cfg        config.Config
	notifier   *email.Notifier
	billing    *billing.Service
	invoicing  *invoicing.Service
	media      media.Store
	pusher   fcm.Pusher
	gemini   *gemini.Client
	// vetLookupRL / vetSuggestRL — anti-scraping / anti-spam (par userId).
	vetLookupRL      *httpx.RateLimiter
	vetSuggestRL     *httpx.RateLimiter
	billitWebhookRL  *httpx.RateLimiter
	authPulse        *authPulse
}

func NewAPI(st *store.Store, tokens *authx.TokenIssuer, cfg config.Config, notifier *email.Notifier, bill *billing.Service, mediaStore media.Store, pusher fcm.Pusher) *API {
	if pusher == nil {
		pusher = fcm.NopPusher{}
	}
	var g *gemini.Client
	if cfg.GeminiAPIKey != "" {
		g = gemini.New(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.GeminiLiteModel)
	}
	var inv *invoicing.Service
	if cfg.BillitEnabled {
		var gw invoicing.Gateway
		if cfg.BillitMockEnabled {
			gw = invoicingmock.New()
		} else {
			gw = billit.NewClient(cfg.BillitBaseURL)
		}
		inv = invoicing.NewService(st, gw, cfg)
	}
	return &API{
		store: st, tokens: tokens, cfg: cfg, notifier: notifier, billing: bill, invoicing: inv, media: mediaStore, pusher: pusher, gemini: g,
		vetLookupRL:     httpx.NewRateLimiter(30, time.Minute),
		vetSuggestRL:    httpx.NewRateLimiter(10, time.Minute),
		billitWebhookRL: httpx.NewRateLimiter(120, time.Minute),
		authPulse:       newAuthPulse(),
	}
}

// TestReplaceNotifier swaps the email notifier (integration tests only).
func (a *API) TestReplaceNotifier(n *email.Notifier) { a.notifier = n }

// TestSetMedia installs a media store (integration tests only).
func (a *API) TestSetMedia(m media.Store) { a.media = m }

// TestSetOpsNotifyEmail sets OPS_NOTIFY_EMAIL (integration tests only).
func (a *API) TestSetOpsNotifyEmail(addr string) { a.cfg.OpsNotifyEmail = addr }

// TestSetAdminStagingSeedEnabled toggles ADMIN_STAGING_SEED_ENABLED (integration tests only).
func (a *API) TestSetAdminStagingSeedEnabled(v bool) { a.cfg.AdminStagingSeedEnabled = v }

// TestSetBillitWebhookSecret sets BILLIT_WEBHOOK_SECRET (integration tests only).
func (a *API) TestSetBillitWebhookSecret(secret string) { a.cfg.BillitWebhookSecret = secret }

func (a *API) Routes(r chi.Router) {
	r.Use(httpx.LocaleMiddleware)
	// Anti brute-force / spam sur les endpoints auth publics (par IP).
	authRL := httpx.NewRateLimiter(a.cfg.AuthRateLimitPerMin, time.Minute)
	r.Group(func(ar chi.Router) {
		ar.Use(authRL.Middleware)
		ar.Post("/auth/login", a.login)
		ar.Post("/auth/register", a.register)
		ar.Post("/auth/register-client", a.registerClient)
		ar.Post("/auth/register-care-pro", a.registerCarePro)
		ar.Post("/auth/confirm-email", a.confirmEmail)
		ar.Post("/auth/resend-confirmation", a.resendConfirmation)
		ar.Post("/auth/forgot-password", a.forgotPassword)
		ar.Post("/auth/reset-password", a.resetPassword)
	})
	r.Post("/auth/refresh", a.refresh)
	a.registerJourneyPublicRoutes(r)
	a.registerAppInviteRoutes(r, authRL.Middleware)
	a.registerPreconsultPublicRoutes(r, authRL.Middleware)
	a.registerDossierSharePublicRoutes(r, authRL.Middleware)
	a.registerCommercialDiscoveryRoutes(r, authRL.Middleware)
	a.registerAuthRoutes(r, authRL.Middleware)
	a.registerBillingRoutes(r)
	a.registerInvoicingRoutes(r)
	a.registerInvoicingWebhookRoutes(r)
	a.registerAdminRoutes(r)
	a.registerBrandAssetAdminRoutes(r)
	a.registerCommissionRoutes(r)
	a.registerCommercialRoutes(r)
	a.registerCommercialManagerRoutes(r)
	a.registerPitchTrainingRoutes(r)
	a.registerProductDigestRoutes(r)
	a.registerAiCrModuleRoutes(r)
	a.registerSupportRoutes(r)
	r.Post("/internal/retention/run", a.internalRunRetentionPurge)
	r.Post("/internal/sales-branches-auto/run", a.internalRunSalesBranchesAuto)
	r.Post("/internal/auth-health/run", a.internalRunAuthHealth)
	r.Post("/internal/pharmacy/expiry-run", a.internalPharmacyExpiryRun)

	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		a.registerPharmacyMedicationRoutes(pr)
		a.registerPharmacyStockRoutes(pr)
		a.registerPharmacyDAFRoutes(pr)
		a.registerPrescriptionRoutes(pr)
		pr.Get("/me", a.me)
		pr.Patch("/me", a.updateMe)
		pr.Post("/me/avatar", a.uploadMyAvatar)
		pr.Patch("/me/password", a.changeMePassword)
		pr.Delete("/me", a.deleteMe)
		pr.Get("/me/export", a.exportMe)
		pr.Patch("/me/locale", a.updateMeLocale)
		a.registerProfileRoutes(pr)
		pr.Get("/me/vets", a.listMyVets)
		pr.Get("/me/vets/lookup", a.lookupVets)
		pr.Post("/me/vets/invite", a.inviteVet)
		pr.Post("/me/vets/suggest", a.suggestVet)
		pr.Get("/vet/link-requests", a.listVetLinkRequests)
		pr.Post("/vet/link-requests/{id}/accept", a.acceptVetLinkRequest)
		pr.Post("/vet/link-requests/{id}/reject", a.rejectVetLinkRequest)
		pr.Get("/vet/visits", a.listVetVisits)
		pr.Get("/vet/schedule", a.getVetSchedule)
		pr.Put("/vet/schedule", a.putVetSchedule)
		pr.Get("/vet/vacations", a.listVetVacations)
		pr.Post("/vet/vacations", a.createVetVacation)
		pr.Delete("/vet/vacations/{id}", a.deleteVetVacation)
		pr.Get("/vet/calendar", a.getVetCalendar)
		pr.Get("/practices/{practiceID}/availability", a.getPracticeAvailability)
		pr.Get("/vet/care-reminders", a.listVetOverdueCare)
		pr.Get("/me/discovery", a.getDiscovery)
		pr.Post("/me/discovery/complete", a.completeDiscovery)
		pr.Put("/me/device-tokens", a.putDeviceToken)
		pr.Get("/me/notification-preferences", a.getClientNotificationPrefs)
		pr.Patch("/me/notification-preferences", a.updateClientNotificationPrefs)
		pr.Get("/me/household", a.getHousehold)
		pr.Get("/clients", a.listClients)
		pr.Post("/vet/clients", a.createVetClient)
		pr.Post("/vet/clients/{clientID}/link", a.linkExistingVetClient)
		pr.Get("/vet/colleagues", a.listPracticeColleagues)
		pr.Get("/clients/{clientID}", a.getClient)
		pr.Get("/clients/{clientID}/overview", a.getClientOverview)
		pr.Post("/clients/{clientID}/send-app-link", a.sendClientAppLink)
		pr.Get("/clients/{clientID}/pets", a.listClientPets)
		pr.Get("/clients/{clientID}/shares", a.listClientShares)
		pr.Post("/clients/{clientID}/shares", a.createClientShare)
		pr.Delete("/clients/{clientID}/shares/{granteeID}", a.deleteClientShare)
		pr.Get("/vet/pets", a.listVetPets)
		pr.Get("/care-pro/clients", a.listCareProClients)
		pr.Get("/care-pro/pets", a.listCareProPets)
		pr.Get("/care-pro/visits", a.listCareProVisits)
		pr.Get("/pets", a.listMyPets)
		pr.Post("/pets", a.createPet)
		pr.Post("/pets/batch", a.createPetsBatch)
		pr.Patch("/pets/{petID}/primary-practice", a.setPetPrimaryPractice)
		pr.Get("/pets/{petID}/care-reminders", a.listCareReminders)
		pr.Post("/pets/{petID}/care-reminders", a.createCareReminder)
		pr.Get("/pets/{petID}/horse-contacts", a.listHorseContacts)
		pr.Post("/pets/{petID}/horse-contacts", a.createHorseContact)
		pr.Delete("/horse-contacts/{id}", a.deleteHorseContact)
		pr.Patch("/horse-contacts/{id}", a.updateHorseContact)
		pr.Get("/pets/{petID}/horse-competitions", a.listHorseCompetitions)
		pr.Post("/pets/{petID}/horse-competitions", a.createHorseCompetition)
		pr.Delete("/horse-competitions/{id}", a.deleteHorseCompetition)
		pr.Patch("/horse-competitions/{id}", a.updateHorseCompetition)
		pr.Get("/pets/{petID}/visits", a.listVisits)
		pr.Post("/pets/{petID}/visits", a.createVisit)
		pr.Put("/pets/{petID}", a.updatePet)
		pr.Post("/pets/{petID}/photo", a.uploadPetPhoto)
		pr.Get("/pets/{petID}/health-book", a.getPetHealthBook)
		pr.Post("/pets/{petID}/health-book", a.uploadPetHealthBook)
		pr.Delete("/pets/{petID}/health-book", a.deletePetHealthBook)
		pr.Get("/pets/{petID}/documents", a.listPetDocuments)
		pr.Post("/pets/{petID}/documents", a.uploadPetDocument)
		pr.Delete("/pets/documents/{documentID}", a.deletePetDocument)
		pr.Get("/pets/{petID}/shares", a.listPetShares)
		pr.Post("/pets/{petID}/shares", a.createPetShare)
		pr.Delete("/pets/{petID}/shares/{granteeID}", a.deletePetShare)
		pr.Post("/pets/{petID}/dossier-shares", a.createPetDossierShare)
		pr.Get("/pets/{petID}", a.getPet)
		pr.Get("/pets/{petID}/timeline", a.petTimeline)
		pr.Post("/pets/{petID}/heartrate/sessions", a.startHeartRate)
		pr.Get("/pets/{petID}/heartrate/sessions", a.listHeartRate)
		pr.Post("/pets/{petID}/heartrate/sessions/seen", a.markPetHeartRateSeen)
		pr.Patch("/heartrate/sessions/{sessionID}", a.completeHeartRate)
		pr.Post("/heartrate/sessions/{sessionID}/validate", a.validateHeartRate)
		pr.Post("/heartrate/sessions/{sessionID}/cancel", a.cancelHeartRate)
		pr.Post("/pets/{petID}/weights", a.createWeightReading)
		pr.Get("/pets/{petID}/weights", a.listWeightReadings)
		pr.Post("/care-reminders/{id}/done", a.markCareReminderDone)
		pr.Post("/care-reminders/{id}/postpone", a.postponeCareReminder)
		pr.Patch("/visits/{id}", a.updateVisit)
		pr.Patch("/visits/{visitID}/location", a.updateVisitLocation)
		pr.Get("/visits/{visitID}/preconsult", a.getVisitPreconsult)
		pr.Put("/visits/{visitID}/preconsult", a.putVisitPreconsult)
		pr.Get("/visits/{visitID}/report", a.getVisitReport)
		pr.Get("/visits/{visitID}/reports", a.listVisitReports)
		pr.Put("/visits/{visitID}/report", a.putVisitReport)
		pr.Get("/visits/{visitID}/report/audio", a.getVisitReportAudio)
		pr.Post("/visits/{visitID}/report/finalize", a.finalizeVisitReport)
		pr.Post("/visits/{visitID}/report/improve", a.improveVisitReport)
		pr.Post("/visits/{visitID}/report/transcribe", a.transcribeVisitReport)
		pr.Get("/messaging/threads", a.listThreads)
		pr.Post("/messaging/threads", a.ensureThread)
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
		a.noteAuthSignal(store.AuthAlertLoginFailSpike, authSpikeLoginFail, "login unauthorized (email inconnu ou erreur)")
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if u.PasswordHash == "" {
		writeErr(w, r, http.StatusUnauthorized, "use_google_sign_in", "use_google_sign_in")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		a.noteAuthSignal(store.AuthAlertLoginFailSpike, authSpikeLoginFail, "login unauthorized (mauvais mot de passe)")
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if (kernel.IsPracticeStaff(u.Role) || u.Role == kernel.RoleClient || u.Role == kernel.RoleCarePro) && u.EmailVerifiedAt == nil {
		a.noteAuthSignal(store.AuthAlertUnverifiedSpike, authSpikeUnverified, "login email_not_verified")
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
	// Compte Pro anonymisé (RGPD) : refuser le refresh malgré un JWT encore valide.
	if store.IsTombstoneEmail(u.Email) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "invalid_token")
		return
	}
	profileID := ""
	if active, err := a.store.GetActiveProfile(r.Context(), u.ID); err == nil {
		profileID = active.ID
	}
	pair, err := a.tokens.IssueProfile(u.ID, u.Email, u.Role, u.PracticeID, profileID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// Le refresh compte comme activité (rétention 3 ans).
	a.store.TouchLastLogin(r.Context(), u.ID)
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
	id, ok := a.requirePracticePerm(w, r, "clients.read")
	if !ok {
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
	id, ok := a.requirePracticePerm(w, r, "clients.read")
	if !ok {
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


func (a *API) getClientOverview(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "clients.read")
	if !ok {
		return
	}
	overview, err := a.store.GetClientOverview(r.Context(), id.PracticeID, chi.URLParam(r, "clientID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "client_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, overview)
}

func (a *API) sendClientAppLink(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	ownerID, ok := a.resolvePracticeInviteOwner(w, r, id)
	if !ok {
		return
	}
	if strings.TrimSpace(a.cfg.PetsAppDownloadURL) == "" && strings.TrimSpace(a.cfg.ProPublicSiteURL) == "" {
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "app_download_url_missing")
		return
	}
	invite, err := a.store.EnsureVetAppInviteCode(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "vet_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// Referral landing (QR/email) so new installs can auto-link to this vet.
	downloadURL := a.appInviteWebURL(invite.Code)
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
	vet, err := a.store.GetUserByID(r.Context(), ownerID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	practiceName := invite.PracticeName
	if practiceName == "" {
		if profile, err := a.store.GetPracticeProfile(r.Context(), id.PracticeID, id.UserID); err == nil {
			practiceName = profile.PracticeName
		}
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
	// App invite covers d0_welcome — avoid duplicate welcome drip.
	_ = a.store.EnrollEmailJourney(r.Context(), client.UserID, time.Now().UTC())
	_ = a.store.RecordEmailSend(r.Context(), client.UserID, "d0_welcome", "skipped", map[string]any{"reason": "app_download_invite"})
	httpx.WriteData(w, http.StatusOK, map[string]string{
		"status":  "sent",
		"email":   client.Email,
		"message": t(r, "success.app_link_sent", nil),
	})
}

func (a *API) listClientPets(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "pets.read")
	if !ok {
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
	if id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	pets, err := a.store.ListPetsAccessibleToClient(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, pets)
}

type petReq struct {
	Name             string   `json:"name"`
	Species          string   `json:"species"`
	Breed            string   `json:"breed"`
	BirthDate        *string  `json:"birthDate"`
	WeightKg         *float64 `json:"weightKg"`
	PhotoURL         string   `json:"photoUrl"`
	LitterTag        string   `json:"litterTag"`
	MicrochipNumber  *string  `json:"microchipNumber"`
	HealthBookNumber *string  `json:"healthBookNumber"`
	Plan             string   `json:"plan"`
	BillingMode      string   `json:"billingMode"`
	SuccessURL       string   `json:"successUrl"`
	CancelURL        string   `json:"cancelUrl"`
	SkipCheckout     bool     `json:"skipCheckout"`
}

type petsBatchReq struct {
	Pets []petReq `json:"pets"`
}

func (a *API) createPet(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	// Practice is optional at create — client may link a vet afterwards.
	practiceID := strings.TrimSpace(id.PracticeID)
	if practiceID == "" {
		resolved, rerr := a.store.ResolveClientPracticeID(r.Context(), id.UserID)
		if rerr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		practiceID = strings.TrimSpace(resolved)
	}
	var req petReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	name := strings.TrimSpace(req.Name)
	species := strings.TrimSpace(req.Species)
	if name == "" || species == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "name_species_required")
		return
	}
	planCode, err := billing.ParsePlanCode(defaultStr(req.Plan, string(billing.PlanTriennial)))
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_plan")
		return
	}
	mode, err := billing.ParseBillingMode(defaultStr(req.BillingMode, string(billing.ModeSubscription)))
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_billing_mode")
		return
	}
	if !billing.SupportsBillingMode(planCode, mode) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_billing_mode")
		return
	}
	plan, planErr := billing.GetPlan(planCode)
	if planErr != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_plan")
		return
	}

	p := store.Pet{
		Name: name, Species: species, Breed: strings.TrimSpace(req.Breed), WeightKg: req.WeightKg,
		PhotoURL: req.PhotoURL, LitterTag: strings.TrimSpace(req.LitterTag),
		OwnerUserID: id.UserID, PracticeID: practiceID, PaymentStatus: "pending_payment",
	}
	if req.MicrochipNumber != nil {
		p.MicrochipNumber = clipPetIDField(*req.MicrochipNumber, maxMicrochipLen)
	}
	if req.HealthBookNumber != nil {
		p.HealthBookNumber = clipPetIDField(*req.HealthBookNumber, maxHealthBookNumberLen)
	}
	if req.BirthDate != nil {
		raw := strings.TrimSpace(*req.BirthDate)
		if raw != "" {
			t, perr := time.Parse("2006-01-02", raw)
			if perr != nil {
				writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_birth_date")
				return
			}
			p.BirthDate = &t
		}
	}
	created, err := a.store.CreatePetWithPendingEntitlement(r.Context(), p, string(planCode), string(mode), plan.AmountCents)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "validation")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// Care reminders are created manually by the client — no default seed.
	a.startPetBillingCheckout(w, r, created, id, createPetBilling{
		SuccessURL: req.SuccessURL, CancelURL: req.CancelURL,
	}, req.SkipCheckout, planCode, mode)
}

func (a *API) createPetsBatch(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	// Practice is optional at create — client may link a vet afterwards.
	practiceID := strings.TrimSpace(id.PracticeID)
	if practiceID == "" {
		resolved, rerr := a.store.ResolveClientPracticeID(r.Context(), id.UserID)
		if rerr != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		practiceID = strings.TrimSpace(resolved)
	}
	var body petsBatchReq
	if err := httpx.DecodeJSON(r, &body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if len(body.Pets) == 0 || len(body.Pets) > 50 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "batch_size")
		return
	}

	// Validate all rows before any insert (all-or-nothing TX).
	items := make([]store.PetWithPendingEntitlement, 0, len(body.Pets))
	for _, req := range body.Pets {
		name := strings.TrimSpace(req.Name)
		species := strings.TrimSpace(req.Species)
		if name == "" || species == "" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "name_species_required")
			return
		}
		planCode, err := billing.ParsePlanCode(defaultStr(req.Plan, string(billing.PlanTriennial)))
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_plan")
			return
		}
		mode, err := billing.ParseBillingMode(defaultStr(req.BillingMode, string(billing.ModeSubscription)))
		if err != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_billing_mode")
			return
		}
		if !billing.SupportsBillingMode(planCode, mode) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_billing_mode")
			return
		}
		plan, planErr := billing.GetPlan(planCode)
		if planErr != nil {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_plan")
			return
		}
		p := store.Pet{
			Name: name, Species: species, Breed: strings.TrimSpace(req.Breed),
			LitterTag:   strings.TrimSpace(req.LitterTag),
			OwnerUserID: id.UserID, PracticeID: practiceID, PaymentStatus: "pending_payment",
		}
		if req.MicrochipNumber != nil {
			p.MicrochipNumber = clipPetIDField(*req.MicrochipNumber, maxMicrochipLen)
		}
		if req.HealthBookNumber != nil {
			p.HealthBookNumber = clipPetIDField(*req.HealthBookNumber, maxHealthBookNumberLen)
		}
		if req.BirthDate != nil {
			raw := strings.TrimSpace(*req.BirthDate)
			if raw != "" {
				t, perr := time.Parse("2006-01-02", raw)
				if perr != nil {
					writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_birth_date")
					return
				}
				p.BirthDate = &t
			}
		}
		items = append(items, store.PetWithPendingEntitlement{
			Pet: p, PlanCode: string(planCode), BillingMode: string(mode), AmountCents: plan.AmountCents,
		})
	}

	created, err := a.store.CreatePetsBatchWithPendingEntitlements(r.Context(), items)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "validation")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// Care reminders are created manually by the client — no default seed.
	httpx.WriteData(w, http.StatusCreated, map[string]any{"pets": created, "count": len(created)})
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
	existing, err := a.store.GetPet(r.Context(), chi.URLParam(r, "petID"))
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	if existing.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	p := store.Pet{
		ID: existing.ID, Name: req.Name, Species: req.Species, Breed: req.Breed,
		WeightKg: req.WeightKg, PhotoURL: req.PhotoURL, OwnerUserID: id.UserID,
		LitterTag: existing.LitterTag,
		MicrochipNumber: existing.MicrochipNumber, HealthBookNumber: existing.HealthBookNumber,
	}
	if tag := strings.TrimSpace(req.LitterTag); tag != "" {
		p.LitterTag = tag
	}
	if req.MicrochipNumber != nil {
		p.MicrochipNumber = clipPetIDField(*req.MicrochipNumber, maxMicrochipLen)
	}
	if req.HealthBookNumber != nil {
		p.HealthBookNumber = clipPetIDField(*req.HealthBookNumber, maxHealthBookNumberLen)
	}
	// Champs omis : conserver les valeurs existantes (édition partielle mobile).
	if strings.TrimSpace(req.PhotoURL) == "" {
		p.PhotoURL = existing.PhotoURL
	}
	if req.WeightKg == nil {
		p.WeightKg = existing.WeightKg
	}
	if req.BirthDate != nil {
		if t, err := time.Parse("2006-01-02", *req.BirthDate); err == nil {
			p.BirthDate = &t
		}
	} else {
		p.BirthDate = existing.BirthDate
	}
	if err := a.store.UpdatePet(r.Context(), p); err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *API) getPet(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	pet, ok := a.requirePetAccess(w, r, chi.URLParam(r, "petID"), id, store.PermRead)
	if !ok {
		return
	}
	if id.Role == kernel.RoleClient {
		if pet.OwnerUserID == id.UserID {
			pet.Permission = string(store.PermFull)
		} else if perm, e := a.store.EffectivePetPermission(r.Context(), pet, id.UserID); e == nil {
			pet.Permission = string(perm)
		}
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
	pet, ok := a.requirePetAccess(w, r, petID, id, store.PermRead)
	if !ok {
		return
	}
	vetView := id.Role == kernel.RoleCarePro || a.allowPracticePerm(r, id, "pets.read")
	ident := store.IdentityOf(id.UserID, id.Role, id.PracticeID)
	canNotes, err := a.store.CanAccessPet(r.Context(), ident, pet, store.PermWriteNotes)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	canFull, err := a.store.CanAccessPet(r.Context(), ident, pet, store.PermFull)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	// Messages: full (or owner/practice vet — CanAccessPet grants any level) only; write_notes alone is not enough.
	includeMessages := canFull
	redactVisitNotes := !canNotes
	items, err := a.store.PetTimelineFiltered(r.Context(), petID, vetView, includeMessages, redactVisitNotes)
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
	if !kernel.SupportsHeartRateControl(pet.Species) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "heartrate_not_supported")
		return
	}
	if !a.requirePremiumAccess(w, r, pet.ID) {
		return
	}
	var req startHRReq
	_ = httpx.DecodeJSON(r, &req)
	var allowed []int
	if strings.TrimSpace(pet.PracticeID) != "" {
		allowed, err = a.store.GetPracticeHeartRateDurations(r.Context(), pet.PracticeID)
		if err != nil {
			allowed = nil
		}
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
	alert, err := a.heartRateDeltaAlert(r.Context(), sess.PetID, bpm)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	sess, err = a.store.CompleteHeartRateSession(r.Context(), chi.URLParam(r, "sessionID"), id.UserID, req.TapCount, bpm, alert)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "session_not_found")
		return
	}
	httpx.WriteData(w, http.StatusOK, sess)
}

type validateHRReq struct {
	Comment *string `json:"comment"`
}

func (a *API) validateHeartRate(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var req validateHRReq
	if err := httpx.DecodeJSON(r, &req); err != nil && !errors.Is(err, io.EOF) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	sess, err := a.store.ValidateHeartRateSession(r.Context(), chi.URLParam(r, "sessionID"), id.UserID, req.Comment)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "session_not_found")
		return
	}
	practiceID := strings.TrimSpace(sess.PracticeID)
	if practiceID == "" {
		practiceID = strings.TrimSpace(id.PracticeID)
	}
	if practiceID == "" {
		if resolved, rerr := a.store.ResolveClientPracticeID(r.Context(), id.UserID); rerr == nil {
			practiceID = resolved
		}
	}
	vetID, _ := a.store.GetVetForClient(r.Context(), id.UserID, practiceID)
	if vetID != "" {
		prefs, _ := a.store.EmailPrefs(r.Context(), vetID)
		if prefs.OnHeartRate {
			vet, _ := a.store.GetUserByID(r.Context(), vetID)
			locale := vet.PreferredLocale
			if locale == "" {
				locale = localeOf(r)
			}
			bpm := 0
			if sess.BPM != nil {
				bpm = *sess.BPM
			}
			if sess.IsAlert {
				_ = a.notifier.SendHeartrateThresholdAlert(vet.Email, locale, bpm)
				_ = a.store.LogNotification(r.Context(), vetID, "heartrate_threshold_alert", map[string]any{"sessionId": sess.ID, "bpm": bpm})
			} else {
				_ = a.notifier.SendHeartrateValidated(vet.Email, locale, bpm)
				_ = a.store.LogNotification(r.Context(), vetID, "heartrate_validated", map[string]any{"sessionId": sess.ID, "bpm": bpm})
			}
		}
	}
	httpx.WriteData(w, http.StatusOK, sess)
}


// heartRateDeltaAlert reports whether bpm rose by at least the species delta
// versus the last validated reading. Unsupported species → false, nil.
func (a *API) heartRateDeltaAlert(ctx context.Context, petID string, bpm int) (bool, error) {
	pet, err := a.store.GetPet(ctx, petID)
	if err != nil {
		return false, err
	}
	delta, ok, err := a.store.GetHeartRateAlertDelta(ctx, pet.Species)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	prev, err := a.store.LastValidatedBPM(ctx, petID)
	if err != nil {
		return false, err
	}
	return kernel.IsHeartRateDeltaAlert(bpm, prev, delta), nil
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
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	if _, ok := a.requirePetAccess(w, r, petID, id, store.PermRead); !ok {
		return
	}
	vetView := id.Role == kernel.RoleCarePro || a.allowPracticePerm(r, id, "pets.read")
	sessions, err := a.store.ListHeartRateSessions(r.Context(), petID, vetView)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, sessions)
}

func (a *API) markPetHeartRateSeen(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "heartrate.validate")
	if !ok {
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	if pet.PracticeID != id.PracticeID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
		return
	}
	n, err := a.store.MarkPetHeartRateSessionsSeen(r.Context(), petID, id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"marked": n})
}

func (a *API) listThreads(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	switch {
	case a.allowPracticePerm(r, id, "messaging"):
		threads, err := a.store.ListThreadSummariesForPractice(r.Context(), id.PracticeID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		httpx.WriteData(w, http.StatusOK, threads)
		return
	case id.Role == kernel.RoleCarePro:
		threads, err := a.store.ListThreadSummariesForVet(r.Context(), id.UserID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		out := make([]store.ThreadSummary, 0, len(threads))
		for _, t := range threads {
			th := store.Thread{
				ID: t.ID, PracticeID: t.PracticeID, ClientUserID: t.ClientUserID,
				VetUserID: t.VetUserID, PetID: t.PetID,
			}
			if a.canAccessThread(r, id, th) {
				out = append(out, t)
			}
		}
		httpx.WriteData(w, http.StatusOK, out)
		return
	case id.Role == kernel.RoleClient:
		threads, err := a.store.ListThreadSummariesForClient(r.Context(), id.UserID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		if threads == nil {
			threads = []store.ThreadSummary{}
		}
		httpx.WriteData(w, http.StatusOK, threads)
		return
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
}

func (a *API) canAccessThread(r *http.Request, id authx.Identity, thread store.Thread) bool {
	if id.UserID == thread.ClientUserID {
		return true
	}
	if kernel.IsPracticeStaff(id.Role) && id.PracticeID != "" && id.PracticeID == thread.PracticeID {
		return a.allowPracticePerm(r, id, "messaging")
	}
	if id.UserID != thread.VetUserID {
		return false
	}
	// care_pro person-scoped threads: ACL must still be active (revoke cuts access).
	if id.Role == kernel.RoleCarePro && thread.PracticeID == "" {
		return a.careProMayAccessThread(r, id, thread)
	}
	return true
}

// careProMayAccessThread re-checks pet_access / client_access for a care_pro thread.
func (a *API) careProMayAccessThread(r *http.Request, id authx.Identity, thread store.Thread) bool {
	if thread.PetID != "" {
		pet, err := a.store.GetPet(r.Context(), thread.PetID)
		if err != nil {
			return false
		}
		ok, err := a.store.CanAccessPet(r.Context(), store.IdentityOf(id.UserID, id.Role, id.PracticeID), pet, store.PermRead)
		return err == nil && ok
	}
	ok, err := a.store.CareProMayMessageClient(r.Context(), id.UserID, thread.ClientUserID)
	return err == nil && ok
}

func (a *API) listMessages(w http.ResponseWriter, r *http.Request) {
	thread, err := a.store.GetThreadByID(r.Context(), chi.URLParam(r, "threadID"))
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "thread_not_found")
		return
	}
	id, _ := authx.FromContext(r.Context())
	if !a.canAccessThread(r, id, thread) {
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
	if !a.canAccessThread(r, id, thread) {
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
	var markErr error
	if a.allowPracticePerm(r, id, "messaging") && id.PracticeID != "" {
		markErr = a.store.MarkAllUnreadForPractice(r.Context(), id.PracticeID)
	} else {
		markErr = a.store.MarkAllUnreadForUser(r.Context(), id.UserID)
	}
	if markErr != nil {
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
	if !a.canAccessThread(r, id, thread) {
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
		prefs, _ := a.store.EmailPrefs(r.Context(), thread.VetUserID)
		if prefs.OnMessage {
			vet, _ := a.store.GetUserByID(r.Context(), thread.VetUserID)
			locale := vet.PreferredLocale
			if locale == "" {
				locale = localeOf(r)
			}
			_ = a.notifier.SendNewMessage(vet.Email, locale, req.Body)
		}
		a.pushNewMessage(thread.VetUserID, thread.ID, req.Body)
	}
	if kernel.IsPracticeStaff(id.Role) || id.Role == kernel.RoleCarePro {
		a.pushNewMessage(thread.ClientUserID, thread.ID, req.Body)
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
	if !a.canAccessThread(r, id, thread) {
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
	preview := body
	if preview == "" {
		if kind == "video" {
			preview = "[video]"
		} else {
			preview = "[image]"
		}
	}
	if id.Role == kernel.RoleClient {
		prefs, _ := a.store.EmailPrefs(r.Context(), thread.VetUserID)
		if prefs.OnMessage {
			vet, _ := a.store.GetUserByID(r.Context(), thread.VetUserID)
			locale := vet.PreferredLocale
			if locale == "" {
				locale = localeOf(r)
			}
			_ = a.notifier.SendNewMessage(vet.Email, locale, preview)
		}
		a.pushNewMessage(thread.VetUserID, thread.ID, preview)
	}
	if kernel.IsPracticeStaff(id.Role) || id.Role == kernel.RoleCarePro {
		a.pushNewMessage(thread.ClientUserID, thread.ID, preview)
	}
	httpx.WriteData(w, http.StatusCreated, msg)
}

type availReq struct {
	Status    kernel.AvailabilityStatus `json:"status"`
	AutoReply string                    `json:"autoReply"`
}

func (a *API) setAvailability(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "messaging")
	if !ok {
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
	id, ok := a.requirePracticePerm(w, r, "clients.read")
	if !ok {
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
	id, ok := a.requirePracticePerm(w, r, "messaging")
	if !ok {
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
	// Consent — acceptation CGU/privacy (checkbox obligatoire côté front, persistée en DB).
	Consent bool `json:"consent"`
	// InviteCode — code parrain commercial (practice.app_invite_codes).
	InviteCode string `json:"inviteCode,omitempty"`
	// AssignedCommercialID — ignored on /auth/register (assignment only via inviteCode).
	AssignedCommercialID string `json:"assignedCommercialId,omitempty"`
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
	if !req.Consent {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "consent_required")
		return
	}
	assignedCommercialID := ""
	if code := store.NormalizeInviteCode(req.InviteCode); code != "" {
		inv, err := a.store.GetAppInviteByCode(r.Context(), code)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_invite_code")
				return
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		switch inv.Role {
		case string(kernel.RoleCommercial), string(kernel.RoleCommercialManager):
			assignedCommercialID = inv.UserID
		default:
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_invite_code")
			return
		}
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
		TermsAccepted: req.Consent, AssignedCommercialID: assignedCommercialID,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	confirmURL := fmt.Sprintf("%s/confirm-email?token=%s", strings.TrimRight(a.cfg.ProPublicSiteURL, "/"), result.Token)
	if err := a.notifier.SendConfirmRegistration(req.Email, locale, req.FullName, confirmURL); err != nil {
		a.reportConfirmEmailFailure(r.Context(), req.Email, err)
		a.noteAuthSignal(store.AuthAlertRegisterFailSpike, authSpikeRegisterFail, "register vet: SMTP confirm fail")
	}
	out := map[string]any{
		"message": t(r, "success.confirm_email_sent", nil),
	}
	// Dev/demo only: never expose the confirmation token outside seeded environments.
	if a.cfg.DevSeedEnabled {
		out["confirmPath"] = "/confirm-email?token=" + result.Token
	}
	httpx.WriteData(w, http.StatusCreated, out)
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
	profileID := ""
	if active, err := a.store.GetActiveProfile(r.Context(), u.ID); err == nil {
		profileID = active.ID
	}
	pair, err := a.tokens.IssueProfile(u.ID, u.Email, u.Role, u.PracticeID, profileID)
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

type resendConfirmationReq struct {
	Email string `json:"email"`
}

func (a *API) resendConfirmation(w http.ResponseWriter, r *http.Request) {
	var req resendConfirmationReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}

	result, err := a.store.RequestEmailConfirmation(r.Context(), req.Email)
	out := map[string]any{
		"message": t(r, "success.confirm_email_sent", nil),
	}
	if err == nil {
		confirmURL := fmt.Sprintf("%s/confirm-email?token=%s",
			strings.TrimRight(a.cfg.ProPublicSiteURL, "/"), result.Token)
		if sendErr := a.notifier.SendConfirmRegistration(result.Email, result.Locale, result.FullName, confirmURL); sendErr != nil {
			a.reportConfirmEmailFailure(r.Context(), result.Email, sendErr)
		}
		if a.cfg.DevSeedEnabled {
			out["confirmPath"] = "/confirm-email?token=" + result.Token
		}
	}
	// Always 200 — do not reveal whether the email exists / is already verified.
	httpx.WriteData(w, http.StatusOK, out)
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
		resetURL := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimRight(a.cfg.ProPublicSiteURL, "/"), result.Token)
		_ = a.notifier.SendPasswordReset(result.Email, result.Locale, result.FullName, resetURL)
		// Dev/demo only: never expose the reset token outside seeded environments.
		if a.cfg.DevSeedEnabled {
			out["resetPath"] = "/reset-password?token=" + result.Token
		}
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
	id, ok := a.requirePracticePerm(w, r, "practice.settings")
	if !ok {
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
	id, ok := a.requirePracticePerm(w, r, "practice.settings")
	if !ok {
		return
	}
	defer r.Body.Close()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	var req store.PracticeProfile
	if err := json.Unmarshal(raw, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	var durationsProbe struct {
		HeartRateDurationsSec *[]int `json:"heartrateDurationsSec"`
	}
	if err := json.Unmarshal(raw, &durationsProbe); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if req.PracticeName == "" || req.Phone == "" || req.ContactEmail == "" ||
		req.AddressLine1 == "" || req.City == "" || req.PostalCode == "" || req.VetFullName == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "profile_fields_required")
		return
	}
	req.PayoutIBAN = normalizeIBAN(req.PayoutIBAN)
	if req.PayoutIBAN != "" && !validIBAN(req.PayoutIBAN) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_iban")
		return
	}
	req.PayoutBIC = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(req.PayoutBIC), " ", ""))
	if !validBIC(req.PayoutBIC) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_bic")
		return
	}
	req.PayoutAccountHolder = strings.TrimSpace(req.PayoutAccountHolder)
	if len(req.PayoutAccountHolder) > 120 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_account_holder")
		return
	}
	req.CompanyLegalName = strings.TrimSpace(req.CompanyLegalName)
	req.VATNumber = strings.TrimSpace(req.VATNumber)
	req.CompanyNumber = strings.TrimSpace(req.CompanyNumber)
	req.LegalForm = strings.TrimSpace(req.LegalForm)
	req.BillingAddressLine1 = strings.TrimSpace(req.BillingAddressLine1)
	req.BillingAddressLine2 = strings.TrimSpace(req.BillingAddressLine2)
	req.BillingPostalCode = strings.TrimSpace(req.BillingPostalCode)
	req.BillingCity = strings.TrimSpace(req.BillingCity)
	// JSON omits default bool to false; treat empty billing address as same-as-practice.
	if !req.BillingSameAsPractice && req.BillingAddressLine1 == "" {
		req.BillingSameAsPractice = true
	}
	var durationsUpdate *[]int
	if durationsProbe.HeartRateDurationsSec != nil {
		normalized := kernel.NormalizeHeartRateDurations(*durationsProbe.HeartRateDurationsSec)
		durationsUpdate = &normalized
	}
	markComplete := r.URL.Query().Get("complete") == "true"
	if err := a.store.UpdatePracticeProfile(r.Context(), id.PracticeID, id.UserID, req, markComplete, durationsUpdate); err != nil {
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
