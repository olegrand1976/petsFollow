package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProfessionalContact struct {
	ID          string    `json:"id"`
	PetID       string    `json:"petId"`
	OwnerUserID string    `json:"ownerUserId"`
	Role        string    `json:"role"`
	FullName    string    `json:"fullName"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Competition struct {
	ID          string    `json:"id"`
	PetID       string    `json:"petId"`
	OwnerUserID string    `json:"ownerUserId"`
	EventDate   string    `json:"eventDate"`
	Title       string    `json:"title"`
	Location    string    `json:"location"`
	Discipline  string    `json:"discipline"`
	Result      string    `json:"result"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (s *Store) ListProfessionalContacts(ctx context.Context, petID, ownerUserID string) ([]ProfessionalContact, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, pet_id::text, owner_user_id::text, role, full_name, phone, email, notes, created_at, updated_at
		FROM care.professional_contacts
		WHERE pet_id=$1 AND owner_user_id=$2
		ORDER BY full_name`, petID, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ProfessionalContact, 0)
	for rows.Next() {
		var c ProfessionalContact
		if err := rows.Scan(&c.ID, &c.PetID, &c.OwnerUserID, &c.Role, &c.FullName, &c.Phone, &c.Email, &c.Notes, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) CreateProfessionalContact(ctx context.Context, petID, ownerUserID, role, fullName, phone, email, notes string) (ProfessionalContact, error) {
	id := uuid.NewString()
	var c ProfessionalContact
	err := s.pool.QueryRow(ctx, `
		INSERT INTO care.professional_contacts (id, pet_id, owner_user_id, role, full_name, phone, email, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id::text, pet_id::text, owner_user_id::text, role, full_name, phone, email, notes, created_at, updated_at`,
		id, petID, ownerUserID, role, fullName, phone, email, notes,
	).Scan(&c.ID, &c.PetID, &c.OwnerUserID, &c.Role, &c.FullName, &c.Phone, &c.Email, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (s *Store) DeleteProfessionalContact(ctx context.Context, id, ownerUserID string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM care.professional_contacts WHERE id=$1 AND owner_user_id=$2`, id, ownerUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpdateProfessionalContact(ctx context.Context, id, ownerUserID, role, fullName, phone, email, notes string) (ProfessionalContact, error) {
	var c ProfessionalContact
	err := s.pool.QueryRow(ctx, `
		UPDATE care.professional_contacts
		SET role=$3, full_name=$4, phone=$5, email=$6, notes=$7, updated_at=NOW()
		WHERE id=$1 AND owner_user_id=$2
		RETURNING id::text, pet_id::text, owner_user_id::text, role, full_name, phone, email, notes, created_at, updated_at`,
		id, ownerUserID, role, fullName, phone, email, notes,
	).Scan(&c.ID, &c.PetID, &c.OwnerUserID, &c.Role, &c.FullName, &c.Phone, &c.Email, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProfessionalContact{}, ErrNotFound
	}
	return c, err
}

func (s *Store) ListCompetitions(ctx context.Context, petID, ownerUserID string) ([]Competition, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, pet_id::text, owner_user_id::text, event_date::text, title, location, discipline, result, notes, created_at, updated_at
		FROM care.competitions
		WHERE pet_id=$1 AND owner_user_id=$2
		ORDER BY event_date DESC`, petID, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Competition, 0)
	for rows.Next() {
		var c Competition
		if err := rows.Scan(&c.ID, &c.PetID, &c.OwnerUserID, &c.EventDate, &c.Title, &c.Location, &c.Discipline, &c.Result, &c.Notes, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) CreateCompetition(ctx context.Context, petID, ownerUserID, eventDate, title, location, discipline, result, notes string) (Competition, error) {
	id := uuid.NewString()
	var c Competition
	err := s.pool.QueryRow(ctx, `
		INSERT INTO care.competitions (id, pet_id, owner_user_id, event_date, title, location, discipline, result, notes)
		VALUES ($1,$2,$3,$4::date,$5,$6,$7,$8,$9)
		RETURNING id::text, pet_id::text, owner_user_id::text, event_date::text, title, location, discipline, result, notes, created_at, updated_at`,
		id, petID, ownerUserID, eventDate, title, location, discipline, result, notes,
	).Scan(&c.ID, &c.PetID, &c.OwnerUserID, &c.EventDate, &c.Title, &c.Location, &c.Discipline, &c.Result, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (s *Store) DeleteCompetition(ctx context.Context, id, ownerUserID string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM care.competitions WHERE id=$1 AND owner_user_id=$2`, id, ownerUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpdateCompetition(ctx context.Context, id, ownerUserID, eventDate, title, location, discipline, result, notes string) (Competition, error) {
	var c Competition
	err := s.pool.QueryRow(ctx, `
		UPDATE care.competitions
		SET event_date=$3::date, title=$4, location=$5, discipline=$6, result=$7, notes=$8, updated_at=NOW()
		WHERE id=$1 AND owner_user_id=$2
		RETURNING id::text, pet_id::text, owner_user_id::text, event_date::text, title, location, discipline, result, notes, created_at, updated_at`,
		id, ownerUserID, eventDate, title, location, discipline, result, notes,
	).Scan(&c.ID, &c.PetID, &c.OwnerUserID, &c.EventDate, &c.Title, &c.Location, &c.Discipline, &c.Result, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Competition{}, ErrNotFound
	}
	return c, err
}

// PetOwnedBy checks pet belongs to owner (and optionally horse species).
func (s *Store) PetOwnedBy(ctx context.Context, petID, ownerUserID string) (species string, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT species FROM pets.pets WHERE id=$1 AND owner_user_id=$2`, petID, ownerUserID).Scan(&species)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return species, err
}
