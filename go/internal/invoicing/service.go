package invoicing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	ErrRelatedDocument     = errors.New("related_document_invalid")
	ErrOrderPersistFailed  = errors.New("order_persist_failed")
	ErrPartnerNotEligible  = errors.New("partner_mark_not_eligible")
	ErrMasterNotConfigured = errors.New("saas_master_not_configured")
	ErrSaasNotEligible     = errors.New("saas_not_eligible")
	ErrSaasBillingDisabled = errors.New("saas_billing_disabled")
	// ErrSaasDisabled: Flux A (SaaS LL-IT-SC → cabinet) en sommeil. La
	// facturation de l'abonnement Pro se fait hors Billit (commercial / compta) ;
	// le code reste en place derrière INVOICING_SAAS_ENABLED.
	ErrSaasDisabled = errors.New("invoicing_saas_disabled")
	// ErrGateway wraps live/mock Billit HTTP failures (mapped to HTTP 502).
	ErrGateway = errors.New("invoicing_gateway")
	// ErrAccountUnverified: garde-fou anti-abus Billit — le compte du cabinet
	// doit valider son téléphone ou son IBAN avant tout envoi. Distinct de
	// ErrGateway : rien n'est parti, le document reste renvoyable tel quel une
	// fois la vérification faite chez Billit, et le véto peut agir lui-même.
	ErrAccountUnverified = errors.New("invoicing_account_unverified")
	// ErrSecrets: practice ApiKey cannot be opened (key rotation) — reconnect Billit.
	ErrSecrets = errors.New("invoicing_secrets_mismatch")
	// ErrRelatedInvoiceNotOnBillit: credit note linked to an invoice never created on Billit.
	ErrRelatedInvoiceNotOnBillit = errors.New("related_invoice_not_on_billit")
	ErrProformaTokenInvalid      = errors.New("proforma_token_invalid")
	ErrProformaAlreadyAccepted   = errors.New("proforma_already_accepted")
	ErrProformaExpired           = errors.New("proforma_expired")
)

// Store is the persistence port used by Service.
type Store interface {
	GetPracticeInvoicingProfile(ctx context.Context, practiceID string) (PracticeParty, bool, error)
	UpsertConnection(ctx context.Context, c Connection, apiSecretRef string) error
	GetConnection(ctx context.Context, practiceID string) (Connection, string, error)
	ListConnections(ctx context.Context, yyyymm int) ([]Connection, error)
	MarkPartnerListed(ctx context.Context, practiceID string) error
	CreateConnectState(ctx context.Context, state, practiceID, userID string, expiresAt time.Time) error
	AssertConnectState(ctx context.Context, state, practiceID string) error
	ConsumeConnectState(ctx context.Context, state, practiceID string) error
	ConsumeConnectStateAndUpsertConnection(ctx context.Context, state string, c Connection, apiSecretRef string) error
	CreateDocument(ctx context.Context, doc Document, lines []Line) (Document, error)
	GetDocument(ctx context.Context, practiceID, docID string) (Document, error)
	ListDocuments(ctx context.Context, practiceID string, limit int) ([]Document, error)
	ClaimEmptyBillitOrder(ctx context.Context, practiceID, docID string) (claimed bool, err error)
	ClaimDocumentForSend(ctx context.Context, practiceID, docID string, yyyymm, quotaLimit int) (doc Document, prevStatus DocStatus, err error)
	ClaimSaasDocumentForSend(ctx context.Context, practiceID, docID string) (doc Document, prevStatus DocStatus, err error)
	IssueProformaForClient(ctx context.Context, practiceID, docID, token string, expiresAt, sentAt time.Time) (Document, error)
	GetProformaByPublicToken(ctx context.Context, token string) (Document, string, error) // doc, practiceName
	AcceptProformaByToken(ctx context.Context, token string, now time.Time) (proforma Document, err error)
	ClearProformaPublicToken(ctx context.Context, practiceID, docID string) error
	GetDocumentByIdempotency(ctx context.Context, practiceID, key string) (Document, error)
	UpdateDocumentExternal(ctx context.Context, practiceID, docID, orderID string, status DocStatus, peppolStatus string, sentAt *time.Time) error
	ApplyBillitWebhookStatus(ctx context.Context, orderID string, status DocStatus, peppolStatus string, sentAt *time.Time, yyyymm int) error
	ListSendingDocuments(ctx context.Context, cutoff time.Time, limit int) ([]SendingDoc, error)
	InsertWebhookEvent(ctx context.Context, provider, eventType, externalID string, payload []byte) (eventID string, duplicate bool, err error)
	MarkWebhookProcessed(ctx context.Context, eventID string, processErr string) error
	DeleteWebhookEvent(ctx context.Context, eventID string) error
	IncrementUsage(ctx context.Context, practiceID string, yyyymm int) error
	UsageForMonth(ctx context.Context, practiceID string, yyyymm int) (int, error)
	ListSaasTargets(ctx context.Context, yyyymm, limit, offset int, optedInOnly bool, allowlist []string) ([]SaasTarget, error)
	IsSaasEligible(ctx context.Context, practiceID string, allowlist []string) (eligible, billingEnabled bool, err error)
	SetSaasBillingEnabled(ctx context.Context, practiceID string, enabled bool) error
}

type Service struct {
	store Store
	gw    Gateway
	cfg   config.Config
}

func NewService(store Store, gw Gateway, cfg config.Config) *Service {
	return &Service{store: store, gw: gw, cfg: cfg}
}

// TestSetSecretsKey updates BILLIT_SECRETS_KEY (integration tests — key rotation).
func (s *Service) TestSetSecretsKey(key string) { s.cfg.BillitSecretsKey = key }

// TestSetSecretsBackend updates BILLIT_SECRETS_BACKEND (integration tests).
func (s *Service) TestSetSecretsBackend(backend string) { s.cfg.BillitSecretsBackend = backend }

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
	// Validate state first without consuming — bad credentials must not burn the token.
	if err := s.store.AssertConnectState(ctx, in.State, practiceID); err != nil {
		return Connection{}, err
	}
	st, err := s.gw.CheckParty(ctx, in.PartyID, in.APIKey)
	if err != nil {
		return Connection{}, fmt.Errorf("%w: %v", ErrGateway, err)
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
	docsIncluded := s.cfg.BillitDefaultDocsIncluded
	if existing, _, gerr := s.store.GetConnection(ctx, practiceID); gerr == nil && existing.DocsIncludedMonthly > 0 {
		docsIncluded = existing.DocsIncludedMonthly
	}
	c := Connection{
		PracticeID:          practiceID,
		BillitPartyID:       in.PartyID,
		Status:              status,
		InvoiceToPartner:    true,
		DocsIncludedMonthly: docsIncluded,
		ConnectedAt:         &now,
		HasAPISecret:        true,
		UpdatedAt:           now,
	}
	if err := s.store.ConsumeConnectStateAndUpsertConnection(ctx, in.State, c, ref); err != nil {
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
		return Connection{}, fmt.Errorf("%w: %v", ErrGateway, err)
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
	Type           DocType      `json:"type"`
	IdempotencyKey string       `json:"idempotencyKey"`
	Counterparty   Counterparty `json:"counterparty"`
	Lines          []Line       `json:"lines"`
	RelatedID      string       `json:"relatedDocumentId"`
	VisitID        string       `json:"visitId"`
	DAFID          string       `json:"dafId"`
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
	if err := ValidateLines(in.Lines); err != nil {
		return Document{}, err
	}
	in.Counterparty = NormalizeCounterparty(in.Counterparty)
	if err := ValidateCounterparty(in.Counterparty); err != nil {
		return Document{}, err
	}
	in.RelatedID = strings.TrimSpace(in.RelatedID)
	if in.Type == DocCreditNote {
		if in.RelatedID == "" {
			return Document{}, ErrRelatedDocument
		}
	}
	if in.RelatedID != "" {
		related, err := s.store.GetDocument(ctx, practiceID, in.RelatedID)
		if err != nil {
			if errors.Is(err, ErrDocNotFound) {
				return Document{}, ErrRelatedDocument
			}
			return Document{}, err
		}
		// Practice docs must not reference Flux A master invoices (wrong Peppol issuer).
		if related.Source == SourceSaasMaster {
			return Document{}, ErrRelatedDocument
		}
		if in.Type == DocCreditNote && related.Type != DocInvoice {
			return Document{}, ErrRelatedDocument
		}
		if in.Type == DocCreditNote {
			switch related.Status {
			case StatusIssued, StatusDelivered, StatusRejected, StatusSending:
				// ok — avoir dès que la facture a un parcours Billit (envoi / livré / rejet)
			default:
				return Document{}, ErrRelatedDocument
			}
		}
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
		VisitID:           strings.TrimSpace(in.VisitID),
		DAFID:             strings.TrimSpace(in.DAFID),
		CreatedBy:         userID,
		CreatedAt:         now,
		UpdatedAt:         now,
		Lines:             in.Lines,
		Source:            SourcePractice,
	}
	return s.store.CreateDocument(ctx, doc, in.Lines)
}

func (s *Service) ListDocuments(ctx context.Context, practiceID string) ([]Document, error) {
	return s.store.ListDocuments(ctx, practiceID, 50)
}

func (s *Service) GetDocument(ctx context.Context, practiceID, docID string) (Document, error) {
	doc, err := s.store.GetDocument(ctx, practiceID, docID)
	if err != nil {
		return Document{}, err
	}
	// SaaS master invoices are ops-only — hide from practice Pro API.
	if doc.Source == SourceSaasMaster {
		return Document{}, ErrDocNotFound
	}
	return doc, nil
}

func (s *Service) SendDocument(ctx context.Context, practiceID, docID string) (Document, error) {
	c, ref, err := s.store.GetConnection(ctx, practiceID)
	if err != nil {
		return Document{}, ErrNotConnected
	}
	if c.Status != ConnActive {
		return Document{}, ErrConnectionNotActive
	}

	peek, err := s.store.GetDocument(ctx, practiceID, docID)
	if err != nil {
		return Document{}, err
	}
	if peek.Type == DocProforma {
		return s.issueProformaToClient(ctx, practiceID, docID)
	}

	yyyymm := currentYYYYMM()
	limit := c.DocsIncludedMonthly
	if limit <= 0 {
		limit = s.cfg.BillitDefaultDocsIncluded
	}
	// Atomic claim + quota lock (serializes Peppol sends per practice).
	doc, prevStatus, err := s.store.ClaimDocumentForSend(ctx, practiceID, docID, yyyymm, limit)
	if err != nil {
		return Document{}, err
	}
	restore := func(orderID string, status DocStatus, peppol string) {
		if status == "" || status == StatusSending {
			status = StatusDraft
		}
		_ = s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, status, peppol, nil)
	}

	key, err := OpenAPIKey(s.cfg.BillitSecretsBackend, s.cfg.BillitSecretsKey, ref)
	if err != nil {
		restore(doc.BillitOrderID, prevStatus, "")
		if errors.Is(err, ErrSecrets) {
			return Document{}, err
		}
		return Document{}, fmt.Errorf("%w: %v", ErrSecrets, err)
	}
	if doc.Type == DocCreditNote {
		if err := s.enrichCreditNoteForBillit(ctx, practiceID, &doc); err != nil {
			restore(doc.BillitOrderID, prevStatus, "")
			return Document{}, err
		}
	}
	orderID := doc.BillitOrderID
	if orderID == "" {
		orderID, err = s.gw.CreateDocument(ctx, c.BillitPartyID, key, doc)
		if err != nil {
			restore("", prevStatus, "")
			return Document{}, fmt.Errorf("%w: %v", ErrGateway, err)
		}
		// Persist order id before Peppol so a crash mid-send still lets webhooks attach.
		if err := s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, StatusSending, "order_created", nil); err != nil {
			// Prefer rejected+orderID (webhook/support/retry) over draft without Billit link.
			if err2 := s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, StatusRejected, "order_persist_failed", nil); err2 != nil {
				restore("", prevStatus, "")
				return Document{}, err
			}
			return Document{}, fmt.Errorf("%w: %v", ErrOrderPersistFailed, err)
		}
	}
	now := time.Now().UTC()
	status := StatusIssued
	peppol := ""
	if PeppolRequired(doc.Type) {
		transport := ResolveTransport(doc.Counterparty)
		if err := s.gw.Send(ctx, c.BillitPartyID, key, orderID, transport); err != nil {
			if errors.Is(err, ErrAccountUnverified) {
				_ = s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, StatusRejected, failedPeppolStatus(transport, "account_unverified"), nil)
				return Document{}, err
			}
			_ = s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, StatusRejected, failedPeppolStatus(transport, "send_failed"), nil)
			return Document{}, fmt.Errorf("%w: %v", ErrGateway, err)
		}
		// Mock: gateway is synchronous — delivered + usage in one TX (same path as live webhook).
		if s.cfg.BillitMockEnabled {
			if err := s.store.ApplyBillitWebhookStatus(ctx, orderID, StatusDelivered, deliveredPeppolStatus(transport), &now, yyyymm); err != nil {
				restore(orderID, prevStatus, "")
				return Document{}, err
			}
			return s.store.GetDocument(ctx, practiceID, docID)
		}
		// Live: stay in sending until Billit webhook confirms delivery.
		status = StatusSending
		peppol = sendingPeppolStatus(transport)
	}
	if err := s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, status, peppol, &now); err != nil {
		if PeppolRequired(doc.Type) && !s.cfg.BillitMockEnabled {
			// Peppol already accepted remotely — keep sending so the webhook can attach.
			return Document{}, err
		}
		restore(orderID, prevStatus, "")
		return Document{}, err
	}
	return s.store.GetDocument(ctx, practiceID, docID)
}

// ReconcileResult résume un passage de réconciliation. Errors nomme la cause des
// échecs : sans elle, une clé Billit tournée ne se lit que comme un compteur
// `failed` constant d'un passage à l'autre.
type ReconcileResult struct {
	Candidates      int      `json:"candidates"`
	Reconciled      int      `json:"reconciled"`
	StillFlying     int      `json:"stillFlying"`
	Failed          int      `json:"failed"`
	Errors          []string `json:"errors,omitempty"`
	ErrorsTruncated bool     `json:"errorsTruncated,omitempty"`
}

const (
	maxReconcileErrors  = 20
	maxReconcileErrLen  = 160
	reconcileDefaultAge = 30 * time.Minute
)

// fail enregistre l'échec d'un document. Le message est tronqué : une erreur
// passerelle transporte un extrait de réponse Billit, et ce résumé finit dans
// les logs Cloud Scheduler.
func (r *ReconcileResult) fail(docID, reason string) {
	r.Failed++
	if len(r.Errors) >= maxReconcileErrors {
		r.ErrorsTruncated = true
		return
	}
	msg := docID + ": " + reason
	if runes := []rune(msg); len(runes) > maxReconcileErrLen {
		msg = string(runes[:maxReconcileErrLen]) + "…"
	}
	r.Errors = append(r.Errors, msg)
}

// ReconcileSending relit chez Billit les documents restés en `sending` : sans ce
// rattrapage, un webhook manqué laisse la facture en vol — et son quota
// consommé — jusqu'au rejet automatique à 7 jours (job de rétention).
// Ne réécrit que sur un statut terminal ; une lecture douteuse ou en erreur
// laisse le document tel quel pour le prochain passage.
func (s *Service) ReconcileSending(ctx context.Context, olderThan time.Duration, limit int) (ReconcileResult, error) {
	var out ReconcileResult
	if !s.Enabled() {
		return out, ErrDisabled
	}
	if olderThan <= 0 {
		olderThan = reconcileDefaultAge
	}
	docs, err := s.store.ListSendingDocuments(ctx, time.Now().UTC().Add(-olderThan), limit)
	if err != nil {
		return out, err
	}
	out.Candidates = len(docs)
	yyyymm := currentYYYYMM()
	now := time.Now().UTC()
	for _, d := range docs {
		c, ref, err := s.store.GetConnection(ctx, d.PracticeID)
		if err != nil {
			out.fail(d.DocumentID, "connection: "+err.Error())
			continue
		}
		if c.Status != ConnActive {
			out.fail(d.DocumentID, "connection_not_active: "+string(c.Status))
			continue
		}
		key, err := OpenAPIKey(s.cfg.BillitSecretsBackend, s.cfg.BillitSecretsKey, ref)
		if err != nil {
			out.fail(d.DocumentID, "api_key: "+err.Error())
			continue
		}
		st, err := s.gw.FetchStatus(ctx, c.BillitPartyID, key, d.BillitOrderID)
		if err != nil {
			out.fail(d.DocumentID, "fetch_status: "+err.Error())
			continue
		}
		if !st.Terminal {
			out.StillFlying++
			continue
		}
		var sentAt *time.Time
		if st.Status == StatusDelivered {
			sentAt = &now
		}
		if err := s.store.ApplyBillitWebhookStatus(ctx, d.BillitOrderID, st.Status, st.PeppolStatus, sentAt, yyyymm); err != nil {
			out.fail(d.DocumentID, "apply_status: "+err.Error())
			continue
		}
		out.Reconciled++
	}
	return out, nil
}

// deliveredPeppolStatus / sendingPeppolStatus keep the audit trail honest: an
// invoice mailed to a particulier never travelled on Peppol.
func deliveredPeppolStatus(t Transport) string {
	if t == TransportSMTP {
		return EmailStatusPrefix + "delivered"
	}
	return "delivered"
}

func sendingPeppolStatus(t Transport) string {
	if t == TransportSMTP {
		return EmailStatusPrefix + "sending"
	}
	return "sending"
}

// failedPeppolStatus applique la même règle aux échecs : « Échec de l'envoi »
// sur une facture de particulier parle du mail, pas du réseau Peppol.
func failedPeppolStatus(t Transport, reason string) string {
	if t == TransportSMTP {
		return EmailStatusPrefix + reason
	}
	return reason
}

// issueProformaToClient emails a magic-link (handler) — no Billit Offer create.
func (s *Service) issueProformaToClient(ctx context.Context, practiceID, docID string) (Document, error) {
	doc, err := s.store.GetDocument(ctx, practiceID, docID)
	if err != nil {
		return Document{}, err
	}
	email := strings.TrimSpace(doc.Counterparty.Email)
	if email == "" || !strings.Contains(email, "@") {
		return Document{}, fmt.Errorf("%w: email_required", ErrInvalidCounterparty)
	}
	token, err := randomToken(32)
	if err != nil {
		return Document{}, err
	}
	now := time.Now().UTC()
	expires := now.Add(ProformaClientTokenTTL)
	out, err := s.store.IssueProformaForClient(ctx, practiceID, docID, token, expires, now)
	if err != nil {
		return Document{}, err
	}
	out.PublicToken = token
	return out, nil
}

// PublicProformaView is the unauthenticated preview payload.
type PublicProformaView struct {
	Status         DocStatus    `json:"status"`
	PracticeName   string       `json:"practiceName"`
	Counterparty   Counterparty `json:"counterparty"`
	Currency       string       `json:"currency"`
	TotalExclCents int64        `json:"totalExclCents"`
	TotalVATCents  int64        `json:"totalVatCents"`
	TotalInclCents int64        `json:"totalInclCents"`
	Lines          []Line       `json:"lines"`
	ExpiresAt      *time.Time   `json:"expiresAt,omitempty"`
	AcceptedAt     *time.Time   `json:"acceptedAt,omitempty"`
	CanAccept      bool         `json:"canAccept"`
}

func (s *Service) GetPublicProforma(ctx context.Context, token string) (PublicProformaView, error) {
	if !s.Enabled() {
		return PublicProformaView{}, ErrDisabled
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return PublicProformaView{}, ErrProformaTokenInvalid
	}
	doc, practiceName, err := s.store.GetProformaByPublicToken(ctx, token)
	if err != nil {
		return PublicProformaView{}, err
	}
	now := time.Now().UTC()
	expired := doc.TokenExpiresAt != nil && now.After(*doc.TokenExpiresAt)
	canAccept := doc.Status == StatusIssued && doc.AcceptedAt == nil && !expired
	return PublicProformaView{
		Status:         doc.Status,
		PracticeName:   practiceName,
		Counterparty:   doc.Counterparty,
		Currency:       doc.Currency,
		TotalExclCents: doc.TotalExclCents,
		TotalVATCents:  doc.TotalVATCents,
		TotalInclCents: doc.TotalInclCents,
		Lines:          doc.Lines,
		ExpiresAt:      doc.TokenExpiresAt,
		AcceptedAt:     doc.AcceptedAt,
		CanAccept:      canAccept,
	}, nil
}

// AcceptProformaResult is returned after client validation + Peppol send attempt.
type AcceptProformaResult struct {
	Proforma Document `json:"proforma"`
	Invoice  Document `json:"invoice"`
}

// AcceptProforma converts an issued proforma into an invoice and sends Peppol.
func (s *Service) AcceptProforma(ctx context.Context, token string) (AcceptProformaResult, error) {
	if !s.Enabled() {
		return AcceptProformaResult{}, ErrDisabled
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return AcceptProformaResult{}, ErrProformaTokenInvalid
	}
	now := time.Now().UTC()
	proforma, err := s.store.AcceptProformaByToken(ctx, token, now)
	already := false
	if errors.Is(err, ErrProformaAlreadyAccepted) {
		already = true
		var practiceName string
		proforma, practiceName, err = s.store.GetProformaByPublicToken(ctx, token)
		_ = practiceName
		if err != nil {
			return AcceptProformaResult{}, err
		}
	} else if err != nil {
		return AcceptProformaResult{}, err
	}

	idemKey := "proforma-accept:" + proforma.ID
	finalize := func(pf Document, inv Document) (AcceptProformaResult, error) {
		// Invalidate magic-link only once Peppol is in flight / done (keep token on draft/rejected for retry).
		if inv.Status != StatusDraft && inv.Status != StatusRejected {
			_ = s.store.ClearProformaPublicToken(ctx, pf.PracticeID, pf.ID)
			if refreshed, err := s.store.GetDocument(ctx, pf.PracticeID, pf.ID); err == nil {
				pf = refreshed
			}
		}
		return AcceptProformaResult{Proforma: pf, Invoice: inv}, nil
	}
	sendIfNeeded := func(pf Document, inv Document) (AcceptProformaResult, error) {
		if inv.Status != StatusDraft && inv.Status != StatusRejected {
			return finalize(pf, inv)
		}
		sent, err := s.SendDocument(ctx, pf.PracticeID, inv.ID)
		if err != nil {
			inv, _ = s.store.GetDocument(ctx, pf.PracticeID, inv.ID)
			return AcceptProformaResult{Proforma: pf, Invoice: inv}, err
		}
		return finalize(pf, sent)
	}

	if already {
		if inv, err := s.store.GetDocumentByIdempotency(ctx, proforma.PracticeID, idemKey); err == nil {
			return sendIfNeeded(proforma, inv)
		}
	}

	invoice, err := s.CreateDocument(ctx, proforma.PracticeID, proforma.CreatedBy, CreateDocumentInput{
		Type:           DocInvoice,
		IdempotencyKey: idemKey,
		Counterparty:   proforma.Counterparty,
		Lines:          proforma.Lines,
		RelatedID:      proforma.ID,
		VisitID:        proforma.VisitID,
		DAFID:          proforma.DAFID,
	})
	if err != nil {
		return AcceptProformaResult{Proforma: proforma}, err
	}
	return sendIfNeeded(proforma, invoice)
}

func randomToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// enrichCreditNoteForBillit requires the related invoice to exist on Billit and
// sets AboutInvoiceNumber when the invoice has a Billit/PF number.
func (s *Service) enrichCreditNoteForBillit(ctx context.Context, practiceID string, doc *Document) error {
	relatedID := strings.TrimSpace(doc.RelatedDocumentID)
	if relatedID == "" {
		return ErrRelatedDocument
	}
	related, err := s.store.GetDocument(ctx, practiceID, relatedID)
	if err != nil {
		if errors.Is(err, ErrDocNotFound) {
			return ErrRelatedDocument
		}
		return err
	}
	if related.Type != DocInvoice || related.Source == SourceSaasMaster {
		return ErrRelatedDocument
	}
	if strings.TrimSpace(related.BillitOrderID) == "" && strings.TrimSpace(related.Number) == "" {
		return ErrRelatedInvoiceNotOnBillit
	}
	if n := strings.TrimSpace(related.Number); n != "" {
		doc.AboutInvoiceNumber = n
	}
	return nil
}

func (s *Service) ListAdminConnections(ctx context.Context) ([]Connection, error) {
	yyyymm := currentYYYYMM()
	items, err := s.store.ListConnections(ctx, yyyymm)
	if err != nil {
		return nil, err
	}
	enabled := s.SaasDraftEnabled()
	for i := range items {
		items[i].SaasDraftEnabled = enabled
		if !s.SaasFluxEnabled() {
			continue
		}
		key := fmt.Sprintf("saas:%s:%d", items[i].PracticeID, yyyymm)
		if d, err := s.store.GetDocumentByIdempotency(ctx, items[i].PracticeID, key); err == nil {
			items[i].SaasDocument = &SaasDocSummary{
				ID:             d.ID,
				Status:         d.Status,
				BillitOrderID:  d.BillitOrderID,
				TotalInclCents: d.TotalInclCents,
				PeppolStatus:   d.PeppolStatus,
			}
		}
	}
	return items, nil
}

const (
	defaultSaasCronLimit = 50
	maxSaasCronLimit     = 100
	maxSaasCronErrors    = 20
	maxSaasCronBatches   = 40 // 40×50 = 2000 practices max per cron call
)

func (s *Service) saasAllowlist() []string {
	return s.cfg.InvoicingSaasAllowlist
}

// ListSaasTargets returns Flux A candidates. optedInOnly=true for cron (billing flag / allowlist).
func (s *Service) ListSaasTargets(ctx context.Context, limit, offset int, optedInOnly bool) ([]SaasTarget, error) {
	if !s.Enabled() {
		return nil, ErrDisabled
	}
	if !s.SaasFluxEnabled() {
		return nil, ErrSaasDisabled
	}
	if limit <= 0 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.store.ListSaasTargets(ctx, saasYYYYMM(), limit, offset, optedInOnly, s.saasAllowlist())
	if err != nil {
		return nil, err
	}
	enabled := s.SaasDraftEnabled()
	for i := range items {
		items[i].SaasDraftEnabled = enabled
	}
	return items, nil
}

// RunMonthlySaasDrafts creates Flux A drafts for opted-in practices (C1 — no Peppol send).
// When all=true, loops offset batches until a short page (scanned < limit) or max batches.
func (s *Service) RunMonthlySaasDrafts(ctx context.Context, limit, offset int, all bool) (SaasDraftRunResult, error) {
	out := SaasDraftRunResult{YYYYMM: saasYYYYMM()}
	if !s.Enabled() {
		return out, ErrDisabled
	}
	if !s.SaasFluxEnabled() {
		return out, ErrSaasDisabled
	}
	if !s.SaasDraftEnabled() {
		return out, ErrMasterNotConfigured
	}
	if limit <= 0 {
		limit = defaultSaasCronLimit
	}
	if limit > maxSaasCronLimit {
		limit = maxSaasCronLimit
	}
	if offset < 0 {
		offset = 0
	}
	out.Limit = limit
	out.Offset = offset

	runBatch := func(off int) (SaasDraftRunResult, error) {
		batch := SaasDraftRunResult{YYYYMM: out.YYYYMM, Limit: limit, Offset: off}
		targets, err := s.ListSaasTargets(ctx, limit, off, true)
		if err != nil {
			return batch, err
		}
		batch.Scanned = len(targets)
		for _, t := range targets {
			hadOrder := t.SaasDocument != nil && t.SaasDocument.BillitOrderID != ""
			_, err := s.CreateSaasDraft(ctx, t.PracticeID, "")
			if err != nil {
				batch.Failed++
				if len(batch.Errors) < maxSaasCronErrors {
					batch.Errors = append(batch.Errors, t.PracticeID+": "+err.Error())
				} else {
					batch.ErrorsTruncated = true
				}
				continue
			}
			if hadOrder {
				batch.Unchanged++
			} else {
				batch.Drafted++
			}
		}
		return batch, nil
	}

	if !all {
		batch, err := runBatch(offset)
		if err != nil {
			return out, err
		}
		batch.Batches = 1
		batch.Done = batch.Scanned < limit
		return batch, nil
	}

	off := offset
	for range maxSaasCronBatches {
		batch, err := runBatch(off)
		if err != nil {
			return out, err
		}
		out.Batches++
		out.Scanned += batch.Scanned
		out.Drafted += batch.Drafted
		out.Unchanged += batch.Unchanged
		out.Failed += batch.Failed
		for _, e := range batch.Errors {
			if len(out.Errors) < maxSaasCronErrors {
				out.Errors = append(out.Errors, e)
			} else {
				out.ErrorsTruncated = true
			}
		}
		if batch.ErrorsTruncated {
			out.ErrorsTruncated = true
		}
		if batch.Scanned < limit {
			out.Done = true
			break
		}
		off += limit
	}
	out.Offset = offset
	if !out.Done && out.Batches >= maxSaasCronBatches {
		out.Done = false
	}
	return out, nil
}

func (s *Service) SetSaasBillingEnabled(ctx context.Context, practiceID string, enabled bool) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	if !s.SaasFluxEnabled() {
		return ErrSaasDisabled
	}
	eligible, _, err := s.store.IsSaasEligible(ctx, practiceID, nil)
	if err != nil {
		return err
	}
	if !eligible {
		return ErrSaasNotEligible
	}
	return s.store.SetSaasBillingEnabled(ctx, practiceID, enabled)
}

func (s *Service) MarkPartnerListed(ctx context.Context, practiceID string) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	return s.store.MarkPartnerListed(ctx, practiceID)
}

// SaasFluxEnabled reports whether Flux A (facturation SaaS LL-IT-SC → cabinet)
// is awake. Off by default: l'abonnement Pro se facture hors Billit.
func (s *Service) SaasFluxEnabled() bool {
	return s.cfg.InvoicingSaasEnabled
}

// SaasDraftEnabled reports whether Flux A master credentials are available (or mock defaults).
func (s *Service) SaasDraftEnabled() bool {
	if !s.Enabled() || !s.SaasFluxEnabled() {
		return false
	}
	party, key := s.masterCredentials()
	return party != "" && key != ""
}

func (s *Service) masterCredentials() (partyID, apiKey string) {
	partyID = strings.TrimSpace(s.cfg.BillitMasterPartyID)
	apiKey = strings.TrimSpace(s.cfg.BillitMasterAPIKey)
	if partyID != "" && apiKey != "" {
		return partyID, apiKey
	}
	if s.cfg.BillitMockEnabled {
		return "party_master_mock", "mock-master-key"
	}
	return "", ""
}

// CreateSaasDraft emits a Flux A draft invoice (LL-IT-SC master → practice) for the current month.
// Idempotent per practice/month via key saas:{practiceId}:{yyyymm}. Does not require a Billit connect.
func (s *Service) CreateSaasDraft(ctx context.Context, practiceID, adminUserID string) (Document, error) {
	if !s.Enabled() {
		return Document{}, ErrDisabled
	}
	if !s.SaasFluxEnabled() {
		return Document{}, ErrSaasDisabled
	}
	partyID, apiKey := s.masterCredentials()
	if partyID == "" || apiKey == "" {
		return Document{}, ErrMasterNotConfigured
	}
	profile, complete, err := s.store.GetPracticeInvoicingProfile(ctx, practiceID)
	if err != nil {
		return Document{}, err
	}
	if !complete {
		return Document{}, ErrProfileIncomplete
	}
	eligible, billingOn, err := s.store.IsSaasEligible(ctx, practiceID, s.saasAllowlist())
	if err != nil {
		return Document{}, err
	}
	if !eligible {
		return Document{}, ErrSaasNotEligible
	}
	if !billingOn {
		return Document{}, ErrSaasBillingDisabled
	}
	country := strings.ToUpper(strings.TrimSpace(profile.Country))
	if country == "" {
		country = CountryFromVAT(profile.VATNumber)
	}
	if country == "" {
		country = "BE"
	}
	// Flux A MVP: BE VAT + 21 % only (no silent 0 % for other countries).
	if country != "BE" {
		return Document{}, fmt.Errorf("%w: saas_be_only", ErrInvalidCounterparty)
	}
	cp := NormalizeCounterparty(Counterparty{
		Name:          profile.LegalName,
		Country:       country,
		VATNumber:     profile.VATNumber,
		CompanyNumber: profile.CompanyNumber,
		Email:         profile.Email,
		Street:        profile.Street,
		City:          profile.City,
		Postal:        profile.Postal,
	})
	if err := ValidateCounterparty(cp); err != nil {
		return Document{}, err
	}
	price := s.cfg.InvoicingSaasPriceEURCents
	if price <= 0 {
		price = 8800
	}
	lines := []Line{{
		Description:        "Abonnement petsFollow Pro",
		Quantity:           1,
		UnitPriceExclCents: int64(price),
		VATPercent:         21.0,
	}}
	if err := ValidateLines(lines); err != nil {
		return Document{}, err
	}
	excl, vat, incl := ComputeTotals(lines)
	yyyymm := saasYYYYMM()
	idem := fmt.Sprintf("saas:%s:%d", practiceID, yyyymm)
	now := time.Now().UTC()
	doc := Document{
		ID:             uuid.NewString(),
		PracticeID:     practiceID,
		Type:           DocInvoice,
		Status:         StatusDraft,
		Source:         SourceSaasMaster,
		IdempotencyKey: idem,
		Counterparty:   cp,
		Currency:       "EUR",
		TotalExclCents: excl,
		TotalVATCents:  vat,
		TotalInclCents: incl,
		CreatedBy:      adminUserID,
		CreatedAt:      now,
		UpdatedAt:      now,
		Lines:          lines,
	}
	saved, err := s.store.CreateDocument(ctx, doc, lines)
	if err != nil {
		return Document{}, err
	}
	if saved.BillitOrderID != "" {
		return saved, nil
	}
	claimed, err := s.store.ClaimEmptyBillitOrder(ctx, practiceID, saved.ID)
	if err != nil {
		return Document{}, err
	}
	if !claimed {
		// Another request holds creating_order — wait briefly for billit_order_id.
		for range 8 {
			d, err := s.store.GetDocument(ctx, practiceID, saved.ID)
			if err != nil {
				return Document{}, err
			}
			if d.BillitOrderID != "" {
				return d, nil
			}
			time.Sleep(40 * time.Millisecond)
		}
		return s.store.GetDocument(ctx, practiceID, saved.ID)
	}
	orderID, err := s.gw.CreateDocument(ctx, partyID, apiKey, saved)
	if err != nil {
		_ = s.store.UpdateDocumentExternal(ctx, practiceID, saved.ID, "", StatusDraft, "saas_draft", nil)
		return Document{}, fmt.Errorf("%w: %v", ErrGateway, err)
	}
	if err := s.store.UpdateDocumentExternal(ctx, practiceID, saved.ID, orderID, StatusDraft, "saas_draft", nil); err != nil {
		if err2 := s.store.UpdateDocumentExternal(ctx, practiceID, saved.ID, orderID, StatusRejected, "order_persist_failed", nil); err2 != nil {
			return Document{}, err
		}
		return Document{}, fmt.Errorf("%w: %v", ErrOrderPersistFailed, err)
	}
	return s.store.GetDocument(ctx, practiceID, saved.ID)
}

// SendSaasDocument sends a Flux A (saas_master) invoice via Billit master Peppol.
// Does not consume the practice monthly Peppol quota (webhook usage skip already in store).
func (s *Service) SendSaasDocument(ctx context.Context, practiceID, docID string) (Document, error) {
	if !s.Enabled() {
		return Document{}, ErrDisabled
	}
	if !s.SaasFluxEnabled() {
		return Document{}, ErrSaasDisabled
	}
	partyID, apiKey := s.masterCredentials()
	if partyID == "" || apiKey == "" {
		return Document{}, ErrMasterNotConfigured
	}

	// Ensure Billit order exists while still draft/rejected (before claim → sending).
	doc, err := s.store.GetDocument(ctx, practiceID, docID)
	if err != nil {
		return Document{}, err
	}
	if doc.Source != SourceSaasMaster {
		return Document{}, ErrDocNotDraft
	}
	switch doc.Status {
	case StatusDraft, StatusRejected:
	case StatusSending:
		return Document{}, ErrSendInProgress
	default:
		return Document{}, ErrDocNotDraft
	}
	if doc.BillitOrderID == "" {
		if err := s.ensureSaasBillitOrder(ctx, practiceID, docID, partyID, apiKey, doc); err != nil {
			return Document{}, err
		}
		doc, err = s.store.GetDocument(ctx, practiceID, docID)
		if err != nil {
			return Document{}, err
		}
		if doc.BillitOrderID == "" {
			return Document{}, fmt.Errorf("%w: order_missing", ErrGateway)
		}
	}

	doc, prevStatus, err := s.store.ClaimSaasDocumentForSend(ctx, practiceID, docID)
	if err != nil {
		return Document{}, err
	}
	restore := func(orderID string, status DocStatus, peppol string) {
		if status == "" || status == StatusSending {
			status = StatusDraft
		}
		_ = s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, status, peppol, nil)
	}

	orderID := doc.BillitOrderID
	if orderID == "" {
		// Should not happen after ensureSaasBillitOrder — fail closed.
		restore("", prevStatus, "saas_draft")
		return Document{}, fmt.Errorf("%w: order_missing", ErrGateway)
	}

	transport := ResolveTransport(doc.Counterparty)
	if err := s.gw.Send(ctx, partyID, apiKey, orderID, transport); err != nil {
		if errors.Is(err, ErrAccountUnverified) {
			_ = s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, StatusRejected, failedPeppolStatus(transport, "account_unverified"), nil)
			return Document{}, err
		}
		_ = s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, StatusRejected, failedPeppolStatus(transport, "send_failed"), nil)
		return Document{}, fmt.Errorf("%w: %v", ErrGateway, err)
	}
	now := time.Now().UTC()
	if s.cfg.BillitMockEnabled {
		if err := s.store.ApplyBillitWebhookStatus(ctx, orderID, StatusDelivered, "delivered", &now, currentYYYYMM()); err != nil {
			restore(orderID, prevStatus, "")
			return Document{}, err
		}
		return s.store.GetDocument(ctx, practiceID, docID)
	}
	if err := s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, StatusSending, "sending", &now); err != nil {
		// Peppol already accepted — keep sending for webhook.
		return Document{}, err
	}
	return s.store.GetDocument(ctx, practiceID, docID)
}

func (s *Service) ensureSaasBillitOrder(ctx context.Context, practiceID, docID, partyID, apiKey string, doc Document) error {
	claimed, err := s.store.ClaimEmptyBillitOrder(ctx, practiceID, docID)
	if err != nil {
		return err
	}
	if !claimed {
		for range 8 {
			d, err := s.store.GetDocument(ctx, practiceID, docID)
			if err != nil {
				return err
			}
			if d.BillitOrderID != "" {
				return nil
			}
			time.Sleep(40 * time.Millisecond)
		}
		return fmt.Errorf("%w: order_create_race", ErrGateway)
	}
	orderID, err := s.gw.CreateDocument(ctx, partyID, apiKey, doc)
	if err != nil {
		_ = s.store.UpdateDocumentExternal(ctx, practiceID, docID, "", doc.Status, "saas_draft", nil)
		return fmt.Errorf("%w: %v", ErrGateway, err)
	}
	status := doc.Status
	if status != StatusDraft && status != StatusRejected {
		status = StatusDraft
	}
	if err := s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, status, "saas_draft", nil); err != nil {
		if err2 := s.store.UpdateDocumentExternal(ctx, practiceID, docID, orderID, StatusRejected, "order_persist_failed", nil); err2 != nil {
			return err
		}
		return fmt.Errorf("%w: %v", ErrOrderPersistFailed, err)
	}
	return nil
}

// ApplyWebhookStatus updates a document from a Billit webhook (by order id).
// On first transition to delivered, increments monthly electronic-doc usage (same TX).
func (s *Service) ApplyWebhookStatus(ctx context.Context, orderID, eventType string, status DocStatus, peppolStatus string) error {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return ErrDocNotFound
	}
	var sentAt *time.Time
	if status == StatusDelivered {
		now := time.Now().UTC()
		sentAt = &now
	}
	_ = eventType
	return s.store.ApplyBillitWebhookStatus(ctx, orderID, status, peppolStatus, sentAt, currentYYYYMM())
}

// RecordWebhook stores the raw event for idempotency; duplicate=true if already seen.
func (s *Service) RecordWebhook(ctx context.Context, eventType, externalID string, payload []byte) (eventID string, duplicate bool, err error) {
	return s.store.InsertWebhookEvent(ctx, "billit", eventType, externalID, payload)
}

func (s *Service) MarkWebhookProcessed(ctx context.Context, eventID string, processErr error) error {
	msg := ""
	if processErr != nil {
		msg = processErr.Error()
	}
	return s.store.MarkWebhookProcessed(ctx, eventID, msg)
}

// ForgetWebhook removes an event so Billit can retry (e.g. document not yet persisted).
func (s *Service) ForgetWebhook(ctx context.Context, eventID string) error {
	return s.store.DeleteWebhookEvent(ctx, eventID)
}

func currentYYYYMM() int {
	t := time.Now().UTC()
	return t.Year()*100 + int(t.Month())
}

// saasYYYYMM is the Flux A billing month in Europe/Brussels (Peppol SaaS period).
func saasYYYYMM() int {
	loc, err := time.LoadLocation("Europe/Brussels")
	if err != nil {
		return currentYYYYMM()
	}
	t := time.Now().In(loc)
	return t.Year()*100 + int(t.Month())
}
