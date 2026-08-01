package store

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
)

const (
	ResearchSignalPreconsult    = "preconsult_syndrome"
	ResearchSignalVisitVolume   = "visit_volume"
	ResearchSignalLabFlag       = "lab_flag"
	ResearchSignalCarePrev      = "care_preventive"
	ResearchSignalAntibioticDAF = "antibiotic_daf"
	ResearchSignalHRAlert       = "hr_alert"

	// ResearchKAnonymity hides sparse postal×species×signal cells from the API.
	ResearchKAnonymity = 5
	researchETLBatch   = 2000
)

type ResearchOptInStatus struct {
	OptedIn   bool       `json:"optedIn"`
	OptedInAt *time.Time `json:"optedInAt,omitempty"`
}

type ResearchOverview struct {
	OptInPracticeCount int            `json:"optInPracticeCount"`
	EventCount         int            `json:"eventCount"`
	WeekCount          int            `json:"weekCount"`
	BySignal           map[string]int `json:"bySignal"`
	BySpecies          map[string]int `json:"bySpecies"`
	ByCountry          map[string]int `json:"byCountry"`
	LatestWeek         *string        `json:"latestWeek,omitempty"`
}

type ResearchHeatmapCell struct {
	PostalCode  string `json:"postalCode"`
	CountryCode string `json:"countryCode"`
	Species     string `json:"species"`
	SignalType  string `json:"signalType"`
	EventCount  int    `json:"eventCount"`
}

type ResearchTimeseriesPoint struct {
	Week       string `json:"week"`
	EventCount int    `json:"eventCount"`
}

type ResearchAlert struct {
	PostalCode  string  `json:"postalCode"`
	CountryCode string  `json:"countryCode"`
	Species     string  `json:"species"`
	SignalType  string  `json:"signalType"`
	Week        string  `json:"week"`
	EventCount  int     `json:"eventCount"`
	BaselineAvg float64 `json:"baselineAvg"`
	ZScore      float64 `json:"zScore"`
}

type ResearchETLResult struct {
	Inserted int `json:"inserted"`
	Skipped  int `json:"skipped"`
}

type etlCursor struct {
	At time.Time
	ID string
}

// ResearchAnonSalt returns the HMAC salt for practice_id_hash.
// Never derived from the ETL scheduler secret. Empty outside seedable envs if unset.
func ResearchAnonSalt(cfg config.Config) string {
	if s := strings.TrimSpace(cfg.ResearchAnonSalt); s != "" {
		return s
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))) {
	case "local", "development", "dev", "test":
		return "petsfollow-research-dev-salt"
	default:
		return ""
	}
}

func HashPracticeID(salt, practiceID string) string {
	mac := hmac.New(sha256.New, []byte(salt))
	_, _ = mac.Write([]byte(practiceID))
	return hex.EncodeToString(mac.Sum(nil))
}

// researchWeekLoc is Europe/Brussels for ISO week boundaries (tzdata embedded in API binary).
var researchWeekLoc = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Brussels")
	if err != nil {
		return time.UTC
	}
	return loc
}()

// isoWeekMonday returns the Monday 00:00 of the ISO week containing t, as a DATE
// (UTC midnight of the Brussels calendar day — matches Postgres DATE semantics).
func isoWeekMonday(t time.Time) time.Time {
	local := t.In(researchWeekLoc)
	weekday := int(local.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, researchWeekLoc).AddDate(0, 0, -(weekday - 1))
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
}

// ResearchIsoWeekMonday is the exported ISO-week Monday (Europe/Brussels) for seed/tests.
func ResearchIsoWeekMonday(t time.Time) time.Time {
	return isoWeekMonday(t)
}

func ageBandFromBirth(birth *time.Time, at time.Time) string {
	if birth == nil || birth.IsZero() || birth.After(at) {
		return "unknown"
	}
	years := at.Sub(*birth).Hours() / (24 * 365.25)
	switch {
	case years < 1:
		return "0-1"
	case years < 7:
		return "1-7"
	default:
		return "7+"
	}
}

func normalizeResearchSpecies(species string) string {
	s := strings.ToLower(strings.TrimSpace(species))
	if s == "" {
		return "unknown"
	}
	return s
}

// ResearchOptInPractice is an admin-facing row (no practice_id_hash).
type ResearchOptInPractice struct {
	PracticeID  string    `json:"practiceId"`
	Name        string    `json:"name"`
	PostalCode  string    `json:"postalCode"`
	City        string    `json:"city"`
	CountryCode string    `json:"countryCode"`
	OptedInAt   time.Time `json:"optedInAt"`
}

func (s *Store) ListResearchOptInPractices(ctx context.Context) ([]ResearchOptInPractice, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, COALESCE(name,''), COALESCE(postal_code,''), COALESCE(city,''),
		       COALESCE(country_code,'BE'), research_opt_in_at
		FROM practice.practices
		WHERE research_opt_in_at IS NOT NULL
		ORDER BY research_opt_in_at DESC, name ASC
		LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchOptInPractice
	for rows.Next() {
		var p ResearchOptInPractice
		if err := rows.Scan(&p.PracticeID, &p.Name, &p.PostalCode, &p.City, &p.CountryCode, &p.OptedInAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if out == nil {
		out = []ResearchOptInPractice{}
	}
	return out, rows.Err()
}

func (s *Store) GetResearchOptIn(ctx context.Context, practiceID string) (ResearchOptInStatus, error) {
	var at *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT research_opt_in_at FROM practice.practices WHERE id = $1`, practiceID).Scan(&at)
	if errors.Is(err, pgx.ErrNoRows) {
		return ResearchOptInStatus{}, ErrNotFound
	}
	if err != nil {
		return ResearchOptInStatus{}, err
	}
	return ResearchOptInStatus{OptedIn: at != nil, OptedInAt: at}, nil
}

func (s *Store) SetResearchOptIn(ctx context.Context, practiceID, userID string, optedIn bool) (ResearchOptInStatus, error) {
	if !optedIn {
		return ResearchOptInStatus{}, fmt.Errorf("use OptOutResearchAndPurge")
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE practice.practices
		SET research_opt_in_at = COALESCE(research_opt_in_at, NOW()),
		    research_opt_in_by = $2
		WHERE id = $1`, practiceID, userID)
	if err != nil {
		return ResearchOptInStatus{}, err
	}
	return s.GetResearchOptIn(ctx, practiceID)
}

// OptOutResearchAndPurge clears opt-in and purges anonymized rows in one transaction.
func (s *Store) OptOutResearchAndPurge(ctx context.Context, salt, practiceID string) (ResearchOptInStatus, error) {
	if strings.TrimSpace(salt) == "" {
		return ResearchOptInStatus{}, fmt.Errorf("research_anon_salt_required")
	}
	h := HashPracticeID(salt, practiceID)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ResearchOptInStatus{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	tag, err := tx.Exec(ctx, `
		UPDATE practice.practices
		SET research_opt_in_at = NULL, research_opt_in_by = NULL
		WHERE id = $1`, practiceID)
	if err != nil {
		return ResearchOptInStatus{}, err
	}
	if tag.RowsAffected() == 0 {
		return ResearchOptInStatus{}, ErrNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM research.anon_events WHERE practice_id_hash = $1`, h); err != nil {
		return ResearchOptInStatus{}, err
	}
	if _, err := tx.Exec(ctx, `TRUNCATE research.weekly_aggregates`); err != nil {
		return ResearchOptInStatus{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO research.weekly_aggregates (event_week, postal_code, country_code, species, signal_type, event_count, updated_at)
		SELECT event_week, postal_code, country_code, species, signal_type, COUNT(*)::int, NOW()
		FROM research.anon_events
		GROUP BY event_week, postal_code, country_code, species, signal_type`); err != nil {
		return ResearchOptInStatus{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ResearchOptInStatus{}, err
	}
	return ResearchOptInStatus{OptedIn: false}, nil
}

func (s *Store) RebuildResearchWeeklyAggregates(ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(87231401)`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `TRUNCATE research.weekly_aggregates`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO research.weekly_aggregates (event_week, postal_code, country_code, species, signal_type, event_count, updated_at)
		SELECT event_week, postal_code, country_code, species, signal_type, COUNT(*)::int, NOW()
		FROM research.anon_events
		GROUP BY event_week, postal_code, country_code, species, signal_type`); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// refreshResearchWeeklyAggregatesForWeeks rebuilds only the given calendar weeks (incremental ETL).
func (s *Store) refreshResearchWeeklyAggregatesForWeeks(ctx context.Context, weeks []time.Time) error {
	if len(weeks) == 0 {
		return nil
	}
	uniq := make(map[string]time.Time, len(weeks))
	for _, w := range weeks {
		key := w.UTC().Format("2006-01-02")
		uniq[key] = time.Date(w.Year(), w.Month(), w.Day(), 0, 0, 0, 0, time.UTC)
	}
	list := make([]time.Time, 0, len(uniq))
	for _, w := range uniq {
		list = append(list, w)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(87231401)`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM research.weekly_aggregates WHERE event_week = ANY($1::date[])`, list); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO research.weekly_aggregates (event_week, postal_code, country_code, species, signal_type, event_count, updated_at)
		SELECT event_week, postal_code, country_code, species, signal_type, COUNT(*)::int, NOW()
		FROM research.anon_events
		WHERE event_week = ANY($1::date[])
		GROUP BY event_week, postal_code, country_code, species, signal_type`, list); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) getETLCursor(ctx context.Context, key string) (etlCursor, error) {
	var c etlCursor
	err := s.pool.QueryRow(ctx, `
		SELECT watermark, COALESCE(watermark_id,'') FROM research.etl_watermarks WHERE source_key = $1`, key).Scan(&c.At, &c.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return etlCursor{At: time.Unix(0, 0).UTC(), ID: ""}, nil
	}
	return c, err
}

func (s *Store) setETLCursor(ctx context.Context, key string, c etlCursor) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO research.etl_watermarks (source_key, watermark, watermark_id, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (source_key) DO UPDATE SET
			watermark = EXCLUDED.watermark,
			watermark_id = EXCLUDED.watermark_id,
			updated_at = NOW()`,
		key, c.At, c.ID)
	return err
}

func cursorAfter(a, b etlCursor) bool {
	if a.At.After(b.At) {
		return true
	}
	return a.At.Equal(b.At) && a.ID > b.ID
}

type anonEventRow struct {
	PracticeID     string // not persisted — opt-in re-check at insert
	EventWeek      time.Time
	PostalCode     string
	City           string
	CountryCode    string
	Species        string
	AgeBand        string
	SignalType     string
	Payload        map[string]any
	SourceHash     string
	PracticeIDHash string
}

func (s *Store) upsertAnonEvents(ctx context.Context, salt string, rows []anonEventRow) (inserted, skipped int, weeks []time.Time, err error) {
	for _, row := range rows {
		payload, _ := json.Marshal(row.Payload)
		if payload == nil {
			payload = []byte("{}")
		}
		// Re-check opt-in at insert time (anti-race with opt-out).
		tag, err := s.pool.Exec(ctx, `
			INSERT INTO research.anon_events (
				id, event_week, postal_code, city, country_code, species, age_band,
				signal_type, payload, source_hash, practice_id_hash
			)
			SELECT $1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10,$11
			WHERE EXISTS (
				SELECT 1 FROM practice.practices pr
				WHERE pr.id = $12::uuid AND pr.research_opt_in_at IS NOT NULL
			)
			ON CONFLICT (source_hash) DO NOTHING`,
			uuid.NewString(), row.EventWeek, row.PostalCode, row.City, row.CountryCode,
			row.Species, row.AgeBand, row.SignalType, string(payload), row.SourceHash, row.PracticeIDHash,
			row.PracticeID,
		)
		if err != nil {
			return inserted, skipped, weeks, err
		}
		if tag.RowsAffected() > 0 {
			inserted++
			weeks = append(weeks, row.EventWeek)
		} else {
			skipped++
		}
	}
	_ = salt
	return inserted, skipped, weeks, nil
}

// RunResearchETL ingests anonymized signals from opted-in practices since watermarks.
func (s *Store) RunResearchETL(ctx context.Context, salt string) (ResearchETLResult, error) {
	if strings.TrimSpace(salt) == "" {
		return ResearchETLResult{}, fmt.Errorf("research_anon_salt_required")
	}
	var total ResearchETLResult
	var weeksTouched []time.Time
	steps := []struct {
		key string
		fn  func(context.Context, string, etlCursor) (int, int, etlCursor, []time.Time, error)
	}{
		{"preconsult", s.etlPreconsult},
		{"visits", s.etlVisits},
		{"lab_flags", s.etlLabFlags},
		{"care", s.etlCare},
		{"daf", s.etlAntibioticDAF},
		{"hr_alert", s.etlHRAlerts},
	}
	for _, step := range steps {
		cur, err := s.getETLCursor(ctx, step.key)
		if err != nil {
			return total, err
		}
		ins, skip, newCur, weeks, err := step.fn(ctx, salt, cur)
		if err != nil {
			return total, fmt.Errorf("%s: %w", step.key, err)
		}
		total.Inserted += ins
		total.Skipped += skip
		weeksTouched = append(weeksTouched, weeks...)
		if cursorAfter(newCur, cur) {
			if err := s.setETLCursor(ctx, step.key, newCur); err != nil {
				return total, err
			}
		}
	}
	if total.Inserted > 0 {
		if err := s.refreshResearchWeeklyAggregatesForWeeks(ctx, weeksTouched); err != nil {
			return total, err
		}
	}
	return total, nil
}

func (s *Store) etlPreconsult(ctx context.Context, salt string, since etlCursor) (int, int, etlCursor, []time.Time, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT pi.id::text, pi.updated_at, pi.answers,
		       pr.id::text, COALESCE(pr.postal_code,''), COALESCE(pr.city,''), COALESCE(pr.country_code,'BE'),
		       COALESCE(p.species,''), p.birth_date
		FROM visits.preconsult_intakes pi
		JOIN visits.visits v ON v.id = pi.visit_id
		JOIN practice.practices pr ON pr.id = v.practice_id AND pr.research_opt_in_at IS NOT NULL
		JOIN pets.pets p ON p.id = v.pet_id
		WHERE pi.status = 'submitted'
		  AND (pi.updated_at, pi.id::text) > ($1::timestamptz, $2::text)
		ORDER BY pi.updated_at ASC, pi.id ASC
		LIMIT $3`, since.At, since.ID, researchETLBatch)
	if err != nil {
		return 0, 0, since, nil, err
	}
	defer rows.Close()
	var out []anonEventRow
	maxCur := since
	for rows.Next() {
		var id, practiceID, postal, city, country, species string
		var updated time.Time
		var answersRaw []byte
		var birth *time.Time
		if err := rows.Scan(&id, &updated, &answersRaw, &practiceID, &postal, &city, &country, &species, &birth); err != nil {
			return 0, 0, since, nil, err
		}
		c := etlCursor{At: updated, ID: id}
		if cursorAfter(c, maxCur) {
			maxCur = c
		}
		var ans PreconsultAnswers
		_ = json.Unmarshal(answersRaw, &ans)
		out = append(out, anonEventRow{
			PracticeID:  practiceID,
			EventWeek:   isoWeekMonday(updated),
			PostalCode:  postal,
			City:        city,
			CountryCode: NormalizeCountryCode(country),
			Species:     normalizeResearchSpecies(species),
			AgeBand:     ageBandFromBirth(birth, updated),
			SignalType:  ResearchSignalPreconsult,
			Payload: map[string]any{
				"urgency":  ans.Urgency,
				"appetite": ans.Appetite,
				"thirst":   ans.Thirst,
				"behavior": ans.Behavior,
				"duration": ans.Duration,
			},
			SourceHash:     "preconsult:" + id,
			PracticeIDHash: HashPracticeID(salt, practiceID),
		})
	}
	if err := rows.Err(); err != nil {
		return 0, 0, since, nil, err
	}
	ins, skip, weeks, err := s.upsertAnonEvents(ctx, salt, out)
	return ins, skip, maxCur, weeks, err
}

func (s *Store) etlVisits(ctx context.Context, salt string, since etlCursor) (int, int, etlCursor, []time.Time, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.id::text, v.created_at, pr.id::text,
		       COALESCE(pr.postal_code,''), COALESCE(pr.city,''), COALESCE(pr.country_code,'BE'),
		       COALESCE(p.species,''), p.birth_date
		FROM visits.visits v
		JOIN practice.practices pr ON pr.id = v.practice_id AND pr.research_opt_in_at IS NOT NULL
		JOIN pets.pets p ON p.id = v.pet_id
		WHERE v.status IN ('confirmed', 'done')
		  AND (v.created_at, v.id::text) > ($1::timestamptz, $2::text)
		ORDER BY v.created_at ASC, v.id ASC
		LIMIT $3`, since.At, since.ID, researchETLBatch)
	if err != nil {
		return 0, 0, since, nil, err
	}
	defer rows.Close()
	var out []anonEventRow
	maxCur := since
	for rows.Next() {
		var id, practiceID, postal, city, country, species string
		var created time.Time
		var birth *time.Time
		if err := rows.Scan(&id, &created, &practiceID, &postal, &city, &country, &species, &birth); err != nil {
			return 0, 0, since, nil, err
		}
		c := etlCursor{At: created, ID: id}
		if cursorAfter(c, maxCur) {
			maxCur = c
		}
		out = append(out, anonEventRow{
			PracticeID:     practiceID,
			EventWeek:      isoWeekMonday(created),
			PostalCode:     postal,
			City:           city,
			CountryCode:    NormalizeCountryCode(country),
			Species:        normalizeResearchSpecies(species),
			AgeBand:        ageBandFromBirth(birth, created),
			SignalType:     ResearchSignalVisitVolume,
			Payload:        map[string]any{},
			SourceHash:     "visit:" + id,
			PracticeIDHash: HashPracticeID(salt, practiceID),
		})
	}
	if err := rows.Err(); err != nil {
		return 0, 0, since, nil, err
	}
	ins, skip, weeks, err := s.upsertAnonEvents(ctx, salt, out)
	return ins, skip, maxCur, weeks, err
}

func (s *Store) etlLabFlags(ctx context.Context, salt string, since etlCursor) (int, int, etlCursor, []time.Time, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT lp.id::text, lr.analyte_code, lp.updated_at, lr.flag, pr.id::text,
		       COALESCE(pr.postal_code,''), COALESCE(pr.city,''), COALESCE(pr.country_code,'BE'),
		       COALESCE(p.species,''), p.birth_date
		FROM labs.panel_results lr
		JOIN labs.panels lp ON lp.id = lr.panel_id
		JOIN pets.pets p ON p.id = lp.pet_id
		JOIN practice.practices pr ON pr.id = COALESCE(lp.practice_id, p.practice_id) AND pr.research_opt_in_at IS NOT NULL
		WHERE lr.flag IN ('low','high')
		  AND COALESCE(lp.practice_id, p.practice_id) IS NOT NULL
		  AND (lp.updated_at, lp.id::text || ':' || lr.analyte_code) > ($1::timestamptz, $2::text)
		ORDER BY lp.updated_at ASC, lp.id ASC, lr.analyte_code ASC
		LIMIT $3`, since.At, since.ID, researchETLBatch)
	if err != nil {
		return 0, 0, since, nil, err
	}
	defer rows.Close()
	var out []anonEventRow
	maxCur := since
	for rows.Next() {
		var panelID, analyte, flag, practiceID, postal, city, country, species string
		var updated time.Time
		var birth *time.Time
		if err := rows.Scan(&panelID, &analyte, &updated, &flag, &practiceID, &postal, &city, &country, &species, &birth); err != nil {
			return 0, 0, since, nil, err
		}
		cursorID := panelID + ":" + analyte
		c := etlCursor{At: updated, ID: cursorID}
		if cursorAfter(c, maxCur) {
			maxCur = c
		}
		out = append(out, anonEventRow{
			PracticeID:     practiceID,
			EventWeek:      isoWeekMonday(updated),
			PostalCode:     postal,
			City:           city,
			CountryCode:    NormalizeCountryCode(country),
			Species:        normalizeResearchSpecies(species),
			AgeBand:        ageBandFromBirth(birth, updated),
			SignalType:     ResearchSignalLabFlag,
			Payload:        map[string]any{"analyte": analyte, "flag": flag},
			SourceHash:     "lab_flag:" + panelID + ":" + analyte,
			PracticeIDHash: HashPracticeID(salt, practiceID),
		})
	}
	if err := rows.Err(); err != nil {
		return 0, 0, since, nil, err
	}
	ins, skip, weeks, err := s.upsertAnonEvents(ctx, salt, out)
	return ins, skip, maxCur, weeks, err
}

func (s *Store) etlCare(ctx context.Context, salt string, since etlCursor) (int, int, etlCursor, []time.Time, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.id::text, c.updated_at, c.type, pr.id::text,
		       COALESCE(pr.postal_code,''), COALESCE(pr.city,''), COALESCE(pr.country_code,'BE'),
		       COALESCE(p.species,''), p.birth_date
		FROM care.reminders c
		JOIN pets.pets p ON p.id = c.pet_id
		JOIN practice.practices pr ON pr.id = c.practice_id AND pr.research_opt_in_at IS NOT NULL
		WHERE c.status = 'done'
		  AND c.type IN ('vaccination','deworming','fecal_egg')
		  AND (c.updated_at, c.id::text) > ($1::timestamptz, $2::text)
		ORDER BY c.updated_at ASC, c.id ASC
		LIMIT $3`, since.At, since.ID, researchETLBatch)
	if err != nil {
		return 0, 0, since, nil, err
	}
	defer rows.Close()
	var out []anonEventRow
	maxCur := since
	for rows.Next() {
		var id, careType, practiceID, postal, city, country, species string
		var at time.Time
		var birth *time.Time
		if err := rows.Scan(&id, &at, &careType, &practiceID, &postal, &city, &country, &species, &birth); err != nil {
			return 0, 0, since, nil, err
		}
		c := etlCursor{At: at, ID: id}
		if cursorAfter(c, maxCur) {
			maxCur = c
		}
		out = append(out, anonEventRow{
			PracticeID:     practiceID,
			EventWeek:      isoWeekMonday(at),
			PostalCode:     postal,
			City:           city,
			CountryCode:    NormalizeCountryCode(country),
			Species:        normalizeResearchSpecies(species),
			AgeBand:        ageBandFromBirth(birth, at),
			SignalType:     ResearchSignalCarePrev,
			Payload:        map[string]any{"type": careType},
			SourceHash:     "care:" + id,
			PracticeIDHash: HashPracticeID(salt, practiceID),
		})
	}
	if err := rows.Err(); err != nil {
		return 0, 0, since, nil, err
	}
	ins, skip, weeks, err := s.upsertAnonEvents(ctx, salt, out)
	return ins, skip, maxCur, weeks, err
}

func (s *Store) etlAntibioticDAF(ctx context.Context, salt string, since etlCursor) (int, int, etlCursor, []time.Time, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.id::text, d.updated_at, pr.id::text,
		       COALESCE(pr.postal_code,''), COALESCE(pr.city,''), COALESCE(pr.country_code,'BE'),
		       COALESCE(p.species,''), p.birth_date
		FROM pharmacy.daf_documents d
		JOIN practice.practices pr ON pr.id = d.practice_id AND pr.research_opt_in_at IS NOT NULL
		LEFT JOIN pets.pets p ON p.id = d.pet_id
		WHERE d.has_antibiotic = true
		  AND d.status = 'finalized'
		  AND (d.updated_at, d.id::text) > ($1::timestamptz, $2::text)
		ORDER BY d.updated_at ASC, d.id ASC
		LIMIT $3`, since.At, since.ID, researchETLBatch)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return 0, 0, since, nil, nil
		}
		return 0, 0, since, nil, err
	}
	defer rows.Close()
	var out []anonEventRow
	maxCur := since
	for rows.Next() {
		var id, practiceID, postal, city, country string
		var species *string
		var updated time.Time
		var birth *time.Time
		if err := rows.Scan(&id, &updated, &practiceID, &postal, &city, &country, &species, &birth); err != nil {
			return 0, 0, since, nil, err
		}
		c := etlCursor{At: updated, ID: id}
		if cursorAfter(c, maxCur) {
			maxCur = c
		}
		sp := "unknown"
		if species != nil {
			sp = normalizeResearchSpecies(*species)
		}
		out = append(out, anonEventRow{
			PracticeID:     practiceID,
			EventWeek:      isoWeekMonday(updated),
			PostalCode:     postal,
			City:           city,
			CountryCode:    NormalizeCountryCode(country),
			Species:        sp,
			AgeBand:        ageBandFromBirth(birth, updated),
			SignalType:     ResearchSignalAntibioticDAF,
			Payload:        map[string]any{},
			SourceHash:     "daf_ab:" + id,
			PracticeIDHash: HashPracticeID(salt, practiceID),
		})
	}
	if err := rows.Err(); err != nil {
		return 0, 0, since, nil, err
	}
	ins, skip, weeks, err := s.upsertAnonEvents(ctx, salt, out)
	return ins, skip, maxCur, weeks, err
}

func (s *Store) etlHRAlerts(ctx context.Context, salt string, since etlCursor) (int, int, etlCursor, []time.Time, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT hs.id::text, hs.validated_at, pr.id::text,
		       COALESCE(pr.postal_code,''), COALESCE(pr.city,''), COALESCE(pr.country_code,'BE'),
		       COALESCE(p.species,''), p.birth_date
		FROM heartrate.sessions hs
		JOIN pets.pets p ON p.id = hs.pet_id
		JOIN practice.practices pr ON pr.id = COALESCE(hs.practice_id, p.practice_id) AND pr.research_opt_in_at IS NOT NULL
		WHERE hs.is_alert = true
		  AND hs.status = 'validated'
		  AND hs.validated_at IS NOT NULL
		  AND COALESCE(hs.practice_id, p.practice_id) IS NOT NULL
		  AND (hs.validated_at, hs.id::text) > ($1::timestamptz, $2::text)
		ORDER BY hs.validated_at ASC, hs.id ASC
		LIMIT $3`, since.At, since.ID, researchETLBatch)
	if err != nil {
		return 0, 0, since, nil, err
	}
	defer rows.Close()
	var out []anonEventRow
	maxCur := since
	for rows.Next() {
		var id, practiceID, postal, city, country, species string
		var at time.Time
		var birth *time.Time
		if err := rows.Scan(&id, &at, &practiceID, &postal, &city, &country, &species, &birth); err != nil {
			return 0, 0, since, nil, err
		}
		c := etlCursor{At: at, ID: id}
		if cursorAfter(c, maxCur) {
			maxCur = c
		}
		out = append(out, anonEventRow{
			PracticeID:     practiceID,
			EventWeek:      isoWeekMonday(at),
			PostalCode:     postal,
			City:           city,
			CountryCode:    NormalizeCountryCode(country),
			Species:        normalizeResearchSpecies(species),
			AgeBand:        ageBandFromBirth(birth, at),
			SignalType:     ResearchSignalHRAlert,
			Payload:        map[string]any{},
			SourceHash:     "hr_alert:" + id,
			PracticeIDHash: HashPracticeID(salt, practiceID),
		})
	}
	if err := rows.Err(); err != nil {
		return 0, 0, since, nil, err
	}
	ins, skip, weeks, err := s.upsertAnonEvents(ctx, salt, out)
	return ins, skip, maxCur, weeks, err
}

func (s *Store) ResearchOverview(ctx context.Context) (ResearchOverview, error) {
	var o ResearchOverview
	o.BySignal = map[string]int{}
	o.BySpecies = map[string]int{}
	o.ByCountry = map[string]int{}
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM practice.practices WHERE research_opt_in_at IS NOT NULL`).Scan(&o.OptInPracticeCount); err != nil {
		return o, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(event_count),0)::int, COUNT(DISTINCT event_week)::int, MAX(event_week)::text
		FROM research.weekly_aggregates`).Scan(&o.EventCount, &o.WeekCount, &o.LatestWeek); err != nil {
		return o, err
	}
	if o.LatestWeek != nil && *o.LatestWeek == "" {
		o.LatestWeek = nil
	}
	sigRows, err := s.pool.Query(ctx, `
		SELECT signal_type, SUM(event_count)::int FROM research.weekly_aggregates GROUP BY signal_type`)
	if err != nil {
		return o, err
	}
	defer sigRows.Close()
	for sigRows.Next() {
		var k string
		var n int
		if err := sigRows.Scan(&k, &n); err != nil {
			return o, err
		}
		o.BySignal[k] = n
	}
	spRows, err := s.pool.Query(ctx, `
		SELECT species, SUM(event_count)::int FROM research.weekly_aggregates GROUP BY species`)
	if err != nil {
		return o, err
	}
	defer spRows.Close()
	for spRows.Next() {
		var k string
		var n int
		if err := spRows.Scan(&k, &n); err != nil {
			return o, err
		}
		o.BySpecies[k] = n
	}
	cRows, err := s.pool.Query(ctx, `
		SELECT country_code, SUM(event_count)::int FROM research.weekly_aggregates GROUP BY country_code`)
	if err != nil {
		return o, err
	}
	defer cRows.Close()
	for cRows.Next() {
		var k string
		var n int
		if err := cRows.Scan(&k, &n); err != nil {
			return o, err
		}
		o.ByCountry[k] = n
	}
	return o, nil
}

func (s *Store) ResearchHeatmap(ctx context.Context, species, signal, from, to string) ([]ResearchHeatmapCell, error) {
	if from != "" {
		if _, err := time.Parse("2006-01-02", from); err != nil {
			return nil, fmt.Errorf("invalid_from_date")
		}
	}
	if to != "" {
		if _, err := time.Parse("2006-01-02", to); err != nil {
			return nil, fmt.Errorf("invalid_to_date")
		}
	}
	// Gate on distinct practices (same bar as Data room) — volume alone is not enough.
	q := `
		SELECT postal_code, country_code, species, signal_type, COUNT(*)::int
		FROM research.anon_events
		WHERE ($1 = '' OR species = $1)
		  AND ($2 = '' OR signal_type = $2)
		  AND ($3 = '' OR event_week >= $3::date)
		  AND ($4 = '' OR event_week <= $4::date)
		GROUP BY postal_code, country_code, species, signal_type
		HAVING COUNT(DISTINCT practice_id_hash) >= $5
		ORDER BY COUNT(*) DESC
		LIMIT 500`
	rows, err := s.pool.Query(ctx, q, species, signal, from, to, ResearchKAnonymity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchHeatmapCell
	for rows.Next() {
		var c ResearchHeatmapCell
		if err := rows.Scan(&c.PostalCode, &c.CountryCode, &c.Species, &c.SignalType, &c.EventCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []ResearchHeatmapCell{}
	}
	return out, rows.Err()
}

func (s *Store) ResearchTimeseries(ctx context.Context, species, signal, country, postal string) ([]ResearchTimeseriesPoint, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT event_week::text, SUM(event_count)::int
		FROM research.weekly_aggregates
		WHERE ($1 = '' OR species = $1)
		  AND ($2 = '' OR signal_type = $2)
		  AND ($3 = '' OR country_code = $3)
		  AND ($4 = '' OR postal_code = $4)
		GROUP BY event_week
		ORDER BY event_week ASC
		LIMIT 104`, species, signal, country, postal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchTimeseriesPoint
	for rows.Next() {
		var p ResearchTimeseriesPoint
		if err := rows.Scan(&p.Week, &p.EventCount); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if out == nil {
		out = []ResearchTimeseriesPoint{}
	}
	return out, rows.Err()
}

// ResearchAlerts returns simple z-score spikes vs prior 8 weeks (same CP+species+signal).
// Current-week cells require ≥ ResearchKAnonymity distinct practice_id_hash.
func (s *Store) ResearchAlerts(ctx context.Context) ([]ResearchAlert, error) {
	rows, err := s.pool.Query(ctx, `
		WITH latest AS (
		  SELECT MAX(event_week) AS w FROM research.weekly_aggregates
		),
		cur AS (
		  SELECT e.postal_code, e.country_code, e.species, e.signal_type, e.event_week,
		         COUNT(*)::int AS event_count
		  FROM research.anon_events e, latest l
		  WHERE e.event_week = l.w
		    AND e.signal_type IN ('preconsult_syndrome', 'lab_flag', 'visit_volume')
		  GROUP BY e.postal_code, e.country_code, e.species, e.signal_type, e.event_week
		  HAVING COUNT(DISTINCT e.practice_id_hash) >= $1
		),
		base AS (
		  SELECT a.postal_code, a.country_code, a.species, a.signal_type,
		         AVG(a.event_count::float) AS avg_c,
		         COALESCE(STDDEV_SAMP(a.event_count::float), 0) AS std_c
		  FROM research.weekly_aggregates a, latest l
		  WHERE a.event_week < l.w
		    AND a.event_week >= l.w - INTERVAL '8 weeks'
		  GROUP BY a.postal_code, a.country_code, a.species, a.signal_type
		)
		SELECT c.postal_code, c.country_code, c.species, c.signal_type, c.event_week::text,
		       c.event_count, COALESCE(b.avg_c, 0),
		       CASE WHEN COALESCE(b.std_c, 0) < 0.5 THEN 0
		            ELSE (c.event_count - b.avg_c) / b.std_c END AS z
		FROM cur c
		LEFT JOIN base b ON b.postal_code = c.postal_code AND b.country_code = c.country_code
		  AND b.species = c.species AND b.signal_type = c.signal_type
		WHERE (
		    (COALESCE(b.std_c, 0) >= 0.5 AND (c.event_count - b.avg_c) / b.std_c >= 2)
		    OR (COALESCE(b.avg_c, 0) > 0 AND c.event_count >= b.avg_c * 2)
		  )
		ORDER BY z DESC NULLS LAST
		LIMIT 50`, ResearchKAnonymity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResearchAlert
	for rows.Next() {
		var a ResearchAlert
		if err := rows.Scan(&a.PostalCode, &a.CountryCode, &a.Species, &a.SignalType, &a.Week,
			&a.EventCount, &a.BaselineAvg, &a.ZScore); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if out == nil {
		out = []ResearchAlert{}
	}
	return out, rows.Err()
}
