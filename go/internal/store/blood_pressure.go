package store

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	MaxBPCommentLen = 500
	MaxBPSiteLen    = 80
	MinSystolic     = 20
	MaxSystolic     = 400
	MinDiastolic    = 10
	MaxDiastolic    = 300
)

var ErrInvalidBloodPressure = errors.New("invalid blood pressure")

type BloodPressureMethod string

const (
	BPMethodDoppler       BloodPressureMethod = "doppler"
	BPMethodOscillometric BloodPressureMethod = "oscillometric"
	BPMethodInvasive      BloodPressureMethod = "invasive"
	BPMethodUnknown       BloodPressureMethod = "unknown"
)

func NormalizeBPMethod(raw string) BloodPressureMethod {
	switch BloodPressureMethod(strings.TrimSpace(strings.ToLower(raw))) {
	case BPMethodDoppler, BPMethodOscillometric, BPMethodInvasive:
		return BloodPressureMethod(strings.TrimSpace(strings.ToLower(raw)))
	default:
		return BPMethodUnknown
	}
}

type BloodPressureReading struct {
	ID            string              `json:"id"`
	PetID         string              `json:"petId"`
	OwnerUserID   string              `json:"ownerUserId"`
	AuthorUserID  string              `json:"authorUserId"`
	PracticeID    string              `json:"practiceId"`
	SystolicMmHg  int                 `json:"systolicMmHg"`
	DiastolicMmHg int                 `json:"diastolicMmHg"`
	MeanMmHg      *int                `json:"meanMmHg,omitempty"`
	Method        BloodPressureMethod `json:"method"`
	Site          *string             `json:"site,omitempty"`
	Comment       *string             `json:"comment,omitempty"`
	RecordedAt    time.Time           `json:"recordedAt"`
}

func NormalizeBPSite(raw *string) *string {
	if raw == nil {
		return nil
	}
	s := strings.TrimSpace(*raw)
	if s == "" {
		return nil
	}
	if utf8.RuneCountInString(s) > MaxBPSiteLen {
		s = string([]rune(s)[:MaxBPSiteLen])
	}
	return &s
}

func NormalizeBPComment(raw *string) *string {
	return NormalizeWeightComment(raw)
}

func ValidateBloodPressure(sys, dia int, mean *int) error {
	if sys < MinSystolic || sys > MaxSystolic {
		return ErrInvalidBloodPressure
	}
	if dia < MinDiastolic || dia > MaxDiastolic {
		return ErrInvalidBloodPressure
	}
	if dia > sys {
		return ErrInvalidBloodPressure
	}
	if mean != nil {
		if *mean < 10 || *mean > 350 || *mean < dia || *mean > sys {
			return ErrInvalidBloodPressure
		}
	}
	return nil
}

func (s *Store) CreateBloodPressureReading(
	ctx context.Context,
	petID, ownerID, authorID, practiceID string,
	sys, dia int,
	mean *int,
	method BloodPressureMethod,
	site, comment *string,
) (BloodPressureReading, error) {
	if err := ValidateBloodPressure(sys, dia, mean); err != nil {
		return BloodPressureReading{}, err
	}
	if method == "" {
		method = BPMethodUnknown
	}
	site = NormalizeBPSite(site)
	comment = NormalizeBPComment(comment)
	reading := BloodPressureReading{
		ID:            uuid.NewString(),
		PetID:         petID,
		OwnerUserID:   ownerID,
		AuthorUserID:  authorID,
		PracticeID:    practiceID,
		SystolicMmHg:  sys,
		DiastolicMmHg: dia,
		MeanMmHg:      mean,
		Method:        method,
		Site:          site,
		Comment:       comment,
		RecordedAt:    time.Now().UTC(),
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO pets.blood_pressure_readings (
			id, pet_id, owner_user_id, author_user_id, practice_id,
			systolic_mmhg, diastolic_mmhg, mean_mmhg, method, site, comment, recorded_at
		) VALUES ($1,$2,$3,$4,NULLIF($5,'')::uuid,$6,$7,$8,$9,$10,$11,$12)
		RETURNING recorded_at`,
		reading.ID, reading.PetID, reading.OwnerUserID, reading.AuthorUserID,
		reading.PracticeID, reading.SystolicMmHg, reading.DiastolicMmHg, reading.MeanMmHg,
		string(reading.Method), reading.Site, reading.Comment, reading.RecordedAt,
	).Scan(&reading.RecordedAt)
	if err != nil {
		return BloodPressureReading{}, err
	}
	return reading, nil
}

func (s *Store) ListBloodPressureReadings(ctx context.Context, petID string) ([]BloodPressureReading, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, pet_id::text, owner_user_id::text, author_user_id::text,
			COALESCE(practice_id::text, ''),
			systolic_mmhg, diastolic_mmhg, mean_mmhg, method, site, comment, recorded_at
		FROM pets.blood_pressure_readings
		WHERE pet_id = $1
		ORDER BY recorded_at DESC`, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BloodPressureReading
	for rows.Next() {
		var r BloodPressureReading
		var method string
		if err := rows.Scan(
			&r.ID, &r.PetID, &r.OwnerUserID, &r.AuthorUserID, &r.PracticeID,
			&r.SystolicMmHg, &r.DiastolicMmHg, &r.MeanMmHg, &method, &r.Site, &r.Comment, &r.RecordedAt,
		); err != nil {
			return nil, err
		}
		r.Method = BloodPressureMethod(method)
		out = append(out, r)
	}
	if out == nil {
		out = []BloodPressureReading{}
	}
	return out, rows.Err()
}
