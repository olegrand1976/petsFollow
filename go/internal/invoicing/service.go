package invoicing

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
)

var (
	ErrDisabled            = errors.New("invoicing_disabled")
	ErrNotConnected        = errors.New("invoicing_not_connected")
	ErrProfileIncomplete   = errors.New("practice_profile_incomplete")
	ErrInvalidState        = errors.New("invalid_connect_state")
	ErrDocNotFound         = errors.New("document_not_found")
	ErrDocNotDraft         = errors.New("document_not_draft")
	ErrPeppolNotForType    = errors.New("peppol_not_allowed_for_type")
	ErrConnectionNotActive = errors.New("connection_not_active")
	ErrInvalidType         = errors.New("invalid_type")
	ErrLinesRequired       = errors.New("lines_required")
	ErrDocsQuotaExceeded   = errors.New("docs_quota_exceeded")
	ErrSendInProgress      = errors.New("document_send_in_progress")
)

// Store is the persistence port used by Service.
type Store interface {
	GetPracticeInvoicingProfile(ctx context.Context, practiceID string) (PracticeParty, bool, error)
	UpsertConnection(ctx context.Context, c Connection, apiSecretRef string) error
	GetConnection(ctx context.Context, practiceID string) (Connection, string, error)
	ListConnections(ctx context.Context) ([]Connection, error)
	MarkPartnerListed(ctx context.Context, practiceID string) error
	CreateConnectState(ctx context.Context, state, practiceID, userID string, expiresAt time.Time) error
	ConsumeConnectState(ctx context.Context, state, practiceID string) error
	CreateDocument(ctx context.Context, doc Document, lines []Line) (Document, error)
	GetDocument(ctx context.Context, practiceID, docID string) (Document, error)
	ListDocuments(ctx context.Context, practiceID string, limit int) ([]Document, error)
	ClaimDocumentForSend(ctx context.Context, practiceID, docID string) (Document, error)
	UpdateDocumentExternal(ctx context.Context, practiceID, docID, orderID string, status DocStatus, peppolStatus string, sentAt *time.Time) error
	IncrementUsage(ctx context.Context, practiceID string, yyyymm int) error
	UsageForMonth(ctx context.Context, practiceID string, yyyymm int) (int, error)
}

type Service struct {
	store Store
	gw    Gateway
	cfg   config.Config
}

func NewService(store Store, gw Gateway, cfg config.Config) *Service {
	return &Service{store: store, gw: gw, cfg: cfg}
}

func (s *Service) Enabled() bool { return s.cfg.BillitEnabled }

func (s *Service) GetConnection(ctx context.Context, practiceID string) (Connection, error) {
	c, _, err := s.store.GetConnection(ctx, practiceID)
	if err != nil {
		if errors.Is(err, ErrNotConnected) {
			return Connection{
				PracticeID:          practiceID,
				Status:              ConnDisconnected,
				InvoiceToPartner:    true,
				DocsIncludedMonthly: s.cfg.BillitDefaultDocsIncluded,
			}, nil
		}
		return Connection{}, err
	}
	c.UsageThisMonth, _ = s.store.UsageForMonth(ctx, practiceID, currentYYYYMM())
	return c, nil
}

type ConnectStartResult struct {
	ResellerURL string    `json:"resellerUrl"`
	State       string    `json:"state"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

func (s *Service) ConnectStart(ctx context.Context, practiceID, userID string) (ConnectStartResult, error) {
	if !s.Enabled() {
		return ConnectStartResult{}, ErrDisabled
	}
	_, complete, err := s.store.GetPracticeInvoicingProfile(ctx, practiceID)
	if err != nil {
		return ConnectStartResult{}, err
	}
	if !complete {
		return ConnectStartResult{}, ErrProfileIncomplete
	}
	state, err := NewConnectState()
	if err != nil {
		return ConnectStartResult{}, err
	}
	exp := time.Now().UTC().Add(2 * time.Hour)
	if err := s.store.CreateConnectState(ctx, state, practiceID, userID, exp); err != nil {
		return ConnectStartResult{}, err
	}

	c, secretRef, err := s.store.GetConnection(ctx, practiceID)
	if err != nil && !errors.Is(err, ErrNotConnected) {
		return ConnectStartResult{}, err
	}
	if errors.Is(err, ErrNotConnected) {
		c = Connection{
			PracticeID:          practiceID,
			InvoiceToPartner:    true,
			DocsIncludedMonthly: s.cfg.BillitDefaultDocsIncluded,
		}
		secretRef = ""
	}
	c.PracticeID = practiceID
	c.Status = ConnPendingRegistration
	if c.DocsIncludedMonthly <= 0 {
		c.DocsIncludedMonthly = s.cfg.BillitDefaultDocsIncluded
	}
	c.InvoiceToPartner = true
	if err := s.store.UpsertConnection(ctx, c, secretRef); err != nil {
		return ConnectStartResult{}, err
	}

	rawURL := strings.TrimSpace(s.cfg.BillitResellerRegisterURL)
	if rawURL == "" {
		rawURL = "https://my.billit.be/"
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ConnectStartResult{}, fmt.Errorf("invalid reseller url: %w", err)
	}
	q := u.Query()
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return ConnectStartResult{
		ResellerURL: u.String(),
		State:       state,
		ExpiresAt:   exp,
	}, nil
}

type ConnectCompleteInput struct {
	State   string `json:"state"`
	PartyID string `json:"partyId"`
	APIKey  string `json:"apiKey"`
}

func (s *Service) ConnectComplete(ctx context.Context, practiceID string, in ConnectCompleteInput) (Connection, error) {
	if !s.Enabled() {
		return Connection{}, ErrDisabled
	}
	in.PartyID = strings.TrimSpace(in.PartyID)
	in.APIKey = strings.TrimSpace(in.APIKey)
	in.State = strings.TrimSpace(in.State)
	if in.PartyID == "" || in.APIKey == "" || in.State == "" {
		return Connection{}, ErrInvalidState
	}
	if err := s.store.ConsumeConnectState(ctx, in.State, practiceID); err != nil {
		return Connection{}, err
	}
	st, err := s.gw.CheckParty(ctx, in.PartyID, in.APIKey)
	if err != nil {
		return Connection{}, err
	}
	ref, err := SealAPIKey(s.cfg.BillitSecretsBackend, s.cfg.BillitSecretsKey, in.APIKey)
	if err != nil {
		return Connection{}, err
	}
	now := time.Now().UTC()
	status := ConnPendingKYC
	if st.Complete {
		status = ConnActive
	}
	c := Connection{
		PracticeID:          practiceID,
		BillitPartyID:       in.PartyID,
		Status:              status,
		InvoiceToPartner:    true,
		DocsIncludedMonthly: s.cfg.BillitDefaultDocsIncluded,
		ConnectedAt:         &now,
		HasAPISecret:        true,
		UpdatedAt:           now,
	}
	if err := s.store.UpsertConnection(ctx, c, ref); err != nil {
		return Connection{}, err
	}
	c.UsageThisMonth, _ = s.store.UsageForMonth(ctx, practiceID, currentYYYYMM())
	return c, nil
}

func (s *Service) ConnectRefresh(ctx context.Context, practiceID string) (Connection, error) {
	c, ref, err := s.store.GetConnection(ctx, practiceID)
	if err != nil {
		return Connection{}, err
	}
	if c.BillitPartyID == "" || ref == "" {
		return c, ErrNotConnected
	}
	key, err := OpenAPIKey(s.cfg.BillitSecretsBackend, s.cfg.BillitSecretsKey, ref)
	if err != nil {
		return Connection{}, err
	}
	st, err := s.gw.CheckParty(ctx, c.BillitPartyID, key)
	if err != nil {
		return Connection{}, err
	}
	if st.Complete {
		c.Status = ConnActive
		c.LastError = ""
	} else {
		c.Status = ConnPendingKYC
		c.LastError = st.Message
	}
	if err := s.store.UpsertConnection(ctx, c, ref); err != nil {
		return Connection{}, err
	}
	c.UsageThisMonth, _ = s.store.UsageForMonth(ctx, practiceID, currentYYYYMM())
	c.HasAPISecret = true
	return c, nil
}

type CreateDocumentInput struct {
	Type           DocType     `json:"type"`
	IdempotencyKey string      `json:"idempotencyKey"`
	Counterparty   Counterparty `json:"counterparty"`
	Lines          []Line      `json:"lines"`
	RelatedID      string      `json:"relatedDocumentId"`
}

func (s *Service) CreateDocument(ctx context.Context, practiceID, userID string, in CreateDocumentInput) (Document, error) {
	if !s.Enabled() {
		return Document{}, ErrDisabled
	}
	c, _, err := s.store.GetConnection(ctx, practiceID)
	if err != nil {
		return Document{}, ErrNotConnected
	}
	if c.Status != ConnActive {
		return Document{}, ErrConnectionNotActive
	}
	switch in.Type {
	case DocInvoice, DocCreditNote, DocProforma:
	default:
		return Document{}, ErrInvalidType
	}
	if len(in.Lines) == 0 {
		return Document{}, ErrLinesRequired
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		in.IdempotencyKey = uuid.NewString()
	}
	excl, vat, incl := ComputeTotals(in.Lines)
	now := time.Now().UTC()
	doc := Document{
		ID:                uuid.NewString(),
		PracticeID:        practiceID,
		Type:              in.Type,
		Status:            StatusDraft,
		IdempotencyKey:    in.IdempotencyKey,
		Counterparty:      in.Counterparty,
		Currency:          "EUR",
		TotalExclCents:    excl,
		TotalVATCents:     vat,
		TotalInclCents:    incl,
		RelatedDocumentID: in.RelatedID,
		CreatedBy:         userID,
		CreatedAt:         now,
		UpdatedAt:         now,
		Lines:             in.Lines,
	}
	return s.store.CreateDocument(ctx, doc, in.Lines)
}

func (s *Service) ListDocuments(ctx context.Context, practiceID string) ([]Document, error) {
	return s.store.ListDocuments(ctx, practiceID, 50)
}

func (s *Service) GetDocument(ctx context.Context, practiceID, docID string) (Document, error) {
	return s.store.GetDocument(ctx, practiceID, docID)
}

func (s *Service) SendDocument(ctx context.Context, practiceID, docID string) (Document, error) {
	c, ref, err := s.store.GetConnection(ctx, practiceID)
	if err != nil {
		return Document{}, ErrNotConnected
	}
	if c.Status != ConnActive {
		return Document{}, ErrConnectionNotActive
	}

	doc, err := s.store.ClaimDocumentForSend(ctx, practiceID, docID)
	if err != nil {
		return Document{}, err
	}

	yyyymm := currentYYYYMM()
	if PeppolRequired(doc.Type) {
		usage, uerr := s.store.UsageForMonth(ctx, practiceID, yyyymm)
		if uerr != nil {
			_ = s.store.UpdateDocumentExternal(ctx, practiceID, docID, doc.BillitOrderID, StatusDraft, "", nil)
			return Document{}, uerr
		}
		limit := c.DocsIncludedMonthly
		if limit <= 0 {
			limit = s.cfg.BillitDefaultDocsIncluded
		}
		if usage >= limit {
			_ = s.store.UpdateDocumentExternal(ctx, practiceID, docID, doc.BillitOrderID, StatusDraft, "", nil)
			return Document{}, ErrDocsQuotaExceeded
		}
	}

	key, err := OpenAPIKey(s.cfg.BillitSecretsBackend, s.cfg.BillitSecretsKey, ref)
	if err != nil {
		_ = s.store.UpdateDocumentExternal(ctx, practiceID, docID, doc.BillitOrderID, StatusDraft, "", nil)
		return Document{}, err
	}
	orderID := doc.BillitOrderID
	if orderID == "" {
		orderID, err = s.gw.CreateDocument(ctx, c.BillitPartyID, key, doc)
		if err != nil {
			_ = s.store.UpdateDocumentExternal(ctx, practiceID, docID, "", StatusDraft, "", nil)
			return Document{}, err
		}
	}
	now := time.Now().UTC()
	status := StatusIssued
	peppol := ""
	if PeppolRequired(doc.Type) {
		if err := s.gw.SendPeppol(ctx, c.BillitPartyID, key, orderID); err != nil {
			_ = s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, StatusRejected, "send_failed", nil)
			return Document{}, err
		}
		status = StatusDelivered
		peppol = "delivered"
		if err := s.store.IncrementUsage(ctx, practiceID, yyyymm); err != nil {
			return Document{}, err
		}
	}
	if err := s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, status, peppol, &now); err != nil {
		return Document{}, err
	}
	return s.store.GetDocument(ctx, practiceID, docID)
}

func (s *Service) ListAdminConnections(ctx context.Context) ([]Connection, error) {
	return s.store.ListConnections(ctx)
}

func (s *Service) MarkPartnerListed(ctx context.Context, practiceID string) error {
	return s.store.MarkPartnerListed(ctx, practiceID)
}

func currentYYYYMM() int {
	t := time.Now().UTC()
	return t.Year()*100 + int(t.Month())
}
