package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type ThreadSummary struct {
	ID                 string     `json:"id"`
	PracticeID         string     `json:"practiceId"`
	ClientUserID       string     `json:"clientUserId"`
	VetUserID          string     `json:"vetUserId"`
	PetID              string     `json:"petId"`
	ClientName         string     `json:"clientName"`
	ClientEmail        string     `json:"clientEmail"`
	PracticeName       string     `json:"practiceName,omitempty"`
	VetFullName        string     `json:"vetFullName,omitempty"`
	PetName            string     `json:"petName,omitempty"`
	LastMessagePreview string     `json:"lastMessagePreview"`
	LastMessageAt      *time.Time `json:"lastMessageAt,omitempty"`
	UnreadCount        int        `json:"unreadCount"`
}

type VetOverview struct {
	ClientCount          int `json:"clientCount"`
	UnreadMessages       int `json:"unreadMessages"`
	RecentSessions7d     int `json:"recentSessions7d"`
	PendingLinkRequests  int `json:"pendingLinkRequests"`
	PendingVisits        int `json:"pendingVisits"`
	OverdueCareCount     int `json:"overdueCareCount"`
	UnreadHeartrate      int `json:"unreadHeartrate"`
}

// ClientOverview aggregates vet-facing KPIs for a single client household.
type ClientOverview struct {
	PetCount         int        `json:"petCount"`
	UnreadHeartrate  int        `json:"unreadHeartrate"`
	AlertSessions7d  int        `json:"alertSessions7d"`
	OverdueCareCount int        `json:"overdueCareCount"`
	PendingVisits    int        `json:"pendingVisits"`
	UpcomingVisitAt  *time.Time `json:"upcomingVisitAt,omitempty"`
	ShareCount       int        `json:"shareCount"`
}

func (s *Store) GetClientByPractice(ctx context.Context, practiceID, clientID string) (ClientSummary, error) {
	var c ClientSummary
	err := s.pool.QueryRow(ctx, `
		SELECT u.id::text, u.email, u.full_name, COALESCE(u.avatar_url,''), COALESCE(u.contact_phone,''), COUNT(p.id)::int
		FROM practice.practice_clients pc
		JOIN identity.users u ON u.id = pc.client_user_id
		LEFT JOIN pets.pets p ON p.owner_user_id = u.id AND p.practice_id = pc.practice_id
		WHERE pc.practice_id = $1 AND pc.client_user_id = $2
		GROUP BY u.id, u.email, u.full_name, u.avatar_url, u.contact_phone`, practiceID, clientID).Scan(
		&c.UserID, &c.Email, &c.FullName, &c.AvatarURL, &c.ContactPhone, &c.PetCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClientSummary{}, ErrNotFound
	}
	return c, err
}

// UpdateClientContactPhoneByPractice updates contact_phone for a client linked to the practice.
// Empty phone clears the field. Returns ErrNotFound if the client is not linked.
func (s *Store) UpdateClientContactPhoneByPractice(ctx context.Context, practiceID, clientID, contactPhone string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE identity.users u
		SET contact_phone = $3
		FROM practice.practice_clients pc
		WHERE u.id = pc.client_user_id
			AND pc.practice_id = $1
			AND pc.client_user_id = $2
			AND u.role = 'client'`, practiceID, clientID, strings.TrimSpace(contactPhone))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetClientOverview(ctx context.Context, practiceID, clientID string) (ClientOverview, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM practice.practice_clients
			WHERE practice_id = $1 AND client_user_id = $2
		)`, practiceID, clientID).Scan(&exists); err != nil {
		return ClientOverview{}, err
	}
	if !exists {
		return ClientOverview{}, ErrNotFound
	}

	var o ClientOverview
	var upcoming *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)::int FROM pets.pets
			 WHERE practice_id = $1 AND owner_user_id = $2),
			(SELECT COUNT(*)::int FROM heartrate.sessions
			 WHERE practice_id = $1 AND owner_user_id = $2
			   AND status = 'validated' AND vet_seen_at IS NULL),
			(SELECT COUNT(*)::int FROM heartrate.sessions
			 WHERE practice_id = $1 AND owner_user_id = $2
			   AND status = 'validated' AND is_alert = TRUE
			   AND validated_at >= NOW() - INTERVAL '7 days'),
			(SELECT COUNT(*)::int FROM care.reminders c
			 JOIN pets.pets p ON p.id = c.pet_id
			 WHERE c.practice_id = $1 AND p.owner_user_id = $2
			   AND c.status = 'pending' AND c.due_at < NOW()),
			(SELECT COUNT(*)::int FROM visits.visits v
			 JOIN pets.pets p ON p.id = v.pet_id
			 WHERE v.practice_id = $1 AND p.owner_user_id = $2
			   AND v.deleted_at IS NULL
			   AND v.pending_action_by = 'vet'
			   AND v.status IN ('requested', 'reschedule_pending')),
			(SELECT MIN(v.scheduled_at) FROM visits.visits v
			 JOIN pets.pets p ON p.id = v.pet_id
			 WHERE v.practice_id = $1 AND p.owner_user_id = $2
			   AND v.deleted_at IS NULL
			   AND v.scheduled_at IS NOT NULL AND v.scheduled_at >= NOW()
			   AND v.status IN ('requested', 'confirmed', 'reschedule_pending')),
			(SELECT COUNT(*)::int FROM practice.client_access a
			 WHERE a.client_user_id = $2
			   AND (a.expires_at IS NULL OR a.expires_at > NOW())
			   AND (
			     a.granted_by_user_id = $2
			     OR EXISTS (
			       SELECT 1 FROM identity.users g
			       WHERE g.id = a.granted_by_user_id AND g.practice_id = $1
			     )
			     OR EXISTS (
			       SELECT 1 FROM identity.users g
			       WHERE g.id = a.grantee_user_id AND g.practice_id = $1
			     )
			   ))`,
		practiceID, clientID).Scan(
		&o.PetCount, &o.UnreadHeartrate, &o.AlertSessions7d,
		&o.OverdueCareCount, &o.PendingVisits, &upcoming, &o.ShareCount,
	)
	if err != nil {
		return ClientOverview{}, err
	}
	o.UpcomingVisitAt = upcoming
	return o, nil
}

func (s *Store) VetOverview(ctx context.Context, practiceID, vetID string) (VetOverview, error) {
	var o VetOverview
	err := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)::int FROM practice.practice_clients WHERE practice_id = $1),
			(SELECT COUNT(*)::int FROM messaging.messages m
			 JOIN messaging.threads t ON t.id = m.thread_id
			 WHERE t.vet_user_id = $2 AND m.sender_user_id <> $2 AND m.read_at IS NULL),
			(SELECT COUNT(*)::int FROM heartrate.sessions
			 WHERE practice_id = $1 AND status = 'validated'
			   AND validated_at >= NOW() - INTERVAL '7 days'),
			(SELECT COUNT(*)::int FROM practice.client_vet_link_requests
			 WHERE vet_user_id = $2 AND status = 'pending'),
			(SELECT COUNT(*)::int FROM visits.visits
			 WHERE practice_id = $1
			   AND deleted_at IS NULL
			   AND pending_action_by = 'vet'
			   AND status IN ('requested', 'reschedule_pending')),
			(SELECT COUNT(*)::int FROM care.reminders
			 WHERE practice_id = $1 AND status = 'pending' AND due_at < NOW()),
			(SELECT COUNT(*)::int FROM heartrate.sessions
			 WHERE practice_id = $1 AND status = 'validated' AND vet_seen_at IS NULL)`,
		practiceID, vetID).Scan(
		&o.ClientCount, &o.UnreadMessages, &o.RecentSessions7d,
		&o.PendingLinkRequests, &o.PendingVisits, &o.OverdueCareCount,
		&o.UnreadHeartrate,
	)
	return o, err
}

const threadSummarySelect = `
		SELECT t.id::text, COALESCE(t.practice_id::text,''), t.client_user_id::text, t.vet_user_id::text,
			COALESCE(t.pet_id::text, ''), u.full_name, u.email,
			COALESCE(p.name, ''),
			COALESCE((
				SELECT CASE
					WHEN COALESCE(m.body, '') <> '' THEN LEFT(m.body, 120)
					WHEN m.media_type = 'video' THEN '[video]'
					WHEN m.media_type = 'image' THEN '[image]'
					ELSE ''
				END
				FROM messaging.messages m
				WHERE m.thread_id = t.id ORDER BY m.created_at DESC LIMIT 1
			), ''),
			(
				SELECT m.created_at
				FROM messaging.messages m
				WHERE m.thread_id = t.id ORDER BY m.created_at DESC LIMIT 1
			),
			COALESCE((
				SELECT COUNT(*)::int FROM messaging.messages m
				WHERE m.thread_id = t.id AND m.sender_user_id = t.client_user_id AND m.read_at IS NULL
			), 0)
		FROM messaging.threads t
		JOIN identity.users u ON u.id = t.client_user_id
		LEFT JOIN pets.pets p ON p.id = t.pet_id`

func (s *Store) ListThreadSummariesForVet(ctx context.Context, vetID string) ([]ThreadSummary, error) {
	rows, err := s.pool.Query(ctx, threadSummarySelect+`
		WHERE t.vet_user_id = $1
		ORDER BY (
			SELECT m.created_at FROM messaging.messages m
			WHERE m.thread_id = t.id ORDER BY m.created_at DESC LIMIT 1
		) DESC NULLS LAST, t.created_at DESC`, vetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanThreadSummaries(rows)
}

// ListThreadSummariesForPractice lists all client threads for a shared cabinet.
func (s *Store) ListThreadSummariesForPractice(ctx context.Context, practiceID string) ([]ThreadSummary, error) {
	rows, err := s.pool.Query(ctx, threadSummarySelect+`
		WHERE t.practice_id = $1
		ORDER BY (
			SELECT m.created_at FROM messaging.messages m
			WHERE m.thread_id = t.id ORDER BY m.created_at DESC LIMIT 1
		) DESC NULLS LAST, t.created_at DESC`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanThreadSummaries(rows)
}

// ListThreadSummariesForClient lists all messaging threads for a client across cabinets.
func (s *Store) ListThreadSummariesForClient(ctx context.Context, clientUserID string) ([]ThreadSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id::text, COALESCE(t.practice_id::text,''), t.client_user_id::text, t.vet_user_id::text,
			COALESCE(t.pet_id::text, ''),
			COALESCE(pr.name, ''),
			COALESCE(v.full_name, ''),
			COALESCE(p.name, ''),
			COALESCE((
				SELECT CASE
					WHEN COALESCE(m.body, '') <> '' THEN LEFT(m.body, 120)
					WHEN m.media_type = 'video' THEN '[video]'
					WHEN m.media_type = 'image' THEN '[image]'
					ELSE ''
				END
				FROM messaging.messages m
				WHERE m.thread_id = t.id ORDER BY m.created_at DESC LIMIT 1
			), ''),
			(
				SELECT m.created_at
				FROM messaging.messages m
				WHERE m.thread_id = t.id ORDER BY m.created_at DESC LIMIT 1
			),
			COALESCE((
				SELECT COUNT(*)::int FROM messaging.messages m
				WHERE m.thread_id = t.id AND m.sender_user_id <> $1 AND m.read_at IS NULL
			), 0)
		FROM messaging.threads t
		LEFT JOIN practice.practices pr ON pr.id = t.practice_id
		LEFT JOIN identity.users v ON v.id = t.vet_user_id
		LEFT JOIN pets.pets p ON p.id = t.pet_id
		WHERE t.client_user_id = $1
		ORDER BY (
			SELECT m.created_at FROM messaging.messages m
			WHERE m.thread_id = t.id ORDER BY m.created_at DESC LIMIT 1
		) DESC NULLS LAST, t.created_at DESC`, clientUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ThreadSummary
	for rows.Next() {
		var t ThreadSummary
		if err := rows.Scan(
			&t.ID, &t.PracticeID, &t.ClientUserID, &t.VetUserID, &t.PetID,
			&t.PracticeName, &t.VetFullName, &t.PetName,
			&t.LastMessagePreview, &t.LastMessageAt, &t.UnreadCount,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func scanThreadSummaries(rows pgx.Rows) ([]ThreadSummary, error) {
	var out []ThreadSummary
	for rows.Next() {
		var t ThreadSummary
		if err := rows.Scan(
			&t.ID, &t.PracticeID, &t.ClientUserID, &t.VetUserID, &t.PetID,
			&t.ClientName, &t.ClientEmail, &t.PetName,
			&t.LastMessagePreview, &t.LastMessageAt, &t.UnreadCount,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
