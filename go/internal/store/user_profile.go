package store

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s *Store) UpdateUserFullName(ctx context.Context, userID, fullName string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE identity.users SET full_name = $2 WHERE id = $1`, userID, fullName)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpdateUserAvatarURL(ctx context.Context, userID, avatarURL string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE identity.users SET avatar_url = $2 WHERE id = $1`, userID, avatarURL)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpdatePetPhotoURL(ctx context.Context, petID, photoURL string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE pets.pets SET photo_url = $2, updated_at = NOW() WHERE id = $1`, petID, photoURL)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ChangeUserPassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	var hash *string
	var mustChange bool
	err := s.pool.QueryRow(ctx, `
		SELECT password_hash, must_change_password FROM identity.users WHERE id = $1`, userID).Scan(&hash, &mustChange)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if hash == nil || *hash == "" {
		return ErrForbidden
	}
	if !mustChange {
		if currentPassword == "" || bcrypt.CompareHashAndPassword([]byte(*hash), []byte(currentPassword)) != nil {
			return ErrForbidden
		}
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE identity.users SET password_hash = $2, must_change_password = false WHERE id = $1`,
		userID, string(newHash))
	return err
}

// ClientAccountArtifacts — références externes à purger après l'effacement DB (RGPD art. 17) :
// objets média (avatar, photos, documents, médias messages, audio CR) et abonnements Stripe.
type ClientAccountArtifacts struct {
	MediaURLs       []string
	MediaObjectKeys []string
	SubscriptionIDs []string
}

func (s *Store) CollectClientAccountArtifacts(ctx context.Context, userID string) (ClientAccountArtifacts, error) {
	var a ClientAccountArtifacts
	appendNonEmpty := func(dst *[]string, v string) {
		if strings.TrimSpace(v) != "" {
			*dst = append(*dst, strings.TrimSpace(v))
		}
	}
	collect := func(dst *[]string, query string) error {
		rows, err := s.pool.Query(ctx, query, userID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var v string
			if err := rows.Scan(&v); err != nil {
				return err
			}
			appendNonEmpty(dst, v)
		}
		return rows.Err()
	}

	var avatar string
	if err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(avatar_url,'') FROM identity.users WHERE id=$1`, userID).Scan(&avatar); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return a, err
	}
	appendNonEmpty(&a.MediaURLs, avatar)

	if err := collect(&a.MediaURLs,
		`SELECT COALESCE(photo_url,'') FROM pets.pets WHERE owner_user_id=$1`); err != nil {
		return a, err
	}
	if err := collect(&a.MediaObjectKeys, `
		SELECT COALESCE(d.object_key,'') FROM pets.pet_documents d
		JOIN pets.pets p ON p.id = d.pet_id WHERE p.owner_user_id=$1`); err != nil {
		return a, err
	}
	if err := collect(&a.MediaURLs, `
		SELECT COALESCE(m.media_url,'') FROM messaging.messages m
		JOIN messaging.threads t ON t.id = m.thread_id WHERE t.client_user_id=$1`); err != nil {
		return a, err
	}
	if err := collect(&a.MediaObjectKeys, `
		SELECT COALESCE(vr.audio_object_key,'') FROM visits.visit_reports vr
		JOIN visits.visits v ON v.id = vr.visit_id
		JOIN pets.pets p ON p.id = v.pet_id WHERE p.owner_user_id=$1`); err != nil {
		return a, err
	}
	if err := collect(&a.SubscriptionIDs, `
		SELECT COALESCE(stripe_subscription_id,'') FROM billing.pet_entitlements
		WHERE owner_user_id=$1 AND status IN ('active','past_due','pending')`); err != nil {
		return a, err
	}
	if err := collect(&a.SubscriptionIDs, `
		SELECT COALESCE(stripe_subscription_id,'') FROM billing.addon_entitlements
		WHERE owner_user_id=$1 AND status = 'active'`); err != nil {
		return a, err
	}
	return a, nil
}

const tombstoneEmailSuffix = "@deleted.petsfollow.invalid"

// IsTombstoneEmail reconnaît un compte Pro anonymisé (login/refresh refusés).
func IsTombstoneEmail(email string) bool {
	return strings.HasSuffix(email, tombstoneEmailSuffix)
}

// DeleteProAccount anonymise un compte Pro (vet / commercial / commercial_manager / care_pro) :
// les données personnelles sont effacées et le login désactivé ; les données cliniques
// rattachées au cabinet (visites, CR) sont conservées pour leur intégrité.
func (s *Store) DeleteProAccount(ctx context.Context, userID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM notifications.device_tokens WHERE user_id = $1`, userID); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE identity.users SET
			email = 'deleted+' || id || '`+tombstoneEmailSuffix+`',
			full_name = 'Compte supprimé',
			password_hash = NULL,
			google_sub = NULL,
			auth_provider = 'password',
			totp_secret = NULL,
			totp_enabled = false,
			avatar_url = NULL,
			email_verified_at = NULL
		WHERE id = $1 AND role IN ('vet','commercial','commercial_manager','care_pro')`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

func (s *Store) DeleteClientAccount(ctx context.Context, userID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM pets.pets WHERE owner_user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM messaging.threads WHERE client_user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM practice.practice_clients WHERE client_user_id = $1`, userID); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `DELETE FROM identity.users WHERE id = $1 AND role = 'client'`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

func (s *Store) UpdateEmailPrefs(ctx context.Context, vetID string, onMessage, onHeartRate, onVisitRequest bool) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO notifications.notification_preferences (vet_user_id, email_on_message, email_on_heartrate, email_on_visit_request)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (vet_user_id) DO UPDATE SET
			email_on_message = EXCLUDED.email_on_message,
			email_on_heartrate = EXCLUDED.email_on_heartrate,
			email_on_visit_request = EXCLUDED.email_on_visit_request`,
		vetID, onMessage, onHeartRate, onVisitRequest)
	return err
}

func (s *Store) GetEmailPrefs(ctx context.Context, vetID string) (VetEmailPrefs, error) {
	return s.EmailPrefs(ctx, vetID)
}
