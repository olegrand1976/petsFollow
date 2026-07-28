package store

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
)

func (s *Store) GetPracticeInvoicingProfile(ctx context.Context, practiceID string) (invoicing.PracticeParty, bool, error) {
	var p invoicing.PracticeParty
	var legal, vat, company, email, street, city, postal string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(company_legal_name,''), COALESCE(vat_number,''), COALESCE(company_number,''),
		       COALESCE(NULLIF(contact_email,''), ''),
		       COALESCE(address_line1,''), COALESCE(city,''), COALESCE(postal_code,'')
		FROM practice.practices WHERE id = $1`, practiceID).Scan(
		&legal, &vat, &company, &email, &street, &city, &postal)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return p, false, nil
		}
		return p, false, err
	}
	p = invoicing.PracticeParty{
		PracticeID:    practiceID,
		LegalName:     legal,
		VATNumber:     vat,
		CompanyNumber: company,
		Email:         email,
		Street:        street,
		City:          city,
		Postal:        postal,
		Country:       invoicing.CountryFromVAT(vat),
	}
	complete := strings.TrimSpace(legal) != "" &&
		strings.TrimSpace(vat) != "" &&
		strings.TrimSpace(company) != "" &&
		strings.TrimSpace(email) != ""
	return p, complete, nil
}

func (s *Store) UpsertConnection(ctx context.Context, c invoicing.Connection, apiSecretRef string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO invoicing.practice_connections (
			practice_id, billit_party_id, status, invoice_to_partner, partner_listed_at,
			docs_included_monthly, api_secret_ref, last_error, connected_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9, now())
		ON CONFLICT (practice_id) DO UPDATE SET
			billit_party_id = EXCLUDED.billit_party_id,
			status = EXCLUDED.status,
			invoice_to_partner = EXCLUDED.invoice_to_partner,
			partner_listed_at = COALESCE(EXCLUDED.partner_listed_at, invoicing.practice_connections.partner_listed_at),
			docs_included_monthly = EXCLUDED.docs_included_monthly,
			api_secret_ref = CASE WHEN EXCLUDED.api_secret_ref = '' THEN invoicing.practice_connections.api_secret_ref ELSE EXCLUDED.api_secret_ref END,
			last_error = EXCLUDED.last_error,
			connected_at = COALESCE(EXCLUDED.connected_at, invoicing.practice_connections.connected_at),
			updated_at = now()`,
		c.PracticeID, nullIfEmpty(c.BillitPartyID), string(c.Status), c.InvoiceToPartner, c.PartnerListedAt,
		c.DocsIncludedMonthly, apiSecretRef, nullIfEmpty(c.LastError), c.ConnectedAt,
	)
	return err
}

func (s *Store) GetConnection(ctx context.Context, practiceID string) (invoicing.Connection, string, error) {
	var c invoicing.Connection
	var status string
	var party, secretRef, lastErr *string
	var partnerListed, connected *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT practice_id, billit_party_id, status, invoice_to_partner, partner_listed_at,
		       docs_included_monthly, api_secret_ref, last_error, connected_at, updated_at
		FROM invoicing.practice_connections WHERE practice_id = $1`, practiceID).Scan(
		&c.PracticeID, &party, &status, &c.InvoiceToPartner, &partnerListed,
		&c.DocsIncludedMonthly, &secretRef, &lastErr, &connected, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoicing.Connection{}, "", invoicing.ErrNotConnected
		}
		return invoicing.Connection{}, "", err
	}
	c.Status = invoicing.ConnectionStatus(status)
	if party != nil {
		c.BillitPartyID = *party
	}
	if lastErr != nil {
		c.LastError = *lastErr
	}
	c.PartnerListedAt = partnerListed
	c.ConnectedAt = connected
	ref := ""
	if secretRef != nil {
		ref = *secretRef
		c.HasAPISecret = ref != ""
	}
	return c, ref, nil
}

func (s *Store) ListConnections(ctx context.Context, yyyymm int) ([]invoicing.Connection, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.practice_id::text,
		       COALESCE(NULLIF(TRIM(p.company_legal_name), ''), NULLIF(TRIM(p.name), ''), ''),
		       COALESCE(p.contact_email, ''),
		       c.billit_party_id, c.status, c.invoice_to_partner, c.partner_listed_at,
		       c.docs_included_monthly, c.api_secret_ref, c.last_error, c.connected_at, c.updated_at,
		       COALESCE(u.doc_count, 0)
		FROM invoicing.practice_connections c
		JOIN practice.practices p ON p.id = c.practice_id
		LEFT JOIN invoicing.usage_monthly u
		  ON u.practice_id = c.practice_id AND u.yyyymm = $1
		ORDER BY c.updated_at DESC`, yyyymm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []invoicing.Connection
	for rows.Next() {
		var c invoicing.Connection
		var status string
		var party, secretRef, lastErr *string
		var partnerListed, connected *time.Time
		if err := rows.Scan(
			&c.PracticeID, &c.PracticeName, &c.ContactEmail,
			&party, &status, &c.InvoiceToPartner, &partnerListed,
			&c.DocsIncludedMonthly, &secretRef, &lastErr, &connected, &c.UpdatedAt,
			&c.UsageThisMonth,
		); err != nil {
			return nil, err
		}
		c.Status = invoicing.ConnectionStatus(status)
		if party != nil {
			c.BillitPartyID = *party
		}
		if lastErr != nil {
			c.LastError = *lastErr
		}
		c.PartnerListedAt = partnerListed
		c.ConnectedAt = connected
		c.HasAPISecret = secretRef != nil && *secretRef != ""
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) MarkPartnerListed(ctx context.Context, practiceID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE invoicing.practice_connections
		SET partner_listed_at = now(), updated_at = now()
		WHERE practice_id = $1
		  AND status = 'active'
		  AND billit_party_id IS NOT NULL
		  AND TRIM(billit_party_id) <> ''`, practiceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		_, _, getErr := s.GetConnection(ctx, practiceID)
		if errors.Is(getErr, invoicing.ErrNotConnected) {
			return invoicing.ErrNotConnected
		}
		if getErr != nil {
			return getErr
		}
		return invoicing.ErrPartnerNotEligible
	}
	return nil
}

func (s *Store) CreateConnectState(ctx context.Context, state, practiceID, userID string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO invoicing.connect_states (state, practice_id, created_by, expires_at)
		VALUES ($1,$2,$3,$4)`, state, practiceID, userID, expiresAt)
	return err
}

// AssertConnectState verifies the OAuth-like state is valid without consuming it.
func (s *Store) AssertConnectState(ctx context.Context, state, practiceID string) error {
	var one int
	err := s.pool.QueryRow(ctx, `
		SELECT 1 FROM invoicing.connect_states
		WHERE state = $1 AND practice_id = $2 AND consumed_at IS NULL AND expires_at > now()`,
		state, practiceID).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoicing.ErrInvalidState
		}
		return err
	}
	return nil
}

func (s *Store) ConsumeConnectState(ctx context.Context, state, practiceID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE invoicing.connect_states
		SET consumed_at = now()
		WHERE state = $1 AND practice_id = $2 AND consumed_at IS NULL AND expires_at > now()`,
		state, practiceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return invoicing.ErrInvalidState
	}
	return nil
}

// ConsumeConnectStateAndUpsertConnection consumes the one-time state and persists the
// connection in the same transaction (avoids burning state if Upsert fails).
func (s *Store) ConsumeConnectStateAndUpsertConnection(ctx context.Context, state string, c invoicing.Connection, apiSecretRef string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE invoicing.connect_states
		SET consumed_at = now()
		WHERE state = $1 AND practice_id = $2 AND consumed_at IS NULL AND expires_at > now()`,
		state, c.PracticeID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return invoicing.ErrInvalidState
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO invoicing.practice_connections (
			practice_id, billit_party_id, status, invoice_to_partner, partner_listed_at,
			docs_included_monthly, api_secret_ref, last_error, connected_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9, now())
		ON CONFLICT (practice_id) DO UPDATE SET
			billit_party_id = EXCLUDED.billit_party_id,
			status = EXCLUDED.status,
			invoice_to_partner = EXCLUDED.invoice_to_partner,
			partner_listed_at = COALESCE(EXCLUDED.partner_listed_at, invoicing.practice_connections.partner_listed_at),
			docs_included_monthly = EXCLUDED.docs_included_monthly,
			api_secret_ref = CASE WHEN EXCLUDED.api_secret_ref = '' THEN invoicing.practice_connections.api_secret_ref ELSE EXCLUDED.api_secret_ref END,
			last_error = EXCLUDED.last_error,
			connected_at = COALESCE(EXCLUDED.connected_at, invoicing.practice_connections.connected_at),
			updated_at = now()`,
		c.PracticeID, nullIfEmpty(c.BillitPartyID), string(c.Status), c.InvoiceToPartner, c.PartnerListedAt,
		c.DocsIncludedMonthly, apiSecretRef, nullIfEmpty(c.LastError), c.ConnectedAt,
	)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// MarkStaleSendingDocuments rejects practice Peppol docs stuck in sending past cutoff (frees quota slots).
// Flux A (saas_master) is ops-owned — never auto-rejected here.
func (s *Store) MarkStaleSendingDocuments(ctx context.Context, cutoff time.Time, limit int) (int, error) {
	if limit <= 0 {
		limit = 200
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET status = 'rejected',
		    peppol_status = 'stale_timeout',
		    updated_at = now()
		WHERE id IN (
			SELECT id FROM invoicing.documents
			WHERE status = 'sending'
			  AND type IN ('invoice', 'credit_note')
			  AND COALESCE(source, 'practice') = 'practice'
			  AND updated_at < $1
			ORDER BY updated_at
			LIMIT $2
		)`, cutoff, limit)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (s *Store) CreateDocument(ctx context.Context, doc invoicing.Document, lines []invoicing.Line) (invoicing.Document, error) {
	cp, err := json.Marshal(doc.Counterparty)
	if err != nil {
		return invoicing.Document{}, err
	}
	var related any
	if doc.RelatedDocumentID != "" {
		related = doc.RelatedDocumentID
	}
	var visitID any
	if doc.VisitID != "" {
		visitID = doc.VisitID
	}
	var dafID any
	if doc.DAFID != "" {
		dafID = doc.DAFID
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return invoicing.Document{}, err
	}
	defer tx.Rollback(ctx)

	source := string(doc.Source)
	if source == "" {
		source = string(invoicing.SourcePractice)
	}
	tag, err := tx.Exec(ctx, `
		INSERT INTO invoicing.documents (
			id, practice_id, type, status, source, number, billit_order_id, idempotency_key,
			counterparty_json, currency, total_excl_cents, total_vat_cents, total_incl_cents,
			related_document_id, visit_id, daf_id, created_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		ON CONFLICT (practice_id, idempotency_key) DO NOTHING`,
		doc.ID, doc.PracticeID, string(doc.Type), string(doc.Status), source, nullIfEmpty(doc.Number), nullIfEmpty(doc.BillitOrderID),
		doc.IdempotencyKey, cp, doc.Currency, doc.TotalExclCents, doc.TotalVATCents, doc.TotalInclCents,
		related, visitID, dafID, nullIfEmpty(doc.CreatedBy), doc.CreatedAt, doc.UpdatedAt,
	)
	if err != nil {
		return invoicing.Document{}, err
	}
	if tag.RowsAffected() == 0 {
		if err := tx.Commit(ctx); err != nil {
			return invoicing.Document{}, err
		}
		return s.GetDocumentByIdempotency(ctx, doc.PracticeID, doc.IdempotencyKey)
	}

	for i, l := range lines {
		meta, _ := json.Marshal(map[string]any{})
		if _, err := tx.Exec(ctx, `
			INSERT INTO invoicing.document_lines (
				id, document_id, position, description, quantity, unit_price_excl_cents, vat_percent, meta_json
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			uuid.NewString(), doc.ID, i+1, l.Description, l.Quantity, l.UnitPriceExclCents, l.VATPercent, meta,
		); err != nil {
			return invoicing.Document{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return invoicing.Document{}, err
	}
	return s.GetDocument(ctx, doc.PracticeID, doc.ID)
}

func (s *Store) ClaimDocumentForSend(ctx context.Context, practiceID, docID string, yyyymm, quotaLimit int) (invoicing.Document, invoicing.DocStatus, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return invoicing.Document{}, "", err
	}
	defer tx.Rollback(ctx)

	// Serialize Peppol sends per practice (quota check + claim).
	var lock int
	err = tx.QueryRow(ctx, `
		SELECT 1 FROM invoicing.practice_connections
		WHERE practice_id = $1 FOR UPDATE`, practiceID).Scan(&lock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoicing.Document{}, "", invoicing.ErrNotConnected
		}
		return invoicing.Document{}, "", err
	}

	var typ, status string
	var source string
	err = tx.QueryRow(ctx, `
		SELECT type, status, COALESCE(source, 'practice') FROM invoicing.documents
		WHERE id = $1 AND practice_id = $2 FOR UPDATE`, docID, practiceID).Scan(&typ, &status, &source)
	if err == nil && source == string(invoicing.SourceSaasMaster) {
		return invoicing.Document{}, invoicing.DocStatus(status), invoicing.ErrDocNotDraft
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoicing.Document{}, "", invoicing.ErrDocNotFound
		}
		return invoicing.Document{}, "", err
	}

	docType := invoicing.DocType(typ)
	prevStatus := invoicing.DocStatus(status)
	switch {
	case docType == invoicing.DocProforma && status == string(invoicing.StatusDraft):
		// ok
	case (docType == invoicing.DocInvoice || docType == invoicing.DocCreditNote) &&
		(status == string(invoicing.StatusDraft) || status == string(invoicing.StatusRejected)):
		// ok — rejected allows Peppol retry; issued is not a Peppol sendable state
	case status == string(invoicing.StatusSending):
		return invoicing.Document{}, prevStatus, invoicing.ErrSendInProgress
	default:
		return invoicing.Document{}, prevStatus, invoicing.ErrDocNotDraft
	}

	if invoicing.PeppolRequired(docType) && quotaLimit > 0 {
		var usage int
		err = tx.QueryRow(ctx, `
			SELECT doc_count FROM invoicing.usage_monthly
			WHERE practice_id = $1 AND yyyymm = $2
			FOR UPDATE`, practiceID, yyyymm).Scan(&usage)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return invoicing.Document{}, prevStatus, err
			}
			usage = 0
			if _, err := tx.Exec(ctx, `
				INSERT INTO invoicing.usage_monthly (practice_id, yyyymm, doc_count)
				VALUES ($1, $2, 0)
				ON CONFLICT (practice_id, yyyymm) DO NOTHING`, practiceID, yyyymm); err != nil {
				return invoicing.Document{}, prevStatus, err
			}
			err = tx.QueryRow(ctx, `
				SELECT doc_count FROM invoicing.usage_monthly
				WHERE practice_id = $1 AND yyyymm = $2
				FOR UPDATE`, practiceID, yyyymm).Scan(&usage)
			if err != nil {
				return invoicing.Document{}, prevStatus, err
			}
		}
		// Live hole fix: in-flight Peppol sends occupy a quota slot until delivered or rejected.
		// Flux A (saas_master) must never occupy practice quota slots.
		var sending int
		err = tx.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM invoicing.documents
			WHERE practice_id = $1
			  AND type IN ('invoice', 'credit_note')
			  AND status = 'sending'
			  AND COALESCE(source, 'practice') = 'practice'
			  AND id <> $2`, practiceID, docID).Scan(&sending)
		if err != nil {
			return invoicing.Document{}, prevStatus, err
		}
		if usage+sending >= quotaLimit {
			return invoicing.Document{}, prevStatus, invoicing.ErrDocsQuotaExceeded
		}
	}

	tag, err := tx.Exec(ctx, `
		UPDATE invoicing.documents
		SET status = 'sending',
		    peppol_status = CASE WHEN status = 'rejected' THEN NULL ELSE peppol_status END,
		    updated_at = now()
		WHERE id = $1 AND practice_id = $2`, docID, practiceID)
	if err != nil {
		return invoicing.Document{}, prevStatus, err
	}
	if tag.RowsAffected() == 0 {
		return invoicing.Document{}, prevStatus, invoicing.ErrDocNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return invoicing.Document{}, prevStatus, err
	}
	doc, err := s.GetDocument(ctx, practiceID, docID)
	return doc, prevStatus, err
}

func (s *Store) GetDocumentByIdempotency(ctx context.Context, practiceID, key string) (invoicing.Document, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM invoicing.documents WHERE practice_id = $1 AND idempotency_key = $2`,
		practiceID, key).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoicing.Document{}, invoicing.ErrDocNotFound
		}
		return invoicing.Document{}, err
	}
	return s.GetDocument(ctx, practiceID, id)
}

func (s *Store) GetDocument(ctx context.Context, practiceID, docID string) (invoicing.Document, error) {
	var doc invoicing.Document
	var typ, status string
	var number, orderID, related, visitID, dafID, peppol, createdBy *string
	var cp []byte
	var sentAt *time.Time
	var source string
	err := s.pool.QueryRow(ctx, `
		SELECT id, practice_id, type, status, COALESCE(source, 'practice'), number, billit_order_id, idempotency_key,
		       counterparty_json, currency, total_excl_cents, total_vat_cents, total_incl_cents,
		       related_document_id, visit_id, daf_id, peppol_status, sent_at, created_by, created_at, updated_at
		FROM invoicing.documents WHERE id = $1 AND practice_id = $2`, docID, practiceID).Scan(
		&doc.ID, &doc.PracticeID, &typ, &status, &source, &number, &orderID, &doc.IdempotencyKey,
		&cp, &doc.Currency, &doc.TotalExclCents, &doc.TotalVATCents, &doc.TotalInclCents,
		&related, &visitID, &dafID, &peppol, &sentAt, &createdBy, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoicing.Document{}, invoicing.ErrDocNotFound
		}
		return invoicing.Document{}, err
	}
	doc.Type = invoicing.DocType(typ)
	doc.Status = invoicing.DocStatus(status)
	doc.Source = invoicing.DocSource(source)
	if number != nil {
		doc.Number = *number
	}
	if orderID != nil {
		doc.BillitOrderID = *orderID
	}
	if related != nil {
		doc.RelatedDocumentID = *related
	}
	if visitID != nil {
		doc.VisitID = *visitID
	}
	if dafID != nil {
		doc.DAFID = *dafID
	}
	if peppol != nil {
		doc.PeppolStatus = *peppol
	}
	if createdBy != nil {
		doc.CreatedBy = *createdBy
	}
	doc.SentAt = sentAt
	_ = json.Unmarshal(cp, &doc.Counterparty)

	rows, err := s.pool.Query(ctx, `
		SELECT description, quantity, unit_price_excl_cents, vat_percent
		FROM invoicing.document_lines WHERE document_id = $1 ORDER BY position`, docID)
	if err != nil {
		return doc, err
	}
	defer rows.Close()
	for rows.Next() {
		var l invoicing.Line
		if err := rows.Scan(&l.Description, &l.Quantity, &l.UnitPriceExclCents, &l.VATPercent); err != nil {
			return doc, err
		}
		doc.Lines = append(doc.Lines, l)
	}
	return doc, rows.Err()
}

func (s *Store) ListDocuments(ctx context.Context, practiceID string, limit int) ([]invoicing.Document, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id FROM invoicing.documents
		WHERE practice_id = $1
		  AND COALESCE(source, 'practice') = 'practice'
		ORDER BY created_at DESC
		LIMIT $2`, practiceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	out := make([]invoicing.Document, 0, len(ids))
	for _, id := range ids {
		d, err := s.GetDocument(ctx, practiceID, id)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (s *Store) UpdateDocumentExternal(ctx context.Context, practiceID, docID, orderID string, status invoicing.DocStatus, peppolStatus string, sentAt *time.Time) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET billit_order_id = COALESCE(NULLIF($3,''), billit_order_id),
		    status = $4,
		    peppol_status = NULLIF($5,''),
		    sent_at = COALESCE($6, sent_at),
		    updated_at = now()
		WHERE id = $1 AND practice_id = $2`,
		docID, practiceID, orderID, string(status), peppolStatus, sentAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return invoicing.ErrDocNotFound
	}
	return nil
}

// ApplyBillitWebhookStatus updates doc status by Billit order id and, on first delivered,
// increments monthly usage in the same transaction (retry-safe with webhook forget+503).
func (s *Store) ApplyBillitWebhookStatus(ctx context.Context, orderID string, status invoicing.DocStatus, peppolStatus string, sentAt *time.Time, yyyymm int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var docID, practiceID, prev, source string
	err = tx.QueryRow(ctx, `
		SELECT id::text, practice_id::text, status, COALESCE(source, 'practice')
		FROM invoicing.documents
		WHERE billit_order_id = $1
		FOR UPDATE`, orderID).Scan(&docID, &practiceID, &prev, &source)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoicing.ErrDocNotFound
		}
		return err
	}
	prevStatus := invoicing.DocStatus(prev)
	applyStatus := invoicing.ResolveWebhookApplyStatus(prevStatus, status)
	applyPeppol := peppolStatus
	applySent := sentAt
	if applyStatus != status {
		// Transition suppressed (monotonic) — keep peppol/sent as-is.
		applyPeppol = ""
		applySent = nil
	}
	tag, err := tx.Exec(ctx, `
		UPDATE invoicing.documents
		SET status = $2,
		    peppol_status = COALESCE(NULLIF($3,''), peppol_status),
		    sent_at = COALESCE($4, sent_at),
		    updated_at = now()
		WHERE id = $1`,
		docID, string(applyStatus), applyPeppol, applySent)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return invoicing.ErrDocNotFound
	}
	// Flux A (saas_master) must never consume the practice Peppol monthly quota.
	if applyStatus == invoicing.StatusDelivered &&
		prevStatus != invoicing.StatusDelivered &&
		source == string(invoicing.SourcePractice) {
		if _, err := tx.Exec(ctx, `
			INSERT INTO invoicing.usage_monthly (practice_id, yyyymm, doc_count)
			VALUES ($1::uuid, $2, 1)
			ON CONFLICT (practice_id, yyyymm)
			DO UPDATE SET doc_count = invoicing.usage_monthly.doc_count + 1`,
			practiceID, yyyymm); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ClaimEmptyBillitOrder serializes Billit order creation for a draft without an order id.
// Stale creating_order claims (>2 min) can be reclaimed after a crash mid-gateway.
func (s *Store) ClaimEmptyBillitOrder(ctx context.Context, practiceID, docID string) (claimed bool, err error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET peppol_status = 'creating_order', updated_at = now()
		WHERE id = $1 AND practice_id = $2
		  AND (billit_order_id IS NULL OR TRIM(billit_order_id) = '')
		  AND (
		    peppol_status IS NULL
		    OR peppol_status = ''
		    OR peppol_status = 'saas_draft'
		    OR (peppol_status = 'creating_order' AND updated_at < now() - interval '2 minutes')
		  )`, docID, practiceID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// ClaimSaasDocumentForSend claims a Flux A document for Peppol send (no practice connection / quota).
func (s *Store) ClaimSaasDocumentForSend(ctx context.Context, practiceID, docID string) (invoicing.Document, invoicing.DocStatus, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return invoicing.Document{}, "", err
	}
	defer tx.Rollback(ctx)

	var typ, status, source string
	err = tx.QueryRow(ctx, `
		SELECT type, status, COALESCE(source, 'practice')
		FROM invoicing.documents
		WHERE id = $1 AND practice_id = $2
		FOR UPDATE`, docID, practiceID).Scan(&typ, &status, &source)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoicing.Document{}, "", invoicing.ErrDocNotFound
		}
		return invoicing.Document{}, "", err
	}
	prevStatus := invoicing.DocStatus(status)
	if source != string(invoicing.SourceSaasMaster) {
		return invoicing.Document{}, prevStatus, invoicing.ErrDocNotDraft
	}
	if typ != string(invoicing.DocInvoice) {
		return invoicing.Document{}, prevStatus, invoicing.ErrPeppolNotForType
	}
	switch status {
	case string(invoicing.StatusDraft), string(invoicing.StatusRejected):
		// ok
	case string(invoicing.StatusSending):
		return invoicing.Document{}, prevStatus, invoicing.ErrSendInProgress
	default:
		return invoicing.Document{}, prevStatus, invoicing.ErrDocNotDraft
	}

	tag, err := tx.Exec(ctx, `
		UPDATE invoicing.documents
		SET status = 'sending',
		    peppol_status = CASE WHEN status = 'rejected' THEN NULL ELSE peppol_status END,
		    updated_at = now()
		WHERE id = $1 AND practice_id = $2`, docID, practiceID)
	if err != nil {
		return invoicing.Document{}, prevStatus, err
	}
	if tag.RowsAffected() == 0 {
		return invoicing.Document{}, prevStatus, invoicing.ErrDocNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return invoicing.Document{}, prevStatus, err
	}
	doc, err := s.GetDocument(ctx, practiceID, docID)
	return doc, prevStatus, err
}

func (s *Store) InsertWebhookEvent(ctx context.Context, provider, eventType, externalID string, payload []byte) (string, bool, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO invoicing.webhook_events (provider, event_type, external_id, payload)
		VALUES ($1, NULLIF($2,''), NULLIF($3,''), $4::jsonb)
		ON CONFLICT (provider, external_id, COALESCE(event_type, ''))
		WHERE external_id IS NOT NULL AND external_id <> ''
		DO NOTHING
		RETURNING id::text`,
		provider, eventType, externalID, string(payload),
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", true, nil
		}
		return "", false, err
	}
	return id, false, nil
}

func (s *Store) MarkWebhookProcessed(ctx context.Context, eventID string, processErr string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE invoicing.webhook_events
		SET processed_at = now(), error = NULLIF($2,'')
		WHERE id = $1`,
		eventID, processErr)
	return err
}

func (s *Store) DeleteWebhookEvent(ctx context.Context, eventID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM invoicing.webhook_events WHERE id = $1`, eventID)
	return err
}

// PurgeOldWebhookEvents deletes Billit webhook audit rows older than cutoff (payloads may hold PII).
func (s *Store) PurgeOldWebhookEvents(ctx context.Context, cutoff time.Time, limit int) (int, error) {
	if limit <= 0 {
		limit = 500
	}
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM invoicing.webhook_events
		WHERE id IN (
			SELECT id FROM invoicing.webhook_events
			WHERE received_at < $1
			ORDER BY received_at
			LIMIT $2
		)`, cutoff, limit)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (s *Store) IncrementUsage(ctx context.Context, practiceID string, yyyymm int) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO invoicing.usage_monthly (practice_id, yyyymm, doc_count)
		VALUES ($1,$2,1)
		ON CONFLICT (practice_id, yyyymm) DO UPDATE SET doc_count = invoicing.usage_monthly.doc_count + 1`,
		practiceID, yyyymm)
	return err
}

func (s *Store) UsageForMonth(ctx context.Context, practiceID string, yyyymm int) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT doc_count FROM invoicing.usage_monthly WHERE practice_id = $1 AND yyyymm = $2`,
		practiceID, yyyymm).Scan(&n)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return n, nil
}

// ListSaasTargets returns active BE practices with a complete fiscal profile (Flux A eligible).
// Active = profile_completed_at + verified staff. optedInOnly filters saas_billing_enabled / allowlist.
func (s *Store) ListSaasTargets(ctx context.Context, yyyymm, limit, offset int, optedInOnly bool, allowlist []string) ([]invoicing.SaasTarget, error) {
	if limit <= 0 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	if allowlist == nil {
		allowlist = []string{}
	}
	yyyymmKey := strconv.Itoa(yyyymm)
	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text,
		       COALESCE(NULLIF(TRIM(p.company_legal_name), ''), NULLIF(TRIM(p.name), ''), ''),
		       COALESCE(p.contact_email, ''),
		       COALESCE(p.vat_number, ''),
		       (c.practice_id IS NOT NULL) AS has_connect,
		       (p.saas_billing_enabled OR p.id = ANY($4::uuid[])) AS saas_billing_enabled,
		       d.id::text, d.status, d.billit_order_id, d.total_incl_cents, d.peppol_status
		FROM practice.practices p
		LEFT JOIN invoicing.practice_connections c ON c.practice_id = p.id
		LEFT JOIN invoicing.documents d
		  ON d.practice_id = p.id
		 AND d.source = 'saas_master'
		 AND d.idempotency_key = 'saas:' || p.id::text || ':' || $1
		WHERE p.profile_completed_at IS NOT NULL
		  AND TRIM(COALESCE(p.company_legal_name, '')) <> ''
		  AND TRIM(COALESCE(p.vat_number, '')) <> ''
		  AND TRIM(COALESCE(p.company_number, '')) <> ''
		  AND TRIM(COALESCE(p.contact_email, '')) <> ''
		  AND UPPER(REPLACE(p.vat_number, ' ', '')) LIKE 'BE%'
		  AND EXISTS (
		    SELECT 1 FROM identity.users u
		    WHERE u.practice_id = p.id
		      AND u.role IN ('vet', 'vet_assistant', 'secretary')
		      AND u.email_verified_at IS NOT NULL
		  )
		  AND (
		    NOT $5::bool
		    OR p.saas_billing_enabled
		    OR p.id = ANY($4::uuid[])
		  )
		ORDER BY COALESCE(NULLIF(TRIM(p.company_legal_name), ''), p.name)
		LIMIT $2 OFFSET $3`, yyyymmKey, limit, offset, allowlist, optedInOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []invoicing.SaasTarget
	for rows.Next() {
		var t invoicing.SaasTarget
		var docID, status, orderID, peppol *string
		var total *int64
		if err := rows.Scan(
			&t.PracticeID, &t.PracticeName, &t.ContactEmail, &t.VATNumber, &t.HasBillitConnect,
			&t.SaasBillingEnabled,
			&docID, &status, &orderID, &total, &peppol,
		); err != nil {
			return nil, err
		}
		if docID != nil && *docID != "" {
			sum := &invoicing.SaasDocSummary{ID: *docID}
			if status != nil {
				sum.Status = invoicing.DocStatus(*status)
			}
			if orderID != nil {
				sum.BillitOrderID = *orderID
			}
			if total != nil {
				sum.TotalInclCents = *total
			}
			if peppol != nil {
				sum.PeppolStatus = *peppol
			}
			t.SaasDocument = sum
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// IsSaasEligible reports fiscal/active BE readiness and whether Flux A billing is opted-in.
func (s *Store) IsSaasEligible(ctx context.Context, practiceID string, allowlist []string) (eligible, billingEnabled bool, err error) {
	if allowlist == nil {
		allowlist = []string{}
	}
	err = s.pool.QueryRow(ctx, `
		SELECT
		  (
		    p.profile_completed_at IS NOT NULL
		    AND TRIM(COALESCE(p.company_legal_name, '')) <> ''
		    AND TRIM(COALESCE(p.vat_number, '')) <> ''
		    AND TRIM(COALESCE(p.company_number, '')) <> ''
		    AND TRIM(COALESCE(p.contact_email, '')) <> ''
		    AND UPPER(REPLACE(p.vat_number, ' ', '')) LIKE 'BE%'
		    AND EXISTS (
		      SELECT 1 FROM identity.users u
		      WHERE u.practice_id = p.id
		        AND u.role IN ('vet', 'vet_assistant', 'secretary')
		        AND u.email_verified_at IS NOT NULL
		    )
		  ) AS eligible,
		  (p.saas_billing_enabled OR p.id = ANY($2::uuid[])) AS billing_enabled
		FROM practice.practices p
		WHERE p.id = $1`, practiceID, allowlist).Scan(&eligible, &billingEnabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, false, nil
		}
		return false, false, err
	}
	return eligible, billingEnabled, nil
}

func (s *Store) SetSaasBillingEnabled(ctx context.Context, practiceID string, enabled bool) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE practice.practices SET saas_billing_enabled = $2 WHERE id = $1`, practiceID, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return invoicing.ErrSaasNotEligible
	}
	return nil
}
