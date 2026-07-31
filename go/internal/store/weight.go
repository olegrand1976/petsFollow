package store

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const MaxWeightCommentLen = 500
const MaxWeightKg = 999.99
const MinWeightKg = 0.01

// ErrInvalidInput is returned for domain validation failures (weight range, etc.).
var ErrInvalidInput = errors.New("invalid input")

type WeightReading struct {
	ID           string    `json:"id"`
	PetID        string    `json:"petId"`
	OwnerUserID  string    `json:"ownerUserId"`
	AuthorUserID string    `json:"authorUserId"`
	PracticeID   string    `json:"practiceId"`
	WeightKg     float64   `json:"weightKg"`
	Comment      *string   `json:"comment,omitempty"`
	RecordedAt   time.Time `json:"recordedAt"`
}

func NormalizeWeightComment(raw *string) *string {
	if raw == nil {
		return nil
	}
	s := strings.TrimSpace(*raw)
	if s == "" {
		return nil
	}
	if utf8.RuneCountInString(s) > MaxWeightCommentLen {
		s = string([]rune(s)[:MaxWeightCommentLen])
	}
	return &s
}

func (s *Store) CreateWeightReading(ctx context.Context, petID, ownerID, authorID, practiceID string, weightKg float64, comment *string) (WeightReading, error) {
	if weightKg < MinWeightKg || weightKg > MaxWeightKg {
		return WeightReading{}, ErrInvalidInput
	}
	comment = NormalizeWeightComment(comment)
	reading := WeightReading{
		ID:           uuid.NewString(),
		PetID:        petID,
		OwnerUserID:  ownerID,
		AuthorUserID: authorID,
		PracticeID:   practiceID,
		WeightKg:     weightKg,
		Comment:      comment,
		RecordedAt:   time.Now().UTC(),
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return WeightReading{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO pets.weight_readings (
			id, pet_id, owner_user_id, author_user_id, practice_id, weight_kg, comment, recorded_at
		) VALUES ($1,$2,$3,$4,NULLIF($5,'')::uuid,$6,$7,$8)
		RETURNING weight_kg::float8, recorded_at`,
		reading.ID, reading.PetID, reading.OwnerUserID, reading.AuthorUserID,
		reading.PracticeID, reading.WeightKg, reading.Comment, reading.RecordedAt,
	).Scan(&reading.WeightKg, &reading.RecordedAt)
	if err != nil {
		return WeightReading{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE pets.pets SET weight_kg = $2, updated_at = NOW() WHERE id = $1`,
		petID, reading.WeightKg); err != nil {
		return WeightReading{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WeightReading{}, err
	}
	return reading, nil
}

func (s *Store) ListWeightReadings(ctx context.Context, petID string) ([]WeightReading, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, pet_id::text, owner_user_id::text, author_user_id::text, practice_id::text,
			weight_kg::float8, comment, recorded_at
		FROM pets.weight_readings
		WHERE pet_id = $1
		ORDER BY recorded_at DESC`, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WeightReading
	for rows.Next() {
		var r WeightReading
		if err := rows.Scan(
			&r.ID, &r.PetID, &r.OwnerUserID, &r.AuthorUserID, &r.PracticeID,
			&r.WeightKg, &r.Comment, &r.RecordedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if out == nil {
		out = []WeightReading{}
	}
	return out, rows.Err()
}
