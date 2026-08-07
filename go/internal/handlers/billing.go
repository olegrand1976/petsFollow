package handlers

import (
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
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
		r.Get("/billing/dev/mock-portal", a.billingMockPortal)
	}
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Use(a.requireTermsAcceptedMiddleware)
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

func (a *API) startAddonCheckout(w http.ResponseWriter, r *http.Request) {
	writeErr(w, r, http.StatusGone, "addon_deprecated", "addon_deprecated")
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
	SuccessURL string `json:"successUrl"`
	CancelURL  string `json:"cancelUrl"`
}

// startPetBillingCheckout assumes pet + pending entitlement are already committed.
// Checkout failures still return 201 with the pet so the client never sees a false "not saved".
func (a *API) startPetBillingCheckout(w http.ResponseWriter, r *http.Request, pet store.Pet, owner authx.Identity, b createPetBilling, skipCheckout bool, planCode billing.PlanCode, mode billing.BillingMode) {
	if pet.Entitlement == nil {
		if ent, e := a.store.GetEntitlementByPetID(r.Context(), pet.ID); e == nil {
			pet.Entitlement = &ent
		}
	}
	if skipCheckout {
		httpx.WriteData(w, http.StatusCreated, map[string]any{"pet": pet})
		return
	}
	u, err := a.store.GetUserByID(r.Context(), owner.UserID)
	if err != nil {
		log.Printf("create pet %s: load owner for checkout: %v", pet.ID, err)
		httpx.WriteData(w, http.StatusCreated, map[string]any{"pet": pet})
		return
	}
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
		log.Printf("create pet %s: checkout after commit: %v", pet.ID, err)
		httpx.WriteData(w, http.StatusCreated, map[string]any{"pet": pet})
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
	// Corps optionnel (successUrl/cancelUrl) : seul un JSON malformé est rejeté.
	if derr := httpx.DecodeJSON(r, &body); derr != nil && derr != io.EOF {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	u, err := a.store.GetUserByID(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
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
	// Corps optionnel (returnUrl) : seul un JSON malformé est rejeté.
	if derr := httpx.DecodeJSON(r, &body); derr != nil && derr != io.EOF {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	portal, err := a.billing.CreatePortalSession(r.Context(), id.UserID, body.ReturnURL)
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, portal)
}

func (a *API) getPetEntitlement(w http.ResponseWriter, r *http.Request) {
	petID := chi.URLParam(r, "petID")
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return
	}
	// Deny-by-default : seuls le propriétaire et un véto du cabinet lisent l'entitlement.
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
	// Reached by a browser redirect, so no bearer token: the signature minted by
	// CreateCheckoutSession is what prevents a stranger granting entitlements.
	if !billing.VerifyMockURL(a.cfg.JWTSigningKey, q) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "invalid_signature")
		return
	}
	successURL := q.Get("success_url")
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
		payload := map[string]string{"status": "completed", "addonId": addonID}
		if writeMockCompleteResult(w, r, successURL) {
			return
		}
		httpx.WriteData(w, http.StatusOK, payload)
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
	payload := map[string]string{"status": "completed", "petId": petID}
	if writeMockCompleteResult(w, r, successURL) {
		return
	}
	httpx.WriteData(w, http.StatusOK, payload)
}

// writeMockCompleteResult serves an HTML success page (browser / Flutter external
// checkout) when success_url is set and the client prefers HTML. Returns true if
// a response was written. Smoke/e2e without success_url keep JSON.
func writeMockCompleteResult(w http.ResponseWriter, r *http.Request, successURL string) bool {
	if successURL == "" || !prefersMockCompleteHTML(r) {
		return false
	}
	if !allowedMockReturnURL(successURL) {
		return false
	}
	safeURL := html.EscapeString(successURL)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta http-equiv="refresh" content="0;url=%s">
<title>petsFollow — payment completed</title>
<style>
body{font-family:system-ui,sans-serif;max-width:28rem;margin:2rem auto;padding:0 1rem;line-height:1.45}
a{color:#0b6e4f;font-weight:600}
</style>
</head><body>
<h1>Payment completed</h1>
<p>Your pet subscription is active.</p>
<p><a href="%s">Return to petsFollow</a></p>
</body></html>`, safeURL, safeURL)
	return true
}

func prefersMockCompleteHTML(r *http.Request) bool {
	if r.URL.Query().Get("format") == "json" {
		return false
	}
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/html") {
		return true
	}
	// Some WebViews / in-app browsers send an empty Accept.
	return accept == ""
}

// allowedMockReturnURL restricts mock success redirects to the app deep-link
// scheme only (blocks open redirects / javascript: / https phishing).
func allowedMockReturnURL(u string) bool {
	return strings.HasPrefix(u, "petsfollow://") && !strings.ContainsAny(u, "<>\"'")
}

// billingMockPortal is the Stripe Customer Portal stand-in when BILLING_MOCK_ENABLED.
func (a *API) billingMockPortal(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if !billing.VerifyMockURL(a.cfg.JWTSigningKey, q) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "invalid_signature")
		return
	}
	// returnURL n'est pas refiltré : il est couvert par la signature ci-dessus,
	// donc il vaut ce que CreatePortalSession a émis et rien d'autre.
	customer := q.Get("customer")
	returnURL := q.Get("return")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>petsFollow — mock billing portal</title>
<style>
body{font-family:system-ui,sans-serif;max-width:28rem;margin:2rem auto;padding:0 1rem;line-height:1.45}
a{color:#0b6e4f}
.meta{color:#666;font-size:.9rem;word-break:break-all}
</style></head><body>
<h1>Mock Stripe portal</h1>
<p>Dev / seed billing — no live Stripe Customer Portal.</p>
<p class="meta">customer: %s</p>
`, html.EscapeString(customer))
	if returnURL != "" {
		_, _ = fmt.Fprintf(w, `<p><a href="%s">Return to app</a></p>`, html.EscapeString(returnURL))
	}
	_, _ = io.WriteString(w, `</body></html>`)
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
