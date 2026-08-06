package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Site is a physical location under a practice (tenant).
type Site struct {
	ID           string    `json:"id"`
	PracticeID   string    `json:"practiceId"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	AddressLine1 string    `json:"addressLine1"`
	AddressLine2 string    `json:"addressLine2"`
	City         string    `json:"city"`
	PostalCode   string    `json:"postalCode"`
	CountryCode  string    `json:"countryCode"`
	Timezone     string    `json:"timezone"`
	IsPrimary    bool      `json:"isPrimary"`
	Active       bool      `json:"active"`
	SortOrder    int       `json:"sortOrder"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// SiteSummary is the lightweight shape exposed on /me and pickers.
type SiteSummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsPrimary bool   `json:"isPrimary"`
	Timezone  string `json:"timezone"`
	Active    bool   `json:"active"`
}

type CreateSiteInput struct {
	Name         string
	Phone        string
	AddressLine1 string
	AddressLine2 string
	City         string
	PostalCode   string
	CountryCode  string
	Timezone     string
	CopyScheduleFromSiteID string // optional template; empty = empty schedule
}

type PatchSiteInput struct {
	Name         *string
	Phone        *string
	AddressLine1 *string
	AddressLine2 *string
	City         *string
	PostalCode   *string
	CountryCode  *string
	Timezone     *string
	IsPrimary    *bool
	Active       *bool
	SortOrder    *int
}

const siteSelectCols = `
	id::text, practice_id::text, name, COALESCE(phone,''), COALESCE(address_line1,''),
	COALESCE(address_line2,''), COALESCE(city,''), COALESCE(postal_code,''),
	COALESCE(country_code,'BE'), COALESCE(timezone,'Europe/Brussels'),
	is_primary, active, sort_order, created_at, updated_at`

func scanSite(row pgx.Row) (Site, error) {
	var s Site
	err := row.Scan(
		&s.ID, &s.PracticeID, &s.Name, &s.Phone, &s.AddressLine1, &s.AddressLine2,
		&s.City, &s.PostalCode, &s.CountryCode, &s.Timezone,
		&s.IsPrimary, &s.Active, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (s *Store) GetSite(ctx context.Context, practiceID, siteID string) (Site, error) {
	site, err := scanSite(s.pool.QueryRow(ctx, `
		SELECT `+siteSelectCols+` FROM practice.sites
		WHERE id = $1 AND practice_id = $2`, siteID, practiceID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Site{}, ErrNotFound
	}
	return site, err
}

func (s *Store) GetPrimarySite(ctx context.Context, practiceID string) (Site, error) {
	site, err := scanSite(s.pool.QueryRow(ctx, `
		SELECT `+siteSelectCols+` FROM practice.sites
		WHERE practice_id = $1 AND is_primary
		LIMIT 1`, practiceID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Site{}, ErrNotFound
	}
	return site, err
}

// ResolveSiteID returns siteID if it belongs to the practice (and is active unless allowInactive),
// or the primary site when siteID is empty (primary must also be active unless allowInactive).
func (s *Store) ResolveSiteID(ctx context.Context, practiceID, siteID string, allowInactive bool) (string, error) {
	siteID = strings.TrimSpace(siteID)
	if siteID == "" {
		primary, err := s.GetPrimarySite(ctx, practiceID)
		if err != nil {
			return "", err
		}
		if !primary.Active && !allowInactive {
			return "", fmt.Errorf("%w: site_inactive", ErrValidation)
		}
		return primary.ID, nil
	}
	site, err := s.GetSite(ctx, practiceID, siteID)
	if err != nil {
		return "", err
	}
	if !site.Active && !allowInactive {
		return "", fmt.Errorf("%w: site_inactive", ErrValidation)
	}
	return site.ID, nil
}

// ResolveBookingSiteID picks the site for a new visit.
// Explicit siteId wins; empty siteId mirrors availability fall-through (sole bookable site).
// When several sites are bookable, empty siteId is rejected (site_required) — same as availability.
// Zero bookable sites → active primary (caller still checks client_booking_enabled).
func (s *Store) ResolveBookingSiteID(ctx context.Context, practiceID, siteID string) (string, error) {
	siteID = strings.TrimSpace(siteID)
	if siteID != "" {
		return s.ResolveSiteID(ctx, practiceID, siteID, false)
	}
	ids, err := s.listBookableSiteIDs(ctx, practiceID)
	if err != nil {
		return "", err
	}
	switch len(ids) {
	case 1:
		return ids[0], nil
	case 0:
		return s.ResolveSiteID(ctx, practiceID, "", false)
	default:
		return "", fmt.Errorf("%w: site_required", ErrValidation)
	}
}

func (s *Store) listBookableSiteIDs(ctx context.Context, practiceID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT s.id::text
		FROM practice.sites s
		JOIN practice.vet_schedule vs ON vs.site_id = s.id
		WHERE s.practice_id = $1 AND s.active AND vs.client_booking_enabled
		ORDER BY s.is_primary DESC, s.sort_order, s.name`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *Store) ListSites(ctx context.Context, practiceID string, includeInactive bool) ([]Site, error) {
	q := `
		SELECT ` + siteSelectCols + ` FROM practice.sites
		WHERE practice_id = $1`
	if !includeInactive {
		q += ` AND active`
	}
	q += ` ORDER BY is_primary DESC, sort_order, name`
	rows, err := s.pool.Query(ctx, q, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Site
	for rows.Next() {
		site, err := scanSite(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, site)
	}
	if out == nil {
		out = []Site{}
	}
	return out, rows.Err()
}

func (s *Store) ListSiteSummaries(ctx context.Context, practiceID string) ([]SiteSummary, error) {
	sites, err := s.ListSites(ctx, practiceID, false)
	if err != nil {
		return nil, err
	}
	out := make([]SiteSummary, 0, len(sites))
	for _, site := range sites {
		out = append(out, SiteSummary{
			ID: site.ID, Name: site.Name, IsPrimary: site.IsPrimary,
			Timezone: site.Timezone, Active: site.Active,
		})
	}
	return out, nil
}

// EnsurePrimarySiteTx creates a primary site inside an existing transaction (seed / provisioning).
func EnsurePrimarySiteTx(ctx context.Context, tx pgx.Tx, practiceID string) (string, error) {
	return insertPrimarySiteTx(ctx, tx, practiceID)
}

// EnsurePrimarySite creates a primary site from practice profile if missing (idempotent).
func (s *Store) EnsurePrimarySite(ctx context.Context, practiceID string) (Site, error) {
	if site, err := s.GetPrimarySite(ctx, practiceID); err == nil {
		return site, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Site{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Site{}, err
	}
	defer tx.Rollback(ctx)
	siteID, err := insertPrimarySiteTx(ctx, tx, practiceID)
	if err != nil {
		return Site{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Site{}, err
	}
	return s.GetSite(ctx, practiceID, siteID)
}

func insertPrimarySiteTx(ctx context.Context, tx pgx.Tx, practiceID string) (string, error) {
	// Race-safe: another tx may have created primary already.
	var existing string
	err := tx.QueryRow(ctx, `
		SELECT id::text FROM practice.sites WHERE practice_id = $1 AND is_primary LIMIT 1`, practiceID,
	).Scan(&existing)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	var name, phone, a1, a2, city, postal, country string
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(NULLIF(TRIM(name),''),'Principal'), COALESCE(phone,''), COALESCE(address_line1,''),
			COALESCE(address_line2,''), COALESCE(city,''), COALESCE(postal_code,''),
			COALESCE(NULLIF(TRIM(country_code),''),'BE')
		FROM practice.practices WHERE id = $1`, practiceID,
	).Scan(&name, &phone, &a1, &a2, &city, &postal, &country)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	tz := defaultScheduleTZ
	_ = tx.QueryRow(ctx, `
		SELECT timezone FROM practice.vet_schedule WHERE practice_id = $1 LIMIT 1`, practiceID,
	).Scan(&tz)
	if strings.TrimSpace(tz) == "" {
		tz = defaultScheduleTZ
	}

	siteID := uuid.NewString()
	_, err = tx.Exec(ctx, `
		INSERT INTO practice.sites (
			id, practice_id, name, phone, address_line1, address_line2, city, postal_code,
			country_code, timezone, is_primary, active, sort_order
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,true,true,0)`,
		siteID, practiceID, name, phone, a1, a2, city, postal, country, tz)
	if err != nil {
		if isUniqueViolation(err) {
			var existing string
			if e2 := tx.QueryRow(ctx, `
				SELECT id::text FROM practice.sites WHERE practice_id = $1 AND is_primary LIMIT 1`, practiceID,
			).Scan(&existing); e2 == nil {
				return existing, nil
			}
		}
		return "", err
	}
	// Ensure a schedule row exists for the primary site.
	_, err = tx.Exec(ctx, `
		INSERT INTO practice.vet_schedule (site_id, practice_id, timezone)
		VALUES ($1, $2, $3)
		ON CONFLICT (site_id) DO NOTHING`, siteID, practiceID, tz)
	return siteID, err
}

func (s *Store) CreateSite(ctx context.Context, practiceID string, in CreateSiteInput) (Site, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return Site{}, fmt.Errorf("%w: name_required", ErrValidation)
	}
	tz := strings.TrimSpace(in.Timezone)
	if tz == "" {
		tz = defaultScheduleTZ
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return Site{}, fmt.Errorf("%w: invalid_timezone", ErrValidation)
	}
	country := strings.TrimSpace(in.CountryCode)
	if country == "" {
		country = "BE"
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Site{}, err
	}
	defer tx.Rollback(ctx)

	siteID := uuid.NewString()
	_, err = tx.Exec(ctx, `
		INSERT INTO practice.sites (
			id, practice_id, name, phone, address_line1, address_line2, city, postal_code,
			country_code, timezone, is_primary, active, sort_order
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,false,true,
			(SELECT COALESCE(MAX(sort_order),0)+1 FROM practice.sites WHERE practice_id = $2))`,
		siteID, practiceID, name, strings.TrimSpace(in.Phone), strings.TrimSpace(in.AddressLine1),
		strings.TrimSpace(in.AddressLine2), strings.TrimSpace(in.City), strings.TrimSpace(in.PostalCode),
		country, tz)
	if err != nil {
		return Site{}, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO practice.vet_schedule (site_id, practice_id, timezone)
		VALUES ($1, $2, $3)`, siteID, practiceID, tz)
	if err != nil {
		return Site{}, err
	}

	templateID := strings.TrimSpace(in.CopyScheduleFromSiteID)
	if templateID != "" && templateID != siteID {
		// Same-practice only (prevent cross-tenant schedule flag copy).
		if _, err := s.GetSite(ctx, practiceID, templateID); err != nil {
			if errors.Is(err, ErrNotFound) {
				return Site{}, fmt.Errorf("%w: invalid_template_site", ErrValidation)
			}
			return Site{}, err
		}
		if err := copyScheduleSlotsTx(ctx, tx, practiceID, templateID, siteID); err != nil {
			return Site{}, err
		}
		var enabled bool
		var duration int
		var vacYear *int
		_ = tx.QueryRow(ctx, `
			SELECT client_booking_enabled, slot_duration_minutes, vacations_declared_year
			FROM practice.vet_schedule WHERE site_id = $1 AND practice_id = $2`, templateID, practiceID,
		).Scan(&enabled, &duration, &vacYear)
		if duration == 0 {
			duration = 30
		}
		_, _ = tx.Exec(ctx, `
			UPDATE practice.vet_schedule
			SET client_booking_enabled = $2, slot_duration_minutes = $3,
			    vacations_declared_year = $4, updated_at = NOW()
			WHERE site_id = $1`, siteID, enabled, duration, vacYear)
	}

	if err := tx.Commit(ctx); err != nil {
		return Site{}, err
	}
	return s.GetSite(ctx, practiceID, siteID)
}

func copyScheduleSlotsTx(ctx context.Context, tx pgx.Tx, practiceID, fromSiteID, toSiteID string) error {
	rows, err := tx.Query(ctx, `
		SELECT weekday, to_char(start_time,'HH24:MI'), to_char(end_time,'HH24:MI')
		FROM practice.vet_schedule_slots WHERE site_id = $1 AND practice_id = $2`, fromSiteID, practiceID)
	if err != nil {
		return err
	}
	type slotRow struct {
		wd       int
		start, end string
	}
	var slots []slotRow
	for rows.Next() {
		var s slotRow
		if err := rows.Scan(&s.wd, &s.start, &s.end); err != nil {
			rows.Close()
			return err
		}
		slots = append(slots, s)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, s := range slots {
		if _, err := tx.Exec(ctx, `
			INSERT INTO practice.vet_schedule_slots (id, practice_id, site_id, weekday, start_time, end_time)
			VALUES ($1,$2,$3,$4,$5::time,$6::time)`,
			uuid.NewString(), practiceID, toSiteID, s.wd, s.start, s.end); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) PatchSite(ctx context.Context, practiceID, siteID string, in PatchSiteInput) (Site, error) {
	if in.Active != nil && !*in.Active {
		return s.DeactivateSite(ctx, practiceID, siteID)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Site{}, err
	}
	defer tx.Rollback(ctx)

	site, err := scanSite(tx.QueryRow(ctx, `
		SELECT `+siteSelectCols+` FROM practice.sites
		WHERE id = $1 AND practice_id = $2 FOR UPDATE`, siteID, practiceID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Site{}, ErrNotFound
	}
	if err != nil {
		return Site{}, err
	}

	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		if n == "" {
			return Site{}, fmt.Errorf("%w: name_required", ErrValidation)
		}
		site.Name = n
	}
	if in.Phone != nil {
		site.Phone = strings.TrimSpace(*in.Phone)
	}
	if in.AddressLine1 != nil {
		site.AddressLine1 = strings.TrimSpace(*in.AddressLine1)
	}
	if in.AddressLine2 != nil {
		site.AddressLine2 = strings.TrimSpace(*in.AddressLine2)
	}
	if in.City != nil {
		site.City = strings.TrimSpace(*in.City)
	}
	if in.PostalCode != nil {
		site.PostalCode = strings.TrimSpace(*in.PostalCode)
	}
	if in.CountryCode != nil {
		c := strings.TrimSpace(*in.CountryCode)
		if c == "" {
			c = "BE"
		}
		site.CountryCode = c
	}
	if in.Timezone != nil {
		tz := strings.TrimSpace(*in.Timezone)
		if tz == "" {
			tz = defaultScheduleTZ
		}
		if _, err := time.LoadLocation(tz); err != nil {
			return Site{}, fmt.Errorf("%w: invalid_timezone", ErrValidation)
		}
		site.Timezone = tz
	}
	if in.SortOrder != nil {
		site.SortOrder = *in.SortOrder
	}
	if in.Active != nil && *in.Active {
		site.Active = true
	}

	makePrimary := in.IsPrimary != nil && *in.IsPrimary && !site.IsPrimary
	if makePrimary {
		if !site.Active {
			return Site{}, fmt.Errorf("%w: cannot_promote_inactive", ErrValidation)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE practice.sites SET is_primary = false, updated_at = NOW()
			WHERE practice_id = $1 AND is_primary`, practiceID); err != nil {
			return Site{}, err
		}
		site.IsPrimary = true
	}

	_, err = tx.Exec(ctx, `
		UPDATE practice.sites SET
			name=$3, phone=$4, address_line1=$5, address_line2=$6, city=$7, postal_code=$8,
			country_code=$9, timezone=$10, is_primary=$11, active=$12, sort_order=$13, updated_at=NOW()
		WHERE id=$1 AND practice_id=$2`,
		siteID, practiceID, site.Name, site.Phone, site.AddressLine1, site.AddressLine2,
		site.City, site.PostalCode, site.CountryCode, site.Timezone, site.IsPrimary, site.Active, site.SortOrder)
	if err != nil {
		return Site{}, err
	}
	_, _ = tx.Exec(ctx, `
		UPDATE practice.vet_schedule SET timezone = $2, updated_at = NOW() WHERE site_id = $1`,
		siteID, site.Timezone)

	if site.IsPrimary {
		if err := syncPracticeContactFromSiteTx(ctx, tx, practiceID, site); err != nil {
			return Site{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Site{}, err
	}
	return s.GetSite(ctx, practiceID, siteID)
}

func syncPracticeContactFromSiteTx(ctx context.Context, tx pgx.Tx, practiceID string, site Site) error {
	_, err := tx.Exec(ctx, `
		UPDATE practice.practices SET
			phone = $2, address_line1 = $3, address_line2 = $4,
			city = $5, postal_code = $6, country_code = $7
		WHERE id = $1`,
		practiceID, site.Phone, site.AddressLine1, site.AddressLine2,
		site.City, site.PostalCode, site.CountryCode)
	return err
}

// DeactivateSite soft-disables a non-primary site when no future agenda visits remain.
func (s *Store) DeactivateSite(ctx context.Context, practiceID, siteID string) (Site, error) {
	site, err := s.GetSite(ctx, practiceID, siteID)
	if err != nil {
		return Site{}, err
	}
	if site.IsPrimary {
		return Site{}, fmt.Errorf("%w: cannot_deactivate_primary", ErrValidation)
	}
	if !site.Active {
		return site, nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Site{}, err
	}
	defer tx.Rollback(ctx)

	// Lock site row to serialize with CreateVisitBooked.
	var locked string
	if err := tx.QueryRow(ctx, `
		SELECT id::text FROM practice.sites
		WHERE id = $1 AND practice_id = $2 FOR UPDATE`, siteID, practiceID,
	).Scan(&locked); err != nil {
		return Site{}, err
	}

	var n int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM visits.visits
		WHERE site_id = $1 AND deleted_at IS NULL
		  AND status IN ('requested', 'confirmed', 'reschedule_pending')
		  AND (
		    (status = 'requested' AND COALESCE(proposed_scheduled_at, scheduled_at) IS NULL)
		    OR COALESCE(proposed_scheduled_at, scheduled_at) > NOW()
		  )`, siteID,
	).Scan(&n)
	if err != nil {
		return Site{}, err
	}
	if n > 0 {
		return Site{}, fmt.Errorf("%w: site_has_future_visits", ErrValidation)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE practice.sites SET active = false, updated_at = NOW()
		WHERE id = $1 AND practice_id = $2`, siteID, practiceID); err != nil {
		return Site{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE practice.vet_schedule SET client_booking_enabled = false, updated_at = NOW()
		WHERE site_id = $1`, siteID); err != nil {
		return Site{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE practice.team_members SET default_site_id = NULL
		WHERE practice_id = $1 AND default_site_id = $2`, practiceID, siteID); err != nil {
		return Site{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Site{}, err
	}
	return s.GetSite(ctx, practiceID, siteID)
}

// GetTeamDefaultSiteID returns the member's preferred site (may be empty).
func (s *Store) GetTeamDefaultSiteID(ctx context.Context, practiceID, userID string) (string, error) {
	var id *string
	err := s.pool.QueryRow(ctx, `
		SELECT default_site_id::text FROM practice.team_members
		WHERE practice_id = $1 AND user_id = $2 AND status = 'active'`, practiceID, userID,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if id == nil {
		return "", nil
	}
	return *id, nil
}

// ClientBookingEnabledForPractice is true if any active site has booking enabled.
func (s *Store) ClientBookingEnabledForPractice(ctx context.Context, practiceID string) (bool, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM practice.vet_schedule vs
		JOIN practice.sites s ON s.id = vs.site_id
		WHERE vs.practice_id = $1 AND s.active AND vs.client_booking_enabled`, practiceID,
	).Scan(&n)
	return n > 0, err
}
