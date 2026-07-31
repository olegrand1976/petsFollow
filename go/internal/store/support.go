package store

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

const (
	SupportSourceNuxtPro         = "nuxt_pro"
	SupportSourceFlutterClient   = "flutter_client"
	SupportSourceFlutterProLight = "flutter_pro_light"
	SupportSourceSystem          = "system"

	SupportStatusOpen       = "open"
	SupportStatusInProgress = "in_progress"
	SupportStatusResolved   = "resolved"
	SupportStatusClosed     = "closed"

	MaxSupportSubjectLen      = 200
	MaxSupportMessageLen      = 8000
	MaxSupportReplyLen        = 8000
	MaxSupportDiagnosticsBytes = 512 * 1024
	SupportTicketsPerHour     = 5
)

var validSupportSources = map[string]bool{
	SupportSourceNuxtPro:         true,
	SupportSourceFlutterClient:   true,
	SupportSourceFlutterProLight: true,
	SupportSourceSystem:          true,
}

var validSupportStatuses = map[string]bool{
	SupportStatusOpen:       true,
	SupportStatusInProgress: true,
	SupportStatusResolved:   true,
	SupportStatusClosed:     true,
}

type SupportTicket struct {
	ID          string          `json:"id"`
	CreatedBy   *string         `json:"createdBy,omitempty"`
	CreatorEmail string         `json:"creatorEmail,omitempty"`
	CreatorName  string         `json:"creatorName,omitempty"`
	CreatorRole  string         `json:"creatorRole,omitempty"`
	Source      string          `json:"source"`
	Subject     string          `json:"subject"`
	Message     string          `json:"message"`
	Status      string          `json:"status"`
	Diagnostics json.RawMessage `json:"diagnostics,omitempty"`
	UserAgent   string          `json:"userAgent,omitempty"`
	AppVersion  string          `json:"appVersion,omitempty"`
	Locale      string          `json:"locale,omitempty"`
	Route       string          `json:"route,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	ReplyCount  int             `json:"replyCount,omitempty"`
}

type SupportTicketReply struct {
	ID         string    `json:"id"`
	TicketID   string    `json:"ticketId"`
	AuthorID   *string   `json:"authorId,omitempty"`
	AuthorName string    `json:"authorName,omitempty"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"createdAt"`
}

type SupportTicketDetail struct {
	SupportTicket
	Replies []SupportTicketReply `json:"replies"`
}

type CreateSupportTicketInput struct {
	CreatedBy   string
	Source      string
	Subject     string
	Message     string
	Diagnostics json.RawMessage
	UserAgent   string
	AppVersion  string
	Locale      string
	Route       string
}

func normalizeSupportSource(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

// ResolveSupportSource dérive la source depuis le rôle JWT (anti-spoof).
// Le client ne peut forcer qu'un sous-ensemble autorisé (véto Pro Light vs Pro web).
func ResolveSupportSource(role kernel.Role, requested string) string {
	req := normalizeSupportSource(requested)
	switch role {
	case kernel.RoleClient:
		return SupportSourceFlutterClient
	case kernel.RoleCarePro:
		return SupportSourceFlutterProLight
	case kernel.RoleVet:
		if req == SupportSourceFlutterProLight {
			return SupportSourceFlutterProLight
		}
		return SupportSourceNuxtPro
	default:
		return SupportSourceNuxtPro
	}
}

func (s *Store) CountRecentSupportTickets(ctx context.Context, userID string, since time.Time) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM ops.support_tickets
		WHERE created_by = $1 AND created_at >= $2`, userID, since).Scan(&n)
	return n, err
}

func (s *Store) CreateSupportTicket(ctx context.Context, in CreateSupportTicketInput) (SupportTicket, error) {
	source := normalizeSupportSource(in.Source)
	if !validSupportSources[source] {
		return SupportTicket{}, ErrValidation
	}
	subject := strings.TrimSpace(in.Subject)
	message := strings.TrimSpace(in.Message)
	if subject == "" || message == "" {
		return SupportTicket{}, ErrValidation
	}
	if utf8.RuneCountInString(subject) > MaxSupportSubjectLen {
		return SupportTicket{}, ErrValidation
	}
	if utf8.RuneCountInString(message) > MaxSupportMessageLen {
		return SupportTicket{}, ErrValidation
	}
	diag := in.Diagnostics
	if len(diag) == 0 {
		diag = json.RawMessage(`{}`)
	}
	if !json.Valid(diag) {
		return SupportTicket{}, ErrValidation
	}
	if len(diag) > MaxSupportDiagnosticsBytes {
		return SupportTicket{}, ErrDiagnosticsTooLarge
	}

	var createdByArg any
	if strings.TrimSpace(in.CreatedBy) != "" {
		createdByArg = strings.TrimSpace(in.CreatedBy)
	}

	var t SupportTicket
	var createdBy *string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO ops.support_tickets (
			created_by, source, subject, message, status, diagnostics,
			user_agent, app_version, locale, route
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10)
		RETURNING id, created_by, source, subject, message, status, diagnostics,
			user_agent, app_version, locale, route, created_at, updated_at`,
		createdByArg, source, subject, message, SupportStatusOpen, []byte(diag),
		strings.TrimSpace(in.UserAgent), strings.TrimSpace(in.AppVersion),
		strings.TrimSpace(in.Locale), strings.TrimSpace(in.Route),
	).Scan(
		&t.ID, &createdBy, &t.Source, &t.Subject, &t.Message, &t.Status, &t.Diagnostics,
		&t.UserAgent, &t.AppVersion, &t.Locale, &t.Route, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return SupportTicket{}, err
	}
	t.CreatedBy = createdBy
	return t, nil
}

const MaxSupportSearchLen = 100

func (s *Store) ListSupportTickets(ctx context.Context, status, source, search string, limit, offset int) ([]SupportTicket, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	status = strings.TrimSpace(status)
	source = normalizeSupportSource(source)
	search = strings.TrimSpace(search)
	if utf8.RuneCountInString(search) > MaxSupportSearchLen {
		runes := []rune(search)
		search = string(runes[:MaxSupportSearchLen])
	}

	args := []any{}
	var conds []string
	if status != "" {
		if !validSupportStatuses[status] {
			return nil, 0, ErrValidation
		}
		args = append(args, status)
		conds = append(conds, "t.status = $"+strconv.Itoa(len(args)))
	}
	if source != "" {
		if !validSupportSources[source] {
			return nil, 0, ErrValidation
		}
		args = append(args, source)
		conds = append(conds, "t.source = $"+strconv.Itoa(len(args)))
	}
	if search != "" {
		args = append(args, "%"+escapeILIKE(search)+"%")
		idx := strconv.Itoa(len(args))
		conds = append(conds, "(t.subject ILIKE $"+idx+" ESCAPE '\\' OR COALESCE(u.email, '') ILIKE $"+idx+" ESCAPE '\\' OR COALESCE(u.full_name, '') ILIKE $"+idx+" ESCAPE '\\')")
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	countQ := `
		SELECT COUNT(*) FROM ops.support_tickets t
		LEFT JOIN identity.users u ON u.id = t.created_by
		` + where
	var total int
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listArgs := append(append([]any{}, args...), limit, offset)
	limIdx := len(args) + 1
	offIdx := len(args) + 2
	listQ := `
		SELECT t.id, t.created_by, COALESCE(u.email, ''), COALESCE(u.full_name, ''), COALESCE(u.role::text, ''),
			t.source, t.subject, t.message, t.status,
			t.user_agent, t.app_version, t.locale, t.route, t.created_at, t.updated_at,
			(SELECT COUNT(*)::int FROM ops.support_ticket_replies r WHERE r.ticket_id = t.id)
		FROM ops.support_tickets t
		LEFT JOIN identity.users u ON u.id = t.created_by
		` + where + `
		ORDER BY t.created_at DESC
		LIMIT $` + strconv.Itoa(limIdx) + ` OFFSET $` + strconv.Itoa(offIdx)
	rows, err := s.pool.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []SupportTicket
	for rows.Next() {
		var t SupportTicket
		var createdBy *string
		if err := rows.Scan(
			&t.ID, &createdBy, &t.CreatorEmail, &t.CreatorName, &t.CreatorRole,
			&t.Source, &t.Subject, &t.Message, &t.Status,
			&t.UserAgent, &t.AppVersion, &t.Locale, &t.Route, &t.CreatedAt, &t.UpdatedAt,
			&t.ReplyCount,
		); err != nil {
			return nil, 0, err
		}
		t.CreatedBy = createdBy
		out = append(out, t)
	}
	if out == nil {
		out = []SupportTicket{}
	}
	return out, total, rows.Err()
}

func escapeILIKE(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(s)
}

func (s *Store) GetSupportTicket(ctx context.Context, id string) (SupportTicketDetail, error) {
	var t SupportTicket
	var createdBy *string
	err := s.pool.QueryRow(ctx, `
		SELECT t.id, t.created_by, COALESCE(u.email, ''), COALESCE(u.full_name, ''), COALESCE(u.role::text, ''),
			t.source, t.subject, t.message, t.status, t.diagnostics,
			t.user_agent, t.app_version, t.locale, t.route, t.created_at, t.updated_at
		FROM ops.support_tickets t
		LEFT JOIN identity.users u ON u.id = t.created_by
		WHERE t.id = $1`, id).Scan(
		&t.ID, &createdBy, &t.CreatorEmail, &t.CreatorName, &t.CreatorRole,
		&t.Source, &t.Subject, &t.Message, &t.Status, &t.Diagnostics,
		&t.UserAgent, &t.AppVersion, &t.Locale, &t.Route, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return SupportTicketDetail{}, ErrNotFound
		}
		return SupportTicketDetail{}, err
	}
	t.CreatedBy = createdBy

	rows, err := s.pool.Query(ctx, `
		SELECT r.id, r.ticket_id, r.author_id, COALESCE(u.full_name, ''), r.body, r.created_at
		FROM ops.support_ticket_replies r
		LEFT JOIN identity.users u ON u.id = r.author_id
		WHERE r.ticket_id = $1
		ORDER BY r.created_at ASC`, id)
	if err != nil {
		return SupportTicketDetail{}, err
	}
	defer rows.Close()

	replies := []SupportTicketReply{}
	for rows.Next() {
		var r SupportTicketReply
		var authorID *string
		if err := rows.Scan(&r.ID, &r.TicketID, &authorID, &r.AuthorName, &r.Body, &r.CreatedAt); err != nil {
			return SupportTicketDetail{}, err
		}
		r.AuthorID = authorID
		replies = append(replies, r)
	}
	if err := rows.Err(); err != nil {
		return SupportTicketDetail{}, err
	}
	return SupportTicketDetail{SupportTicket: t, Replies: replies}, nil
}

func (s *Store) UpdateSupportTicketStatus(ctx context.Context, id, status string) (SupportTicket, error) {
	status = strings.TrimSpace(status)
	if !validSupportStatuses[status] {
		return SupportTicket{}, ErrValidation
	}
	var t SupportTicket
	var createdBy *string
	err := s.pool.QueryRow(ctx, `
		UPDATE ops.support_tickets
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, created_by, source, subject, message, status,
			user_agent, app_version, locale, route, created_at, updated_at`,
		id, status,
	).Scan(
		&t.ID, &createdBy, &t.Source, &t.Subject, &t.Message, &t.Status,
		&t.UserAgent, &t.AppVersion, &t.Locale, &t.Route, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return SupportTicket{}, ErrNotFound
		}
		return SupportTicket{}, err
	}
	t.CreatedBy = createdBy
	return t, nil
}

func (s *Store) AddSupportTicketReply(ctx context.Context, ticketID, authorID, body string) (SupportTicketReply, SupportTicket, error) {
	body = strings.TrimSpace(body)
	if body == "" || utf8.RuneCountInString(body) > MaxSupportReplyLen {
		return SupportTicketReply{}, SupportTicket{}, ErrValidation
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return SupportTicketReply{}, SupportTicket{}, err
	}
	defer tx.Rollback(ctx)

	var t SupportTicket
	var createdBy *string
	err = tx.QueryRow(ctx, `
		SELECT id, created_by, source, subject, message, status,
			user_agent, app_version, locale, route, created_at, updated_at
		FROM ops.support_tickets WHERE id = $1 FOR UPDATE`, ticketID).Scan(
		&t.ID, &createdBy, &t.Source, &t.Subject, &t.Message, &t.Status,
		&t.UserAgent, &t.AppVersion, &t.Locale, &t.Route, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return SupportTicketReply{}, SupportTicket{}, ErrNotFound
		}
		return SupportTicketReply{}, SupportTicket{}, err
	}
	t.CreatedBy = createdBy

	newStatus := t.Status
	if t.Status == SupportStatusOpen {
		newStatus = SupportStatusInProgress
	}
	err = tx.QueryRow(ctx, `
		UPDATE ops.support_tickets
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING status, updated_at`, ticketID, newStatus).Scan(&t.Status, &t.UpdatedAt)
	if err != nil {
		return SupportTicketReply{}, SupportTicket{}, err
	}

	var reply SupportTicketReply
	var authorIDPtr *string
	err = tx.QueryRow(ctx, `
		INSERT INTO ops.support_ticket_replies (ticket_id, author_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, ticket_id, author_id, body, created_at`,
		ticketID, authorID, body,
	).Scan(&reply.ID, &reply.TicketID, &authorIDPtr, &reply.Body, &reply.CreatedAt)
	if err != nil {
		return SupportTicketReply{}, SupportTicket{}, err
	}
	reply.AuthorID = authorIDPtr

	if err := tx.Commit(ctx); err != nil {
		return SupportTicketReply{}, SupportTicket{}, err
	}
	return reply, t, nil
}

// AnonymizeUserSupportTickets clears PII on tickets (standalone, uses pool).
func (s *Store) AnonymizeUserSupportTickets(ctx context.Context, userID string) error {
	return anonymizeUserSupportTicketsExec(ctx, s.pool, userID)
}

type supportExecer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func anonymizeUserSupportTicketsExec(ctx context.Context, db supportExecer, userID string) error {
	_, err := db.Exec(ctx, `
		UPDATE ops.support_tickets SET
			created_by = NULL,
			subject = '[anonymized]',
			message = '[anonymized]',
			diagnostics = '{}'::jsonb,
			user_agent = '',
			app_version = '',
			locale = '',
			route = '',
			updated_at = NOW()
		WHERE created_by = $1`, userID)
	return err
}

// ListSupportTicketsForExport returns tickets + replies for RGPD export.
func (s *Store) ListSupportTicketsForExport(ctx context.Context, userID string) (json.RawMessage, error) {
	var raw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(jsonb_agg(row_to_json(x) ORDER BY x.created_at), '[]'::jsonb)
		FROM (
			SELECT t.id, t.source, t.subject, t.message, t.status, t.diagnostics,
				t.user_agent, t.app_version, t.locale, t.route,
				t.created_at, t.updated_at,
				COALESCE((
					SELECT jsonb_agg(jsonb_build_object(
						'id', r.id, 'body', r.body, 'createdAt', r.created_at
					) ORDER BY r.created_at)
					FROM ops.support_ticket_replies r WHERE r.ticket_id = t.id
				), '[]'::jsonb) AS replies
			FROM ops.support_tickets t
			WHERE t.created_by = $1
		) x`, userID).Scan(&raw)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

// SupportTicketStats — agrégats inbox ops (admin / DEV).
type SupportTicketStats struct {
	ByStatus       map[string]int `json:"byStatus"`
	BySource       map[string]int `json:"bySource"`
	OpenOlderThan24h int          `json:"openOlderThan24h"`
	OpenOlderThan7d  int          `json:"openOlderThan7d"`
	Total          int            `json:"total"`
}

func (s *Store) SupportTicketStats(ctx context.Context) (SupportTicketStats, error) {
	out := SupportTicketStats{
		ByStatus: map[string]int{},
		BySource: map[string]int{},
	}
	rows, err := s.pool.Query(ctx, `
		SELECT status, COUNT(*)::int FROM ops.support_tickets GROUP BY status`)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var st string
		var n int
		if err := rows.Scan(&st, &n); err != nil {
			return out, err
		}
		out.ByStatus[st] = n
		out.Total += n
	}
	if err := rows.Err(); err != nil {
		return out, err
	}

	srcRows, err := s.pool.Query(ctx, `
		SELECT source, COUNT(*)::int FROM ops.support_tickets GROUP BY source`)
	if err != nil {
		return out, err
	}
	defer srcRows.Close()
	for srcRows.Next() {
		var src string
		var n int
		if err := srcRows.Scan(&src, &n); err != nil {
			return out, err
		}
		out.BySource[src] = n
	}
	if err := srcRows.Err(); err != nil {
		return out, err
	}

	now := time.Now().UTC()
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM ops.support_tickets
		WHERE status IN ('open', 'in_progress') AND created_at < $1`,
		now.Add(-24*time.Hour)).Scan(&out.OpenOlderThan24h)
	if err != nil {
		return out, err
	}
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM ops.support_tickets
		WHERE status IN ('open', 'in_progress') AND created_at < $1`,
		now.Add(-7*24*time.Hour)).Scan(&out.OpenOlderThan7d)
	return out, err
}
