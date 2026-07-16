package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

type Service struct {
	store   *store.Store
	gateway Gateway
	cfg     config.Config
}

func NewService(st *store.Store, cfg config.Config) *Service {
	var gw Gateway
	if LiveEnabled(cfg.StripeSecretKey, cfg.BillingMockEnabled) {
		gw = NewLiveGateway(cfg.StripeSecretKey, cfg.StripeWebhookSecret)
	} else {
		gw = NewMockGateway(cfg.StripeWebhookSecret, cfg.APIPublicURL)
	}
	return &Service{store: st, gateway: gw, cfg: cfg}
}

func NewServiceWithGateway(st *store.Store, cfg config.Config, gw Gateway) *Service {
	return &Service{store: st, gateway: gw, cfg: cfg}
}

func (s *Service) ListPlans() []PlanOffer {
	return s.ListPlansForLocale("fr")
}

func (s *Service) ListPlansForLocale(locale string) []PlanOffer {
	var offers []PlanOffer
	for _, plan := range AllPlansForLocale(locale) {
		for _, mode := range []BillingMode{ModeOneTime, ModeSubscription} {
			offers = append(offers, PlanOffer{
				Plan:        plan,
				BillingMode: mode,
				Summary:     PlanSummaryForLocale(plan, mode, locale),
			})
		}
	}
	return offers
}

type StartCheckoutInput struct {
	PetID         string
	OwnerUserID   string
	OwnerEmail    string
	PlanCode      PlanCode
	BillingMode   BillingMode
	SuccessURL    string
	CancelURL     string
}

func (s *Service) StartCheckout(ctx context.Context, in StartCheckoutInput) (CheckoutSession, error) {
	if _, err := GetPlan(in.PlanCode); err != nil {
		return CheckoutSession{}, err
	}
	mode, err := ParseBillingMode(string(in.BillingMode))
	if err != nil {
		return CheckoutSession{}, err
	}
	priceID := s.priceID(in.PlanCode, mode)
	if priceID == "" && !s.cfg.BillingMockEnabled {
		return CheckoutSession{}, fmt.Errorf("missing stripe price for %s/%s", in.PlanCode, mode)
	}
	if priceID == "" {
		priceID = fmt.Sprintf("price_mock_%s_%s", in.PlanCode, mode)
	}

	ent, err := s.store.GetEntitlementByPetID(ctx, in.PetID)
	if err != nil {
		return CheckoutSession{}, err
	}
	if ent.Status != string(StatusPending) {
		return CheckoutSession{}, fmt.Errorf("entitlement not pending")
	}

	customerID, _ := s.store.GetStripeCustomerID(ctx, in.OwnerUserID)
	checkoutMode := "payment"
	if mode == ModeSubscription {
		checkoutMode = "subscription"
	}
	successURL := in.SuccessURL
	if successURL == "" {
		successURL = s.cfg.StripeSuccessURL
	}
	cancelURL := in.CancelURL
	if cancelURL == "" {
		cancelURL = s.cfg.StripeCancelURL
	}

	sess, err := s.gateway.CreateCheckoutSession(ctx, CheckoutRequest{
		PriceID:       priceID,
		Mode:          checkoutMode,
		CustomerID:    customerID,
		CustomerEmail: in.OwnerEmail,
		SuccessURL:    successURL,
		CancelURL:     cancelURL,
		Metadata: map[string]string{
			"pet_id":       in.PetID,
			"owner_user_id": in.OwnerUserID,
			"plan_code":    string(in.PlanCode),
			"billing_mode": string(mode),
		},
	})
	if err != nil {
		return CheckoutSession{}, err
	}
	if customerID == "" {
		// Live gateway may have created customer; mock does not persist it.
	}
	if err := s.store.SetEntitlementCheckoutSession(ctx, in.PetID, sess.ID); err != nil {
		return CheckoutSession{}, err
	}
	return sess, nil
}

func (s *Service) CreatePortalSession(ctx context.Context, userID, returnURL string) (PortalSession, error) {
	customerID, err := s.store.GetStripeCustomerID(ctx, userID)
	if err != nil || customerID == "" {
		return PortalSession{}, fmt.Errorf("no stripe customer")
	}
	if returnURL == "" {
		returnURL = s.cfg.StripeSuccessURL
	}
	return s.gateway.CreatePortalSession(ctx, customerID, returnURL)
}

func (s *Service) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	event, err := s.gateway.VerifyWebhook(payload, signature)
	if err != nil {
		return err
	}
	processed, err := s.store.IsStripeEventProcessed(ctx, event.ID)
	if err != nil {
		return err
	}
	if processed {
		return nil
	}

	switch event.Type {
	case "checkout.session.completed":
		err = s.handleCheckoutCompleted(ctx, event)
	case "invoice.paid":
		err = s.handleInvoicePaid(ctx, event)
	case "invoice.payment_failed":
		err = s.handleInvoicePaymentFailed(ctx, event)
	case "customer.subscription.updated":
		err = s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		err = s.handleSubscriptionDeleted(ctx, event)
	default:
		err = nil
	}
	if err != nil {
		return err
	}
	return s.store.RecordStripeEvent(ctx, event.ID, event.Type)
}

func (s *Service) MockCompleteCheckout(ctx context.Context, petID, ownerUserID, planCode, billingMode, sessionID string) error {
	payload, sig, err := BuildTestWebhookPayload(s.cfg.StripeWebhookSecret, "checkout.session.completed", map[string]any{
		"id":                sessionID,
		"payment_status":    "paid",
		"customer":          "cus_mock_" + ownerUserID,
		"subscription":      nil,
		"payment_intent":    "pi_mock_" + petID,
		"metadata": map[string]any{
			"pet_id":        petID,
			"owner_user_id": ownerUserID,
			"plan_code":     planCode,
			"billing_mode":  billingMode,
		},
	})
	if err != nil {
		return err
	}
	if billingMode == string(ModeSubscription) {
		obj := map[string]any{
			"id":             sessionID,
			"payment_status": "paid",
			"customer":       "cus_mock_" + ownerUserID,
			"subscription":   "sub_mock_" + petID,
			"payment_intent": "pi_mock_" + petID,
			"metadata": map[string]any{
				"pet_id":        petID,
				"owner_user_id": ownerUserID,
				"plan_code":     planCode,
				"billing_mode":  billingMode,
			},
		}
		payload, sig, err = BuildTestWebhookPayload(s.cfg.StripeWebhookSecret, "checkout.session.completed", obj)
		if err != nil {
			return err
		}
	}
	return s.HandleWebhook(ctx, payload, sig)
}

type StartAddonCheckoutInput struct {
	AddonID     string
	OwnerUserID string
	OwnerEmail  string
	AddonCode   AddonCode
	SuccessURL  string
	CancelURL   string
}

func (s *Service) StartAddonCheckout(ctx context.Context, in StartAddonCheckoutInput) (CheckoutSession, error) {
	addon, err := GetAddon(in.AddonCode)
	if err != nil {
		return CheckoutSession{}, err
	}
	priceID := s.addonPriceID(in.AddonCode)
	if priceID == "" && !s.cfg.BillingMockEnabled {
		return CheckoutSession{}, fmt.Errorf("missing stripe price for addon %s", in.AddonCode)
	}
	if priceID == "" {
		priceID = fmt.Sprintf("price_mock_addon_%s", in.AddonCode)
	}

	customerID, _ := s.store.GetStripeCustomerID(ctx, in.OwnerUserID)
	successURL := in.SuccessURL
	if successURL == "" {
		successURL = s.cfg.StripeSuccessURL
	}
	cancelURL := in.CancelURL
	if cancelURL == "" {
		cancelURL = s.cfg.StripeCancelURL
	}
	_ = addon
	return s.gateway.CreateCheckoutSession(ctx, CheckoutRequest{
		PriceID:       priceID,
		Mode:          "payment",
		CustomerID:    customerID,
		CustomerEmail: in.OwnerEmail,
		SuccessURL:    successURL,
		CancelURL:     cancelURL,
		Metadata: map[string]string{
			"kind":          "addon",
			"addon_id":      in.AddonID,
			"addon_code":    string(in.AddonCode),
			"owner_user_id": in.OwnerUserID,
		},
	})
}

func (s *Service) MockCompleteAddonCheckout(ctx context.Context, addonID, ownerUserID, addonCode, sessionID string) error {
	payload, sig, err := BuildTestWebhookPayload(s.cfg.StripeWebhookSecret, "checkout.session.completed", map[string]any{
		"id":             sessionID,
		"payment_status": "paid",
		"customer":       "cus_mock_" + ownerUserID,
		"subscription":   nil,
		"payment_intent": "pi_mock_addon_" + addonID,
		"metadata": map[string]any{
			"kind":          "addon",
			"addon_id":      addonID,
			"addon_code":    addonCode,
			"owner_user_id": ownerUserID,
		},
	})
	if err != nil {
		return err
	}
	return s.HandleWebhook(ctx, payload, sig)
}

func (s *Service) handleAddonCheckoutCompleted(ctx context.Context, obj map[string]any, meta map[string]string) error {
	addonID := meta["addon_id"]
	if addonID == "" {
		return fmt.Errorf("missing addon_id in checkout metadata")
	}
	now := time.Now()
	addon, err := s.store.GetAddonEntitlement(ctx, addonID)
	if err != nil {
		return err
	}
	addonDef, err := GetAddon(AddonCode(addon.AddonCode))
	if err != nil {
		return err
	}
	validUntil := AddonValidUntil(now, addonDef)
	customerID, _ := asString(obj["customer"])
	piID, _ := asString(obj["payment_intent"])
	sessionID, _ := asString(obj["id"])
	if addon.OwnerUserID != "" && customerID != "" {
		_ = s.store.UpsertStripeCustomer(ctx, addon.OwnerUserID, customerID)
	}
	if err := s.store.ActivateAddonEntitlement(ctx, addonID, now, validUntil, sessionID, piID); err != nil {
		return err
	}
	if err := s.store.AccrueCommercialForAddon(ctx, addonID); err != nil {
		fmt.Printf("commercial addon accrual failed for addon %s: %v\n", addonID, err)
	}
	return nil
}

func (s *Service) addonPriceID(code AddonCode) string {
	switch code {
	case AddonFamily:
		return s.cfg.StripePriceAddonFamily
	case AddonCarePlus:
		return s.cfg.StripePriceAddonCarePlus
	case AddonHorse:
		return s.cfg.StripePriceAddonHorse
	default:
		return ""
	}
}

func (s *Service) handleCheckoutCompleted(ctx context.Context, event StripeEvent) error {
	obj := objectMap(event)
	meta := metadataMap(obj)
	if meta["kind"] == "addon" {
		return s.handleAddonCheckoutCompleted(ctx, obj, meta)
	}
	petID := meta["pet_id"]
	ownerUserID := meta["owner_user_id"]
	planCode, _ := ParsePlanCode(meta["plan_code"])
	_ , _ = ParseBillingMode(meta["billing_mode"])
	if petID == "" {
		return fmt.Errorf("missing pet_id in checkout metadata")
	}
	plan, err := GetPlan(planCode)
	if err != nil {
		return err
	}
	now := time.Now()
	validUntil := ValidUntil(now, plan)
	customerID, _ := asString(obj["customer"])
	subID, _ := asString(obj["subscription"])
	piID, _ := asString(obj["payment_intent"])
	sessionID, _ := asString(obj["id"])

	if ownerUserID != "" && customerID != "" {
		_ = s.store.UpsertStripeCustomer(ctx, ownerUserID, customerID)
	}
	if err := s.store.ActivateEntitlement(ctx, store.ActivateEntitlementParams{
		PetID:                 petID,
		Status:                string(StatusActive),
		ValidFrom:             now,
		ValidUntil:            validUntil,
		StripeCheckoutSession: sessionID,
		StripePaymentIntent:   piID,
		StripeSubscription:    subID,
	}); err != nil {
		return err
	}
	_ = s.store.EnsureDefaultCommissionTiers(ctx)
	// Accrual must not fail the Stripe webhook once entitlement is active.
	if err := s.store.AccrueCommissionForPetActivation(ctx, petID); err != nil {
		fmt.Printf("commission accrual failed for pet %s: %v\n", petID, err)
	}
	return nil
}

func (s *Service) handleInvoicePaid(ctx context.Context, event StripeEvent) error {
	obj := objectMap(event)
	subID, _ := asString(obj["subscription"])
	if subID == "" {
		return nil
	}
	ent, err := s.store.GetEntitlementBySubscriptionID(ctx, subID)
	if err != nil {
		return nil
	}
	plan, err := GetPlan(PlanCode(ent.PlanCode))
	if err != nil {
		return err
	}
	now := time.Now()
	validUntil := ValidUntil(now, plan)
	return s.store.ActivateEntitlement(ctx, store.ActivateEntitlementParams{
		PetID:      ent.PetID,
		Status:     string(StatusActive),
		ValidFrom:  now,
		ValidUntil: validUntil,
	})
}

func (s *Service) handleInvoicePaymentFailed(ctx context.Context, event StripeEvent) error {
	obj := objectMap(event)
	subID, _ := asString(obj["subscription"])
	if subID == "" {
		return nil
	}
	ent, err := s.store.GetEntitlementBySubscriptionID(ctx, subID)
	if err != nil {
		return nil
	}
	return s.store.UpdateEntitlementStatus(ctx, ent.PetID, string(StatusPastDue))
}

func (s *Service) handleSubscriptionUpdated(ctx context.Context, event StripeEvent) error {
	obj := objectMap(event)
	subID, _ := asString(obj["id"])
	status, _ := asString(obj["status"])
	ent, err := s.store.GetEntitlementBySubscriptionID(ctx, subID)
	if err != nil {
		return nil
	}
	switch status {
	case "active", "trialing":
		return s.store.UpdateEntitlementStatus(ctx, ent.PetID, string(StatusActive))
	case "past_due", "unpaid":
		return s.store.UpdateEntitlementStatus(ctx, ent.PetID, string(StatusPastDue))
	case "canceled":
		return s.store.UpdateEntitlementStatus(ctx, ent.PetID, string(StatusCancelled))
	default:
		return nil
	}
}

func (s *Service) handleSubscriptionDeleted(ctx context.Context, event StripeEvent) error {
	obj := objectMap(event)
	subID, _ := asString(obj["id"])
	ent, err := s.store.GetEntitlementBySubscriptionID(ctx, subID)
	if err != nil {
		return nil
	}
	return s.store.UpdateEntitlementStatus(ctx, ent.PetID, string(StatusCancelled))
}

func (s *Service) PetHasPremiumAccess(ctx context.Context, petID string) (bool, error) {
	return s.store.HasActiveEntitlement(ctx, petID)
}

func (s *Service) priceID(plan PlanCode, mode BillingMode) string {
	switch {
	case plan == PlanAnnual && mode == ModeOneTime:
		return s.cfg.StripePriceAnnualOnetime
	case plan == PlanTriennial && mode == ModeOneTime:
		return s.cfg.StripePriceTriennialOnetime
	case plan == PlanQuinquennial && mode == ModeOneTime:
		return s.cfg.StripePriceQuinquennialOnetime
	case plan == PlanAnnual && mode == ModeSubscription:
		return s.cfg.StripePriceAnnualSub
	case plan == PlanTriennial && mode == ModeSubscription:
		return s.cfg.StripePriceTriennialSub
	case plan == PlanQuinquennial && mode == ModeSubscription:
		return s.cfg.StripePriceQuinquennialSub
	default:
		return ""
	}
}

func objectMap(event StripeEvent) map[string]any {
	if obj, ok := event.Data["object"].(map[string]any); ok {
		return obj
	}
	return map[string]any{}
}

func metadataMap(obj map[string]any) map[string]string {
	out := map[string]string{}
	raw, ok := obj["metadata"].(map[string]any)
	if !ok {
		return out
	}
	for k, v := range raw {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

func asString(v any) (string, bool) {
	s, ok := v.(string)
	return s, ok && s != ""
}
