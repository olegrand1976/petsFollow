package store

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
)

const (
	defaultNearbyLimit    = 5
	defaultNearbyRadiusKm = 80.0
	maxNearbyLimit        = 20
)

// haversineKmSQL is the Postgres expression for distance in km between
// ($1,$2) = (lat,lng) and (base_lat, base_lng). Used once via subquery.
const haversineKmSQL = `(6371 * acos(
	LEAST(1.0, GREATEST(-1.0,
		cos(radians($1)) * cos(radians(base_lat)) *
		cos(radians(base_lng) - radians($2)) +
		sin(radians($1)) * sin(radians(base_lat))
	))
))`

// NearbyCommercial is a public discovery row (no email/phone/IBAN).
type NearbyCommercial struct {
	UserID     string  `json:"userId"`
	FullName   string  `json:"fullName"`
	City       string  `json:"city"`
	DistanceKm float64 `json:"distanceKm,omitempty"`
	PostalCode string  `json:"postalCode,omitempty"`
	InviteCode string  `json:"inviteCode,omitempty"`
}

// NearbyCommercialsQuery filters discovery by GPS and/or postal code.
type NearbyCommercialsQuery struct {
	Lat        *float64
	Lng        *float64
	PostalCode string
	Limit      int
	RadiusKm   float64
}

// ListNearbyCommercials returns commercials ordered by distance (GPS) or postal prefix.
// Only role=commercial (not commercial_manager) — signup assignment targets field reps.
func (s *Store) ListNearbyCommercials(ctx context.Context, q NearbyCommercialsQuery) ([]NearbyCommercial, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = defaultNearbyLimit
	}
	if limit > maxNearbyLimit {
		limit = maxNearbyLimit
	}
	radius := q.RadiusKm
	if radius <= 0 {
		radius = defaultNearbyRadiusKm
	}
	postal := strings.TrimSpace(q.PostalCode)

	hasGPS := q.Lat != nil && q.Lng != nil &&
		!math.IsNaN(*q.Lat) && !math.IsNaN(*q.Lng) &&
		*q.Lat >= -90 && *q.Lat <= 90 && *q.Lng >= -180 && *q.Lng <= 180

	if hasGPS {
		return s.listNearbyByGPS(ctx, *q.Lat, *q.Lng, radius, limit)
	}
	if postal != "" {
		return s.listNearbyByPostal(ctx, postal, limit)
	}
	return []NearbyCommercial{}, nil
}

func (s *Store) listNearbyByGPS(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]NearbyCommercial, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, full_name, city, postal_code, distance_km, invite_code FROM (
			SELECT u.id, u.full_name, COALESCE(u.base_city,'') AS city, COALESCE(u.base_postal_code,'') AS postal_code,
				(6371 * acos(
					LEAST(1.0, GREATEST(-1.0,
						cos(radians($1)) * cos(radians(u.base_lat)) *
						cos(radians(u.base_lng) - radians($2)) +
						sin(radians($1)) * sin(radians(u.base_lat))
					))
				)) AS distance_km,
				COALESCE(ic.code,'') AS invite_code
			FROM identity.users u
			LEFT JOIN practice.app_invite_codes ic ON ic.user_id = u.id
			WHERE u.role = 'commercial'
			  AND u.base_lat IS NOT NULL AND u.base_lng IS NOT NULL
		) d
		WHERE distance_km <= $3
		ORDER BY distance_km ASC, full_name ASC
		LIMIT $4`,
		lat, lng, radiusKm, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanNearbyCommercials(ctx, rows, true)
}

func (s *Store) listNearbyByPostal(ctx context.Context, postal string, limit int) ([]NearbyCommercial, error) {
	digits := digitsOnly(postal)
	prefix := digits
	if len(prefix) > 2 {
		prefix = prefix[:2]
	}
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.full_name, COALESCE(u.base_city,''), COALESCE(u.base_postal_code,''),
			0::float8 AS distance_km, COALESCE(ic.code,'')
		FROM identity.users u
		LEFT JOIN practice.app_invite_codes ic ON ic.user_id = u.id
		WHERE u.role = 'commercial'
		  AND (
			u.base_lat IS NOT NULL AND u.base_lng IS NOT NULL
			OR COALESCE(u.base_postal_code,'') <> ''
		  )
		  AND (
			regexp_replace(COALESCE(u.base_postal_code,''), '[^0-9]', '', 'g') = $1
			OR (
				length($2) >= 2
				AND left(regexp_replace(COALESCE(u.base_postal_code,''), '[^0-9]', '', 'g'), 2) = $2
			)
		  )
		ORDER BY
			CASE WHEN regexp_replace(COALESCE(u.base_postal_code,''), '[^0-9]', '', 'g') = $1 THEN 0 ELSE 1 END,
			u.full_name ASC
		LIMIT $3`,
		digits, prefix, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanNearbyCommercials(ctx, rows, false)
}

func (s *Store) scanNearbyCommercials(ctx context.Context, rows pgx.Rows, withDistance bool) ([]NearbyCommercial, error) {
	out := make([]NearbyCommercial, 0)
	for rows.Next() {
		var c NearbyCommercial
		var dist float64
		if err := rows.Scan(&c.UserID, &c.FullName, &c.City, &c.PostalCode, &dist, &c.InviteCode); err != nil {
			return nil, err
		}
		if withDistance {
			c.DistanceKm = math.Round(dist*10) / 10
		}
		if c.InviteCode == "" {
			if inv, err := s.EnsureAppInviteCode(ctx, c.UserID); err == nil {
				c.InviteCode = inv.Code
			}
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// CommercialBaseLocation is the commercial's public base area.
// JSON tags aligned with payout profile / admin PATCH (baseLat, baseCity, …).
type CommercialBaseLocation struct {
	Lat        *float64 `json:"baseLat,omitempty"`
	Lng        *float64 `json:"baseLng,omitempty"`
	City       string   `json:"baseCity"`
	PostalCode string   `json:"basePostalCode"`
}

// IsAssignableCommercial returns true if userID is an active field commercial
// (not commercial_manager — managers are not offered at signup nearby).
func (s *Store) IsAssignableCommercial(ctx context.Context, userID string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM identity.users
			WHERE id=$1 AND role='commercial'
		)`, userID).Scan(&ok)
	return ok, err
}

// canOwnCommercialReferral — commercial or commercial_manager may own a client referral
// (QR claim, standalone create). Nearby discovery still uses IsAssignableCommercial.
func (s *Store) canOwnCommercialReferral(ctx context.Context, userID string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM identity.users
			WHERE id=$1 AND role IN ('commercial','commercial_manager')
		)`, userID).Scan(&ok)
	return ok, err
}

// LinkClientCommercialReferral soft-links a client to a commercial (first wins).
func (s *Store) LinkClientCommercialReferral(ctx context.Context, clientUserID, commercialUserID string) error {
	ok, err := s.canOwnCommercialReferral(ctx, commercialUserID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO practice.commercial_referrals (client_user_id, commercial_user_id, invite_code, updated_at)
		VALUES ($1, $2, '', NOW())
		ON CONFLICT (client_user_id) DO NOTHING`,
		clientUserID, commercialUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		_ = s.RecordFiliationEvent(ctx, FiliationEventInput{
			EventType:        FiliationEventClientReferral,
			CommercialUserID: commercialUserID,
			ClientUserID:     clientUserID,
			ActorUserID:      commercialUserID,
			Meta:             map[string]any{"source": "link_referral"},
		})
	}
	return nil
}

// UpdateCommercialBaseLocation sets base_* for a commercial or manager user.
func (s *Store) UpdateCommercialBaseLocation(ctx context.Context, userID string, loc CommercialBaseLocation) error {
	city := strings.TrimSpace(loc.City)
	postal := strings.TrimSpace(loc.PostalCode)
	if len(city) > 120 {
		city = city[:120]
	}
	if len(postal) > 20 {
		postal = postal[:20]
	}
	var lat, lng any
	if loc.Lat != nil && loc.Lng != nil {
		if *loc.Lat < -90 || *loc.Lat > 90 || *loc.Lng < -180 || *loc.Lng > 180 {
			return ErrValidation
		}
		lat, lng = *loc.Lat, *loc.Lng
	} else if loc.Lat == nil && loc.Lng == nil {
		lat, lng = nil, nil
	} else {
		return ErrValidation
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE identity.users
		SET base_lat=$2, base_lng=$3, base_city=$4, base_postal_code=$5
		WHERE id=$1 AND role IN ('commercial', 'commercial_manager')`,
		userID, lat, lng, city, postal)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetCommercialBaseLocation returns base location fields.
func (s *Store) GetCommercialBaseLocation(ctx context.Context, userID string) (CommercialBaseLocation, error) {
	var loc CommercialBaseLocation
	var lat, lng *float64
	err := s.pool.QueryRow(ctx, `
		SELECT base_lat, base_lng, COALESCE(base_city,''), COALESCE(base_postal_code,'')
		FROM identity.users
		WHERE id=$1 AND role IN ('commercial', 'commercial_manager')`, userID).Scan(
		&lat, &lng, &loc.City, &loc.PostalCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return CommercialBaseLocation{}, ErrNotFound
	}
	if err != nil {
		return CommercialBaseLocation{}, err
	}
	loc.Lat, loc.Lng = lat, lng
	return loc, nil
}
