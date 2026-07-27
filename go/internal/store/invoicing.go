package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
)

func (s *Store) GetPracticeInvoicingProfile(ctx context.Context, practiceID string) (invoicing.PracticeParty, bool, error) {
	var p invoicing.PracticeParty
	var legal, vat, company, email string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(company_legal_name,''), COALESCE(vat_number,''), COALESCE(company_number,''),
		       COALESCE(NULLIF(contact_email,''), '')
		FROM practice.practices WHERE id = $1`, practiceID).Scan(&legal, &vat, &company, &email)
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
	}
	complete := strings.TrimSpace(legal) != "" &&
		strings.TrimSpace(vat) != "" &&
		strings.TrimSpace(company) != ""
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

func (s *Store) ListConnections(ctx context.Context) ([]invoicing.Connection, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT practice_id, billit_party_id, status, invoice_to_partner, partner_listed_at,
		       docs_included_monthly, api_secret_ref, last_error, connected_at, updated_at
		FROM invoicing.practice_connections
		ORDER BY updated_at DESC`)
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
			&c.PracticeID, &party, &status, &c.InvoiceToPartner, &partnerListed,
			&c.DocsIncludedMonthly, &secretRef, &lastErr, &connected, &c.UpdatedAt,
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
		WHERE practice_id = $1`, practiceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return invoicing.ErrNotConnected
	}
	return nil
}

func (s *Store) CreateConnectState(ctx context.Context, state, practiceID, userID string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO invoicing.connect_states (state, practice_id, created_by, expires_at)
		VALUES ($1,$2,$3,$4)`, state, practiceID, userID, expiresAt)
	return err
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

func (s *Store) CreateDocument(ctx context.Context, doc invoicing.Document, lines []invoicing.Line) (invoicing.Document, error) {
	cp, err := json.Marshal(doc.Counterparty)
	if err != nil {
		return invoicing.Document{}, err
	}
	var related any
	if doc.RelatedDocumentID != "" {
		related = doc.RelatedDocumentID
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return invoicing.Document{}, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		INSERT INTO invoicing.documents (
			id, practice_id, type, status, number, billit_order_id, idempotency_key,
			counterparty_json, currency, total_excl_cents, total_vat_cents, total_incl_cents,
			related_document_id, created_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (practice_id, idempotency_key) DO NOTHING`,
		doc.ID, doc.PracticeID, string(doc.Type), string(doc.Status), nullIfEmpty(doc.Number), nullIfEmpty(doc.BillitOrderID),
		doc.IdempotencyKey, cp, doc.Currency, doc.TotalExclCents, doc.TotalVATCents, doc.TotalInclCents,
		related, nullIfEmpty(doc.CreatedBy), doc.CreatedAt, doc.UpdatedAt,
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

func (s *Store) ClaimDocumentForSend(ctx context.Context, practiceID, docID string) (invoicing.Document, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE invoicing.documents
		SET status = 'sending', updated_at = now()
		WHERE id = $1 AND practice_id = $2 AND status IN ('draft', 'issued')`,
		docID, practiceID)
	if err != nil {
		return invoicing.Document{}, err
	}
	if tag.RowsAffected() == 0 {
		// Distinguish missing vs already sending/done
		var status string
		err := s.pool.QueryRow(ctx, `
			SELECT status FROM invoicing.documents WHERE id = $1 AND practice_id = $2`,
			docID, practiceID).Scan(&status)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return invoicing.Document{}, invoicing.ErrDocNotFound
			}
			return invoicing.Document{}, err
		}
		if status == string(invoicing.StatusSending) {
			return invoicing.Document{}, invoicing.ErrSendInProgress
		}
		return invoicing.Document{}, invoicing.ErrDocNotDraft
	}
	return s.GetDocument(ctx, practiceID, docID)
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
	var number, orderID, related, peppol, createdBy *string
	var cp []byte
	var sentAt *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT id, practice_id, type, status, number, billit_order_id, idempotency_key,
		       counterparty_json, currency, total_excl_cents, total_vat_cents, total_incl_cents,
		       related_document_id, peppol_status, sent_at, created_by, created_at, updated_at
		FROM invoicing.documents WHERE id = $1 AND practice_id = $2`, docID, practiceID).Scan(
		&doc.ID, &doc.PracticeID, &typ, &status, &number, &orderID, &doc.IdempotencyKey,
		&cp, &doc.Currency, &doc.TotalExclCents, &doc.TotalVATCents, &doc.TotalInclCents,
		&related, &peppol, &sentAt, &createdBy, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoicing.Document{}, invoicing.ErrDocNotFound
		}
		return invoicing.Document{}, err
	}
	doc.Type = invoicing.DocType(typ)
	doc.Status = invoicing.DocStatus(status)
	if number != nil {
		doc.Number = *number
	}
	if orderID != nil {
		doc.BillitOrderID = *orderID
	}
	if related != nil {
		doc.RelatedDocumentID = *related
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
