package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Filiation event types (audit trail).
const (
	FiliationEventVetAssigned          = "vet_assigned"
	FiliationEventVetUnassigned        = "vet_unassigned"
	FiliationEventClientReferral       = "client_referral"
	FiliationEventPracticeClientLinked = "practice_client_linked"
)

const (
	DefaultFiliationEventLimit = 100
	MaxFiliationEventLimit     = 500
)

// FiliationEvent is one attribution mutation.
type FiliationEvent struct {
	ID               string         `json:"id"`
	EventType        string         `json:"eventType"`
	CommercialUserID string         `json:"commercialUserId,omitempty"`
	CommercialName   string         `json:"commercialName,omitempty"`
	VetUserID        string         `json:"vetUserId,omitempty"`
	VetName          string         `json:"vetName,omitempty"`
	ClientUserID     string         `json:"clientUserId,omitempty"`
	ClientName       string         `json:"clientName,omitempty"`
	PracticeID       string         `json:"practiceId,omitempty"`
	PracticeName     string         `json:"practiceName,omitempty"`
	ActorUserID      string         `json:"actorUserId,omitempty"`
	InviteCode       string         `json:"inviteCode,omitempty"`
	Meta             map[string]any `json:"meta,omitempty"`
	CreatedAt        time.Time      `json:"createdAt"`
}

// FiliationEventInput is used to append an audit row (best-effort).
type FiliationEventInput struct {
	EventType        string
	CommercialUserID string
	VetUserID        string
	ClientUserID     string
	PracticeID       string
	ActorUserID      string
	InviteCode       string
	Meta             map[string]any
}

// FiliationEventFilter scopes ListFiliationEvents.
type FiliationEventFilter struct {
	CommercialIDs []string // nil = all; empty = none
	ClientUserID  string
	VetUserID     string
	EventType     string
	Limit         int
	Offset        int
}

// FiliationEventPage is a capped ListFiliationEvents result.
type FiliationEventPage struct {
	Items     []FiliationEvent `json:"items"`
	Limit     int              `json:"limit"`
	Offset    int              `json:"offset"`
	Truncated bool             `json:"truncated"`
}

// RecordPracticeClientLinkedEvent best-effort audit after a cabinet↔client link.
func (s *Store) RecordPracticeClientLinkedEvent(ctx context.Context, practiceID, clientUserID, vetUserID, actorUserID string, meta map[string]any) {
	var assignedComm string
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(assigned_commercial_id::text,'') FROM identity.users WHERE id=$1`, vetUserID).Scan(&assignedComm)
	_ = s.RecordFiliationEvent(ctx, FiliationEventInput{
		EventType:        FiliationEventPracticeClientLinked,
		CommercialUserID: assignedComm,
		VetUserID:        vetUserID,
		ClientUserID:     clientUserID,
		PracticeID:       practiceID,
		ActorUserID:      actorUserID,
		Meta:             meta,
	})
}

// RecordFiliationEvent appends an audit row. Errors are ignored by callers (best-effort).
func (s *Store) RecordFiliationEvent(ctx context.Context, in FiliationEventInput) error {
	if strings.TrimSpace(in.EventType) == "" {
		return nil
	}
	nullUUID := func(id string) any {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil
		}
		return id
	}
	metaJSON := []byte("{}")
	if in.Meta != nil {
		if b, err := json.Marshal(in.Meta); err == nil {
			metaJSON = b
		}
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO practice.filiation_events (
			id, event_type, commercial_user_id, vet_user_id, client_user_id,
			practice_id, actor_user_id, invite_code, meta
		) VALUES ($1, $2, $3::uuid, $4::uuid, $5::uuid, $6::uuid, $7::uuid, $8, $9::jsonb)`,
		uuid.NewString(), in.EventType,
		nullUUID(in.CommercialUserID), nullUUID(in.VetUserID), nullUUID(in.ClientUserID),
		nullUUID(in.PracticeID), nullUUID(in.ActorUserID), strings.TrimSpace(in.InviteCode),
		metaJSON,
	)
	return err
}

// ListFiliationEvents returns recent attribution events (newest first), paginated.
func (s *Store) ListFiliationEvents(ctx context.Context, f FiliationEventFilter) (FiliationEventPage, error) {
	empty := FiliationEventPage{Items: []FiliationEvent{}, Limit: DefaultFiliationEventLimit}
	if f.CommercialIDs != nil && len(f.CommercialIDs) == 0 {
		return empty, nil
	}
	limit := f.Limit
	if limit <= 0 {
		limit = DefaultFiliationEventLimit
	}
	if limit > MaxFiliationEventLimit {
		limit = MaxFiliationEventLimit
	}
	offset := max(f.Offset, 0)
	empty.Limit = limit
	empty.Offset = offset

	args := make([]any, 0, 8)
	argN := 1
	where := []string{"TRUE"}

	if f.CommercialIDs != nil {
		where = append(where, fmt.Sprintf("e.commercial_user_id = ANY($%d::uuid[])", argN))
		args = append(args, f.CommercialIDs)
		argN++
	}
	if cid := strings.TrimSpace(f.ClientUserID); cid != "" {
		where = append(where, fmt.Sprintf("e.client_user_id = $%d::uuid", argN))
		args = append(args, cid)
		argN++
	}
	if vid := strings.TrimSpace(f.VetUserID); vid != "" {
		where = append(where, fmt.Sprintf("e.vet_user_id = $%d::uuid", argN))
		args = append(args, vid)
		argN++
	}
	if et := strings.TrimSpace(f.EventType); et != "" {
		where = append(where, fmt.Sprintf("e.event_type = $%d", argN))
		args = append(args, et)
		argN++
	}

	limitArg := argN
	args = append(args, limit+1)
	argN++
	offsetArg := argN
	args = append(args, offset)

	sql := fmt.Sprintf(`
		SELECT e.id::text, e.event_type,
			COALESCE(e.commercial_user_id::text,''), COALESCE(c.full_name,''),
			COALESCE(e.vet_user_id::text,''), COALESCE(v.full_name,''),
			COALESCE(e.client_user_id::text,''), COALESCE(cli.full_name,''),
			COALESCE(e.practice_id::text,''), COALESCE(pr.name,''),
			COALESCE(e.actor_user_id::text,''), COALESCE(e.invite_code,''),
			COALESCE(e.meta, '{}'::jsonb),
			e.created_at
		FROM practice.filiation_events e
		LEFT JOIN identity.users c ON c.id = e.commercial_user_id
		LEFT JOIN identity.users v ON v.id = e.vet_user_id
		LEFT JOIN identity.users cli ON cli.id = e.client_user_id
		LEFT JOIN practice.practices pr ON pr.id = e.practice_id
		WHERE %s
		ORDER BY e.created_at DESC, e.id DESC
		LIMIT $%d OFFSET $%d`,
		strings.Join(where, " AND "), limitArg, offsetArg)

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return empty, err
	}
	defer rows.Close()

	out := make([]FiliationEvent, 0)
	for rows.Next() {
		var e FiliationEvent
		var metaRaw []byte
		if err := rows.Scan(
			&e.ID, &e.EventType,
			&e.CommercialUserID, &e.CommercialName,
			&e.VetUserID, &e.VetName,
			&e.ClientUserID, &e.ClientName,
			&e.PracticeID, &e.PracticeName,
			&e.ActorUserID, &e.InviteCode,
			&metaRaw,
			&e.CreatedAt,
		); err != nil {
			return empty, err
		}
		if len(metaRaw) > 0 {
			_ = json.Unmarshal(metaRaw, &e.Meta)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return empty, err
	}
	truncated := len(out) > limit
	if truncated {
		out = out[:limit]
	}
	return FiliationEventPage{
		Items:     out,
		Limit:     limit,
		Offset:    offset,
		Truncated: truncated,
	}, nil
}
