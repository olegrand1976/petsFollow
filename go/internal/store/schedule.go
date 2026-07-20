package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const defaultScheduleTZ = "Europe/Brussels"

type ScheduleSlot struct {
	ID        string `json:"id,omitempty"`
	Weekday   int    `json:"weekday"` // 0=Sunday … 6=Saturday (Go time.Weekday)
	StartTime string `json:"startTime"` // HH:MM
	EndTime   string `json:"endTime"`   // HH:MM
}

type VetSchedule struct {
	PracticeID             string         `json:"practiceId"`
	ClientBookingEnabled   bool           `json:"clientBookingEnabled"`
	SlotDurationMinutes    int            `json:"slotDurationMinutes"`
	VacationsDeclaredYear  *int           `json:"vacationsDeclaredYear,omitempty"`
	Timezone               string         `json:"timezone"`
	Slots                  []ScheduleSlot `json:"slots"`
	VacationsConfiguredForYear bool       `json:"vacationsConfiguredForYear"`
}

type Vacation struct {
	ID        string `json:"id"`
	PracticeID string `json:"practiceId"`
	StartsOn  string `json:"startsOn"` // YYYY-MM-DD
	EndsOn    string `json:"endsOn"`
	Label     string `json:"label,omitempty"`
}

type AvailableSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func (s *Store) GetVetSchedule(ctx context.Context, practiceID string) (VetSchedule, error) {
	var out VetSchedule
	out.PracticeID = practiceID
	out.SlotDurationMinutes = 30
	out.Timezone = defaultScheduleTZ
	var vacYear *int
	err := s.pool.QueryRow(ctx, `
		SELECT client_booking_enabled, slot_duration_minutes, vacations_declared_year, timezone
		FROM practice.vet_schedule WHERE practice_id = $1`, practiceID,
	).Scan(&out.ClientBookingEnabled, &out.SlotDurationMinutes, &vacYear, &out.Timezone)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return VetSchedule{}, err
	}
	out.VacationsDeclaredYear = vacYear
	year := time.Now().Year()
	if vacYear != nil && *vacYear >= year {
		out.VacationsConfiguredForYear = true
	} else {
		var vacCount int
		_ = s.pool.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM practice.vet_vacations
			WHERE practice_id = $1 AND EXTRACT(YEAR FROM starts_on) = $2`, practiceID, year,
		).Scan(&vacCount)
		out.VacationsConfiguredForYear = vacCount > 0
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id::text, weekday, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI')
		FROM practice.vet_schedule_slots WHERE practice_id = $1
		ORDER BY weekday, start_time`, practiceID)
	if err != nil {
		return VetSchedule{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var sl ScheduleSlot
		if err := rows.Scan(&sl.ID, &sl.Weekday, &sl.StartTime, &sl.EndTime); err != nil {
			return VetSchedule{}, err
		}
		out.Slots = append(out.Slots, sl)
	}
	if out.Slots == nil {
		out.Slots = []ScheduleSlot{}
	}
	return out, rows.Err()
}

func (s *Store) PutVetSchedule(ctx context.Context, practiceID string, clientBooking bool, duration int, vacYear *int, slots []ScheduleSlot) (VetSchedule, error) {
	if duration != 15 && duration != 30 && duration != 60 {
		return VetSchedule{}, fmt.Errorf("%w: invalid_duration", ErrValidation)
	}
	for _, sl := range slots {
		if sl.Weekday < 0 || sl.Weekday > 6 {
			return VetSchedule{}, fmt.Errorf("%w: invalid_weekday", ErrValidation)
		}
		if sl.StartTime == "" || sl.EndTime == "" || sl.StartTime >= sl.EndTime {
			return VetSchedule{}, fmt.Errorf("%w: invalid_slot_times", ErrValidation)
		}
	}
	if clientBooking && len(slots) == 0 {
		return VetSchedule{}, fmt.Errorf("%w: schedule_incomplete", ErrValidation)
	}
	if len(slots) == 0 {
		clientBooking = false
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return VetSchedule{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO practice.vet_schedule (practice_id, client_booking_enabled, slot_duration_minutes, vacations_declared_year, timezone, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (practice_id) DO UPDATE SET
			client_booking_enabled = EXCLUDED.client_booking_enabled,
			slot_duration_minutes = EXCLUDED.slot_duration_minutes,
			vacations_declared_year = EXCLUDED.vacations_declared_year,
			updated_at = NOW()`,
		practiceID, clientBooking, duration, vacYear, defaultScheduleTZ)
	if err != nil {
		return VetSchedule{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM practice.vet_schedule_slots WHERE practice_id = $1`, practiceID); err != nil {
		return VetSchedule{}, err
	}
	for _, sl := range slots {
		if _, err := tx.Exec(ctx, `
			INSERT INTO practice.vet_schedule_slots (id, practice_id, weekday, start_time, end_time)
			VALUES ($1, $2, $3, $4::time, $5::time)`,
			uuid.NewString(), practiceID, sl.Weekday, sl.StartTime, sl.EndTime); err != nil {
			return VetSchedule{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return VetSchedule{}, err
	}
	return s.GetVetSchedule(ctx, practiceID)
}

func (s *Store) ListVacations(ctx context.Context, practiceID string) ([]Vacation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, practice_id::text, to_char(starts_on, 'YYYY-MM-DD'), to_char(ends_on, 'YYYY-MM-DD'), COALESCE(label,'')
		FROM practice.vet_vacations WHERE practice_id = $1
		ORDER BY starts_on`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Vacation
	for rows.Next() {
		var v Vacation
		if err := rows.Scan(&v.ID, &v.PracticeID, &v.StartsOn, &v.EndsOn, &v.Label); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if out == nil {
		out = []Vacation{}
	}
	return out, rows.Err()
}

func (s *Store) CreateVacation(ctx context.Context, practiceID, startsOn, endsOn, label string) (Vacation, error) {
	id := uuid.NewString()
	var v Vacation
	err := s.pool.QueryRow(ctx, `
		INSERT INTO practice.vet_vacations (id, practice_id, starts_on, ends_on, label)
		VALUES ($1, $2, $3::date, $4::date, NULLIF($5,''))
		RETURNING id::text, practice_id::text, to_char(starts_on, 'YYYY-MM-DD'), to_char(ends_on, 'YYYY-MM-DD'), COALESCE(label,'')`,
		id, practiceID, startsOn, endsOn, label,
	).Scan(&v.ID, &v.PracticeID, &v.StartsOn, &v.EndsOn, &v.Label)
	return v, err
}

func (s *Store) DeleteVacation(ctx context.Context, practiceID, vacationID string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM practice.vet_vacations WHERE id = $1 AND practice_id = $2`, vacationID, practiceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListPracticeVisitsInRange(ctx context.Context, practiceID string, from, to time.Time) ([]Visit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id::text, v.pet_id::text, v.practice_id::text, v.scheduled_at, v.status,
			COALESCE(v.notes,''), v.source, v.created_at,
			COALESCE(p.name,''), COALESCE(u.full_name,''), p.owner_user_id::text,
			v.duration_minutes, v.proposed_scheduled_at, v.pending_action_by
		FROM visits.visits v
		JOIN pets.pets p ON p.id = v.pet_id
		JOIN identity.users u ON u.id = p.owner_user_id
		WHERE v.practice_id = $1
		  AND v.status IN ('requested', 'confirmed', 'reschedule_pending')
		  AND (
			(v.scheduled_at IS NOT NULL AND v.scheduled_at >= $2 AND v.scheduled_at < $3)
			OR (v.proposed_scheduled_at IS NOT NULL AND v.proposed_scheduled_at >= $2 AND v.proposed_scheduled_at < $3)
			OR (v.scheduled_at IS NULL AND v.created_at >= $2 AND v.created_at < $3 AND v.status = 'requested')
		  )
		ORDER BY COALESCE(v.scheduled_at, v.proposed_scheduled_at, v.created_at)`, practiceID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVisitsFull(rows)
}

func (s *Store) ClientBookingEnabled(ctx context.Context, practiceID string) (bool, int, error) {
	var enabled bool
	var duration int
	err := s.pool.QueryRow(ctx, `
		SELECT client_booking_enabled, slot_duration_minutes
		FROM practice.vet_schedule WHERE practice_id = $1`, practiceID,
	).Scan(&enabled, &duration)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, 30, nil
	}
	return enabled, duration, err
}

func (s *Store) IsOnVacation(ctx context.Context, practiceID string, day time.Time) (bool, error) {
	sched, err := s.GetVetSchedule(ctx, practiceID)
	if err != nil {
		return false, err
	}
	loc, err := time.LoadLocation(sched.Timezone)
	if err != nil {
		loc, _ = time.LoadLocation(defaultScheduleTZ)
	}
	localDay := day.In(loc).Format("2006-01-02")
	var n int
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM practice.vet_vacations
		WHERE practice_id = $1 AND starts_on <= $2::date AND ends_on >= $2::date`,
		practiceID, localDay,
	).Scan(&n)
	return n > 0, err
}

func (s *Store) HasVisitOverlap(ctx context.Context, practiceID string, start time.Time, durationMin int, excludeVisitID string) (bool, error) {
	end := start.Add(time.Duration(durationMin) * time.Minute)
	var n int
	// Busy interval = proposed_scheduled_at when reschedule_pending, else scheduled_at.
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM visits.visits
		WHERE practice_id = $1
		  AND status IN ('requested', 'confirmed', 'reschedule_pending')
		  AND ($4 = '' OR id::text <> $4)
		  AND COALESCE(proposed_scheduled_at, scheduled_at) IS NOT NULL
		  AND COALESCE(proposed_scheduled_at, scheduled_at) < $3
		  AND COALESCE(proposed_scheduled_at, scheduled_at)
		      + (COALESCE(duration_minutes, $5) || ' minutes')::interval > $2`,
		practiceID, start, end, excludeVisitID, durationMin,
	).Scan(&n)
	return n > 0, err
}


// ListAvailableSlots returns bookable slots between from and to (exclusive end day).
func (s *Store) ListAvailableSlots(ctx context.Context, practiceID string, from, to time.Time) ([]AvailableSlot, error) {
	sched, err := s.GetVetSchedule(ctx, practiceID)
	if err != nil {
		return nil, err
	}
	if !sched.ClientBookingEnabled || len(sched.Slots) == 0 {
		return []AvailableSlot{}, nil
	}
	loc, err := time.LoadLocation(sched.Timezone)
	if err != nil {
		loc, _ = time.LoadLocation(defaultScheduleTZ)
	}
	duration := time.Duration(sched.SlotDurationMinutes) * time.Minute
	vacations, err := s.ListVacations(ctx, practiceID)
	if err != nil {
		return nil, err
	}
	vacSet := map[string]bool{}
	for _, v := range vacations {
		start, _ := time.ParseInLocation("2006-01-02", v.StartsOn, loc)
		end, _ := time.ParseInLocation("2006-01-02", v.EndsOn, loc)
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			vacSet[d.Format("2006-01-02")] = true
		}
	}

	busy, err := s.ListPracticeVisitsInRange(ctx, practiceID, from, to)
	if err != nil {
		return nil, err
	}

	var out []AvailableSlot
	now := time.Now().In(loc)
	for day := time.Date(from.In(loc).Year(), from.In(loc).Month(), from.In(loc).Day(), 0, 0, 0, 0, loc); day.Before(to); day = day.AddDate(0, 0, 1) {
		if vacSet[day.Format("2006-01-02")] {
			continue
		}
		wd := int(day.Weekday())
		for _, sl := range sched.Slots {
			if sl.Weekday != wd {
				continue
			}
			sh, sm := parseHM(sl.StartTime)
			eh, em := parseHM(sl.EndTime)
			slotStart := time.Date(day.Year(), day.Month(), day.Day(), sh, sm, 0, 0, loc)
			slotEnd := time.Date(day.Year(), day.Month(), day.Day(), eh, em, 0, 0, loc)
			for t := slotStart; !t.Add(duration).After(slotEnd); t = t.Add(duration) {
				if t.Before(now) {
					continue
				}
				endT := t.Add(duration)
				if overlapsBusy(busy, t, endT, sched.SlotDurationMinutes) {
					continue
				}
				out = append(out, AvailableSlot{Start: t.UTC(), End: endT.UTC()})
			}
		}
	}
	return out, nil
}

func parseHM(s string) (h, m int) {
	var hh, mm int
	fmt.Sscanf(s, "%d:%d", &hh, &mm)
	return hh, mm
}

func overlapsBusy(visits []Visit, start, end time.Time, defaultDur int) bool {
	for _, v := range visits {
		busyAt := v.ScheduledAt
		if v.Status == "reschedule_pending" && v.ProposedScheduledAt != nil {
			busyAt = v.ProposedScheduledAt
		}
		if busyAt == nil {
			continue
		}
		dur := defaultDur
		if v.DurationMinutes != nil {
			dur = *v.DurationMinutes
		}
		vs := busyAt.UTC()
		ve := vs.Add(time.Duration(dur) * time.Minute)
		if start.Before(ve) && end.After(vs) {
			return true
		}
	}
	return false
}

func (s *Store) ListVetsForVisitAlert(ctx context.Context, practiceID, clientUserID string) ([]User, error) {
	var vetID string
	err := s.pool.QueryRow(ctx, `
		SELECT vet_user_id::text FROM practice.practice_clients
		WHERE practice_id = $1 AND client_user_id = $2`, practiceID, clientUserID,
	).Scan(&vetID)
	if err == nil {
		u, err := s.GetUserByID(ctx, vetID)
		if err != nil {
			return nil, err
		}
		return []User{u}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT `+userSelectCols+` FROM identity.users WHERE practice_id = $1 AND role = 'vet'`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
