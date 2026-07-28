package store

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/olegrand1976/petsFollow/go/internal/prescription"
)

type Prescription struct {
	ID             string          `json:"id"`
	PracticeID     string          `json:"practiceId"`
	VeterinaryID   string          `json:"veterinaryId"`
	VeterinaryName string          `json:"veterinaryName,omitempty"`
	PetID          string          `json:"petId"`
	PetName        string          `json:"petName,omitempty"`
	PetSpecies     string          `json:"petSpecies,omitempty"`
	OwnerID        string          `json:"ownerId"`
	OwnerName      string          `json:"ownerName,omitempty"`
	PracticeName   string          `json:"practiceName,omitempty"`
	CountryCode    string          `json:"countryCode"`
	DateIssued     *string         `json:"dateIssued,omitempty"`
	ValidUntil     *string         `json:"validUntil,omitempty"`
	SignatureID    string          `json:"signatureId,omitempty"`
	PDFObjectKey   string          `json:"pdfObjectKey,omitempty"`
	PDFSHA256      string          `json:"pdfSha256,omitempty"`
	Status         string          `json:"status"`
	PaperFormat    string          `json:"paperFormat"`
	Medications    json.RawMessage `json:"medications"`
	Notes          string          `json:"notes,omitempty"`
	CreatedAt      string          `json:"createdAt,omitempty"`
	UpdatedAt      string          `json:"updatedAt,omitempty"`
}

type PrescriptionPatch struct {
	Medications *json.RawMessage
	Notes       *string
	PaperFormat *string
	ValidUntil  *time.Time
	ClearValid  bool
}

const prescriptionSelectCols = `
		SELECT p.id::text, p.practice_id::text, p.veterinary_id::text, COALESCE(v.full_name,''),
		       p.pet_id::text, COALESCE(pet.name,''), COALESCE(pet.species,''),
		       p.owner_id::text, COALESCE(o.full_name,''), COALESCE(pr.name,''),
		       p.country_code, p.date_issued, p.valid_until,
		       COALESCE(p.signature_id::text,''), COALESCE(p.pdf_object_key,''), COALESCE(p.pdf_sha256,''),
		       p.status, p.paper_format, p.medications, COALESCE(p.notes,''),
		       p.created_at, p.updated_at
		FROM prescriptions.prescriptions p
		JOIN identity.users v ON v.id = p.veterinary_id
		JOIN identity.users o ON o.id = p.owner_id
		JOIN pets.pets pet ON pet.id = p.pet_id
		JOIN practice.practices pr ON pr.id = p.practice_id`

// CreatePrescriptionDraft inserts a draft. Caller must have already loaded and ACL-checked pet.
func (s *Store) CreatePrescriptionDraft(
	ctx context.Context,
	practiceID, veterinaryID string,
	pet Pet,
	meds []prescription.Medication,
	notes, paperFormat string,
	validUntil *time.Time,
) (Prescription, error) {
	if pet.ID == "" || pet.PracticeID == "" || pet.PracticeID != practiceID {
		return Prescription{}, ErrForbidden
	}
	notesNorm, err := prescription.NormalizeNotes(notes)
	if err != nil {
		return Prescription{}, err
	}
	contact, err := s.GetPracticeContact(ctx, practiceID)
	if err != nil {
		return Prescription{}, err
	}
	country := NormalizeCountryCode(contact.CountryCode)
	medJSON, err := prescription.MedicationsJSON(meds)
	if err != nil {
		return Prescription{}, err
	}
	format, err := prescription.NormalizePaperFormat(paperFormat)
	if err != nil {
		return Prescription{}, err
	}
	id := uuid.NewString()
	_, err = s.pool.Exec(ctx, `
		INSERT INTO prescriptions.prescriptions (
			id, practice_id, veterinary_id, pet_id, owner_id, country_code,
			valid_until, status, paper_format, medications, notes
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,'draft',$8,$9::jsonb,$10
		)`,
		id, practiceID, veterinaryID, pet.ID, pet.OwnerUserID, country,
		validUntil, format, string(medJSON), notesNorm,
	)
	if err != nil {
		return Prescription{}, err
	}
	return s.GetPrescription(ctx, practiceID, id)
}

func (s *Store) ListPrescriptions(ctx context.Context, practiceID, status, petID string) ([]Prescription, error) {
	q := prescriptionSelectCols + ` WHERE p.practice_id = $1`
	args := []any{practiceID}
	n := 2
	if status != "" {
		q += ` AND p.status = $` + strconv.Itoa(n)
		args = append(args, status)
		n++
	}
	if petID != "" {
		q += ` AND p.pet_id = $` + strconv.Itoa(n)
		args = append(args, petID)
	}
	q += ` ORDER BY p.created_at DESC LIMIT 200`
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Prescription
	for rows.Next() {
		doc, err := scanPrescription(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, doc)
	}
	return out, rows.Err()
}

func (s *Store) GetPrescription(ctx context.Context, practiceID, id string) (Prescription, error) {
	row := s.pool.QueryRow(ctx, prescriptionSelectCols+`
		WHERE p.practice_id = $1 AND p.id = $2`, practiceID, id)
	doc, err := scanPrescription(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Prescription{}, prescription.ErrNotFound
	}
	return doc, err
}

func (s *Store) PatchPrescriptionDraft(ctx context.Context, practiceID, id string, patch PrescriptionPatch) (Prescription, error) {
	cur, err := s.GetPrescription(ctx, practiceID, id)
	if err != nil {
		return Prescription{}, err
	}
	if cur.Status != prescription.StatusDraft {
		return Prescription{}, prescription.ErrNotDraft
	}
	notes := cur.Notes
	if patch.Notes != nil {
		notes, err = prescription.NormalizeNotes(*patch.Notes)
		if err != nil {
			return Prescription{}, err
		}
	}
	format := cur.PaperFormat
	if patch.PaperFormat != nil {
		format, err = prescription.NormalizePaperFormat(*patch.PaperFormat)
		if err != nil {
			return Prescription{}, err
		}
	}
	medJSON := cur.Medications
	if patch.Medications != nil {
		list, nerr := prescription.NormalizeMedications(*patch.Medications)
		if nerr != nil {
			return Prescription{}, nerr
		}
		medJSON, err = prescription.MedicationsJSON(list)
		if err != nil {
			return Prescription{}, err
		}
	}

	touchValid := patch.ClearValid || patch.ValidUntil != nil
	var tag pgconn.CommandTag
	if touchValid {
		var validUntil any
		if !patch.ClearValid && patch.ValidUntil != nil {
			validUntil = *patch.ValidUntil
		}
		tag, err = s.pool.Exec(ctx, `
			UPDATE prescriptions.prescriptions SET
				notes = $3, paper_format = $4, medications = $5::jsonb,
				valid_until = $6, updated_at = now()
			WHERE practice_id = $1 AND id = $2 AND status = 'draft'`,
			practiceID, id, notes, format, string(medJSON), validUntil,
		)
	} else {
		tag, err = s.pool.Exec(ctx, `
			UPDATE prescriptions.prescriptions SET
				notes = $3, paper_format = $4, medications = $5::jsonb, updated_at = now()
			WHERE practice_id = $1 AND id = $2 AND status = 'draft'`,
			practiceID, id, notes, format, string(medJSON),
		)
	}
	if err != nil {
		return Prescription{}, err
	}
	if tag.RowsAffected() == 0 {
		cur2, gerr := s.GetPrescription(ctx, practiceID, id)
		if gerr != nil {
			return Prescription{}, gerr
		}
		if cur2.Status != prescription.StatusDraft {
			return Prescription{}, prescription.ErrNotDraft
		}
		return Prescription{}, prescription.ErrNotFound
	}
	return s.GetPrescription(ctx, practiceID, id)
}

func (s *Store) DeletePrescriptionDraft(ctx context.Context, practiceID, id string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM prescriptions.prescriptions
		WHERE practice_id = $1 AND id = $2 AND status = 'draft'`, practiceID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		cur, gerr := s.GetPrescription(ctx, practiceID, id)
		if gerr != nil {
			return gerr
		}
		if cur.Status != prescription.StatusDraft {
			return prescription.ErrNotDraft
		}
		return prescription.ErrNotFound
	}
	return nil
}

type prescriptionRow interface {
	Scan(dest ...any) error
}

func scanPrescription(row prescriptionRow) (Prescription, error) {
	var p Prescription
	var dateIssued, validUntil *time.Time
	var createdAt, updatedAt time.Time
	var meds []byte
	err := row.Scan(
		&p.ID, &p.PracticeID, &p.VeterinaryID, &p.VeterinaryName,
		&p.PetID, &p.PetName, &p.PetSpecies,
		&p.OwnerID, &p.OwnerName, &p.PracticeName,
		&p.CountryCode, &dateIssued, &validUntil,
		&p.SignatureID, &p.PDFObjectKey, &p.PDFSHA256,
		&p.Status, &p.PaperFormat, &meds, &p.Notes,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return Prescription{}, err
	}
	p.Medications = json.RawMessage(meds)
	if dateIssued != nil {
		s := dateIssued.UTC().Format(time.RFC3339)
		p.DateIssued = &s
	}
	if validUntil != nil {
		s := validUntil.UTC().Format(time.RFC3339)
		p.ValidUntil = &s
	}
	p.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	p.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	return p, nil
}
