package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PetStudy struct {
	ID               string    `json:"id"`
	PetID            string    `json:"petId"`
	PracticeID       string    `json:"practiceId"`
	OrthancStudyID   string    `json:"orthancStudyId"`
	StudyInstanceUID string    `json:"studyInstanceUid"`
	OrthancSeriesID  string    `json:"orthancSeriesId"`
	Description      string    `json:"description"`
	Modality         string    `json:"modality"`
	UploadedByUserID string    `json:"uploadedByUserId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
}

type CreatePetStudyInput struct {
	PetID            string
	PracticeID       string
	OrthancStudyID   string
	StudyInstanceUID string
	OrthancSeriesID  string
	Description      string
	Modality         string
	UploadedByUserID string
}

func scanPetStudy(row pgx.Row) (PetStudy, error) {
	var out PetStudy
	err := row.Scan(&out.ID, &out.PetID, &out.PracticeID, &out.OrthancStudyID, &out.StudyInstanceUID,
		&out.OrthancSeriesID, &out.Description, &out.Modality, &out.UploadedByUserID, &out.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PetStudy{}, ErrNotFound
	}
	return out, err
}

const petStudySelect = `
	SELECT id, pet_id, practice_id, orthanc_study_id, study_instance_uid,
		orthanc_series_id, description, modality, COALESCE(uploaded_by_user_id::text,''), created_at
	FROM imaging.pet_studies`

func (s *Store) CreatePetStudy(ctx context.Context, in CreatePetStudyInput) (PetStudy, error) {
	if strings.TrimSpace(in.OrthancStudyID) == "" {
		return PetStudy{}, ErrValidation
	}
	existing, err := s.FindPetStudyByOrthancStudy(ctx, in.PracticeID, in.OrthancStudyID)
	if err == nil {
		if existing.PetID != in.PetID {
			return PetStudy{}, ErrConflict
		}
		return scanPetStudy(s.pool.QueryRow(ctx, `
			UPDATE imaging.pet_studies SET
				study_instance_uid = $3,
				orthanc_series_id = $4,
				description = $5,
				modality = $6
			WHERE practice_id = $1 AND orthanc_study_id = $2
			RETURNING id, pet_id, practice_id, orthanc_study_id, study_instance_uid,
				orthanc_series_id, description, modality, COALESCE(uploaded_by_user_id::text,''), created_at
		`, in.PracticeID, in.OrthancStudyID, in.StudyInstanceUID, in.OrthancSeriesID, in.Description, in.Modality))
	}
	if !errors.Is(err, ErrNotFound) {
		return PetStudy{}, err
	}
	id := uuid.NewString()
	return scanPetStudy(s.pool.QueryRow(ctx, `
		INSERT INTO imaging.pet_studies (
			id, pet_id, practice_id, orthanc_study_id, study_instance_uid,
			orthanc_series_id, description, modality, uploaded_by_user_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9,'')::uuid)
		RETURNING id, pet_id, practice_id, orthanc_study_id, study_instance_uid,
			orthanc_series_id, description, modality, COALESCE(uploaded_by_user_id::text,''), created_at
	`, id, in.PetID, in.PracticeID, in.OrthancStudyID, in.StudyInstanceUID,
		in.OrthancSeriesID, in.Description, in.Modality, in.UploadedByUserID))
}

func (s *Store) ListPetStudies(ctx context.Context, petID, practiceID string) ([]PetStudy, error) {
	rows, err := s.pool.Query(ctx, petStudySelect+`
		WHERE pet_id = $1 AND practice_id = $2
		ORDER BY created_at DESC
	`, petID, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PetStudy
	for rows.Next() {
		var st PetStudy
		if err := rows.Scan(&st.ID, &st.PetID, &st.PracticeID, &st.OrthancStudyID, &st.StudyInstanceUID,
			&st.OrthancSeriesID, &st.Description, &st.Modality, &st.UploadedByUserID, &st.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (s *Store) GetPetStudy(ctx context.Context, studyRowID, practiceID string) (PetStudy, error) {
	return scanPetStudy(s.pool.QueryRow(ctx, petStudySelect+`
		WHERE id = $1 AND practice_id = $2
	`, studyRowID, practiceID))
}

// FindPetStudyByOrthancStudy — tenant gate for Orthanc study proxy.
func (s *Store) FindPetStudyByOrthancStudy(ctx context.Context, practiceID, orthancStudyID string) (PetStudy, error) {
	return scanPetStudy(s.pool.QueryRow(ctx, petStudySelect+`
		WHERE practice_id = $1 AND orthanc_study_id = $2
	`, practiceID, orthancStudyID))
}

// FindPetStudyByOrthancSeries — tenant gate for Orthanc series proxy.
func (s *Store) FindPetStudyByOrthancSeries(ctx context.Context, practiceID, orthancSeriesID string) (PetStudy, error) {
	return scanPetStudy(s.pool.QueryRow(ctx, petStudySelect+`
		WHERE practice_id = $1 AND orthanc_series_id = $2
	`, practiceID, orthancSeriesID))
}

// ListPetStudyOrthancIDsForOwner returns Orthanc study IDs for pets owned by userID (RGPD purge).
func (s *Store) ListPetStudyOrthancIDsForOwner(ctx context.Context, ownerUserID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ps.orthanc_study_id
		FROM imaging.pet_studies ps
		JOIN pets.pets p ON p.id = ps.pet_id
		WHERE p.owner_user_id = $1 AND ps.orthanc_study_id <> ''
	`, ownerUserID)
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
