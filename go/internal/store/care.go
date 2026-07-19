package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CareReminder struct {
	ID              string    `json:"id"`
	PetID           string    `json:"petId"`
	PracticeID      string    `json:"practiceId"`
	Type            string    `json:"type"`
	Title           string    `json:"title"`
	DueAt           time.Time `json:"dueAt"`
	Status          string    `json:"status"`
	Notes           string    `json:"notes,omitempty"`
	RecurrenceDays  *int      `json:"recurrenceDays,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
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
		// farrier / fecal_egg require Horse pack (SeedHorsePackReminders).
	},
}

var horsePackCareTemplates = []careReminderTemplate{
	{Type: "farrier", Title: "Maréchal-ferrant", Days: 42},
	{Type: "fecal_egg", Title: "Coproscopie", Days: 365},
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

// SeedHorsePackReminders adds farrier/copro reminders for all horse pets of an owner (idempotent by type).
func (s *Store) SeedHorsePackReminders(ctx context.Context, ownerUserID string) error {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, practice_id::text FROM pets.pets
		WHERE owner_user_id=$1 AND species='horse'`, ownerUserID)
	if err != nil {
		return err
	}
	defer rows.Close()
	now := time.Now()
	for rows.Next() {
		var petID, practiceID string
		if err := rows.Scan(&petID, &practiceID); err != nil {
			return err
		}
		for _, tpl := range horsePackCareTemplates {
			var exists bool
			if err := s.pool.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM care.reminders
					WHERE pet_id=$1 AND type=$2 AND status='pending'
				)`, petID, tpl.Type).Scan(&exists); err != nil {
				return err
			}
			if exists {
				continue
			}
			if err := s.insertCareReminder(ctx, petID, practiceID, tpl.Type, tpl.Title, now.AddDate(0, 0, tpl.Days)); err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (s *Store) insertCareReminder(ctx context.Context, petID, practiceID, reminderType, title string, dueAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO care.reminders (id, pet_id, practice_id, type, title, due_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')`,
		uuid.NewString(), petID, practiceID, reminderType, title, dueAt)
	return err
}

const careReminderSelect = `
	id::text, pet_id::text, practice_id::text, type, title, due_at, status,
	COALESCE(notes,''), recurrence_days, created_at, updated_at`

func (s *Store) ListCareReminders(ctx context.Context, petID string) ([]CareReminder, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+careReminderSelect+`
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
		if err := rows.Scan(&r.ID, &r.PetID, &r.PracticeID, &r.Type, &r.Title, &r.DueAt, &r.Status,
			&r.Notes, &r.RecurrenceDays, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) CreateCareReminder(ctx context.Context, petID, practiceID, reminderType, title string, dueAt time.Time) (CareReminder, error) {
	return s.CreateCareReminderFull(ctx, petID, practiceID, reminderType, title, dueAt, "", nil)
}

func (s *Store) CreateCareReminderFull(ctx context.Context, petID, practiceID, reminderType, title string, dueAt time.Time, notes string, recurrenceDays *int) (CareReminder, error) {
	id := uuid.NewString()
	var r CareReminder
	err := s.pool.QueryRow(ctx, `
		INSERT INTO care.reminders (id, pet_id, practice_id, type, title, due_at, status, notes, recurrence_days)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, $8)
		RETURNING `+careReminderSelect,
		id, petID, practiceID, reminderType, title, dueAt, notes, recurrenceDays,
	).Scan(&r.ID, &r.PetID, &r.PracticeID, &r.Type, &r.Title, &r.DueAt, &r.Status,
		&r.Notes, &r.RecurrenceDays, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func (s *Store) GetCareReminder(ctx context.Context, id string) (CareReminder, error) {
	var r CareReminder
	err := s.pool.QueryRow(ctx, `
		SELECT `+careReminderSelect+` FROM care.reminders WHERE id = $1`, id,
	).Scan(&r.ID, &r.PetID, &r.PracticeID, &r.Type, &r.Title, &r.DueAt, &r.Status,
		&r.Notes, &r.RecurrenceDays, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return CareReminder{}, ErrNotFound
	}
	return r, err
}

func scanCareReminderRow(r *CareReminder) []any {
	return []any{&r.ID, &r.PetID, &r.PracticeID, &r.Type, &r.Title, &r.DueAt, &r.Status,
		&r.Notes, &r.RecurrenceDays, &r.CreatedAt, &r.UpdatedAt}
}

func (s *Store) MarkCareReminderDone(ctx context.Context, id, ownerUserID string) (CareReminder, error) {
	var r CareReminder
	err := s.pool.QueryRow(ctx, `
		UPDATE care.reminders cr SET status = 'done', updated_at = NOW()
		FROM pets.pets p
		WHERE cr.id = $1 AND cr.pet_id = p.id AND p.owner_user_id = $2 AND cr.status = 'pending'
		RETURNING cr.id::text, cr.pet_id::text, cr.practice_id::text, cr.type, cr.title, cr.due_at, cr.status,
			COALESCE(cr.notes,''), cr.recurrence_days, cr.created_at, cr.updated_at`,
		id, ownerUserID,
	).Scan(scanCareReminderRow(&r)...)
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
		RETURNING `+careReminderSelect,
		id, practiceID,
	).Scan(scanCareReminderRow(&r)...)
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
		RETURNING cr.id::text, cr.pet_id::text, cr.practice_id::text, cr.type, cr.title, cr.due_at, cr.status,
			COALESCE(cr.notes,''), cr.recurrence_days, cr.created_at, cr.updated_at`,
		id, ownerUserID, days,
	).Scan(scanCareReminderRow(&r)...)
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
		RETURNING `+careReminderSelect,
		id, practiceID, days,
	).Scan(scanCareReminderRow(&r)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return CareReminder{}, ErrNotFound
	}
	return r, err
}

// ListCarePlusEmailCandidates returns pending Care+ owner reminders due within daysAhead (J-3 / J0 hook).
// Callers (cron / notifier) should send email; this is the query side only.
func (s *Store) ListCarePlusEmailCandidates(ctx context.Context, daysAhead int) ([]CareReminder, error) {
	if daysAhead < 0 {
		daysAhead = 0
	}
	rows, err := s.pool.Query(ctx, `
		SELECT cr.id::text, cr.pet_id::text, cr.practice_id::text, cr.type, cr.title, cr.due_at, cr.status,
			COALESCE(cr.notes,''), cr.recurrence_days, cr.created_at, cr.updated_at
		FROM care.reminders cr
		JOIN pets.pets p ON p.id = cr.pet_id
		JOIN billing.addon_entitlements ae ON ae.owner_user_id = p.owner_user_id
			AND ae.addon_code = 'care_plus' AND ae.status = 'active'
			AND (ae.valid_until IS NULL OR ae.valid_until > NOW())
		WHERE cr.status = 'pending'
			AND cr.due_at::date <= (CURRENT_DATE + ($1 || ' days')::interval)::date
			AND cr.due_at::date >= CURRENT_DATE
		ORDER BY cr.due_at ASC
		LIMIT 500`, daysAhead)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCareReminders(rows)
}

func (s *Store) ListOverdueCareReminders(ctx context.Context, practiceID string) ([]CareReminder, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+careReminderSelect+`
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
