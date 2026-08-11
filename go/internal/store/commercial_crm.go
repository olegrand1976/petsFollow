package store

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProspectEventKind string

const (
	EventNote         ProspectEventKind = "note"
	EventCall         ProspectEventKind = "call"
	EventMeeting      ProspectEventKind = "meeting"
	EventEmailSent    ProspectEventKind = "email_sent"
	EventStatusChange ProspectEventKind = "status_change"
	EventTaskDone     ProspectEventKind = "task_done"
	EventSystem       ProspectEventKind = "system"
)

func ValidProspectEventKind(k string) bool {
	switch ProspectEventKind(k) {
	case EventNote, EventCall, EventMeeting, EventEmailSent, EventStatusChange, EventTaskDone, EventSystem:
		return true
	default:
		return false
	}
}

type ActivityKind string

const (
	ActivityCall     ActivityKind = "call"
	ActivityEmail    ActivityKind = "email"
	ActivityFollowUp ActivityKind = "follow_up"
	ActivityMeeting  ActivityKind = "meeting"
	ActivityOther    ActivityKind = "other"
)

func ValidActivityKind(k string) bool {
	switch ActivityKind(k) {
	case ActivityCall, ActivityEmail, ActivityFollowUp, ActivityMeeting, ActivityOther:
		return true
	default:
		return false
	}
}

func ValidActivityStatus(s string) bool {
	switch s {
	case "open", "done", "cancelled":
		return true
	default:
		return false
	}
}

type ProspectEvent struct {
	ID          string         `json:"id"`
	ProspectID  string         `json:"prospectId"`
	ActorUserID string         `json:"actorUserId,omitempty"`
	ActorName   string         `json:"actorName,omitempty"`
	Kind        string         `json:"kind"`
	Body        string         `json:"body"`
	Meta        map[string]any `json:"meta,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
}

type Activity struct {
	ID             string     `json:"id"`
	ProspectID     string     `json:"prospectId"`
	AssigneeUserID string     `json:"assigneeUserId"`
	AssigneeName   string     `json:"assigneeName,omitempty"`
	CreatedBy      string     `json:"createdBy,omitempty"`
	Kind           string     `json:"kind"`
	Title          string     `json:"title"`
	DueAt          *time.Time `json:"dueAt,omitempty"`
	DoneAt         *time.Time `json:"doneAt,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	PracticeName   string     `json:"practiceName,omitempty"`
	ContactName    string     `json:"contactName,omitempty"`
}

type ActivityInput struct {
	ProspectID     string
	AssigneeUserID string
	CreatedBy      string
	Kind           string
	Title          string
	DueAt          *time.Time
	Status         string
}

type AgendaItem struct {
	ID             string    `json:"id"`
	Source         string    `json:"source"` // appointment | activity
	ProspectID     string    `json:"prospectId"`
	PracticeName   string    `json:"practiceName"`
	ContactName    string    `json:"contactName,omitempty"`
	Title          string    `json:"title"`
	StartsAt       time.Time `json:"startsAt"`
	Kind           string    `json:"kind,omitempty"`
	CommercialID   string    `json:"commercialUserId,omitempty"`
	CommercialName string    `json:"commercialName,omitempty"`
}

type ProspectDetail struct {
	Prospect   Prospect        `json:"prospect"`
	Events     []ProspectEvent `json:"events"`
	Activities []Activity      `json:"activities"`
	Emails     []EmailSend     `json:"emails"`
}

func (s *Store) CreateProspectEvent(ctx context.Context, prospectID, actorUserID, kind, body string, meta map[string]any) (ProspectEvent, error) {
	if !ValidProspectEventKind(kind) {
		return ProspectEvent{}, errors.New("invalid_event_kind")
	}
	id := uuid.NewString()
	metaJSON, _ := json.Marshal(meta)
	if meta == nil {
		metaJSON = []byte("{}")
	}
	var actor any
	if strings.TrimSpace(actorUserID) != "" {
		actor = actorUserID
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO sales.prospect_events (id, prospect_id, actor_user_id, kind, body, meta)
		VALUES ($1,$2,$3,$4,$5,$6::jsonb)
		RETURNING id::text, prospect_id::text, COALESCE(actor_user_id::text,''), kind, body, meta, created_at`,
		id, prospectID, actor, kind, body, string(metaJSON))
	return scanProspectEvent(row.Scan)
}

func (s *Store) ListProspectEvents(ctx context.Context, prospectID string, limit int) ([]ProspectEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.pool.Query(ctx, `
		SELECT e.id::text, e.prospect_id::text, COALESCE(e.actor_user_id::text,''), e.kind, e.body, e.meta, e.created_at,
			COALESCE(u.full_name,'')
		FROM sales.prospect_events e
		LEFT JOIN identity.users u ON u.id = e.actor_user_id
		WHERE e.prospect_id=$1
		ORDER BY e.created_at DESC
		LIMIT $2`, prospectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ProspectEvent, 0)
	for rows.Next() {
		var e ProspectEvent
		var metaBytes []byte
		if err := rows.Scan(&e.ID, &e.ProspectID, &e.ActorUserID, &e.Kind, &e.Body, &metaBytes, &e.CreatedAt, &e.ActorName); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(metaBytes, &e.Meta)
		out = append(out, e)
	}
	return out, rows.Err()
}

func scanProspectEvent(scan func(dest ...any) error) (ProspectEvent, error) {
	var e ProspectEvent
	var metaBytes []byte
	err := scan(&e.ID, &e.ProspectID, &e.ActorUserID, &e.Kind, &e.Body, &metaBytes, &e.CreatedAt)
	if err != nil {
		return ProspectEvent{}, err
	}
	_ = json.Unmarshal(metaBytes, &e.Meta)
	return e, nil
}

func (s *Store) CreateActivity(ctx context.Context, in ActivityInput) (Activity, error) {
	kind := strings.TrimSpace(in.Kind)
	if kind == "" {
		kind = string(ActivityFollowUp)
	}
	if !ValidActivityKind(kind) {
		return Activity{}, errors.New("invalid_activity_kind")
	}
	status := strings.TrimSpace(in.Status)
	if status == "" {
		status = "open"
	}
	if !ValidActivityStatus(status) {
		return Activity{}, errors.New("invalid_activity_status")
	}
	id := uuid.NewString()
	var createdBy any
	if strings.TrimSpace(in.CreatedBy) != "" {
		createdBy = in.CreatedBy
	}
	var due any
	if in.DueAt != nil {
		due = *in.DueAt
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO sales.activities (
			id, prospect_id, assignee_user_id, created_by, kind, title, due_at, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id::text, prospect_id::text, COALESCE(assignee_user_id::text,''), COALESCE(created_by::text,''),
			kind, title, due_at, done_at, status, created_at, updated_at`,
		id, in.ProspectID, in.AssigneeUserID, createdBy, kind, strings.TrimSpace(in.Title), due, status)
	return scanActivity(row.Scan)
}

func (s *Store) UpdateActivity(ctx context.Context, id string, title, status string, dueAt *time.Time, clearDue bool) (Activity, error) {
	existing, err := s.GetActivity(ctx, id)
	if err != nil {
		return Activity{}, err
	}
	if strings.TrimSpace(title) == "" {
		title = existing.Title
	}
	if strings.TrimSpace(status) == "" {
		status = existing.Status
	}
	if !ValidActivityStatus(status) {
		return Activity{}, errors.New("invalid_activity_status")
	}
	var due any
	if clearDue {
		due = nil
	} else if dueAt != nil {
		due = *dueAt
	} else if existing.DueAt != nil {
		due = *existing.DueAt
	}
	var doneAt any
	if status == "done" {
		if existing.DoneAt != nil {
			doneAt = *existing.DoneAt
		} else {
			doneAt = time.Now().UTC()
		}
	} else {
		doneAt = nil
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE sales.activities SET
			title=$2, status=$3, due_at=$4, done_at=$5, updated_at=NOW()
		WHERE id=$1
		RETURNING id::text, prospect_id::text, COALESCE(assignee_user_id::text,''), COALESCE(created_by::text,''),
			kind, title, due_at, done_at, status, created_at, updated_at`,
		id, title, status, due, doneAt)
	a, err := scanActivity(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return Activity{}, ErrNotFound
	}
	return a, err
}

func (s *Store) GetActivity(ctx context.Context, id string) (Activity, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, prospect_id::text, COALESCE(assignee_user_id::text,''), COALESCE(created_by::text,''),
			kind, title, due_at, done_at, status, created_at, updated_at
		FROM sales.activities WHERE id=$1`, id)
	a, err := scanActivity(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return Activity{}, ErrNotFound
	}
	return a, err
}

func scanActivity(scan func(dest ...any) error) (Activity, error) {
	var a Activity
	var due, done *time.Time
	err := scan(&a.ID, &a.ProspectID, &a.AssigneeUserID, &a.CreatedBy,
		&a.Kind, &a.Title, &due, &done, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	a.DueAt = due
	a.DoneAt = done
	return a, err
}

func (s *Store) ListActivitiesByProspect(ctx context.Context, prospectID string, openOnly bool) ([]Activity, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id::text, a.prospect_id::text, COALESCE(a.assignee_user_id::text,''), COALESCE(a.created_by::text,''),
			a.kind, a.title, a.due_at, a.done_at, a.status, a.created_at, a.updated_at,
			COALESCE(u.full_name,''), COALESCE(p.practice_name,''), COALESCE(p.contact_name,'')
		FROM sales.activities a
		LEFT JOIN identity.users u ON u.id = a.assignee_user_id
		JOIN sales.prospects p ON p.id = a.prospect_id
		WHERE a.prospect_id=$1 AND ($2 = false OR a.status='open')
		ORDER BY COALESCE(a.due_at, a.created_at) ASC`, prospectID, openOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanActivityList(rows)
}

func (s *Store) ListActivitiesForAssignee(ctx context.Context, assigneeUserID, status string, overdueOnly bool, limit int) ([]Activity, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
		SELECT a.id::text, a.prospect_id::text, COALESCE(a.assignee_user_id::text,''), COALESCE(a.created_by::text,''),
			a.kind, a.title, a.due_at, a.done_at, a.status, a.created_at, a.updated_at,
			COALESCE(u.full_name,''), COALESCE(p.practice_name,''), COALESCE(p.contact_name,'')
		FROM sales.activities a
		LEFT JOIN identity.users u ON u.id = a.assignee_user_id
		JOIN sales.prospects p ON p.id = a.prospect_id
		WHERE a.assignee_user_id=$1
		  AND ($2='' OR a.status=$2)
		  AND (
			NOT $3::boolean
			OR (a.status='open' AND a.due_at IS NOT NULL AND a.due_at < NOW())
		  )
		ORDER BY COALESCE(a.due_at, a.created_at) ASC
		LIMIT $4`, assigneeUserID, status, overdueOnly, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanActivityList(rows)
}

func (s *Store) ListTeamActivities(ctx context.Context, managerUserID, commercialUserID, status string, overdueOnly bool, limit int) ([]Activity, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
		SELECT a.id::text, a.prospect_id::text, COALESCE(a.assignee_user_id::text,''), COALESCE(a.created_by::text,''),
			a.kind, a.title, a.due_at, a.done_at, a.status, a.created_at, a.updated_at,
			COALESCE(u.full_name,''), COALESCE(p.practice_name,''), COALESCE(p.contact_name,'')
		FROM sales.activities a
		LEFT JOIN identity.users u ON u.id = a.assignee_user_id
		JOIN sales.prospects p ON p.id = a.prospect_id
		WHERE a.assignee_user_id IN (
			SELECT id FROM identity.users WHERE role='commercial' AND manager_user_id=$1
			UNION SELECT $1::uuid
		)
		  AND ($2='' OR a.assignee_user_id::text=$2)
		  AND ($3='' OR a.status=$3)
		  AND (
			NOT $4::boolean
			OR (a.status='open' AND a.due_at IS NOT NULL AND a.due_at < NOW())
		  )
		ORDER BY COALESCE(a.due_at, a.created_at) ASC
		LIMIT $5`, managerUserID, commercialUserID, status, overdueOnly, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanActivityList(rows)
}

func (s *Store) CountOverdueActivitiesForManager(ctx context.Context, managerUserID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM sales.activities a
		WHERE a.status='open' AND a.due_at IS NOT NULL AND a.due_at < NOW()
		  AND a.assignee_user_id IN (
			SELECT id FROM identity.users WHERE role='commercial' AND manager_user_id=$1
			UNION SELECT $1::uuid
		  )`, managerUserID).Scan(&n)
	return n, err
}

func scanActivityList(rows pgx.Rows) ([]Activity, error) {
	out := make([]Activity, 0)
	for rows.Next() {
		var a Activity
		var due, done *time.Time
		if err := rows.Scan(&a.ID, &a.ProspectID, &a.AssigneeUserID, &a.CreatedBy,
			&a.Kind, &a.Title, &due, &done, &a.Status, &a.CreatedAt, &a.UpdatedAt,
			&a.AssigneeName, &a.PracticeName, &a.ContactName); err != nil {
			return nil, err
		}
		a.DueAt = due
		a.DoneAt = done
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) ListAgenda(ctx context.Context, commercialUserID string, from, to time.Time) ([]AgendaItem, error) {
	items := make([]AgendaItem, 0)

	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text, p.practice_name, COALESCE(p.contact_name,''), p.appointment_at,
			COALESCE(p.commercial_user_id::text,''), COALESCE(u.full_name,'')
		FROM sales.prospects p
		LEFT JOIN identity.users u ON u.id = p.commercial_user_id
		WHERE p.commercial_user_id=$1
		  AND p.appointment_at IS NOT NULL
		  AND p.appointment_at >= $2 AND p.appointment_at < $3
		  AND p.appointment_outcome IN ('','scheduled')
		ORDER BY p.appointment_at ASC`, commercialUserID, from, to)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var it AgendaItem
		var starts time.Time
		if err := rows.Scan(&it.ProspectID, &it.PracticeName, &it.ContactName, &starts, &it.CommercialID, &it.CommercialName); err != nil {
			rows.Close()
			return nil, err
		}
		it.ID = "appt-" + it.ProspectID
		it.Source = "appointment"
		it.Title = it.PracticeName
		it.StartsAt = starts
		it.Kind = "meeting"
		items = append(items, it)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	arows, err := s.pool.Query(ctx, `
		SELECT a.id::text, a.prospect_id::text, p.practice_name, COALESCE(p.contact_name,''),
			a.title, a.due_at, a.kind, COALESCE(a.assignee_user_id::text,''), COALESCE(u.full_name,'')
		FROM sales.activities a
		JOIN sales.prospects p ON p.id = a.prospect_id
		LEFT JOIN identity.users u ON u.id = a.assignee_user_id
		WHERE a.assignee_user_id=$1
		  AND a.status='open'
		  AND a.kind='meeting'
		  AND a.due_at IS NOT NULL
		  AND a.due_at >= $2 AND a.due_at < $3
		ORDER BY a.due_at ASC`, commercialUserID, from, to)
	if err != nil {
		return nil, err
	}
	defer arows.Close()
	for arows.Next() {
		var it AgendaItem
		var starts time.Time
		if err := arows.Scan(&it.ID, &it.ProspectID, &it.PracticeName, &it.ContactName,
			&it.Title, &starts, &it.Kind, &it.CommercialID, &it.CommercialName); err != nil {
			return nil, err
		}
		it.Source = "activity"
		it.StartsAt = starts
		items = append(items, it)
	}
	if err := arows.Err(); err != nil {
		return nil, err
	}
	sortAgendaItems(items)
	return items, nil
}

func sortAgendaItems(items []AgendaItem) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].StartsAt.Before(items[j].StartsAt)
	})
}

func (s *Store) ListTeamAgenda(ctx context.Context, managerUserID, commercialUserID string, from, to time.Time) ([]AgendaItem, error) {
	items := make([]AgendaItem, 0)
	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text, p.practice_name, COALESCE(p.contact_name,''), p.appointment_at,
			COALESCE(p.commercial_user_id::text,''), COALESCE(u.full_name,'')
		FROM sales.prospects p
		LEFT JOIN identity.users u ON u.id = p.commercial_user_id
		WHERE (
			p.commercial_user_id=$1
			OR p.commercial_user_id IN (SELECT id FROM identity.users WHERE role='commercial' AND manager_user_id=$1)
		)
		  AND ($2='' OR p.commercial_user_id::text=$2)
		  AND p.appointment_at IS NOT NULL
		  AND p.appointment_at >= $3 AND p.appointment_at < $4
		  AND p.appointment_outcome IN ('','scheduled')
		ORDER BY p.appointment_at ASC`, managerUserID, commercialUserID, from, to)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var it AgendaItem
		var starts time.Time
		if err := rows.Scan(&it.ProspectID, &it.PracticeName, &it.ContactName, &starts, &it.CommercialID, &it.CommercialName); err != nil {
			rows.Close()
			return nil, err
		}
		it.ID = "appt-" + it.ProspectID
		it.Source = "appointment"
		it.Title = it.PracticeName
		it.StartsAt = starts
		it.Kind = "meeting"
		items = append(items, it)
	}
	rows.Close()

	arows, err := s.pool.Query(ctx, `
		SELECT a.id::text, a.prospect_id::text, p.practice_name, COALESCE(p.contact_name,''),
			a.title, a.due_at, a.kind, COALESCE(a.assignee_user_id::text,''), COALESCE(u.full_name,'')
		FROM sales.activities a
		JOIN sales.prospects p ON p.id = a.prospect_id
		LEFT JOIN identity.users u ON u.id = a.assignee_user_id
		WHERE a.assignee_user_id IN (
			SELECT id FROM identity.users WHERE role='commercial' AND manager_user_id=$1
			UNION SELECT $1::uuid
		)
		  AND ($2='' OR a.assignee_user_id::text=$2)
		  AND a.status='open' AND a.kind='meeting'
		  AND a.due_at IS NOT NULL
		  AND a.due_at >= $3 AND a.due_at < $4
		ORDER BY a.due_at ASC`, managerUserID, commercialUserID, from, to)
	if err != nil {
		return nil, err
	}
	defer arows.Close()
	for arows.Next() {
		var it AgendaItem
		var starts time.Time
		if err := arows.Scan(&it.ID, &it.ProspectID, &it.PracticeName, &it.ContactName,
			&it.Title, &starts, &it.Kind, &it.CommercialID, &it.CommercialName); err != nil {
			return nil, err
		}
		it.Source = "activity"
		it.StartsAt = starts
		items = append(items, it)
	}
	if err := arows.Err(); err != nil {
		return nil, err
	}
	sortAgendaItems(items)
	return items, nil
}

func (s *Store) GetProspectDetail(ctx context.Context, prospectID string) (ProspectDetail, error) {
	p, err := s.GetProspectByID(ctx, prospectID)
	if err != nil {
		return ProspectDetail{}, err
	}
	events, err := s.ListProspectEvents(ctx, prospectID, 100)
	if err != nil {
		return ProspectDetail{}, err
	}
	acts, err := s.ListActivitiesByProspect(ctx, prospectID, false)
	if err != nil {
		return ProspectDetail{}, err
	}
	mails, err := s.ListEmailSendsByProspect(ctx, prospectID)
	if err != nil {
		return ProspectDetail{}, err
	}
	if len(mails) > 20 {
		mails = mails[:20]
	}
	return ProspectDetail{Prospect: p, Events: events, Activities: acts, Emails: mails}, nil
}
