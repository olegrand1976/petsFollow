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
	SupportStatusToTest     = "to_test"
	SupportStatusDone       = "done"
	SupportStatusClosed     = "closed"

	MaxSupportSubjectLen       = 200
	MaxSupportMessageLen       = 8000
	MaxSupportReplyLen         = 8000
	MaxSupportDiagnosticsBytes = 512 * 1024
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
	SupportStatusToTest:     true,
	SupportStatusDone:       true,
	SupportStatusClosed:     true,
}

type SupportTicket struct {
	ID           string          `json:"id"`
	CreatedBy    *string         `json:"createdBy,omitempty"`
	CreatorEmail string          `json:"creatorEmail,omitempty"`
	CreatorName  string          `json:"creatorName,omitempty"`
	CreatorRole  string          `json:"creatorRole,omitempty"`
	Source       string          `json:"source"`
	Subject      string          `json:"subject"`
	Message      string          `json:"message"`
	Status       string          `json:"status"`
	Diagnostics  json.RawMessage `json:"diagnostics,omitempty"`
	UserAgent    string          `json:"userAgent,omitempty"`
	AppVersion   string          `json:"appVersion,omitempty"`
	Locale       string          `json:"locale,omitempty"`
	Route        string          `json:"route,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
	ReplyCount   int             `json:"replyCount,omitempty"`
}

type SupportTicketReply struct {
	ID         string    `json:"id"`
	TicketID   string    `json:"ticketId"`
	AuthorID   *string   `json:"authorId,omitempty"`
	AuthorName string    `json:"authorName,omitempty"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"createdAt"`
}

type SupportTicketStatusEvent struct {
	ID          string    `json:"id"`
	TicketID    string    `json:"ticketId"`
	FromStatus  *string   `json:"fromStatus"`
	ToStatus    string    `json:"toStatus"`
	ChangedBy   *string   `json:"changedBy,omitempty"`
	ChangedName string    `json:"changedName,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type SupportTicketAttachment struct {
	ID           string    `json:"id"`
	TicketID     string    `json:"ticketId"`
	UploadedBy   *string   `json:"uploadedBy,omitempty"`
	UploaderName string    `json:"uploaderName,omitempty"`
	FileName     string    `json:"fileName"`
	ContentType  string    `json:"contentType"`
	SizeBytes    int64     `json:"sizeBytes"`
	ObjectKey    string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type SupportTicketDetail struct {
	SupportTicket
	Replies       []SupportTicketReply       `json:"replies"`
	StatusHistory []SupportTicketStatusEvent `json:"statusHistory"`
	Attachments   []SupportTicketAttachment  `json:"attachments"`
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

type CreateSupportTicketAttachmentInput struct {
	TicketID    string
	UploadedBy  string
	FileName    string
	ContentType string
	SizeBytes   int64
	ObjectKey   string
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

func (s *Store) CreateSupportTicket(ctx context.Context, in CreateSupportTicketInput) (SupportTicket, error) {
	source := normalizeSupportSource(in.Source)
	subject := strings.TrimSpace(in.Subject)
	message := strings.TrimSpace(in.Message)
	if !validSupportSources[source] || subject == "" || message == "" ||
		utf8.RuneCountInString(subject) > MaxSupportSubjectLen ||
		utf8.RuneCountInString(message) > MaxSupportMessageLen {
		return SupportTicket{}, ErrValidation
	}
	diagnostics := in.Diagnostics
	if len(diagnostics) == 0 {
		diagnostics = json.RawMessage(`{}`)
	}
	if !json.Valid(diagnostics) {
		return SupportTicket{}, ErrValidation
	}
	if len(diagnostics) > MaxSupportDiagnosticsBytes {
		return SupportTicket{}, ErrDiagnosticsTooLarge
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return SupportTicket{}, err
	}
	defer tx.Rollback(ctx)

	var createdBy any
	if userID := strings.TrimSpace(in.CreatedBy); userID != "" {
		createdBy = userID
	}
	var ticket SupportTicket
	var ticketCreatedBy *string
	err = tx.QueryRow(ctx, `
		INSERT INTO ops.support_tickets (
			created_by, source, subject, message, status, diagnostics,
			user_agent, app_version, locale, route
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10)
		RETURNING id, created_by, source, subject, message, status, diagnostics,
			user_agent, app_version, locale, route, created_at, updated_at`,
		createdBy, source, subject, message, SupportStatusOpen, []byte(diagnostics),
		strings.TrimSpace(in.UserAgent), strings.TrimSpace(in.AppVersion),
		strings.TrimSpace(in.Locale), strings.TrimSpace(in.Route),
	).Scan(
		&ticket.ID, &ticketCreatedBy, &ticket.Source, &ticket.Subject, &ticket.Message,
		&ticket.Status, &ticket.Diagnostics, &ticket.UserAgent, &ticket.AppVersion,
		&ticket.Locale, &ticket.Route, &ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		return SupportTicket{}, err
	}
	ticket.CreatedBy = ticketCreatedBy
	if err := insertSupportStatusEventTx(ctx, tx, ticket.ID, nil, SupportStatusOpen, in.CreatedBy); err != nil {
		return SupportTicket{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return SupportTicket{}, err
	}
	return ticket, nil
}

func insertSupportStatusEventTx(ctx context.Context, tx pgx.Tx, ticketID string, fromStatus *string, toStatus, changedBy string) error {
	var changedByArg any
	if userID := strings.TrimSpace(changedBy); userID != "" {
		changedByArg = userID
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO ops.support_ticket_status_events (
			ticket_id, from_status, to_status, changed_by
		) VALUES ($1, $2, $3, $4)`,
		ticketID, fromStatus, toStatus, changedByArg,
	)
	return err
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
		search = string([]rune(search)[:MaxSupportSearchLen])
	}

	args := []any{}
	var conditions []string
	if status != "" {
		if !validSupportStatuses[status] {
			return nil, 0, ErrValidation
		}
		args = append(args, status)
		conditions = append(conditions, "t.status = $"+strconv.Itoa(len(args)))
	}
	if source != "" {
		if !validSupportSources[source] {
			return nil, 0, ErrValidation
		}
		args = append(args, source)
		conditions = append(conditions, "t.source = $"+strconv.Itoa(len(args)))
	}
	if search != "" {
		args = append(args, "%"+escapeILIKE(search)+"%")
		index := strconv.Itoa(len(args))
		conditions = append(conditions, "(t.subject ILIKE $"+index+" ESCAPE '\\' OR COALESCE(u.email, '') ILIKE $"+index+" ESCAPE '\\' OR COALESCE(u.full_name, '') ILIKE $"+index+" ESCAPE '\\')")
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM ops.support_tickets t
		LEFT JOIN identity.users u ON u.id = t.created_by `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listArgs := append(append([]any{}, args...), limit, offset)
	limitIndex := len(args) + 1
	offsetIndex := len(args) + 2
	rows, err := s.pool.Query(ctx, `
		SELECT t.id, t.created_by, COALESCE(u.email, ''), COALESCE(u.full_name, ''), COALESCE(u.role::text, ''),
			t.source, t.subject, t.message, t.status,
			t.user_agent, t.app_version, t.locale, t.route, t.created_at, t.updated_at,
			(SELECT COUNT(*)::int FROM ops.support_ticket_replies r WHERE r.ticket_id = t.id)
		FROM ops.support_tickets t
		LEFT JOIN identity.users u ON u.id = t.created_by `+where+`
		ORDER BY t.created_at DESC
		LIMIT $`+strconv.Itoa(limitIndex)+` OFFSET $`+strconv.Itoa(offsetIndex), listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tickets := []SupportTicket{}
	for rows.Next() {
		var ticket SupportTicket
		var createdBy *string
		if err := rows.Scan(
			&ticket.ID, &createdBy, &ticket.CreatorEmail, &ticket.CreatorName, &ticket.CreatorRole,
			&ticket.Source, &ticket.Subject, &ticket.Message, &ticket.Status,
			&ticket.UserAgent, &ticket.AppVersion, &ticket.Locale, &ticket.Route,
			&ticket.CreatedAt, &ticket.UpdatedAt, &ticket.ReplyCount,
		); err != nil {
			return nil, 0, err
		}
		ticket.CreatedBy = createdBy
		tickets = append(tickets, ticket)
	}
	return tickets, total, rows.Err()
}

func escapeILIKE(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (s *Store) GetSupportTicket(ctx context.Context, id string) (SupportTicketDetail, error) {
	var ticket SupportTicket
	var createdBy *string
	err := s.pool.QueryRow(ctx, `
		SELECT t.id, t.created_by, COALESCE(u.email, ''), COALESCE(u.full_name, ''), COALESCE(u.role::text, ''),
			t.source, t.subject, t.message, t.status, t.diagnostics,
			t.user_agent, t.app_version, t.locale, t.route, t.created_at, t.updated_at
		FROM ops.support_tickets t
		LEFT JOIN identity.users u ON u.id = t.created_by
		WHERE t.id = $1`, id).Scan(
		&ticket.ID, &createdBy, &ticket.CreatorEmail, &ticket.CreatorName, &ticket.CreatorRole,
		&ticket.Source, &ticket.Subject, &ticket.Message, &ticket.Status, &ticket.Diagnostics,
		&ticket.UserAgent, &ticket.AppVersion, &ticket.Locale, &ticket.Route,
		&ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return SupportTicketDetail{}, ErrNotFound
		}
		return SupportTicketDetail{}, err
	}
	ticket.CreatedBy = createdBy

	replies, err := s.listSupportTicketReplies(ctx, id)
	if err != nil {
		return SupportTicketDetail{}, err
	}
	history, err := s.listSupportTicketStatusEvents(ctx, id)
	if err != nil {
		return SupportTicketDetail{}, err
	}
	attachments, err := s.listSupportTicketAttachments(ctx, id)
	if err != nil {
		return SupportTicketDetail{}, err
	}
	return SupportTicketDetail{
		SupportTicket: ticket,
		Replies:       replies,
		StatusHistory: history,
		Attachments:   attachments,
	}, nil
}

func (s *Store) listSupportTicketReplies(ctx context.Context, ticketID string) ([]SupportTicketReply, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.id, r.ticket_id, r.author_id, COALESCE(u.full_name, ''), r.body, r.created_at
		FROM ops.support_ticket_replies r
		LEFT JOIN identity.users u ON u.id = r.author_id
		WHERE r.ticket_id = $1
		ORDER BY r.created_at ASC`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	replies := []SupportTicketReply{}
	for rows.Next() {
		var reply SupportTicketReply
		if err := rows.Scan(
			&reply.ID, &reply.TicketID, &reply.AuthorID, &reply.AuthorName, &reply.Body, &reply.CreatedAt,
		); err != nil {
			return nil, err
		}
		replies = append(replies, reply)
	}
	return replies, rows.Err()
}

func (s *Store) listSupportTicketStatusEvents(ctx context.Context, ticketID string) ([]SupportTicketStatusEvent, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.id, e.ticket_id, e.from_status, e.to_status, e.changed_by,
			COALESCE(u.full_name, ''), e.created_at
		FROM ops.support_ticket_status_events e
		LEFT JOIN identity.users u ON u.id = e.changed_by
		WHERE e.ticket_id = $1
		ORDER BY e.created_at ASC, e.id ASC`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []SupportTicketStatusEvent{}
	for rows.Next() {
		var event SupportTicketStatusEvent
		if err := rows.Scan(
			&event.ID, &event.TicketID, &event.FromStatus, &event.ToStatus,
			&event.ChangedBy, &event.ChangedName, &event.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) listSupportTicketAttachments(ctx context.Context, ticketID string) ([]SupportTicketAttachment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.ticket_id, a.uploaded_by, COALESCE(u.full_name, ''),
			a.file_name, a.content_type, a.size_bytes, a.object_key, a.created_at
		FROM ops.support_ticket_attachments a
		LEFT JOIN identity.users u ON u.id = a.uploaded_by
		WHERE a.ticket_id = $1
		ORDER BY a.created_at ASC, a.id ASC`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attachments := []SupportTicketAttachment{}
	for rows.Next() {
		var attachment SupportTicketAttachment
		if err := rows.Scan(
			&attachment.ID, &attachment.TicketID, &attachment.UploadedBy, &attachment.UploaderName,
			&attachment.FileName, &attachment.ContentType, &attachment.SizeBytes,
			&attachment.ObjectKey, &attachment.CreatedAt,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}
	return attachments, rows.Err()
}

// UpdateSupportTicketStatus records an effective status change. The optional form
// keeps legacy callers compiling while new callers must provide the actor ID.
// previous is the status before the update (equal to ticket.Status when unchanged).
func (s *Store) UpdateSupportTicketStatus(ctx context.Context, id, status string, changedBy ...string) (SupportTicket, string, error) {
	status = strings.TrimSpace(status)
	if !validSupportStatuses[status] {
		return SupportTicket{}, "", ErrValidation
	}
	actorID := ""
	if len(changedBy) > 0 {
		actorID = changedBy[0]
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return SupportTicket{}, "", err
	}
	defer tx.Rollback(ctx)

	var ticket SupportTicket
	var createdBy *string
	err = tx.QueryRow(ctx, `
		SELECT id, created_by, source, subject, message, status,
			user_agent, app_version, locale, route, created_at, updated_at
		FROM ops.support_tickets
		WHERE id = $1
		FOR UPDATE`, id).Scan(
		&ticket.ID, &createdBy, &ticket.Source, &ticket.Subject, &ticket.Message, &ticket.Status,
		&ticket.UserAgent, &ticket.AppVersion, &ticket.Locale, &ticket.Route,
		&ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return SupportTicket{}, "", ErrNotFound
		}
		return SupportTicket{}, "", err
	}
	ticket.CreatedBy = createdBy
	previous := ticket.Status
	if ticket.Status == status {
		if err := tx.Commit(ctx); err != nil {
			return SupportTicket{}, "", err
		}
		return ticket, previous, nil
	}

	if err := tx.QueryRow(ctx, `
		UPDATE ops.support_tickets
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING status, updated_at`, id, status,
	).Scan(&ticket.Status, &ticket.UpdatedAt); err != nil {
		return SupportTicket{}, "", err
	}
	if err := insertSupportStatusEventTx(ctx, tx, id, &previous, status, actorID); err != nil {
		return SupportTicket{}, "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return SupportTicket{}, "", err
	}
	return ticket, previous, nil
}

// AddSupportTicketReply appends an admin comment. previousStatus is the ticket
// status before the reply (open→in_progress may advance it).
func (s *Store) AddSupportTicketReply(ctx context.Context, ticketID, authorID, body string) (SupportTicketReply, SupportTicket, string, error) {
	body = strings.TrimSpace(body)
	if body == "" || utf8.RuneCountInString(body) > MaxSupportReplyLen {
		return SupportTicketReply{}, SupportTicket{}, "", ErrValidation
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return SupportTicketReply{}, SupportTicket{}, "", err
	}
	defer tx.Rollback(ctx)

	var ticket SupportTicket
	var createdBy *string
	err = tx.QueryRow(ctx, `
		SELECT id, created_by, source, subject, message, status,
			user_agent, app_version, locale, route, created_at, updated_at
		FROM ops.support_tickets
		WHERE id = $1
		FOR UPDATE`, ticketID).Scan(
		&ticket.ID, &createdBy, &ticket.Source, &ticket.Subject, &ticket.Message, &ticket.Status,
		&ticket.UserAgent, &ticket.AppVersion, &ticket.Locale, &ticket.Route,
		&ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return SupportTicketReply{}, SupportTicket{}, "", ErrNotFound
		}
		return SupportTicketReply{}, SupportTicket{}, "", err
	}
	ticket.CreatedBy = createdBy
	previousStatus := ticket.Status

	if ticket.Status == SupportStatusOpen {
		previous := ticket.Status
		if err := tx.QueryRow(ctx, `
			UPDATE ops.support_tickets
			SET status = $2, updated_at = NOW()
			WHERE id = $1
			RETURNING status, updated_at`, ticketID, SupportStatusInProgress,
		).Scan(&ticket.Status, &ticket.UpdatedAt); err != nil {
			return SupportTicketReply{}, SupportTicket{}, "", err
		}
		if err := insertSupportStatusEventTx(ctx, tx, ticketID, &previous, ticket.Status, authorID); err != nil {
			return SupportTicketReply{}, SupportTicket{}, "", err
		}
	}

	var reply SupportTicketReply
	if err := tx.QueryRow(ctx, `
		INSERT INTO ops.support_ticket_replies (ticket_id, author_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, ticket_id, author_id, body, created_at`,
		ticketID, authorID, body,
	).Scan(&reply.ID, &reply.TicketID, &reply.AuthorID, &reply.Body, &reply.CreatedAt); err != nil {
		return SupportTicketReply{}, SupportTicket{}, "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return SupportTicketReply{}, SupportTicket{}, "", err
	}
	return reply, ticket, previousStatus, nil
}

func (s *Store) CreateSupportTicketAttachment(ctx context.Context, in CreateSupportTicketAttachmentInput) (SupportTicketAttachment, error) {
	in.TicketID = strings.TrimSpace(in.TicketID)
	in.UploadedBy = strings.TrimSpace(in.UploadedBy)
	in.FileName = strings.TrimSpace(in.FileName)
	in.ContentType = strings.TrimSpace(in.ContentType)
	in.ObjectKey = strings.TrimSpace(in.ObjectKey)
	if in.TicketID == "" || in.FileName == "" || in.ObjectKey == "" || in.SizeBytes < 0 {
		return SupportTicketAttachment{}, ErrValidation
	}

	var attachment SupportTicketAttachment
	err := s.pool.QueryRow(ctx, `
		INSERT INTO ops.support_ticket_attachments (
			ticket_id, uploaded_by, file_name, content_type, size_bytes, object_key
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, ticket_id, uploaded_by, file_name, content_type, size_bytes, object_key, created_at`,
		in.TicketID, in.UploadedBy, in.FileName, in.ContentType, in.SizeBytes, in.ObjectKey,
	).Scan(
		&attachment.ID, &attachment.TicketID, &attachment.UploadedBy, &attachment.FileName,
		&attachment.ContentType, &attachment.SizeBytes, &attachment.ObjectKey, &attachment.CreatedAt,
	)
	if err != nil {
		return SupportTicketAttachment{}, err
	}
	return attachment, nil
}

func (s *Store) GetSupportTicketAttachment(ctx context.Context, id string) (SupportTicketAttachment, error) {
	var attachment SupportTicketAttachment
	err := s.pool.QueryRow(ctx, `
		SELECT a.id, a.ticket_id, a.uploaded_by, COALESCE(u.full_name, ''),
			a.file_name, a.content_type, a.size_bytes, a.object_key, a.created_at
		FROM ops.support_ticket_attachments a
		LEFT JOIN identity.users u ON u.id = a.uploaded_by
		WHERE a.id = $1`, id,
	).Scan(
		&attachment.ID, &attachment.TicketID, &attachment.UploadedBy, &attachment.UploaderName,
		&attachment.FileName, &attachment.ContentType, &attachment.SizeBytes,
		&attachment.ObjectKey, &attachment.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return SupportTicketAttachment{}, ErrNotFound
	}
	return attachment, err
}

func (s *Store) ListSupportAttachmentObjectKeysForUser(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT COALESCE(a.object_key, '')
		FROM ops.support_ticket_attachments a
		JOIN ops.support_tickets t ON t.id = a.ticket_id
		WHERE t.created_by = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := []string{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		if key = strings.TrimSpace(key); key != "" {
			keys = append(keys, key)
		}
	}
	return keys, rows.Err()
}

// AnonymizeUserSupportTickets clears ticket PII and all user references.
func (s *Store) AnonymizeUserSupportTickets(ctx context.Context, userID string) error {
	return anonymizeUserSupportTicketsExec(ctx, s.pool, userID)
}

type supportExecer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func anonymizeUserSupportTicketsExec(ctx context.Context, db supportExecer, userID string) error {
	if _, err := db.Exec(ctx, `
		UPDATE ops.support_ticket_attachments a
		SET file_name = '[anonymized]',
			content_type = '',
			size_bytes = 0,
			object_key = ''
		FROM ops.support_tickets t
		WHERE t.id = a.ticket_id AND t.created_by = $1`, userID); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, `
		UPDATE ops.support_tickets
		SET created_by = NULL,
			subject = '[anonymized]',
			message = '[anonymized]',
			diagnostics = '{}'::jsonb,
			user_agent = '',
			app_version = '',
			locale = '',
			route = '',
			updated_at = NOW()
		WHERE created_by = $1`, userID); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, `UPDATE ops.support_ticket_replies SET author_id = NULL WHERE author_id = $1`, userID); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, `UPDATE ops.support_ticket_status_events SET changed_by = NULL WHERE changed_by = $1`, userID); err != nil {
		return err
	}
	_, err := db.Exec(ctx, `UPDATE ops.support_ticket_attachments SET uploaded_by = NULL WHERE uploaded_by = $1`, userID)
	return err
}

// ListSupportTicketsForExport returns tickets, replies, status history, and attachment metadata for RGPD export.
func (s *Store) ListSupportTicketsForExport(ctx context.Context, userID string) (json.RawMessage, error) {
	var raw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(jsonb_agg(row_to_json(ticket) ORDER BY ticket.created_at), '[]'::jsonb)
		FROM (
			SELECT t.id, t.source, t.subject, t.message, t.status, t.diagnostics,
				t.user_agent, t.app_version, t.locale, t.route, t.created_at, t.updated_at,
				COALESCE((
					SELECT jsonb_agg(jsonb_build_object(
						'id', r.id, 'body', r.body, 'createdAt', r.created_at
					) ORDER BY r.created_at)
					FROM ops.support_ticket_replies r
					WHERE r.ticket_id = t.id
				), '[]'::jsonb) AS replies,
				COALESCE((
					SELECT jsonb_agg(jsonb_build_object(
						'id', e.id, 'fromStatus', e.from_status, 'toStatus', e.to_status,
						'changedBy', e.changed_by, 'changedName', COALESCE(u.full_name, ''),
						'createdAt', e.created_at
					) ORDER BY e.created_at, e.id)
					FROM ops.support_ticket_status_events e
					LEFT JOIN identity.users u ON u.id = e.changed_by
					WHERE e.ticket_id = t.id
				), '[]'::jsonb) AS status_history,
				COALESCE((
					SELECT jsonb_agg(jsonb_build_object(
						'id', a.id, 'uploadedBy', a.uploaded_by,
						'uploaderName', COALESCE(u.full_name, ''), 'fileName', a.file_name,
						'contentType', a.content_type, 'sizeBytes', a.size_bytes, 'createdAt', a.created_at
					) ORDER BY a.created_at, a.id)
					FROM ops.support_ticket_attachments a
					LEFT JOIN identity.users u ON u.id = a.uploaded_by
					WHERE a.ticket_id = t.id
				), '[]'::jsonb) AS attachments
			FROM ops.support_tickets t
			WHERE t.created_by = $1
		) ticket`, userID).Scan(&raw)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

// SupportTicketStats — agrégats inbox ops, avec vieillissement des tickets actifs.
type SupportTicketStats struct {
	ByStatus         map[string]int `json:"byStatus"`
	BySource         map[string]int `json:"bySource"`
	OpenOlderThan24h int            `json:"openOlderThan24h"`
	OpenOlderThan7d  int            `json:"openOlderThan7d"`
	Total            int            `json:"total"`
}

func (s *Store) SupportTicketStats(ctx context.Context) (SupportTicketStats, error) {
	out := SupportTicketStats{
		ByStatus: map[string]int{},
		BySource: map[string]int{},
	}
	rows, err := s.pool.Query(ctx, `SELECT status, COUNT(*)::int FROM ops.support_tickets GROUP BY status`)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return out, err
		}
		out.ByStatus[status] = count
		out.Total += count
	}
	if err := rows.Err(); err != nil {
		return out, err
	}

	sourceRows, err := s.pool.Query(ctx, `SELECT source, COUNT(*)::int FROM ops.support_tickets GROUP BY source`)
	if err != nil {
		return out, err
	}
	defer sourceRows.Close()
	for sourceRows.Next() {
		var source string
		var count int
		if err := sourceRows.Scan(&source, &count); err != nil {
			return out, err
		}
		out.BySource[source] = count
	}
	if err := sourceRows.Err(); err != nil {
		return out, err
	}

	activeStatuses := []string{SupportStatusOpen, SupportStatusInProgress, SupportStatusToTest}
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM ops.support_tickets
		WHERE status = ANY($1) AND created_at < NOW() - INTERVAL '24 hours'`, activeStatuses,
	).Scan(&out.OpenOlderThan24h); err != nil {
		return out, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM ops.support_tickets
		WHERE status = ANY($1) AND created_at < NOW() - INTERVAL '7 days'`, activeStatuses,
	).Scan(&out.OpenOlderThan7d); err != nil {
		return out, err
	}
	return out, nil
}

// SupportOpsRecipient is an admin/dev mailbox for support fan-out emails.
type SupportOpsRecipient struct {
	UserID string
	Email  string
	Name   string
	Locale string
	Role   string
}

// ListSupportOpsRecipients returns users who have an admin and/or DEV profile
// (independent of the currently active users.role after profile switch), with a
// real mailbox (skips empty, *.petsfollow.test, and tombstone addresses).
func (s *Store) ListSupportOpsRecipients(ctx context.Context) ([]SupportOpsRecipient, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text,
			u.email,
			COALESCE(u.full_name, ''),
			COALESCE(NULLIF(TRIM(u.preferred_locale), ''), 'fr'),
			CASE
				WHEN EXISTS (
					SELECT 1 FROM identity.profiles p
					WHERE p.user_id = u.id AND p.role = 'admin'
				) THEN 'admin'
				ELSE 'dev'
			END
		FROM identity.users u
		WHERE EXISTS (
			SELECT 1 FROM identity.profiles p
			WHERE p.user_id = u.id AND p.role IN ('admin', 'dev')
		)
		  AND u.email IS NOT NULL AND TRIM(u.email) <> ''
		  AND LOWER(u.email) NOT LIKE '%@petsfollow.test'
		  AND LOWER(u.email) NOT LIKE '%@deleted.petsfollow.invalid'
		ORDER BY 5, u.email`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []SupportOpsRecipient{}
	for rows.Next() {
		var r SupportOpsRecipient
		if err := rows.Scan(&r.UserID, &r.Email, &r.Name, &r.Locale, &r.Role); err != nil {
			return nil, err
		}
		r.Email = strings.TrimSpace(r.Email)
		if r.Email == "" {
			continue
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
