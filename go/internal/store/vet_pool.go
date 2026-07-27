package store

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// UnassignedVetRow is a vet without assigned_commercial_id (admin pool).
type UnassignedVetRow struct {
	UserID       string    `json:"userId"`
	FullName     string    `json:"fullName"`
	Email        string    `json:"email"`
	PracticeID   string    `json:"practiceId,omitempty"`
	PracticeName string    `json:"practiceName"`
	City         string    `json:"city"`
	PostalCode   string    `json:"postalCode"`
	CreatedAt    time.Time `json:"createdAt"`
}

// CommercialSuggestion is a scored commercial candidate for assigning a vet.
type CommercialSuggestion struct {
	UserID       string   `json:"userId"`
	FullName     string   `json:"fullName"`
	Email        string   `json:"email"`
	City         string   `json:"city,omitempty"`
	PostalCode   string   `json:"postalCode,omitempty"`
	Score        int      `json:"score"`
	ZoneScore    int      `json:"zoneScore"`
	ActivityScore int     `json:"activityScore"`
	Reasons      []string `json:"reasons"`
	Note         string   `json:"note"`
	DistanceKm   *float64 `json:"distanceKm,omitempty"`
	AssignedVets int      `json:"assignedVets"`
}

// ListUnassignedVets returns vets with no assigned commercial.
func (s *Store) ListUnassignedVets(ctx context.Context) ([]UnassignedVetRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.full_name, u.email,
			COALESCE(u.practice_id::text,''), COALESCE(pr.name,''),
			COALESCE(pr.city,''), COALESCE(pr.postal_code,''), u.created_at
		FROM identity.users u
		LEFT JOIN practice.practices pr ON pr.id = u.practice_id
		WHERE u.role = 'vet' AND u.assigned_commercial_id IS NULL
		ORDER BY u.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]UnassignedVetRow, 0)
	for rows.Next() {
		var v UnassignedVetRow
		if err := rows.Scan(&v.UserID, &v.FullName, &v.Email, &v.PracticeID, &v.PracticeName,
			&v.City, &v.PostalCode, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// UnassignVetFromCommercial clears assigned_commercial_id for a vet.
func (s *Store) UnassignVetFromCommercial(ctx context.Context, vetUserID string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE identity.users SET assigned_commercial_id = NULL
		WHERE id = $1 AND role = 'vet'`, vetUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type commercialScoreSeed struct {
	UserID         string
	FullName       string
	Email          string
	City           string
	PostalCode     string
	BaseLat        *float64
	BaseLng        *float64
	LastLoginAt    *time.Time
	AssignedVets   int
	Contacts30d    int
	StaleCount     int
}

// SuggestCommercialsForVet returns top 5 commercials by deterministic zone+activity score.
func (s *Store) SuggestCommercialsForVet(ctx context.Context, vetUserID string) ([]CommercialSuggestion, error) {
	var practicePostal string
	var practiceLat, practiceLng *float64
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(pr.postal_code,''),
			NULL::float8, NULL::float8
		FROM identity.users u
		LEFT JOIN practice.practices pr ON pr.id = u.practice_id
		WHERE u.id = $1 AND u.role = 'vet'`, vetUserID).Scan(&practicePostal, &practiceLat, &practiceLng)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.full_name, u.email,
			COALESCE(u.base_city,''), COALESCE(u.base_postal_code,''),
			u.base_lat, u.base_lng, u.last_login_at,
			COALESCE((SELECT COUNT(*)::int FROM identity.users v
				WHERE v.role='vet' AND v.assigned_commercial_id = u.id), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p
				WHERE p.commercial_user_id = u.id
				  AND p.updated_at >= NOW() - INTERVAL '30 days'), 0),
			COALESCE((SELECT COUNT(*)::int FROM sales.prospects p
				WHERE p.commercial_user_id = u.id
				  AND p.status NOT IN ('converted','lost')
				  AND GREATEST(p.status_changed_at, COALESCE(p.last_contacted_at, p.status_changed_at), p.updated_at)
				      < NOW() - make_interval(days => $1)), 0)
		FROM identity.users u
		WHERE u.role = 'commercial'
		ORDER BY u.full_name`, ProspectInactiveThresholdDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seeds := make([]commercialScoreSeed, 0)
	for rows.Next() {
		var c commercialScoreSeed
		if err := rows.Scan(&c.UserID, &c.FullName, &c.Email, &c.City, &c.PostalCode,
			&c.BaseLat, &c.BaseLng, &c.LastLoginAt, &c.AssignedVets, &c.Contacts30d, &c.StaleCount); err != nil {
			return nil, err
		}
		seeds = append(seeds, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	vetDigits := digitsOnly(practicePostal)
	vetPrefix := vetDigits
	if len(vetPrefix) > 2 {
		vetPrefix = vetPrefix[:2]
	}

	out := make([]CommercialSuggestion, 0, len(seeds))
	for _, c := range seeds {
		sug := scoreCommercialForVet(c, vetDigits, vetPrefix, practiceLat, practiceLng)
		out = append(out, sug)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].FullName < out[j].FullName
	})
	if len(out) > 5 {
		out = out[:5]
	}
	return out, nil
}

func scoreCommercialForVet(c commercialScoreSeed, vetDigits, vetPrefix string, practiceLat, practiceLng *float64) CommercialSuggestion {
	reasons := make([]string, 0, 6)
	zone := 0
	commDigits := digitsOnly(c.PostalCode)
	commPrefix := commDigits
	if len(commPrefix) > 2 {
		commPrefix = commPrefix[:2]
	}
	if vetDigits != "" && commDigits != "" {
		if vetDigits == commDigits {
			zone = 50
			reasons = append(reasons, "postal_exact")
		} else if len(vetPrefix) >= 2 && vetPrefix == commPrefix {
			zone = 30
			reasons = append(reasons, "postal_prefix")
		}
	}

	var distKm *float64
	if practiceLat != nil && practiceLng != nil && c.BaseLat != nil && c.BaseLng != nil {
		d := haversineKm(*practiceLat, *practiceLng, *c.BaseLat, *c.BaseLng)
		rounded := math.Round(d*10) / 10
		distKm = &rounded
		bonus := 0
		switch {
		case d <= 10:
			bonus = 20
			reasons = append(reasons, "distance_near")
		case d <= 30:
			bonus = 10
			reasons = append(reasons, "distance_mid")
		case d <= 50:
			bonus = 5
			reasons = append(reasons, "distance_far")
		}
		zone += bonus
		if zone > 50 {
			zone = 50
		}
	}

	activity := 0
	now := time.Now()
	if c.LastLoginAt != nil {
		days := now.Sub(*c.LastLoginAt).Hours() / 24
		switch {
		case days <= 7:
			activity += 20
			reasons = append(reasons, "login_recent")
		case days <= 30:
			activity += 10
			reasons = append(reasons, "login_month")
		}
	}
	contactPts := c.Contacts30d * 3
	if contactPts > 15 {
		contactPts = 15
	}
	if contactPts > 0 {
		activity += contactPts
		reasons = append(reasons, "contacts_30d")
	}
	stalePts := 15 - c.StaleCount*3
	if stalePts < 0 {
		stalePts = 0
	}
	if c.StaleCount == 0 {
		reasons = append(reasons, "low_stale")
	}
	activity += stalePts
	loadPenalty := c.AssignedVets * 2
	if loadPenalty > 10 {
		loadPenalty = 10
	}
	activity -= loadPenalty
	if loadPenalty == 0 {
		reasons = append(reasons, "load_light")
	}
	if activity < 0 {
		activity = 0
	}
	if activity > 50 {
		activity = 50
	}

	note := fmt.Sprintf("Score %d (zone %d + activité %d). %s — %s.",
		zone+activity, zone, activity, c.FullName, strings.TrimSpace(c.City+" "+c.PostalCode))
	if distKm != nil {
		note = fmt.Sprintf("%s Distance ~%.0f km.", note, *distKm)
	}

	return CommercialSuggestion{
		UserID:        c.UserID,
		FullName:      c.FullName,
		Email:         c.Email,
		City:          c.City,
		PostalCode:    c.PostalCode,
		Score:         zone + activity,
		ZoneScore:     zone,
		ActivityScore: activity,
		Reasons:       reasons,
		Note:          note,
		DistanceKm:    distKm,
		AssignedVets:  c.AssignedVets,
	}
}

func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const r = 6371.0
	toRad := math.Pi / 180
	dLat := (lat2 - lat1) * toRad
	dLng := (lng2 - lng1) * toRad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*toRad)*math.Cos(lat2*toRad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * r * math.Asin(math.Min(1, math.Sqrt(a)))
}
