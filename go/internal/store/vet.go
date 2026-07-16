package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type ThreadSummary struct {
	ID                  string `json:"id"`
	PracticeID          string `json:"practiceId"`
	ClientUserID        string `json:"clientUserId"`
	VetUserID           string `json:"vetUserId"`
	PetID               string `json:"petId"`
	ClientName          string `json:"clientName"`
	ClientEmail         string `json:"clientEmail"`
	LastMessagePreview  string `json:"lastMessagePreview"`
	UnreadCount         int    `json:"unreadCount"`
}

type VetOverview struct {
	ClientCount          int `json:"clientCount"`
	UnreadMessages       int `json:"unreadMessages"`
	RecentSessions7d     int `json:"recentSessions7d"`
	PendingLinkRequests  int `json:"pendingLinkRequests"`
	PendingVisits        int `json:"pendingVisits"`
	OverdueCareCount     int `json:"overdueCareCount"`
}

func (s *Store) GetClientByPractice(ctx context.Context, practiceID, clientID string) (ClientSummary, error) {
	var c ClientSummary
	err := s.pool.QueryRow(ctx, `
		SELECT u.id::text, u.email, u.full_name, COALESCE(u.avatar_url,''), COUNT(p.id)::int
		FROM practice.practice_clients pc
		JOIN identity.users u ON u.id = pc.client_user_id
		LEFT JOIN pets.pets p ON p.owner_user_id = u.id AND p.practice_id = pc.practice_id
		WHERE pc.practice_id = $1 AND pc.client_user_id = $2
		GROUP BY u.id, u.email, u.full_name, u.avatar_url`, practiceID, clientID).Scan(
		&c.UserID, &c.Email, &c.FullName, &c.AvatarURL, &c.PetCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClientSummary{}, ErrNotFound
	}
	return c, err
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
			 WHERE practice_id = $1 AND status = 'requested'),
			(SELECT COUNT(*)::int FROM care.reminders
			 WHERE practice_id = $1 AND status = 'pending' AND due_at < NOW())`,
		practiceID, vetID).Scan(
		&o.ClientCount, &o.UnreadMessages, &o.RecentSessions7d,
		&o.PendingLinkRequests, &o.PendingVisits, &o.OverdueCareCount,
	)
	return o, err
}

func (s *Store) ListThreadSummariesForVet(ctx context.Context, vetID string) ([]ThreadSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id::text, t.practice_id::text, t.client_user_id::text, t.vet_user_id::text,
			COALESCE(t.pet_id::text, ''), u.full_name, u.email,
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
			COALESCE((
				SELECT COUNT(*)::int FROM messaging.messages m
				WHERE m.thread_id = t.id AND m.sender_user_id <> t.vet_user_id AND m.read_at IS NULL
			), 0)
		FROM messaging.threads t
		JOIN identity.users u ON u.id = t.client_user_id
		WHERE t.vet_user_id = $1
		ORDER BY t.created_at DESC`, vetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ThreadSummary
	for rows.Next() {
		var t ThreadSummary
		if err := rows.Scan(
			&t.ID, &t.PracticeID, &t.ClientUserID, &t.VetUserID, &t.PetID,
			&t.ClientName, &t.ClientEmail, &t.LastMessagePreview, &t.UnreadCount,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
