package billing

import "context"

type CheckoutRequest struct {
	PriceID       string
	Mode          string
	CustomerID    string
	CustomerEmail string
	SuccessURL    string
	CancelURL     string
	Metadata      map[string]string
}

type CheckoutSession struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type PortalSession struct {
	URL string `json:"url"`
}

type StripeEvent struct {
	ID   string
	Type string
	Data map[string]any
}

type Gateway interface {
	CreateCheckoutSession(ctx context.Context, req CheckoutRequest) (CheckoutSession, error)
	CreatePortalSession(ctx context.Context, customerID, returnURL string) (PortalSession, error)
	VerifyWebhook(payload []byte, signature string) (StripeEvent, error)
}
