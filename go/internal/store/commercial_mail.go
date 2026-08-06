package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const CommercialMailDailyLimit = 30

type EmailTemplateCategory string

const (
	EmailTplIntro        EmailTemplateCategory = "intro"
	EmailTplRDV          EmailTemplateCategory = "rdv"
	EmailTplNurture      EmailTemplateCategory = "nurture"
	EmailTplPost         EmailTemplateCategory = "post"
	EmailTplReactivation EmailTemplateCategory = "reactivation"
)

func ValidEmailTemplateCategory(c string) bool {
	switch EmailTemplateCategory(c) {
	case EmailTplIntro, EmailTplRDV, EmailTplNurture, EmailTplPost, EmailTplReactivation:
		return true
	default:
		return false
	}
}

type EmailTemplate struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Locale    string    `json:"locale"`
	Subject   string    `json:"subject"`
	BodyHTML  string    `json:"bodyHtml"`
	IsActive  bool      `json:"isActive"`
	UpdatedBy string    `json:"updatedBy,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type EmailTemplateInput struct {
	Slug     string
	Name     string
	Category string
	Locale   string
	Subject  string
	BodyHTML string
	IsActive *bool
}

type EmailClick struct {
	ID         string     `json:"id"`
	SendID     string     `json:"sendId"`
	ClickToken string     `json:"clickToken,omitempty"`
	TargetURL  string     `json:"targetUrl"`
	ClickedAt  *time.Time `json:"clickedAt,omitempty"`
	ClickCount int        `json:"clickCount"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type EmailSend struct {
	ID               string       `json:"id"`
	ProspectID       string       `json:"prospectId"`
	TemplateID       string       `json:"templateId,omitempty"`
	TemplateName     string       `json:"templateName,omitempty"`
	CommercialUserID string       `json:"commercialUserId"`
	CommercialName   string       `json:"commercialName,omitempty"`
	ToEmail          string       `json:"toEmail"`
	Subject          string       `json:"subject"`
	BodyHTMLRendered string       `json:"bodyHtmlRendered,omitempty"`
	Status           string       `json:"status"`
	Error            string       `json:"error,omitempty"`
	OpenToken        string       `json:"-"`
	OpenedAt         *time.Time   `json:"openedAt,omitempty"`
	OpenCount        int          `json:"openCount"`
	SentAt           *time.Time   `json:"sentAt,omitempty"`
	CreatedAt        time.Time    `json:"createdAt"`
	PracticeName     string       `json:"practiceName,omitempty"`
	ContactName      string       `json:"contactName,omitempty"`
	Clicks           []EmailClick `json:"clicks,omitempty"`
	ClickTotal       int          `json:"clickTotal"`
}

type EmailSendCreate struct {
	ProspectID       string
	TemplateID       string
	CommercialUserID string
	ToEmail          string
	Subject          string
	BodyHTMLRendered string
	Status           string
	Error            string
	OpenToken        string
	Clicks           []EmailClickCreate
}

type EmailClickCreate struct {
	ClickToken string
	TargetURL  string
}

func (s *Store) ListEmailTemplates(ctx context.Context, activeOnly bool) ([]EmailTemplate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, slug, name, category, locale, subject, body_html, is_active,
			COALESCE(updated_by::text,''), created_at, updated_at
		FROM sales.email_templates
		WHERE ($1 = false OR is_active = true)
		ORDER BY category ASC, name ASC`, activeOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmailTemplate, 0)
	for rows.Next() {
		t, err := scanEmailTemplate(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) GetEmailTemplate(ctx context.Context, id string) (EmailTemplate, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, slug, name, category, locale, subject, body_html, is_active,
			COALESCE(updated_by::text,''), created_at, updated_at
		FROM sales.email_templates WHERE id=$1`, id)
	t, err := scanEmailTemplate(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return EmailTemplate{}, ErrNotFound
	}
	return t, err
}

func (s *Store) GetEmailTemplateBySlug(ctx context.Context, slug string) (EmailTemplate, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, slug, name, category, locale, subject, body_html, is_active,
			COALESCE(updated_by::text,''), created_at, updated_at
		FROM sales.email_templates WHERE slug=$1`, slug)
	t, err := scanEmailTemplate(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return EmailTemplate{}, ErrNotFound
	}
	return t, err
}

func (s *Store) CreateEmailTemplate(ctx context.Context, in EmailTemplateInput, updatedBy string) (EmailTemplate, error) {
	id := uuid.NewString()
	locale := strings.TrimSpace(in.Locale)
	if locale == "" {
		locale = "fr"
	}
	cat := strings.TrimSpace(in.Category)
	if cat == "" {
		cat = string(EmailTplIntro)
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	var updBy any
	if strings.TrimSpace(updatedBy) != "" {
		updBy = updatedBy
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO sales.email_templates (
			id, slug, name, category, locale, subject, body_html, is_active, updated_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id::text, slug, name, category, locale, subject, body_html, is_active,
			COALESCE(updated_by::text,''), created_at, updated_at`,
		id, strings.TrimSpace(in.Slug), strings.TrimSpace(in.Name), cat, locale,
		strings.TrimSpace(in.Subject), in.BodyHTML, active, updBy)
	t, err := scanEmailTemplate(row.Scan)
	if isUniqueViolation(err) {
		return EmailTemplate{}, ErrConflict
	}
	return t, err
}

func (s *Store) UpdateEmailTemplate(ctx context.Context, id string, in EmailTemplateInput, updatedBy string) (EmailTemplate, error) {
	existing, err := s.GetEmailTemplate(ctx, id)
	if err != nil {
		return EmailTemplate{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = existing.Name
	}
	cat := strings.TrimSpace(in.Category)
	if cat == "" {
		cat = existing.Category
	}
	locale := strings.TrimSpace(in.Locale)
	if locale == "" {
		locale = existing.Locale
	}
	subject := strings.TrimSpace(in.Subject)
	if subject == "" {
		subject = existing.Subject
	}
	body := in.BodyHTML
	if strings.TrimSpace(body) == "" {
		body = existing.BodyHTML
	}
	active := existing.IsActive
	if in.IsActive != nil {
		active = *in.IsActive
	}
	var updBy any
	if strings.TrimSpace(updatedBy) != "" {
		updBy = updatedBy
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE sales.email_templates SET
			name=$2, category=$3, locale=$4, subject=$5, body_html=$6, is_active=$7,
			updated_by=$8, updated_at=NOW()
		WHERE id=$1
		RETURNING id::text, slug, name, category, locale, subject, body_html, is_active,
			COALESCE(updated_by::text,''), created_at, updated_at`,
		id, name, cat, locale, subject, body, active, updBy)
	t, err := scanEmailTemplate(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return EmailTemplate{}, ErrNotFound
	}
	return t, err
}

// UpsertEmailTemplateBySlug seeds or refreshes a catalog template (idempotent).
func (s *Store) UpsertEmailTemplateBySlug(ctx context.Context, in EmailTemplateInput) (EmailTemplate, error) {
	existing, err := s.GetEmailTemplateBySlug(ctx, in.Slug)
	if err == nil {
		return s.UpdateEmailTemplate(ctx, existing.ID, in, "")
	}
	if !errors.Is(err, ErrNotFound) {
		return EmailTemplate{}, err
	}
	return s.CreateEmailTemplate(ctx, in, "")
}

func scanEmailTemplate(scan func(dest ...any) error) (EmailTemplate, error) {
	var t EmailTemplate
	err := scan(&t.ID, &t.Slug, &t.Name, &t.Category, &t.Locale, &t.Subject, &t.BodyHTML,
		&t.IsActive, &t.UpdatedBy, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (s *Store) CountCommercialMailsToday(ctx context.Context, commercialUserID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM sales.email_sends
		WHERE commercial_user_id=$1 AND created_at >= date_trunc('day', NOW())`, commercialUserID).Scan(&n)
	return n, err
}

func (s *Store) CreateEmailSend(ctx context.Context, in EmailSendCreate) (EmailSend, error) {
	id := uuid.NewString()
	status := in.Status
	if status == "" {
		status = "sent"
	}
	var tplID any
	if strings.TrimSpace(in.TemplateID) != "" {
		tplID = in.TemplateID
	}
	var sentAt any
	if status == "sent" {
		sentAt = time.Now().UTC()
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return EmailSend{}, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		INSERT INTO sales.email_sends (
			id, prospect_id, template_id, commercial_user_id, to_email, subject,
			body_html_rendered, status, error, open_token, sent_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id::text, prospect_id::text, COALESCE(template_id::text,''), commercial_user_id::text,
			to_email, subject, body_html_rendered, status, error, open_token,
			opened_at, open_count, sent_at, created_at`,
		id, in.ProspectID, tplID, in.CommercialUserID, in.ToEmail, in.Subject,
		in.BodyHTMLRendered, status, in.Error, in.OpenToken, sentAt)
	send, err := scanEmailSendCore(row.Scan)
	if err != nil {
		return EmailSend{}, err
	}

	clicks := make([]EmailClick, 0, len(in.Clicks))
	for _, c := range in.Clicks {
		cid := uuid.NewString()
		crow := tx.QueryRow(ctx, `
			INSERT INTO sales.email_clicks (id, send_id, click_token, target_url)
			VALUES ($1,$2,$3,$4)
			RETURNING id::text, send_id::text, click_token, target_url, clicked_at, click_count, created_at`,
			cid, id, c.ClickToken, c.TargetURL)
		ec, err := scanEmailClick(crow.Scan)
		if err != nil {
			return EmailSend{}, err
		}
		clicks = append(clicks, ec)
	}
	if err := tx.Commit(ctx); err != nil {
		return EmailSend{}, err
	}
	send.Clicks = clicks
	send.ClickTotal = 0
	return send, nil
}

func scanEmailSendCore(scan func(dest ...any) error) (EmailSend, error) {
	var e EmailSend
	var opened, sent *time.Time
	err := scan(&e.ID, &e.ProspectID, &e.TemplateID, &e.CommercialUserID, &e.ToEmail, &e.Subject,
		&e.BodyHTMLRendered, &e.Status, &e.Error, &e.OpenToken, &opened, &e.OpenCount, &sent, &e.CreatedAt)
	if err != nil {
		return EmailSend{}, err
	}
	e.OpenedAt = opened
	e.SentAt = sent
	return e, nil
}

func scanEmailClick(scan func(dest ...any) error) (EmailClick, error) {
	var c EmailClick
	var clicked *time.Time
	err := scan(&c.ID, &c.SendID, &c.ClickToken, &c.TargetURL, &clicked, &c.ClickCount, &c.CreatedAt)
	c.ClickedAt = clicked
	return c, err
}

func (s *Store) ListEmailSendsByProspect(ctx context.Context, prospectID string) ([]EmailSend, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT s.id::text, s.prospect_id::text, COALESCE(s.template_id::text,''), s.commercial_user_id::text,
			s.to_email, s.subject, s.status, s.error, s.opened_at, s.open_count, s.sent_at, s.created_at,
			COALESCE(t.name,''), COALESCE(u.full_name,''),
			COALESCE(p.practice_name,''), COALESCE(p.contact_name,''),
			COALESCE((SELECT SUM(c.click_count) FROM sales.email_clicks c WHERE c.send_id = s.id), 0)::int
		FROM sales.email_sends s
		LEFT JOIN sales.email_templates t ON t.id = s.template_id
		LEFT JOIN identity.users u ON u.id = s.commercial_user_id
		JOIN sales.prospects p ON p.id = s.prospect_id
		WHERE s.prospect_id=$1
		ORDER BY s.created_at DESC`, prospectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmailSendList(rows)
}

func (s *Store) ListEmailSends(ctx context.Context, commercialUserID, filterCommercialID string, limit, offset int) ([]EmailSend, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	ownerFilter := commercialUserID
	if strings.TrimSpace(filterCommercialID) != "" {
		ownerFilter = filterCommercialID
	}
	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM sales.email_sends WHERE commercial_user_id=$1`, ownerFilter).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT s.id::text, s.prospect_id::text, COALESCE(s.template_id::text,''), s.commercial_user_id::text,
			s.to_email, s.subject, s.status, s.error, s.opened_at, s.open_count, s.sent_at, s.created_at,
			COALESCE(t.name,''), COALESCE(u.full_name,''),
			COALESCE(p.practice_name,''), COALESCE(p.contact_name,''),
			COALESCE((SELECT SUM(c.click_count) FROM sales.email_clicks c WHERE c.send_id = s.id), 0)::int
		FROM sales.email_sends s
		LEFT JOIN sales.email_templates t ON t.id = s.template_id
		LEFT JOIN identity.users u ON u.id = s.commercial_user_id
		JOIN sales.prospects p ON p.id = s.prospect_id
		WHERE s.commercial_user_id=$1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3`, ownerFilter, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanEmailSendList(rows)
	return items, total, err
}

func scanEmailSendList(rows pgx.Rows) ([]EmailSend, error) {
	out := make([]EmailSend, 0)
	for rows.Next() {
		var e EmailSend
		var opened, sent *time.Time
		if err := rows.Scan(
			&e.ID, &e.ProspectID, &e.TemplateID, &e.CommercialUserID,
			&e.ToEmail, &e.Subject, &e.Status, &e.Error, &opened, &e.OpenCount, &sent, &e.CreatedAt,
			&e.TemplateName, &e.CommercialName, &e.PracticeName, &e.ContactName, &e.ClickTotal,
		); err != nil {
			return nil, err
		}
		e.OpenedAt = opened
		e.SentAt = sent
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) GetEmailSend(ctx context.Context, id string) (EmailSend, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT s.id::text, s.prospect_id::text, COALESCE(s.template_id::text,''), s.commercial_user_id::text,
			s.to_email, s.subject, s.body_html_rendered, s.status, s.error, s.open_token,
			s.opened_at, s.open_count, s.sent_at, s.created_at,
			COALESCE(t.name,''), COALESCE(u.full_name,''),
			COALESCE(p.practice_name,''), COALESCE(p.contact_name,'')
		FROM sales.email_sends s
		LEFT JOIN sales.email_templates t ON t.id = s.template_id
		LEFT JOIN identity.users u ON u.id = s.commercial_user_id
		JOIN sales.prospects p ON p.id = s.prospect_id
		WHERE s.id=$1`, id)
	var e EmailSend
	var opened, sent *time.Time
	err := row.Scan(
		&e.ID, &e.ProspectID, &e.TemplateID, &e.CommercialUserID,
		&e.ToEmail, &e.Subject, &e.BodyHTMLRendered, &e.Status, &e.Error, &e.OpenToken,
		&opened, &e.OpenCount, &sent, &e.CreatedAt,
		&e.TemplateName, &e.CommercialName, &e.PracticeName, &e.ContactName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return EmailSend{}, ErrNotFound
	}
	if err != nil {
		return EmailSend{}, err
	}
	e.OpenedAt = opened
	e.SentAt = sent

	crows, err := s.pool.Query(ctx, `
		SELECT id::text, send_id::text, click_token, target_url, clicked_at, click_count, created_at
		FROM sales.email_clicks WHERE send_id=$1 ORDER BY created_at ASC`, id)
	if err != nil {
		return EmailSend{}, err
	}
	defer crows.Close()
	clicks := make([]EmailClick, 0)
	total := 0
	for crows.Next() {
		c, err := scanEmailClick(crows.Scan)
		if err != nil {
			return EmailSend{}, err
		}
		c.ClickToken = "" // hide public token in authenticated detail
		clicks = append(clicks, c)
		total += c.ClickCount
	}
	e.Clicks = clicks
	e.ClickTotal = total
	e.OpenToken = ""
	return e, crows.Err()
}

func (s *Store) RecordEmailOpen(ctx context.Context, openToken string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE sales.email_sends SET
			opened_at = COALESCE(opened_at, NOW()),
			open_count = open_count + 1
		WHERE open_token=$1`, openToken)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) RecordEmailClick(ctx context.Context, clickToken string) (string, error) {
	var target string
	err := s.pool.QueryRow(ctx, `
		UPDATE sales.email_clicks SET
			clicked_at = COALESCE(clicked_at, NOW()),
			click_count = click_count + 1
		WHERE click_token=$1
		RETURNING target_url`, clickToken).Scan(&target)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return target, err
}

func (s *Store) SetProspectEmailOptOutByOpenToken(ctx context.Context, openToken string) (string, error) {
	var prospectID string
	err := s.pool.QueryRow(ctx, `
		UPDATE sales.prospects p SET email_opt_out = true, updated_at = NOW()
		FROM sales.email_sends s
		WHERE s.open_token=$1 AND s.prospect_id = p.id
		RETURNING p.id::text`, openToken).Scan(&prospectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return prospectID, err
}

func (s *Store) TouchProspectContacted(ctx context.Context, prospectID, actorUserID string) error {
	var prevStatus string
	err := s.pool.QueryRow(ctx, `SELECT status FROM sales.prospects WHERE id=$1`, prospectID).Scan(&prevStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE sales.prospects SET
			last_contacted_at = NOW(),
			first_contacted_at = COALESCE(first_contacted_at, NOW()),
			status = CASE WHEN status = 'new' THEN 'contacted' ELSE status END,
			status_changed_at = CASE WHEN status = 'new' THEN NOW() ELSE status_changed_at END,
			updated_at = NOW()
		WHERE id=$1`, prospectID)
	if err != nil {
		return err
	}
	if prevStatus == "new" {
		_, _ = s.CreateProspectEvent(ctx, prospectID, actorUserID, string(EventStatusChange),
			"Statut : new → contacted", map[string]any{"from": "new", "to": "contacted", "auto": true})
	}
	return nil
}

func (s *Store) IsManagerOfCommercial(ctx context.Context, managerUserID, commercialUserID string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM identity.users
			WHERE id=$1 AND role='commercial' AND manager_user_id=$2
		)`, commercialUserID, managerUserID).Scan(&ok)
	return ok, err
}
