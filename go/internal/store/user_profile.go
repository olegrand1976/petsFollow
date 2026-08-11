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

func (s *Store) UpdateUserContactPhone(ctx context.Context, userID, contactPhone string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE identity.users SET contact_phone = $2 WHERE id = $1`, userID, strings.TrimSpace(contactPhone))
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
	// Comme au reset : changer son mot de passe révoque les tokens déjà émis.
	// L'appelant réémet une paire pour son propre appareil (changeMePassword),
	// les autres sessions tombent à leur prochain refresh.
	_, err = s.pool.Exec(ctx, `
		UPDATE identity.users
		SET password_hash = $2, must_change_password = false, token_version = token_version + 1
		WHERE id = $1`,
		userID, string(newHash))
	return err
}

// ClientAccountArtifacts — références externes à purger après l'effacement DB (RGPD art. 17) :
// objets média (avatar, photos, documents, médias messages, audio CR) et abonnements Stripe.
type ClientAccountArtifacts struct {
	MediaURLs       []string
	MediaObjectKeys []string
	SubscriptionIDs []string
	OrthancStudyIDs []string
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
		SELECT COALESCE(health_book_pdf_object_key,'') FROM pets.pets WHERE owner_user_id=$1`); err != nil {
		return a, err
	}
	if err := collect(&a.MediaURLs, `
		SELECT COALESCE(health_book_pdf_url,'') FROM pets.pets WHERE owner_user_id=$1`); err != nil {
		return a, err
	}
	if err := collect(&a.MediaObjectKeys, `
		SELECT COALESCE(d.object_key,'') FROM pets.documents d
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
	if err := collect(&a.MediaObjectKeys, `
		SELECT COALESCE(object_key,'') FROM pets.dossier_share_tokens WHERE owner_user_id=$1`); err != nil {
		return a, err
	}
	if err := collect(&a.MediaObjectKeys, `
		SELECT COALESCE(object_key,'') FROM pets.consultation_share_tokens WHERE owner_user_id=$1`); err != nil {
		return a, err
	}
	if err := collect(&a.MediaObjectKeys, `
		SELECT COALESCE(pdf_object_key,'') FROM prescriptions.prescriptions WHERE owner_id=$1`); err != nil {
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
	if err := collect(&a.OrthancStudyIDs, `
		SELECT DISTINCT ps.orthanc_study_id
		FROM imaging.pet_studies ps
		JOIN pets.pets p ON p.id = ps.pet_id
		WHERE p.owner_user_id = $1 AND ps.orthanc_study_id <> ''`); err != nil {
		return a, err
	}
	return a, nil
}

const tombstoneEmailSuffix = "@deleted.petsfollow.invalid"

// IsTombstoneEmail reconnaît un compte Pro anonymisé (login/refresh refusés).
func IsTombstoneEmail(email string) bool {
	return strings.HasSuffix(email, tombstoneEmailSuffix)
}

// purgeClientOwnedDataExec efface les données détenues en tant que client
// (pets, messagerie, liens cabinet, referrals) sans supprimer la ligne users.
// Utilisé par DeleteClientAccount et DeleteProAccount (dual profil care_pro).
func purgeClientOwnedDataExec(ctx context.Context, tx pgx.Tx, userID string) error {
	stmts := []string{
		`DELETE FROM client_ai.triage_sessions WHERE user_id = $1`,
		`DELETE FROM client_ai.usage_events WHERE user_id = $1`,
		`DELETE FROM pets.pets WHERE owner_user_id = $1`,
		`DELETE FROM messaging.threads WHERE client_user_id = $1`,
		`DELETE FROM practice.vet_leads WHERE client_user_id = $1`,
		`DELETE FROM practice.practice_clients WHERE client_user_id = $1`,
		`DELETE FROM practice.client_referrals
			WHERE referred_client_user_id = $1 OR sponsor_client_user_id = $1`,
		`DELETE FROM practice.commercial_referrals WHERE client_user_id = $1`,
		`DELETE FROM practice.filiation_events
			WHERE client_user_id = $1 OR actor_user_id = $1`,
	}
	for _, q := range stmts {
		if _, err := tx.Exec(ctx, q, userID); err != nil {
			return err
		}
	}
	return redactPharmacyUserDataExec(ctx, tx, userID)
}

// redactPharmacyUserDataExec clears pharmacy PII / PHI trails tied to a user while
// keeping practice operational records (DAF numbers, stock, orders).
func redactPharmacyUserDataExec(ctx context.Context, tx pgx.Tx, userID string) error {
	if _, err := tx.Exec(ctx, `
		UPDATE pharmacy.job_audit ja
		SET request_json = '{}'::jsonb,
		    response_json = NULL,
		    error = CASE WHEN COALESCE(error,'') = '' THEN error ELSE 'redacted' END
		FROM pharmacy.daf_documents d
		WHERE ja.entity_id = d.id
		  AND ja.job_type = 'vamreg'
		  AND (d.prescriber_user_id = $1 OR d.client_user_id = $1)`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE pharmacy.daf_documents
		SET client_user_id = NULL, updated_at = now()
		WHERE client_user_id = $1`, userID); err != nil {
		return err
	}
	for _, q := range []string{
		`UPDATE pharmacy.purchase_orders SET created_by = NULL WHERE created_by = $1`,
		`UPDATE pharmacy.delivery_notes SET created_by = NULL WHERE created_by = $1`,
		`UPDATE pharmacy.inventory_sessions SET created_by = NULL WHERE created_by = $1`,
		`UPDATE pharmacy.inventory_sessions SET closed_by = NULL WHERE closed_by = $1`,
		`SELECT pharmacy.rgpd_null_stock_movement_created_by($1::uuid)`,
		`UPDATE pharmacy.medication_prices SET updated_by = NULL WHERE updated_by = $1`,
	} {
		if _, err := tx.Exec(ctx, q, userID); err != nil {
			return err
		}
	}
	return nil
}

// DeleteProAccount anonymise un compte Pro (vet / assistant / secretary /
// commercial / commercial_manager / care_pro) : données personnelles effacées,
// login désactivé ; données cliniques cabinet (visites, CR) conservées.
// Si le compte possède aussi des données client (dual profil), elles sont purgées.
func (s *Store) DeleteProAccount(ctx context.Context, userID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM notifications.device_tokens WHERE user_id = $1`, userID); err != nil {
		return err
	}
	// Drop live attribution before tombstone so Resolve/Accrue/List cannot pay or
	// surface a deleted commercial (assigned_commercial_id + commercial_referrals).
	if _, err := tx.Exec(ctx, `
		UPDATE identity.users SET assigned_commercial_id = NULL
		WHERE assigned_commercial_id = $1 AND role = 'vet'`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM practice.commercial_referrals WHERE commercial_user_id = $1`, userID); err != nil {
		return err
	}
	// Dual profil : purger pets / threads client avant tombstone (art. 17).
	if err := purgeClientOwnedDataExec(ctx, tx, userID); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE identity.users SET
			email = 'deleted+' || id || '`+tombstoneEmailSuffix+`',
			full_name = 'Compte supprimé',
			first_name = '',
			last_name = '',
			address = '',
			national_registry_number = '',
			billing_vat_number = '',
			billing_company_number = '',
			billing_street = '',
			billing_city = '',
			billing_postal = '',
			billing_country = '',
			billing_customer_kind = '',
			password_hash = NULL,
			google_sub = NULL,
			auth_provider = 'password',
			totp_secret = NULL,
			totp_enabled = false,
			avatar_url = NULL,
			email_verified_at = NULL,
			assigned_commercial_id = NULL,
			contact_phone = ''
		WHERE id = $1 AND role IN (
			'vet','vet_assistant','secretary','commercial','commercial_manager','care_pro'
		)`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	if err := anonymizeUserSupportTicketsExec(ctx, tx, userID); err != nil {
		return err
	}
	// Attribution restante (commercial / vet) — actor/client déjà purgés ci-dessus.
	if _, err := tx.Exec(ctx, `
		DELETE FROM practice.filiation_events
		WHERE commercial_user_id = $1 OR vet_user_id = $1`, userID); err != nil {
		return err
	}
	// Unlink authorship on practice invoices (cabinet keeps counterparty / fiscal records).
	if _, err := tx.Exec(ctx, `
		UPDATE invoicing.documents SET created_by = NULL WHERE created_by = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM invoicing.connect_states WHERE created_by = $1`, userID); err != nil {
		return err
	}
	if err := redactPharmacyUserDataExec(ctx, tx, userID); err != nil {
		return err
	}
	// CR clinical rows stay; purge multi-agent audit trails (steps/citations PHI).
	if _, err := tx.Exec(ctx, `DELETE FROM rag.improve_runs WHERE user_id = $1`, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) DeleteClientAccount(ctx context.Context, userID string) error {
	walkin, err := s.IsWalkinPlaceholderUser(ctx, userID)
	if err != nil {
		return err
	}
	if walkin {
		return ErrWalkinImmutable
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := anonymizeUserSupportTicketsExec(ctx, tx, userID); err != nil {
		return err
	}
	if err := purgeClientOwnedDataExec(ctx, tx, userID); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `
		DELETE FROM identity.users
		WHERE id = $1 AND role = 'client' AND COALESCE(is_walkin_placeholder, false) = false`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

// AcceptUserTerms horodate le consentement CGU/privacy (clients provisionnés / import).
func (s *Store) AcceptUserTerms(ctx context.Context, userID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE identity.users SET terms_accepted_at = COALESCE(terms_accepted_at, NOW())
		WHERE id = $1`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UserHasAcceptedTerms — true si terms_accepted_at est renseigné.
func (s *Store) UserHasAcceptedTerms(ctx context.Context, userID string) (bool, error) {
	var accepted bool
	err := s.pool.QueryRow(ctx, `
		SELECT terms_accepted_at IS NOT NULL FROM identity.users WHERE id = $1`, userID).Scan(&accepted)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, ErrNotFound
	}
	return accepted, err
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
