package invoicing

import (
	"context"
	"math"
	"time"
)

type DocType string

const (
	DocInvoice    DocType = "invoice"
	DocCreditNote DocType = "credit_note"
	DocProforma   DocType = "proforma"
)

type DocStatus string

const (
	StatusDraft     DocStatus = "draft"
	StatusIssued    DocStatus = "issued"
	StatusSending   DocStatus = "sending"
	StatusDelivered DocStatus = "delivered"
	StatusRejected  DocStatus = "rejected"
	StatusCancelled DocStatus = "cancelled"
)

type ConnectionStatus string

const (
	ConnDisconnected         ConnectionStatus = "disconnected"
	ConnPendingRegistration  ConnectionStatus = "pending_registration"
	ConnPendingKYC           ConnectionStatus = "pending_kyc"
	ConnActive               ConnectionStatus = "active"
	ConnSuspended            ConnectionStatus = "suspended"
	ConnError                ConnectionStatus = "error"
)

type Counterparty struct {
	Name      string `json:"name"`
	VATNumber string `json:"vatNumber,omitempty"`
	Email     string `json:"email,omitempty"`
	Street    string `json:"street,omitempty"`
	City      string `json:"city,omitempty"`
	Postal    string `json:"postal,omitempty"`
	Country   string `json:"country,omitempty"`
}

type Line struct {
	Description        string  `json:"description"`
	Quantity           float64 `json:"quantity"`
	UnitPriceExclCents int64   `json:"unitPriceExclCents"`
	VATPercent         float64 `json:"vatPercent"`
}

type Document struct {
	ID                string       `json:"id"`
	PracticeID        string       `json:"practiceId"`
	Type              DocType     `json:"type"`
	Status            DocStatus    `json:"status"`
	Number            string       `json:"number,omitempty"`
	BillitOrderID     string       `json:"billitOrderId,omitempty"`
	IdempotencyKey    string       `json:"idempotencyKey"`
	Counterparty      Counterparty `json:"counterparty"`
	Currency          string       `json:"currency"`
	TotalExclCents    int64        `json:"totalExclCents"`
	TotalVATCents     int64        `json:"totalVatCents"`
	TotalInclCents    int64        `json:"totalInclCents"`
	RelatedDocumentID string       `json:"relatedDocumentId,omitempty"`
	PeppolStatus      string       `json:"peppolStatus,omitempty"`
	SentAt            *time.Time   `json:"sentAt,omitempty"`
	CreatedBy         string       `json:"createdBy,omitempty"`
	CreatedAt         time.Time    `json:"createdAt"`
	UpdatedAt         time.Time    `json:"updatedAt"`
	Lines             []Line       `json:"lines,omitempty"`
}

type Connection struct {
	PracticeID           string           `json:"practiceId"`
	BillitPartyID        string           `json:"billitPartyId,omitempty"`
	Status               ConnectionStatus `json:"status"`
	InvoiceToPartner     bool             `json:"invoiceToPartner"`
	PartnerListedAt      *time.Time       `json:"partnerListedAt,omitempty"`
	DocsIncludedMonthly  int              `json:"docsIncludedMonthly"`
	UsageThisMonth       int              `json:"usageThisMonth"`
	LastError            string           `json:"lastError,omitempty"`
	ConnectedAt          *time.Time       `json:"connectedAt,omitempty"`
	HasAPISecret         bool             `json:"hasApiSecret"`
	UpdatedAt            time.Time        `json:"updatedAt"`
}

type PracticeParty struct {
	PracticeID   string
	LegalName    string
	VATNumber    string
	CompanyNumber string
	Email        string
}

type GatewayStatus struct {
	Complete bool
	Message  string
}

// Gateway abstracts Billit (live or mock).
type Gateway interface {
	EnsureParty(ctx context.Context, practice PracticeParty) (partyID string, err error)
	CheckParty(ctx context.Context, partyID, apiKey string) (GatewayStatus, error)
	CreateDocument(ctx context.Context, partyID, apiKey string, doc Document) (externalID string, err error)
	SendPeppol(ctx context.Context, partyID, apiKey, externalID string) error
}

func ComputeTotals(lines []Line) (excl, vat, incl int64) {
	for _, l := range lines {
		lineExcl := int64(math.Round(float64(l.UnitPriceExclCents) * l.Quantity))
		lineVAT := int64(math.Round(float64(lineExcl) * l.VATPercent / 100.0))
		excl += lineExcl
		vat += lineVAT
	}
	incl = excl + vat
	return excl, vat, incl
}

func PeppolRequired(t DocType) bool {
	return t == DocInvoice || t == DocCreditNote
}
