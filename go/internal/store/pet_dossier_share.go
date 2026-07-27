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
	DossierShareTTL        = 24 * time.Hour
	MaxDossierSharesPerDay = 10
)

type DossierShareToken struct {
	ID                string
	Token             string
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
	// Joined for public meta
	PetName string
}

type CreateDossierShareInput struct {
	PetID            string
	OwnerUserID      string
	RecipientEmail   string
	Locale           string
	CommercialUserID string
	CommercialName   string
	CommercialPhone  string
	RegisterURL      string
}

type DossierVisitRow struct {
	ID            string
	ScheduledAt   *time.Time
	Status        string
	Notes         string
	Source        string
	CreatedAt     time.Time
	PracticeName  string
	ProConsulted  string
	ReportExcerpt string
}

// CreateDossierShareToken inserts a share under an advisory lock and enforces MaxDossierSharesPerDay.
func (s *Store) CreateDossierShareToken(ctx context.Context, in CreateDossierShareInput) (DossierShareToken, error) {
	email := strings.TrimSpace(strings.ToLower(in.RecipientEmail))
	if email == "" || !strings.Contains(email, "@") {
		return DossierShareToken{}, ErrValidation
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
	t := DossierShareToken{
		ID:              uuid.NewString(),
		Token:           uuid.NewString(),
		PetID:           in.PetID,
		OwnerUserID:     in.OwnerUserID,
		RecipientEmail:  email,
		Locale:          locale,
		CommercialName:  strings.TrimSpace(in.CommercialName),
		CommercialPhone: strings.TrimSpace(in.CommercialPhone),
		RegisterURL:     strings.TrimSpace(in.RegisterURL),
		ExpiresAt:       time.Now().UTC().Add(DossierShareTTL),
	}
	if commercialID != nil {
		t.CommercialUserID = *commercialID
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return DossierShareToken{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "dossier-share:"+in.OwnerUserID); err != nil {
		return DossierShareToken{}, err
	}
	var n int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM pets.dossier_share_tokens
		WHERE owner_user_id = $1 AND created_at >= date_trunc('day', NOW() AT TIME ZONE 'UTC')`,
		in.OwnerUserID).Scan(&n); err != nil {
		return DossierShareToken{}, err
	}
	if n >= MaxDossierSharesPerDay {
		return DossierShareToken{}, ErrConflict
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO pets.dossier_share_tokens (
			id, token, pet_id, owner_user_id, recipient_email, locale,
			commercial_user_id, commercial_name, commercial_phone, register_url, expires_at
		) VALUES (
			$1, $2, $3::uuid, $4::uuid, $5, $6,
			$7, $8, $9, $10, $11
		)
		RETURNING created_at`,
		t.ID, t.Token, t.PetID, t.OwnerUserID, t.RecipientEmail, t.Locale,
		commercialID, t.CommercialName, t.CommercialPhone, t.RegisterURL, t.ExpiresAt,
	).Scan(&t.CreatedAt)
	if err != nil {
		return DossierShareToken{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return DossierShareToken{}, err
	}
	return t, nil
}

func (s *Store) GetDossierShareByToken(ctx context.Context, token string) (DossierShareToken, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return DossierShareToken{}, ErrNotFound
	}
	var t DossierShareToken
	var commercialID string
	err := s.pool.QueryRow(ctx, `
		SELECT t.id::text, t.token, t.pet_id::text, t.owner_user_id::text, t.recipient_email,
			t.locale, COALESCE(t.commercial_user_id::text,''), t.commercial_name, t.commercial_phone,
			t.register_url, t.expires_at, t.created_at, t.first_downloaded_at, t.download_count,
			COALESCE(t.object_key,''), t.cached_at, COALESCE(p.name,'')
		FROM pets.dossier_share_tokens t
		JOIN pets.pets p ON p.id = t.pet_id
		WHERE t.token = $1`, token).Scan(
		&t.ID, &t.Token, &t.PetID, &t.OwnerUserID, &t.RecipientEmail,
		&t.Locale, &commercialID, &t.CommercialName, &t.CommercialPhone,
		&t.RegisterURL, &t.ExpiresAt, &t.CreatedAt, &t.FirstDownloadedAt, &t.DownloadCount,
		&t.ObjectKey, &t.CachedAt, &t.PetName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DossierShareToken{}, ErrNotFound
	}
	if err != nil {
		return DossierShareToken{}, err
	}
	t.CommercialUserID = commercialID
	return t, nil
}

// LockDossierShareForDownload locks the row for the transaction lifetime (caller must commit/rollback).
// Prevents concurrent ZIP rebuilds for the same token.
func (s *Store) LockDossierShareForDownload(ctx context.Context, tx pgx.Tx, token string) (DossierShareToken, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return DossierShareToken{}, ErrNotFound
	}
	var t DossierShareToken
	var commercialID string
	err := tx.QueryRow(ctx, `
		SELECT t.id::text, t.token, t.pet_id::text, t.owner_user_id::text, t.recipient_email,
			t.locale, COALESCE(t.commercial_user_id::text,''), t.commercial_name, t.commercial_phone,
			t.register_url, t.expires_at, t.created_at, t.first_downloaded_at, t.download_count,
			COALESCE(t.object_key,''), t.cached_at, COALESCE(p.name,'')
		FROM pets.dossier_share_tokens t
		JOIN pets.pets p ON p.id = t.pet_id
		WHERE t.token = $1
		FOR UPDATE OF t`, token).Scan(
		&t.ID, &t.Token, &t.PetID, &t.OwnerUserID, &t.RecipientEmail,
		&t.Locale, &commercialID, &t.CommercialName, &t.CommercialPhone,
		&t.RegisterURL, &t.ExpiresAt, &t.CreatedAt, &t.FirstDownloadedAt, &t.DownloadCount,
		&t.ObjectKey, &t.CachedAt, &t.PetName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DossierShareToken{}, ErrNotFound
	}
	if err != nil {
		return DossierShareToken{}, err
	}
	t.CommercialUserID = commercialID
	return t, nil
}

func (s *Store) MarkDossierShareDownloadedTx(ctx context.Context, tx pgx.Tx, id, objectKey string) error {
	_, err := tx.Exec(ctx, `
		UPDATE pets.dossier_share_tokens SET
			first_downloaded_at = COALESCE(first_downloaded_at, NOW()),
			download_count = download_count + 1,
			object_key = CASE WHEN $2 <> '' THEN $2 ELSE object_key END,
			cached_at = CASE WHEN $2 <> '' THEN NOW() ELSE cached_at END
		WHERE id = $1`, id, strings.TrimSpace(objectKey))
	return err
}

func (s *Store) MarkDossierShareDownloaded(ctx context.Context, id, objectKey string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE pets.dossier_share_tokens SET
			first_downloaded_at = COALESCE(first_downloaded_at, NOW()),
			download_count = download_count + 1,
			object_key = CASE WHEN $2 <> '' THEN $2 ELSE object_key END,
			cached_at = CASE WHEN $2 <> '' THEN NOW() ELSE cached_at END
		WHERE id = $1`, id, strings.TrimSpace(objectKey))
	return err
}

// TakeDossierShareObjectKey returns the current object_key and clears it atomically.
func (s *Store) TakeDossierShareObjectKey(ctx context.Context, id string) (string, error) {
	var key string
	err := s.pool.QueryRow(ctx, `
		WITH prev AS (
			SELECT id, object_key FROM pets.dossier_share_tokens
			WHERE id = $1 AND COALESCE(object_key,'') <> ''
			FOR UPDATE
		)
		UPDATE pets.dossier_share_tokens t
		SET object_key = '', cached_at = NULL
		FROM prev
		WHERE t.id = prev.id
		RETURNING prev.object_key`, id).Scan(&key)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return strings.TrimSpace(key), err
}

// PurgeExpiredDossierShareCaches clears object_key on expired rows and returns keys to delete from media.
func (s *Store) PurgeExpiredDossierShareCaches(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		WITH doomed AS (
			SELECT id, object_key FROM pets.dossier_share_tokens
			WHERE expires_at < NOW() AND COALESCE(object_key,'') <> ''
			ORDER BY expires_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE pets.dossier_share_tokens t
		SET object_key = '', cached_at = NULL
		FROM doomed d
		WHERE t.id = d.id
		RETURNING d.object_key`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		if strings.TrimSpace(k) != "" {
			keys = append(keys, k)
		}
	}
	return keys, rows.Err()
}

// ListVisitsForDossier returns visits with practice name and consulted pro (CR author or practice vet).
func (s *Store) ListVisitsForDossier(ctx context.Context, petID string) ([]DossierVisitRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id::text, v.scheduled_at, v.status, COALESCE(v.notes,''), v.source, v.created_at,
			COALESCE(pr.name, ''),
			COALESCE(
				(SELECT u.full_name FROM visits.visit_reports vr
				 JOIN identity.users u ON u.id = vr.author_user_id
				 WHERE vr.visit_id = v.id AND vr.status = 'finalized'
				 ORDER BY vr.finalized_at DESC NULLS LAST LIMIT 1),
				(SELECT u.full_name FROM identity.users u
				 WHERE u.practice_id = v.practice_id AND u.role = 'vet'
				 ORDER BY u.created_at ASC LIMIT 1),
				''
			),
			COALESCE(
				(SELECT LEFT(NULLIF(TRIM(COALESCE(vr.improved_text, vr.body_text)), ''), 280)
				 FROM visits.visit_reports vr
				 WHERE vr.visit_id = v.id AND vr.status = 'finalized'
				 ORDER BY vr.finalized_at DESC NULLS LAST LIMIT 1),
				''
			)
		FROM visits.visits v
		LEFT JOIN practice.practices pr ON pr.id = v.practice_id
		WHERE v.pet_id = $1
		ORDER BY COALESCE(v.scheduled_at, v.created_at) DESC`, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DossierVisitRow
	for rows.Next() {
		var r DossierVisitRow
		if err := rows.Scan(
			&r.ID, &r.ScheduledAt, &r.Status, &r.Notes, &r.Source, &r.CreatedAt,
			&r.PracticeName, &r.ProConsulted, &r.ReportExcerpt,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if out == nil {
		out = []DossierVisitRow{}
	}
	return out, rows.Err()
}
