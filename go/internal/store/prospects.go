package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Prospect struct {
	ID                 string    `json:"id"`
	CommercialUserID   string    `json:"commercialUserId"`
	PracticeName       string    `json:"practiceName"`
	ContactName        string    `json:"contactName"`
	ContactEmail       string    `json:"contactEmail"`
	ContactPhone       string    `json:"contactPhone"`
	City               string    `json:"city"`
	Notes              string    `json:"notes"`
	Source             string    `json:"source"`
	ReferringVetUserID string    `json:"referringVetUserId,omitempty"`
	Status             string    `json:"status"`
	StatusChangedAt    time.Time `json:"statusChangedAt"`
	DaysInStatus       int       `json:"daysInStatus"`
	CreatedAt          time.Time `json:"createdAt"`
	CommercialName     string    `json:"commercialName,omitempty"`
	CommercialEmail    string    `json:"commercialEmail,omitempty"`
}

type ProspectInput struct {
	PracticeName string
	ContactName  string
	ContactEmail string
	ContactPhone string
	City         string
	Notes        string
	Status       string
	Source       string
}

func ValidProspectStatus(status string) bool {
	switch status {
	case "new", "contacted", "qualified", "converted", "lost":
		return true
	default:
		return false
	}
}

func ValidProspectSource(source string) bool {
	switch source {
	case "commercial", "vet_referral":
		return true
	default:
		return false
	}
}

func (s *Store) CreateProspect(ctx context.Context, commercialUserID string, in ProspectInput) (Prospect, error) {
	status := in.Status
	if status == "" {
		status = "new"
	}
	source := in.Source
	if source == "" {
		source = "commercial"
	}
	if !ValidProspectSource(source) {
		return Prospect{}, errors.New("invalid prospect source")
	}
	id := uuid.NewString()
	var p Prospect
	err := s.pool.QueryRow(ctx, `
		INSERT INTO sales.prospects (
			id, commercial_user_id, practice_name, contact_name, contact_email, contact_phone,
			city, notes, source, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id::text, commercial_user_id::text, practice_name, contact_name, contact_email, contact_phone,
			city, notes, source, COALESCE(referring_vet_user_id::text,''), status, status_changed_at, created_at`,
		id, commercialUserID, in.PracticeName, in.ContactName, in.ContactEmail, in.ContactPhone,
		in.City, in.Notes, source, status).Scan(
		&p.ID, &p.CommercialUserID, &p.PracticeName, &p.ContactName, &p.ContactEmail, &p.ContactPhone,
		&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt)
	if err != nil {
		return Prospect{}, err
	}
	p.DaysInStatus = daysSince(p.StatusChangedAt)
	return p, nil
}

// CreateVetReferralProspect creates a prospect for the commercial assigned to the vet.
func (s *Store) CreateVetReferralProspect(ctx context.Context, vetUserID string, in ProspectInput) (Prospect, error) {
	commercialID, err := s.GetAssignedCommercialID(ctx, vetUserID)
	if err != nil {
		return Prospect{}, err
	}
	if commercialID == "" {
		return Prospect{}, ErrNotFound
	}
	status := in.Status
	if status == "" {
		status = "new"
	}
	id := uuid.NewString()
	var p Prospect
	err = s.pool.QueryRow(ctx, `
		INSERT INTO sales.prospects (
			id, commercial_user_id, practice_name, contact_name, contact_email, contact_phone,
			city, notes, source, referring_vet_user_id, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'vet_referral',$9,$10)
		RETURNING id::text, commercial_user_id::text, practice_name, contact_name, contact_email, contact_phone,
			city, notes, source, COALESCE(referring_vet_user_id::text,''), status, status_changed_at, created_at`,
		id, commercialID, in.PracticeName, in.ContactName, in.ContactEmail, in.ContactPhone,
		in.City, in.Notes, vetUserID, status).Scan(
		&p.ID, &p.CommercialUserID, &p.PracticeName, &p.ContactName, &p.ContactEmail, &p.ContactPhone,
		&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt)
	if err != nil {
		return Prospect{}, err
	}
	p.DaysInStatus = daysSince(p.StatusChangedAt)
	return p, nil
}

func (s *Store) GetProspect(ctx context.Context, commercialUserID, id string) (Prospect, error) {
	var p Prospect
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, commercial_user_id::text, practice_name, contact_name, contact_email, contact_phone,
			city, notes, source, COALESCE(referring_vet_user_id::text,''), status, status_changed_at, created_at
		FROM sales.prospects WHERE id=$1 AND commercial_user_id=$2`, id, commercialUserID).Scan(
		&p.ID, &p.CommercialUserID, &p.PracticeName, &p.ContactName, &p.ContactEmail, &p.ContactPhone,
		&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Prospect{}, ErrNotFound
	}
	if err != nil {
		return Prospect{}, err
	}
	p.DaysInStatus = daysSince(p.StatusChangedAt)
	return p, nil
}

func (s *Store) UpdateProspect(ctx context.Context, commercialUserID, id string, in ProspectInput) (Prospect, error) {
	status := in.Status
	if status == "" {
		status = "new"
	}
	var p Prospect
	err := s.pool.QueryRow(ctx, `
		UPDATE sales.prospects SET
			practice_name=$3, contact_name=$4, contact_email=$5, contact_phone=$6, city=$7, notes=$8,
			status=$9,
			status_changed_at=CASE WHEN status <> $9 THEN NOW() ELSE status_changed_at END,
			updated_at=NOW()
		WHERE id=$1 AND commercial_user_id=$2
		RETURNING id::text, commercial_user_id::text, practice_name, contact_name, contact_email, contact_phone,
			city, notes, source, COALESCE(referring_vet_user_id::text,''), status, status_changed_at, created_at`,
		id, commercialUserID, in.PracticeName, in.ContactName, in.ContactEmail, in.ContactPhone, in.City, in.Notes, status).Scan(
		&p.ID, &p.CommercialUserID, &p.PracticeName, &p.ContactName, &p.ContactEmail, &p.ContactPhone,
		&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Prospect{}, ErrNotFound
	}
	if err != nil {
		return Prospect{}, err
	}
	p.DaysInStatus = daysSince(p.StatusChangedAt)
	return p, nil
}

func (s *Store) DeleteProspect(ctx context.Context, commercialUserID, id string) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM sales.prospects WHERE id=$1 AND commercial_user_id=$2`, id, commercialUserID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListProspects(ctx context.Context, commercialUserID, statusFilter string) ([]Prospect, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, commercial_user_id::text, practice_name, contact_name, contact_email, contact_phone,
			city, notes, source, COALESCE(referring_vet_user_id::text,''), status, status_changed_at, created_at
		FROM sales.prospects
		WHERE commercial_user_id=$1 AND ($2='' OR status=$2)
		ORDER BY created_at DESC`, commercialUserID, statusFilter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProspects(rows)
}

func (s *Store) ListAllProspects(ctx context.Context, statusFilter string) ([]Prospect, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id::text, p.commercial_user_id::text, p.practice_name, p.contact_name, p.contact_email, p.contact_phone,
			p.city, p.notes, p.source, COALESCE(p.referring_vet_user_id::text,''), p.status, p.status_changed_at, p.created_at,
			COALESCE(u.full_name,''), COALESCE(u.email,'')
		FROM sales.prospects p
		JOIN identity.users u ON u.id = p.commercial_user_id
		WHERE ($1='' OR p.status=$1)
		ORDER BY p.created_at DESC`, statusFilter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Prospect, 0)
	for rows.Next() {
		var p Prospect
		if err := rows.Scan(&p.ID, &p.CommercialUserID, &p.PracticeName, &p.ContactName, &p.ContactEmail, &p.ContactPhone,
			&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt,
			&p.CommercialName, &p.CommercialEmail); err != nil {
			return nil, err
		}
		p.DaysInStatus = daysSince(p.StatusChangedAt)
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanProspects(rows pgx.Rows) ([]Prospect, error) {
	out := make([]Prospect, 0)
	for rows.Next() {
		var p Prospect
		if err := rows.Scan(&p.ID, &p.CommercialUserID, &p.PracticeName, &p.ContactName, &p.ContactEmail, &p.ContactPhone,
			&p.City, &p.Notes, &p.Source, &p.ReferringVetUserID, &p.Status, &p.StatusChangedAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.DaysInStatus = daysSince(p.StatusChangedAt)
		out = append(out, p)
	}
	return out, rows.Err()
}

func daysSince(t time.Time) int {
	d := int(time.Since(t).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}
