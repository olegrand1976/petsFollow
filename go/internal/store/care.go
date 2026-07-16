package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CareReminder struct {
	ID         string    `json:"id"`
	PetID      string    `json:"petId"`
	PracticeID string    `json:"practiceId"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	DueAt      time.Time `json:"dueAt"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type careReminderTemplate struct {
	Type  string
	Title string
	Days  int
}

var defaultCareBySpecies = map[string][]careReminderTemplate{
	"dog": {
		{Type: "vaccination", Title: "Vaccination", Days: 365},
		{Type: "deworming", Title: "Vermifuge", Days: 90},
		{Type: "vet_check", Title: "Contrôle vétérinaire", Days: 180},
		{Type: "dental", Title: "Soins dentaires", Days: 365},
	},
	"cat": {
		{Type: "vaccination", Title: "Vaccination", Days: 365},
		{Type: "deworming", Title: "Vermifuge", Days: 90},
		{Type: "vet_check", Title: "Contrôle vétérinaire", Days: 180},
		{Type: "dental", Title: "Soins dentaires", Days: 365},
	},
	"horse": {
		{Type: "vaccination", Title: "Vaccination", Days: 365},
		{Type: "deworming", Title: "Vermifuge", Days: 90},
		{Type: "vet_check", Title: "Contrôle vétérinaire", Days: 180},
		{Type: "dental", Title: "Soins dentaires", Days: 365},
		{Type: "farrier", Title: "Maréchal-ferrant", Days: 42},
		{Type: "fecal_egg", Title: "Coproscopie", Days: 90},
	},
}

func (s *Store) SeedDefaultCareReminders(ctx context.Context, petID, practiceID, species string) error {
	templates, ok := defaultCareBySpecies[species]
	if !ok {
		templates = defaultCareBySpecies["dog"]
	}
	now := time.Now()
	for _, tpl := range templates {
		if err := s.insertCareReminder(ctx, petID, practiceID, tpl.Type, tpl.Title, now.AddDate(0, 0, tpl.Days)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) insertCareReminder(ctx context.Context, petID, practiceID, reminderType, title string, dueAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO care.reminders (id, pet_id, practice_id, type, title, due_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')`,
		uuid.NewString(), petID, practiceID, reminderType, title, dueAt)
	return err
}

func (s *Store) ListCareReminders(ctx context.Context, petID string) ([]CareReminder, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, pet_id::text, practice_id::text, type, title, due_at, status, created_at, updated_at
		FROM care.reminders WHERE pet_id = $1
		ORDER BY due_at ASC`, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCareReminders(rows)
}

func scanCareReminders(rows pgx.Rows) ([]CareReminder, error) {
	var out []CareReminder
	for rows.Next() {
		var r CareReminder
		if err := rows.Scan(&r.ID, &r.PetID, &r.PracticeID, &r.Type, &r.Title, &r.DueAt, &r.Status, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) CreateCareReminder(ctx context.Context, petID, practiceID, reminderType, title string, dueAt time.Time) (CareReminder, error) {
	id := uuid.NewString()
	var r CareReminder
	err := s.pool.QueryRow(ctx, `
		INSERT INTO care.reminders (id, pet_id, practice_id, type, title, due_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
		RETURNING id::text, pet_id::text, practice_id::text, type, title, due_at, status, created_at, updated_at`,
		id, petID, practiceID, reminderType, title, dueAt,
	).Scan(&r.ID, &r.PetID, &r.PracticeID, &r.Type, &r.Title, &r.DueAt, &r.Status, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func (s *Store) GetCareReminder(ctx context.Context, id string) (CareReminder, error) {
	var r CareReminder
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, pet_id::text, practice_id::text, type, title, due_at, status, created_at, updated_at
		FROM care.reminders WHERE id = $1`, id,
	).Scan(&r.ID, &r.PetID, &r.PracticeID, &r.Type, &r.Title, &r.DueAt, &r.Status, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return CareReminder{}, ErrNotFound
	}
	return r, err
}

func (s *Store) MarkCareReminderDone(ctx context.Context, id, ownerUserID string) (CareReminder, error) {
	var r CareReminder
	err := s.pool.QueryRow(ctx, `
		UPDATE care.reminders cr SET status = 'done', updated_at = NOW()
		FROM pets.pets p
		WHERE cr.id = $1 AND cr.pet_id = p.id AND p.owner_user_id = $2 AND cr.status = 'pending'
		RETURNING cr.id::text, cr.pet_id::text, cr.practice_id::text, cr.type, cr.title, cr.due_at, cr.status, cr.created_at, cr.updated_at`,
		id, ownerUserID,
	).Scan(&r.ID, &r.PetID, &r.PracticeID, &r.Type, &r.Title, &r.DueAt, &r.Status, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return CareReminder{}, ErrNotFound
	}
	return r, err
}

func (s *Store) MarkCareReminderDoneByPractice(ctx context.Context, id, practiceID string) (CareReminder, error) {
	var r CareReminder
	err := s.pool.QueryRow(ctx, `
		UPDATE care.reminders
		SET status = 'done', updated_at = NOW()
		WHERE id = $1 AND practice_id = $2 AND status = 'pending'
		RETURNING id::text, pet_id::text, practice_id::text, type, title, due_at, status, created_at, updated_at`,
		id, practiceID,
	).Scan(&r.ID, &r.PetID, &r.PracticeID, &r.Type, &r.Title, &r.DueAt, &r.Status, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return CareReminder{}, ErrNotFound
	}
	return r, err
}

func (s *Store) PostponeCareReminder(ctx context.Context, id, ownerUserID string, days int) (CareReminder, error) {
	var r CareReminder
	err := s.pool.QueryRow(ctx, `
		UPDATE care.reminders cr SET due_at = due_at + ($3 || ' days')::interval, updated_at = NOW()
		FROM pets.pets p
		WHERE cr.id = $1 AND cr.pet_id = p.id AND p.owner_user_id = $2 AND cr.status = 'pending'
		RETURNING cr.id::text, cr.pet_id::text, cr.practice_id::text, cr.type, cr.title, cr.due_at, cr.status, cr.created_at, cr.updated_at`,
		id, ownerUserID, days,
	).Scan(&r.ID, &r.PetID, &r.PracticeID, &r.Type, &r.Title, &r.DueAt, &r.Status, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return CareReminder{}, ErrNotFound
	}
	return r, err
}

func (s *Store) PostponeCareReminderByPractice(ctx context.Context, id, practiceID string, days int) (CareReminder, error) {
	var r CareReminder
	err := s.pool.QueryRow(ctx, `
		UPDATE care.reminders
		SET due_at = due_at + ($3 || ' days')::interval, updated_at = NOW()
		WHERE id = $1 AND practice_id = $2 AND status = 'pending'
		RETURNING id::text, pet_id::text, practice_id::text, type, title, due_at, status, created_at, updated_at`,
		id, practiceID, days,
	).Scan(&r.ID, &r.PetID, &r.PracticeID, &r.Type, &r.Title, &r.DueAt, &r.Status, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return CareReminder{}, ErrNotFound
	}
	return r, err
}

func (s *Store) ListOverdueCareReminders(ctx context.Context, practiceID string) ([]CareReminder, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, pet_id::text, practice_id::text, type, title, due_at, status, created_at, updated_at
		FROM care.reminders
		WHERE practice_id = $1 AND status = 'pending' AND due_at < NOW()
		ORDER BY due_at ASC
		LIMIT 100`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCareReminders(rows)
}
