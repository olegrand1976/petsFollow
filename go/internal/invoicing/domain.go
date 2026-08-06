package invoicing

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
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
	Name               string `json:"name"`
	VATNumber          string `json:"vatNumber,omitempty"`
	CompanyNumber      string `json:"companyNumber,omitempty"` // BE BCE (optional)
	SIRET              string `json:"siret,omitempty"`         // FR
	SIREN              string `json:"siren,omitempty"`         // FR
	CodiceDestinatario string `json:"codiceDestinatario,omitempty"` // IT SDI (7 chars)
	PEC                string `json:"pec,omitempty"`                 // IT certified email
	TaxID              string `json:"taxId,omitempty"`               // ES NIF/CIF
	Email              string `json:"email,omitempty"`
	Street             string `json:"street,omitempty"`
	City               string `json:"city,omitempty"`
	Postal             string `json:"postal,omitempty"`
	Country            string `json:"country,omitempty"` // ISO-3166-1 alpha-2
}

type Line struct {
	Description        string  `json:"description"`
	Quantity           float64 `json:"quantity"`
	UnitPriceExclCents int64   `json:"unitPriceExclCents"`
	VATPercent         float64 `json:"vatPercent"`
}

type DocSource string

const (
	SourcePractice   DocSource = "practice"
	SourceSaasMaster DocSource = "saas_master"
)

type Document struct {
	ID                string       `json:"id"`
	PracticeID        string       `json:"practiceId"`
	Type              DocType      `json:"type"`
	Status            DocStatus    `json:"status"`
	Source            DocSource    `json:"source,omitempty"`
	Number            string       `json:"number,omitempty"`
	BillitOrderID     string       `json:"billitOrderId,omitempty"`
	IdempotencyKey    string       `json:"idempotencyKey"`
	Counterparty      Counterparty `json:"counterparty"`
	Currency          string       `json:"currency"`
	TotalExclCents    int64        `json:"totalExclCents"`
	TotalVATCents     int64        `json:"totalVatCents"`
	TotalInclCents    int64        `json:"totalInclCents"`
	RelatedDocumentID string       `json:"relatedDocumentId,omitempty"`
	VisitID           string       `json:"visitId,omitempty"`
	DAFID             string       `json:"dafId,omitempty"`
	PeppolStatus      string       `json:"peppolStatus,omitempty"`
	SentAt            *time.Time   `json:"sentAt,omitempty"`
	CreatedBy         string       `json:"createdBy,omitempty"`
	CreatedAt         time.Time    `json:"createdAt"`
	UpdatedAt         time.Time    `json:"updatedAt"`
	Lines             []Line       `json:"lines,omitempty"`
	// AboutInvoiceNumber is set at send-time for CreditNotes (Billit AboutInvoiceNumber). Not persisted.
	AboutInvoiceNumber string `json:"-"`
}

type Connection struct {
	PracticeID          string           `json:"practiceId"`
	PracticeName        string           `json:"practiceName,omitempty"`
	ContactEmail        string           `json:"contactEmail,omitempty"`
	BillitPartyID       string           `json:"billitPartyId,omitempty"`
	Status              ConnectionStatus `json:"status"`
	InvoiceToPartner    bool             `json:"invoiceToPartner"`
	PartnerListedAt     *time.Time       `json:"partnerListedAt,omitempty"`
	DocsIncludedMonthly int              `json:"docsIncludedMonthly"`
	UsageThisMonth      int              `json:"usageThisMonth"`
	LastError           string           `json:"lastError,omitempty"`
	ConnectedAt         *time.Time       `json:"connectedAt,omitempty"`
	HasAPISecret        bool             `json:"hasApiSecret"`
	UpdatedAt           time.Time        `json:"updatedAt"`
	// SaasDraftEnabled is set by admin list (Flux A master available).
	SaasDraftEnabled bool `json:"saasDraftEnabled,omitempty"`
	// SaasDocument is the current-month Flux A invoice (if any).
	SaasDocument *SaasDocSummary `json:"saasDocument,omitempty"`
}

// SaasDocSummary is a compact view of the monthly master→practice invoice for admin UI.
type SaasDocSummary struct {
	ID             string    `json:"id"`
	Status         DocStatus `json:"status"`
	BillitOrderID  string    `json:"billitOrderId,omitempty"`
	TotalInclCents int64     `json:"totalInclCents"`
	PeppolStatus   string    `json:"peppolStatus,omitempty"`
}

// SaasTarget is a practice eligible for Flux A (active BE cabinet with complete fiscal profile).
type SaasTarget struct {
	PracticeID          string          `json:"practiceId"`
	PracticeName        string          `json:"practiceName,omitempty"`
	ContactEmail        string          `json:"contactEmail,omitempty"`
	VATNumber           string          `json:"vatNumber,omitempty"`
	HasBillitConnect    bool            `json:"hasBillitConnect"`
	SaasBillingEnabled  bool            `json:"saasBillingEnabled"`
	SaasDraftEnabled    bool            `json:"saasDraftEnabled,omitempty"`
	SaasDocument        *SaasDocSummary `json:"saasDocument,omitempty"`
}

// SaasDraftRunResult summarizes a C1 cron pass (draft only).
type SaasDraftRunResult struct {
	YYYYMM          int      `json:"yyyymm"`
	Limit           int      `json:"limit"`
	Offset          int      `json:"offset"`
	Batches         int      `json:"batches,omitempty"`
	Scanned         int      `json:"scanned"`
	Drafted         int      `json:"drafted"`
	Unchanged       int      `json:"unchanged"`
	Failed          int      `json:"failed"`
	Errors          []string `json:"errors,omitempty"`
	ErrorsTruncated bool     `json:"errorsTruncated,omitempty"`
	Done            bool     `json:"done,omitempty"` // true when all batches exhausted (all mode)
}

type PracticeParty struct {
	PracticeID    string
	LegalName     string
	VATNumber     string
	CompanyNumber string
	Email         string
	Street        string
	City          string
	Postal        string
	Country       string // ISO-2, default BE when empty
}

type GatewayStatus struct {
	Complete bool
	Message  string
}

var (
	ErrInvalidCounterparty = errors.New("invalid_counterparty")
	ErrInvalidLines        = errors.New("invalid_lines")
	reDigits               = regexp.MustCompile(`^\d+$`)
	reBEVat                = regexp.MustCompile(`(?i)^BE0?\d{9,10}$`)
	reITCodice             = regexp.MustCompile(`(?i)^[A-Z0-9]{7}$`)
)

// validBelgianVAT checks format + mod-97 checksum (aligné Billit / SPF Finances).
func validBelgianVAT(vat string) bool {
	v := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(vat), " ", ""))
	if !reBEVat.MatchString(v) {
		return false
	}
	digits := strings.TrimPrefix(v, "BE")
	if len(digits) == 9 {
		digits = "0" + digits
	}
	if len(digits) != 10 || !reDigits.MatchString(digits) {
		return false
	}
	base, err1 := strconv.ParseInt(digits[:8], 10, 64)
	check, err2 := strconv.ParseInt(digits[8:], 10, 64)
	if err1 != nil || err2 != nil {
		return false
	}
	return 97-(base%97) == check
}

// CountryFromVAT returns the ISO-2 prefix of a VAT number (e.g. BE0123… → BE).
// Empty / non-prefixed VAT → "" (caller chooses default).
func CountryFromVAT(vat string) string {
	v := strings.ToUpper(strings.TrimSpace(vat))
	v = strings.ReplaceAll(v, " ", "")
	if len(v) < 2 {
		return ""
	}
	if v[0] >= 'A' && v[0] <= 'Z' && v[1] >= 'A' && v[1] <= 'Z' {
		return v[:2]
	}
	return ""
}

// NormalizeCounterparty trims fields and uppercases Country.
func NormalizeCounterparty(c Counterparty) Counterparty {
	c.Name = strings.TrimSpace(c.Name)
	c.VATNumber = strings.TrimSpace(c.VATNumber)
	c.CompanyNumber = strings.TrimSpace(c.CompanyNumber)
	c.SIRET = strings.TrimSpace(c.SIRET)
	c.SIREN = strings.TrimSpace(c.SIREN)
	c.CodiceDestinatario = strings.ToUpper(strings.TrimSpace(c.CodiceDestinatario))
	c.PEC = strings.TrimSpace(c.PEC)
	c.TaxID = strings.TrimSpace(c.TaxID)
	c.Email = strings.TrimSpace(c.Email)
	c.Street = strings.TrimSpace(c.Street)
	c.City = strings.TrimSpace(c.City)
	c.Postal = strings.TrimSpace(c.Postal)
	c.Country = strings.ToUpper(strings.TrimSpace(c.Country))
	if c.SIREN == "" && len(c.SIRET) == 14 && reDigits.MatchString(c.SIRET) {
		c.SIREN = c.SIRET[:9]
	}
	return c
}

// ValidateCounterparty enforces multi-country fiscal rules for electronic invoicing.
func ValidateCounterparty(c Counterparty) error {
	c = NormalizeCounterparty(c)
	if c.Name == "" {
		return fmt.Errorf("%w: name_required", ErrInvalidCounterparty)
	}
	if len(c.Country) != 2 {
		return fmt.Errorf("%w: country_required", ErrInvalidCounterparty)
	}
	switch c.Country {
	case "BE":
		if c.VATNumber == "" {
			return fmt.Errorf("%w: be_vat_required", ErrInvalidCounterparty)
		}
		if !validBelgianVAT(c.VATNumber) {
			return fmt.Errorf("%w: be_vat_invalid", ErrInvalidCounterparty)
		}
		// Peppol BE needs a postal address on the customer party.
		if c.Street == "" || c.City == "" || c.Postal == "" {
			return fmt.Errorf("%w: be_address_required", ErrInvalidCounterparty)
		}
	case "FR":
		if c.VATNumber == "" {
			return fmt.Errorf("%w: fr_vat_required", ErrInvalidCounterparty)
		}
		if c.SIRET == "" && c.SIREN == "" {
			return fmt.Errorf("%w: fr_siret_or_siren_required", ErrInvalidCounterparty)
		}
		if c.SIRET != "" && (len(c.SIRET) != 14 || !reDigits.MatchString(c.SIRET)) {
			return fmt.Errorf("%w: fr_siret_invalid", ErrInvalidCounterparty)
		}
		if c.SIREN != "" && (len(c.SIREN) != 9 || !reDigits.MatchString(c.SIREN)) {
			return fmt.Errorf("%w: fr_siren_invalid", ErrInvalidCounterparty)
		}
	case "IT":
		hasCodice := c.CodiceDestinatario != ""
		hasPEC := c.PEC != ""
		if hasCodice == hasPEC {
			return fmt.Errorf("%w: it_codice_xor_pec", ErrInvalidCounterparty)
		}
		if hasCodice && !reITCodice.MatchString(c.CodiceDestinatario) {
			return fmt.Errorf("%w: it_codice_invalid", ErrInvalidCounterparty)
		}
		if hasPEC && !strings.Contains(c.PEC, "@") {
			return fmt.Errorf("%w: it_pec_invalid", ErrInvalidCounterparty)
		}
		if c.VATNumber == "" {
			return fmt.Errorf("%w: it_vat_required", ErrInvalidCounterparty)
		}
	case "ES":
		if c.VATNumber == "" && c.TaxID == "" {
			return fmt.Errorf("%w: es_vat_or_nif_required", ErrInvalidCounterparty)
		}
	default:
		// Other countries: name + country only (Billit/Peppol may still need VAT later).
	}
	return nil
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

// ValidateLines rejects empty descriptions, non-positive qty/price, or out-of-range VAT.
// Caller must ensure len(lines) > 0 (ErrLinesRequired).
func ValidateLines(lines []Line) error {
	for _, l := range lines {
		if strings.TrimSpace(l.Description) == "" {
			return fmt.Errorf("%w: description_required", ErrInvalidLines)
		}
		if l.Quantity <= 0 || math.IsNaN(l.Quantity) || math.IsInf(l.Quantity, 0) {
			return fmt.Errorf("%w: quantity_invalid", ErrInvalidLines)
		}
		if l.UnitPriceExclCents <= 0 {
			return fmt.Errorf("%w: unit_price_invalid", ErrInvalidLines)
		}
		if l.VATPercent < 0 || l.VATPercent > 100 || math.IsNaN(l.VATPercent) || math.IsInf(l.VATPercent, 0) {
			return fmt.Errorf("%w: vat_percent_invalid", ErrInvalidLines)
		}
	}
	return nil
}

// ResolveWebhookApplyStatus enforces monotonic Peppol lifecycle:
// delivered/cancelled never regress; rejected ignores sending/unknown but accepts late delivered.
func ResolveWebhookApplyStatus(prev, incoming DocStatus) DocStatus {
	switch prev {
	case StatusDelivered:
		return StatusDelivered
	case StatusCancelled:
		return StatusCancelled
	case StatusRejected:
		switch incoming {
		case StatusDelivered, StatusRejected, StatusCancelled:
			return incoming
		default:
			return StatusRejected
		}
	default:
		return incoming
	}
}

func PeppolRequired(t DocType) bool {
	return t == DocInvoice || t == DocCreditNote
}
