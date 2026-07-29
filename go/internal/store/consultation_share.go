package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	ConsultationShareTTL        = 24 * time.Hour
	MaxConsultationSharesPerDay = 10
)

type ConsultationShareToken struct {
	ID                string
	Token             string
	VisitID           string
	PetID             string
	OwnerUserID       string
	RecipientEmail    string
	Locale            string
	CommercialUserID  string
	CommercialName    string
	CommercialPhone   string
	RegisterURL       string
	ExpiresAt         time.Time
	CreatedAt         time.Time
	FirstDownloadedAt *time.Time
	DownloadCount     int
	ObjectKey         string
	CachedAt          *time.Time
	PetName           string
}

type CreateConsultationShareInput struct {
	VisitID          string
	PetID            string
	OwnerUserID      string
	RecipientEmail   string
	Locale           string
	CommercialUserID string
	CommercialName   string
	CommercialPhone  string
	RegisterURL      string
}

// CreateConsultationShareToken inserts a share under an advisory lock and enforces MaxConsultationSharesPerDay.
func (s *Store) CreateConsultationShareToken(ctx context.Context, in CreateConsultationShareInput) (ConsultationShareToken, error) {
	email := strings.TrimSpace(strings.ToLower(in.RecipientEmail))
	if email == "" || !strings.Contains(email, "@") {
		return ConsultationShareToken{}, ErrValidation
	}
	locale := strings.TrimSpace(in.Locale)
	if locale == "" {
		locale = "fr"
	}
	var commercialID *string
	if strings.TrimSpace(in.CommercialUserID) != "" {
		id := strings.TrimSpace(in.CommercialUserID)
		commercialID = &id
	}
	t := ConsultationShareToken{
		ID:              uuid.NewString(),
		Token:           uuid.NewString(),
		VisitID:         in.VisitID,
		PetID:           in.PetID,
		OwnerUserID:     in.OwnerUserID,
		RecipientEmail:  email,
		Locale:          locale,
		CommercialName:  strings.TrimSpace(in.CommercialName),
		CommercialPhone: strings.TrimSpace(in.CommercialPhone),
		RegisterURL:     strings.TrimSpace(in.RegisterURL),
		ExpiresAt:       time.Now().UTC().Add(ConsultationShareTTL),
	}
	if commercialID != nil {
		t.CommercialUserID = *commercialID
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ConsultationShareToken{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "consultation-share:"+in.OwnerUserID); err != nil {
		return ConsultationShareToken{}, err
	}
	var n int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM pets.consultation_share_tokens
		WHERE owner_user_id = $1 AND created_at >= date_trunc('day', NOW() AT TIME ZONE 'UTC')`,
		in.OwnerUserID).Scan(&n); err != nil {
		return ConsultationShareToken{}, err
	}
	if n >= MaxConsultationSharesPerDay {
		return ConsultationShareToken{}, ErrConflict
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO pets.consultation_share_tokens (
			id, token, visit_id, pet_id, owner_user_id, recipient_email, locale,
			commercial_user_id, commercial_name, commercial_phone, register_url, expires_at
		) VALUES (
			$1, $2, $3::uuid, $4::uuid, $5::uuid, $6, $7,
			$8, $9, $10, $11, $12
		)
		RETURNING created_at`,
		t.ID, t.Token, t.VisitID, t.PetID, t.OwnerUserID, t.RecipientEmail, t.Locale,
		commercialID, t.CommercialName, t.CommercialPhone, t.RegisterURL, t.ExpiresAt,
	).Scan(&t.CreatedAt)
	if err != nil {
		return ConsultationShareToken{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ConsultationShareToken{}, err
	}
	return t, nil
}

func (s *Store) GetConsultationShareByToken(ctx context.Context, token string) (ConsultationShareToken, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return ConsultationShareToken{}, ErrNotFound
	}
	var t ConsultationShareToken
	var commercialID string
	err := s.pool.QueryRow(ctx, `
		SELECT t.id::text, t.token, t.visit_id::text, t.pet_id::text, t.owner_user_id::text, t.recipient_email,
			t.locale, COALESCE(t.commercial_user_id::text,''), t.commercial_name, t.commercial_phone,
			t.register_url, t.expires_at, t.created_at, t.first_downloaded_at, t.download_count,
			COALESCE(t.object_key,''), t.cached_at, COALESCE(p.name,'')
		FROM pets.consultation_share_tokens t
		JOIN pets.pets p ON p.id = t.pet_id
		WHERE t.token = $1`, token).Scan(
		&t.ID, &t.Token, &t.VisitID, &t.PetID, &t.OwnerUserID, &t.RecipientEmail,
		&t.Locale, &commercialID, &t.CommercialName, &t.CommercialPhone,
		&t.RegisterURL, &t.ExpiresAt, &t.CreatedAt, &t.FirstDownloadedAt, &t.DownloadCount,
		&t.ObjectKey, &t.CachedAt, &t.PetName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ConsultationShareToken{}, ErrNotFound
	}
	if err != nil {
		return ConsultationShareToken{}, err
	}
	t.CommercialUserID = commercialID
	return t, nil
}

func (s *Store) LockConsultationShareForDownload(ctx context.Context, tx pgx.Tx, token string) (ConsultationShareToken, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return ConsultationShareToken{}, ErrNotFound
	}
	var t ConsultationShareToken
	var commercialID string
	err := tx.QueryRow(ctx, `
		SELECT t.id::text, t.token, t.visit_id::text, t.pet_id::text, t.owner_user_id::text, t.recipient_email,
			t.locale, COALESCE(t.commercial_user_id::text,''), t.commercial_name, t.commercial_phone,
			t.register_url, t.expires_at, t.created_at, t.first_downloaded_at, t.download_count,
			COALESCE(t.object_key,''), t.cached_at, COALESCE(p.name,'')
		FROM pets.consultation_share_tokens t
		JOIN pets.pets p ON p.id = t.pet_id
		WHERE t.token = $1
		FOR UPDATE OF t`, token).Scan(
		&t.ID, &t.Token, &t.VisitID, &t.PetID, &t.OwnerUserID, &t.RecipientEmail,
		&t.Locale, &commercialID, &t.CommercialName, &t.CommercialPhone,
		&t.RegisterURL, &t.ExpiresAt, &t.CreatedAt, &t.FirstDownloadedAt, &t.DownloadCount,
		&t.ObjectKey, &t.CachedAt, &t.PetName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ConsultationShareToken{}, ErrNotFound
	}
	if err != nil {
		return ConsultationShareToken{}, err
	}
	t.CommercialUserID = commercialID
	return t, nil
}

func (s *Store) MarkConsultationShareDownloadedTx(ctx context.Context, tx pgx.Tx, id, objectKey string) error {
	_, err := tx.Exec(ctx, `
		UPDATE pets.consultation_share_tokens SET
			first_downloaded_at = COALESCE(first_downloaded_at, NOW()),
			download_count = download_count + 1,
			object_key = CASE WHEN $2 <> '' THEN $2 ELSE object_key END,
			cached_at = CASE WHEN $2 <> '' THEN NOW() ELSE cached_at END
		WHERE id = $1`, id, strings.TrimSpace(objectKey))
	return err
}

func (s *Store) DeleteConsultationShareByID(ctx context.Context, id string) (objectKey string, err error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", ErrValidation
	}
	err = s.pool.QueryRow(ctx, `
		DELETE FROM pets.consultation_share_tokens
		WHERE id = $1
		RETURNING COALESCE(object_key, '')`, id).Scan(&objectKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return strings.TrimSpace(objectKey), err
}

// PurgeExpiredConsultationShares deletes expired share rows and returns cached PDF keys.
func (s *Store) PurgeExpiredConsultationShares(ctx context.Context, ownerUserID string, limit int) ([]string, int, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := s.pool.Query(ctx, `
		DELETE FROM pets.consultation_share_tokens t
		USING (
			SELECT id FROM pets.consultation_share_tokens
			WHERE expires_at < NOW()
			  AND owner_user_id = COALESCE(NULLIF($1, '')::uuid, owner_user_id)
			ORDER BY expires_at ASC
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		) d
		WHERE t.id = d.id
		RETURNING COALESCE(t.object_key, '')`, strings.TrimSpace(ownerUserID), limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var keys []string
	deleted := 0
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, 0, err
		}
		deleted++
		if strings.TrimSpace(k) != "" {
			keys = append(keys, k)
		}
	}
	return keys, deleted, rows.Err()
}

// ConsultationShareCacheStale is true when a final report was finalized after the cached PDF.
func (s *Store) ConsultationShareCacheStale(ctx context.Context, visitID string, cachedAt *time.Time) (bool, error) {
	if cachedAt == nil || cachedAt.IsZero() {
		return true, nil
	}
	var newer bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM visits.visit_reports
			WHERE visit_id = $1::uuid AND status = 'final'
			  AND finalized_at IS NOT NULL AND finalized_at > $2
		)`, visitID, *cachedAt).Scan(&newer)
	return newer, err
}

func (s *Store) ClearConsultationShareObjectKeyTx(ctx context.Context, tx pgx.Tx, id string) error {
	_, err := tx.Exec(ctx, `
		UPDATE pets.consultation_share_tokens
		SET object_key = '', cached_at = NULL
		WHERE id = $1`, id)
	return err
}
