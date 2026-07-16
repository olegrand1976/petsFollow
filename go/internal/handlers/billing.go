package handlers

import (
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/billing"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) registerBillingRoutes(r chi.Router) {
	r.Get("/billing/plans", a.listBillingPlans)
	r.Get("/billing/addons", a.listBillingAddons)
	r.Post("/billing/webhooks/stripe", a.stripeWebhook)
	if a.cfg.BillingMockEnabled {
		r.Get("/billing/dev/mock-complete", a.billingMockComplete)
	}
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Post("/pets/{petID}/billing/checkout", a.resumePetCheckout)
		pr.Post("/pets/{petID}/billing/portal", a.petBillingPortal)
		pr.Get("/pets/{petID}/entitlement", a.getPetEntitlement)
		pr.Get("/billing/my-addons", a.listMyAddons)
		pr.Post("/billing/addons/checkout", a.startAddonCheckout)
	})
}

func (a *API) listBillingAddons(w http.ResponseWriter, r *http.Request) {
	locale := localeOf(r)
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"addons": billing.AllAddonsForLocale(locale),
	})
}

func (a *API) listMyAddons(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	addons, err := a.store.ListOwnerAddons(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, addons)
}

type addonCheckoutReq struct {
	Addon      string `json:"addon"`
	SuccessURL string `json:"successUrl"`
	CancelURL  string `json:"cancelUrl"`
}

func (a *API) startAddonCheckout(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	var body addonCheckoutReq
	if err := httpx.DecodeJSON(r, &body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	code, err := billing.ParseAddonCode(body.Addon)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_addon")
		return
	}
	addon, _ := billing.GetAddon(code)
	ent, err := a.store.CreateAddonEntitlement(r.Context(), id.UserID, string(code), addon.AmountCents)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	u, _ := a.store.GetUserByID(r.Context(), id.UserID)
	sess, err := a.billing.StartAddonCheckout(r.Context(), billing.StartAddonCheckoutInput{
		AddonID:     ent.ID,
		OwnerUserID: id.UserID,
		OwnerEmail:  u.Email,
		AddonCode:   code,
		SuccessURL:  body.SuccessURL,
		CancelURL:   body.CancelURL,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, map[string]any{
		"addonId":     ent.ID,
		"checkoutUrl": sess.URL,
		"sessionId":   sess.ID,
	})
}

func (a *API) listBillingPlans(w http.ResponseWriter, r *http.Request) {
	locale := localeOf(r)
	plans := a.billing.ListPlansForLocale(locale)
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"plans":       billing.AllPlansForLocale(locale),
		"offers":      plans,
		"recommended": billing.PlanTriennial,
		"defaultMode": billing.ModeSubscription,
	})
}

type createPetBilling struct {
	Plan        string `json:"plan"`
	BillingMode string `json:"billingMode"`
	SuccessURL  string `json:"successUrl"`
	CancelURL   string `json:"cancelUrl"`
}

func (a *API) startPetBillingCheckout(w http.ResponseWriter, r *http.Request, pet store.Pet, owner authx.Identity, b createPetBilling) {
	planCode, err := billing.ParsePlanCode(defaultStr(b.Plan, string(billing.PlanTriennial)))
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_plan")
		return
	}
	mode, err := billing.ParseBillingMode(defaultStr(b.BillingMode, string(billing.ModeSubscription)))
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_billing_mode")
		return
	}
	plan, _ := billing.GetPlan(planCode)
	_, err = a.store.CreateEntitlement(r.Context(), pet.ID, owner.UserID, string(planCode), string(mode), plan.AmountCents)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	u, _ := a.store.GetUserByID(r.Context(), owner.UserID)
	sess, err := a.billing.StartCheckout(r.Context(), billing.StartCheckoutInput{
		PetID:       pet.ID,
		OwnerUserID: owner.UserID,
		OwnerEmail:  u.Email,
		PlanCode:    planCode,
		BillingMode: mode,
		SuccessURL:  b.SuccessURL,
		CancelURL:   b.CancelURL,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, map[string]any{
		"pet":         pet,
		"checkoutUrl": sess.URL,
		"sessionId":   sess.ID,
	})
}

func (a *API) resumePetCheckout(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil || pet.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	ent, err := a.store.GetEntitlementByPetID(r.Context(), petID)
	if err != nil || ent.Status != string(billing.StatusPending) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "no_pending_payment")
		return
	}
	var body createPetBilling
	_ = httpx.DecodeJSON(r, &body)
	u, _ := a.store.GetUserByID(r.Context(), id.UserID)
	sess, err := a.billing.StartCheckout(r.Context(), billing.StartCheckoutInput{
		PetID:       pet.ID,
		OwnerUserID: id.UserID,
		OwnerEmail:  u.Email,
		PlanCode:    billing.PlanCode(ent.PlanCode),
		BillingMode: billing.BillingMode(ent.BillingMode),
		SuccessURL:  body.SuccessURL,
		CancelURL:   body.CancelURL,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"checkoutUrl": sess.URL, "sessionId": sess.ID})
}

func (a *API) petBillingPortal(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleClient {
		writeErr(w, r, http.StatusForbidden, "forbidden", "client_only")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil || pet.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	ent, err := a.store.GetEntitlementByPetID(r.Context(), petID)
	if err != nil || ent.BillingMode != string(billing.ModeSubscription) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "not_a_subscription")
		return
	}
	var body struct {
		ReturnURL string `json:"returnUrl"`
	}
	_ = httpx.DecodeJSON(r, &body)
	portal, err := a.billing.CreatePortalSession(r.Context(), id.UserID, body.ReturnURL)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, portal)
}

func (a *API) getPetEntitlement(w http.ResponseWriter, r *http.Request) {
	petID := chi.URLParam(r, "petID")
	id, _ := authx.FromContext(r.Context())
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	if id.Role == kernel.RoleClient && pet.OwnerUserID != id.UserID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
		return
	}
	if id.Role == kernel.RoleVet && pet.PracticeID != id.PracticeID {
		writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
		return
	}
	ent, err := a.store.GetEntitlementByPetID(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "no_entitlement")
		return
	}
	httpx.WriteData(w, http.StatusOK, ent)
}

func (a *API) stripeWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_body")
		return
	}
	if err := a.billing.HandleWebhook(r.Context(), payload, r.Header.Get("Stripe-Signature")); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) billingMockComplete(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if addonID := q.Get("addon_id"); addonID != "" {
		ownerUserID := q.Get("owner_user_id")
		addonCode := q.Get("addon_code")
		sessionID := defaultStr(q.Get("session_id"), "cs_mock_addon_dev")
		if ownerUserID == "" || addonCode == "" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
			return
		}
		if err := a.billing.MockCompleteAddonCheckout(r.Context(), addonID, ownerUserID, addonCode, sessionID); err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "completed", "addonId": addonID})
		return
	}
	petID := q.Get("pet_id")
	ownerUserID := q.Get("owner_user_id")
	planCode := defaultStr(q.Get("plan_code"), string(billing.PlanTriennial))
	billingMode := defaultStr(q.Get("billing_mode"), string(billing.ModeSubscription))
	sessionID := defaultStr(q.Get("session_id"), "cs_mock_dev")
	if petID == "" || ownerUserID == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "pet_owner_required")
		return
	}
	if err := a.billing.MockCompleteCheckout(r.Context(), petID, ownerUserID, planCode, billingMode, sessionID); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "completed", "petId": petID})
}

func (a *API) requirePremiumAccess(w http.ResponseWriter, r *http.Request, petID string) bool {
	ok, err := a.billing.PetHasPremiumAccess(r.Context(), petID)
	if err != nil || !ok {
		writeErr(w, r, http.StatusPaymentRequired, "payment_required", "payment_required")
		return false
	}
	return true
}

func defaultStr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func parseAdminRange(r *http.Request) (time.Time, time.Time) {
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now().Add(24 * time.Hour)
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t
		}
	}
	return from, to
}
