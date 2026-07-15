package billing

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/stripe/stripe-go/v81"
	billingportal "github.com/stripe/stripe-go/v81/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/webhook"
)

type LiveGateway struct {
	WebhookSecret string
}

func NewLiveGateway(secretKey, webhookSecret string) *LiveGateway {
	if secretKey == "" {
		secretKey = os.Getenv("STRIPE_SECRET_KEY")
	}
	stripe.Key = secretKey
	return &LiveGateway{WebhookSecret: webhookSecret}
}

func LiveEnabled(secretKey string, mockEnabled bool) bool {
	if mockEnabled {
		return false
	}
	if secretKey == "" {
		secretKey = os.Getenv("STRIPE_SECRET_KEY")
	}
	return secretKey != "" && (strings.HasPrefix(secretKey, "sk_live_") || strings.HasPrefix(secretKey, "sk_test_"))
}

func (g *LiveGateway) CreateCheckoutSession(_ context.Context, req CheckoutRequest) (CheckoutSession, error) {
	customerID := req.CustomerID
	if customerID == "" && req.CustomerEmail != "" {
		c, err := customer.New(&stripe.CustomerParams{Email: stripe.String(req.CustomerEmail)})
		if err != nil {
			return CheckoutSession{}, fmt.Errorf("stripe create customer: %w", err)
		}
		customerID = c.ID
	}
	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(req.Mode),
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: stripe.String(req.PriceID), Quantity: stripe.Int64(1)},
		},
		Metadata: req.Metadata,
	}
	if customerID != "" {
		params.Customer = stripe.String(customerID)
	}
	sess, err := checkoutsession.New(params)
	if err != nil {
		return CheckoutSession{}, fmt.Errorf("stripe checkout session: %w", err)
	}
	return CheckoutSession{ID: sess.ID, URL: sess.URL}, nil
}

func (g *LiveGateway) CreatePortalSession(_ context.Context, customerID, returnURL string) (PortalSession, error) {
	sess, err := billingportal.New(&stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	})
	if err != nil {
		return PortalSession{}, fmt.Errorf("stripe portal session: %w", err)
	}
	return PortalSession{URL: sess.URL}, nil
}

func (g *LiveGateway) VerifyWebhook(payload []byte, signature string) (StripeEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, g.WebhookSecret)
	if err != nil {
		return StripeEvent{}, err
	}
	data := map[string]any{}
	if obj, ok := event.Data.Object["id"]; ok {
		data["object"] = event.Data.Object
		_ = obj
	} else {
		data["object"] = event.Data.Object
	}
	return StripeEvent{ID: event.ID, Type: string(event.Type), Data: data}, nil
}

var _ Gateway = (*LiveGateway)(nil)
