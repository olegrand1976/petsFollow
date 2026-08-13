package store

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EidReading is an audit row for Belgian eID prefill attempts (no raw NISS).
type EidReading struct {
	ID         string
	PracticeID string
	UserID     string
	Tool       string
	Success    bool
	FieldsRead []string
	NISSHash   string
	ErrorCode  string
}

// InsertEidReading stores a hashed audit trail for an eID read attempt.
func (s *Store) InsertEidReading(ctx context.Context, in EidReading) error {
	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = uuid.NewString()
	}
	fields := in.FieldsRead
	if fields == nil {
		fields = []string{}
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO practice.eid_readings (
			id, practice_id, user_id, tool, success, fields_read, niss_hash, error_code
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, in.PracticeID, in.UserID, strings.TrimSpace(in.Tool), in.Success, fields,
		strings.TrimSpace(in.NISSHash), strings.TrimSpace(in.ErrorCode),
	)
	return err
}

// PurgeOldEidReadings deletes audit rows older than cutoff (RGPD minimisation).
func (s *Store) PurgeOldEidReadings(ctx context.Context, cutoff time.Time, limit int) (int, error) {
	if limit < 1 {
		limit = 500
	}
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM practice.eid_readings
		WHERE id IN (
			SELECT id FROM practice.eid_readings WHERE created_at < $1 ORDER BY created_at LIMIT $2
		)`, cutoff, limit)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}
