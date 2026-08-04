package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// SmsLogEntry journalise un SMS client (audit + idempotence rappel).
// Le corps du message n'est jamais stocké : kind + locale déterminent le template.
type SmsLogEntry struct {
	ID                string     `json:"id"`
	UserID            string     `json:"userId"`
	VisitID           string     `json:"visitId,omitempty"`
	Kind              string     `json:"kind"`
	ToPhone           string     `json:"toPhone"`
	Locale            string     `json:"locale"`
	ScheduledFor      *time.Time `json:"scheduledFor,omitempty"`
	ProviderMessageID string     `json:"providerMessageId,omitempty"`
	Status            string     `json:"status"`
	Error             string     `json:"error,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
}

// InsertSmsLog écrit une ligne terminale (sent|dry_run|skipped|error) pour les
// kinds événementiels (confirmation, reprogrammation) et les skips de rappel.
func (s *Store) InsertSmsLog(ctx context.Context, e SmsLogEntry) (string, error) {
	var visitID any
	if e.VisitID != "" {
		visitID = e.VisitID
	}
	var id string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO notifications.sms_log
			(user_id, visit_id, kind, to_phone, locale, scheduled_for, provider_message_id, status, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id::text`,
		e.UserID, visitID, e.Kind, e.ToPhone, e.Locale, e.ScheduledFor, e.ProviderMessageID, e.Status, e.Error,
	).Scan(&id)
	return id, err
}

// ClaimVisitReminderSms réserve l'envoi du rappel J-1 pour (visite, créneau).
// L'index unique partiel sms_log_reminder_once arbitre la concurrence :
// claimed=false signifie qu'un autre run a déjà traité ce rappel (at-most-once).
func (s *Store) ClaimVisitReminderSms(ctx context.Context, visitID string, scheduledFor time.Time, userID, toPhone, locale string) (string, bool, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO notifications.sms_log (user_id, visit_id, kind, to_phone, locale, scheduled_for, status)
		VALUES ($1, $2, 'visit_reminder', $3, $4, $5, 'sending')
		ON CONFLICT (visit_id, scheduled_for) WHERE kind = 'visit_reminder' DO NOTHING
		RETURNING id::text`,
		userID, visitID, toPhone, locale, scheduledFor,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

// FinalizeSmsLog fixe l'issue d'un envoi préalablement réservé (status sending).
func (s *Store) FinalizeSmsLog(ctx context.Context, id, status, providerMessageID, errMsg string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE notifications.sms_log
		SET status = $2, provider_message_id = $3, error = $4
		WHERE id = $1`, id, status, providerMessageID, errMsg)
	return err
}

// VisitReminderCandidate est une visite confirmée éligible au rappel J-1.
type VisitReminderCandidate struct {
	VisitID     string
	PetID       string
	PetName     string
	OwnerUserID string
	PracticeID  string
	ScheduledAt time.Time
}

// ListVisitsForReminder liste les visites confirmées dans [from, to) sans rappel
// déjà journalisé pour leur créneau. Les sessions walk-in sont exclues.
func (s *Store) ListVisitsForReminder(ctx context.Context, from, to time.Time, limit int) ([]VisitReminderCandidate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id::text, v.pet_id::text, p.name, p.owner_user_id::text, v.practice_id::text, v.scheduled_at
		FROM visits.visits v
		JOIN pets.pets p ON p.id = v.pet_id
		WHERE v.status = 'confirmed'
		  AND COALESCE(v.consultation_session, FALSE) = FALSE
		  AND v.scheduled_at >= $1 AND v.scheduled_at < $2
		  AND NOT EXISTS (
			SELECT 1 FROM notifications.sms_log l
			WHERE l.visit_id = v.id AND l.kind = 'visit_reminder' AND l.scheduled_for = v.scheduled_at
		  )
		ORDER BY v.scheduled_at
		LIMIT $3`, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VisitReminderCandidate
	for rows.Next() {
		var c VisitReminderCandidate
		if err := rows.Scan(&c.VisitID, &c.PetID, &c.PetName, &c.OwnerUserID, &c.PracticeID, &c.ScheduledAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ApplySmsDeliveryReport rapproche un DLR Telnyx de la ligne de log par
// provider_message_id. Retourne false si aucun envoi ne correspond (relivraison
// d'un message émis avant la mise en place du webhook, ou message d'un autre env).
func (s *Store) ApplySmsDeliveryReport(ctx context.Context, providerMessageID, status, errMsg string, deliveredAt *time.Time) (bool, error) {
	if providerMessageID == "" {
		return false, nil
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE notifications.sms_log
		SET delivery_status = $2,
			delivery_error = $3,
			delivered_at = COALESCE($4, delivered_at)
		WHERE provider_message_id = $1`,
		providerMessageID, status, errMsg, deliveredAt)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// SmsInboundEntry est un SMS entrant journalisé.
type SmsInboundEntry struct {
	UserID            string
	FromPhone         string
	ToPhone           string
	Body              string
	Command           string
	ProviderMessageID string
	ViaFailover       bool
}

// InsertSmsInbound journalise un SMS entrant. inserted=false signale une
// relivraison Telnyx déjà traitée (index unique sur provider_message_id).
func (s *Store) InsertSmsInbound(ctx context.Context, e SmsInboundEntry) (bool, error) {
	var userID any
	if e.UserID != "" {
		userID = e.UserID
	}
	var providerID any
	if e.ProviderMessageID != "" {
		providerID = e.ProviderMessageID
	}
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO notifications.sms_inbound
			(user_id, from_phone, to_phone, body, command, provider_message_id, via_failover)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6, ''), $7)
		ON CONFLICT (provider_message_id) WHERE provider_message_id <> '' DO NOTHING`,
		userID, e.FromPhone, e.ToPhone, e.Body, e.Command, providerID, e.ViaFailover)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// FindUserIDByContactPhone rapproche un numéro E.164 entrant d'un compte.
// contact_phone est du texte libre (« 0470 12 34 56 ») : on compare les chiffres
// seuls sur les 9 derniers, ce qui absorbe indicatif national vs international.
// Retourne "" si zéro ou plusieurs comptes correspondent (jamais de rattachement
// arbitraire — un STOP ambigu est journalisé sans couper le mauvais client).
func (s *Store) FindUserIDByContactPhone(ctx context.Context, e164 string) (string, error) {
	digits := onlyDigits(e164)
	if len(digits) < 9 {
		return "", nil
	}
	suffix := digits[len(digits)-9:]
	rows, err := s.pool.Query(ctx, `
		SELECT id::text FROM identity.users
		WHERE contact_phone <> ''
		  AND regexp_replace(contact_phone, '[^0-9]', '', 'g') LIKE '%' || $1
		LIMIT 2`, suffix)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if len(ids) != 1 {
		return "", nil
	}
	return ids[0], nil
}

func onlyDigits(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			out = append(out, s[i])
		}
	}
	return string(out)
}

// SetClientSMSPref bascule le canal SMS d'un client (opt-out STOP / opt-in START).
// Upsert : un client sans ligne de préférences en obtient une aux défauts.
func (s *Store) SetClientSMSPref(ctx context.Context, userID string, enabled bool) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO notifications.client_preferences (user_id, sms)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET sms = EXCLUDED.sms`, userID, enabled)
	return err
}

// PurgeOldSmsInbound supprime les SMS entrants plus anciens que cutoff.
func (s *Store) PurgeOldSmsInbound(ctx context.Context, cutoff time.Time, limit int) (int, error) {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM notifications.sms_inbound
		WHERE id IN (
			SELECT id FROM notifications.sms_inbound WHERE created_at < $1 ORDER BY created_at LIMIT $2
		)`, cutoff, limit)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

// PurgeOldSmsLogs supprime les lignes plus anciennes que cutoff (rétention).
func (s *Store) PurgeOldSmsLogs(ctx context.Context, cutoff time.Time, limit int) (int, error) {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM notifications.sms_log
		WHERE id IN (
			SELECT id FROM notifications.sms_log WHERE created_at < $1 ORDER BY created_at LIMIT $2
		)`, cutoff, limit)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}
