package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Sentinel conflicts for imaging.pet_studies (compatible with errors.Is(..., ErrConflict) via handler checks).
var (
	ErrPacsStudyOtherPet      = errors.New("pacs_study_other_pet")
	ErrPacsStudyOtherPractice = errors.New("pacs_study_other_practice")
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

func (s *Store) classifyPetStudyConflict(existing PetStudy, in CreatePetStudyInput) error {
	if existing.PracticeID != in.PracticeID {
		return ErrPacsStudyOtherPractice
	}
	if existing.PetID != in.PetID {
		return ErrPacsStudyOtherPet
	}
	return nil
}

func (s *Store) updatePetStudyMeta(ctx context.Context, in CreatePetStudyInput) (PetStudy, error) {
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

func (s *Store) CreatePetStudy(ctx context.Context, in CreatePetStudyInput) (PetStudy, error) {
	if strings.TrimSpace(in.OrthancStudyID) == "" {
		return PetStudy{}, ErrValidation
	}
	existing, err := s.FindPetStudyByOrthancStudyID(ctx, in.OrthancStudyID)
	if err == nil {
		if cerr := s.classifyPetStudyConflict(existing, in); cerr != nil {
			return PetStudy{}, cerr
		}
		return s.updatePetStudyMeta(ctx, in)
	}
	if !errors.Is(err, ErrNotFound) {
		return PetStudy{}, err
	}
	id := uuid.NewString()
	row, err := scanPetStudy(s.pool.QueryRow(ctx, `
		INSERT INTO imaging.pet_studies (
			id, pet_id, practice_id, orthanc_study_id, study_instance_uid,
			orthanc_series_id, description, modality, uploaded_by_user_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9,'')::uuid)
		RETURNING id, pet_id, practice_id, orthanc_study_id, study_instance_uid,
			orthanc_series_id, description, modality, COALESCE(uploaded_by_user_id::text,''), created_at
	`, id, in.PetID, in.PracticeID, in.OrthancStudyID, in.StudyInstanceUID,
		in.OrthancSeriesID, in.Description, in.Modality, in.UploadedByUserID))
	if err == nil {
		return row, nil
	}
	if !isUniqueViolation(err) {
		return PetStudy{}, err
	}
	// Race: another practice/pet won the insert — classify from winner row.
	winner, findErr := s.FindPetStudyByOrthancStudyID(ctx, in.OrthancStudyID)
	if findErr != nil {
		return PetStudy{}, ErrConflict
	}
	if cerr := s.classifyPetStudyConflict(winner, in); cerr != nil {
		return PetStudy{}, cerr
	}
	return s.updatePetStudyMeta(ctx, in)
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

// FindPetStudyByOrthancStudyID looks up any practice binding for an Orthanc study (global uniqueness).
func (s *Store) FindPetStudyByOrthancStudyID(ctx context.Context, orthancStudyID string) (PetStudy, error) {
	return scanPetStudy(s.pool.QueryRow(ctx, petStudySelect+`
		WHERE orthanc_study_id = $1
	`, orthancStudyID))
}

// FindPetStudyByOrthancSeries — tenant gate for Orthanc series proxy.
func (s *Store) FindPetStudyByOrthancSeries(ctx context.Context, practiceID, orthancSeriesID string) (PetStudy, error) {
	return scanPetStudy(s.pool.QueryRow(ctx, petStudySelect+`
		WHERE practice_id = $1 AND orthanc_series_id = $2
	`, practiceID, orthancSeriesID))
}

// ListAllPetStudies returns imaging links (admin ops — prune orphans).
// Hard cap 10_000; caller should treat len==limit as possibly truncated.
func (s *Store) ListAllPetStudies(ctx context.Context, limit int) ([]PetStudy, error) {
	if limit <= 0 || limit > 10000 {
		limit = 10000
	}
	rows, err := s.pool.Query(ctx, petStudySelect+`
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
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

// DeletePetStudyByID removes an imaging.pet_studies row (admin prune / ops).
func (s *Store) DeletePetStudyByID(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM imaging.pet_studies WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// PlaygroundPacsClient is a seed client (@petsfollow.test) with at least one pet — admin PACS picker.
type PlaygroundPacsClient struct {
	UserID   string `json:"userId"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	PetCount int    `json:"petCount"`
}

// ListPlaygroundPacsClients returns demo clients for the admin PACS client→pet picker (no arbitrary PHI).
func (s *Store) ListPlaygroundPacsClients(ctx context.Context) ([]PlaygroundPacsClient, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id::text, u.email, COALESCE(u.full_name, ''), COUNT(p.id)::int
		FROM identity.users u
		JOIN pets.pets p ON p.owner_user_id = u.id
		WHERE u.role = 'client'
		  AND u.email LIKE '%@petsfollow.test'
		  AND u.email NOT LIKE '%@deleted.petsfollow.invalid'
		GROUP BY u.id, u.email, u.full_name
		HAVING COUNT(p.id) > 0
		ORDER BY u.full_name, u.email
		LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PlaygroundPacsClient
	for rows.Next() {
		var c PlaygroundPacsClient
		if err := rows.Scan(&c.UserID, &c.Email, &c.FullName, &c.PetCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// PetStudyComment is an append-only clinical note on an imaging.pet_studies row.
type PetStudyComment struct {
	ID        string                `json:"id"`
	Body      string                `json:"body"`
	CreatedAt time.Time             `json:"createdAt"`
	Author    PetStudyCommentAuthor `json:"author"`
}

type PetStudyCommentAuthor struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// GetPetStudyForPet returns a study row scoped to pet + practice (anti-IDOR).
func (s *Store) GetPetStudyForPet(ctx context.Context, studyRowID, petID, practiceID string) (PetStudy, error) {
	return scanPetStudy(s.pool.QueryRow(ctx, petStudySelect+`
		WHERE id = $1 AND pet_id = $2 AND practice_id = $3
	`, studyRowID, petID, practiceID))
}

// ListPetStudyComments returns comments newest-first.
func (s *Store) ListPetStudyComments(ctx context.Context, petStudyID string) ([]PetStudyComment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.id::text, c.body, c.created_at,
			c.author_user_id::text, COALESCE(u.full_name, '')
		FROM imaging.pet_study_comments c
		JOIN identity.users u ON u.id = c.author_user_id
		WHERE c.pet_study_id = $1
		ORDER BY c.created_at DESC
	`, petStudyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PetStudyComment
	for rows.Next() {
		var c PetStudyComment
		if err := rows.Scan(&c.ID, &c.Body, &c.CreatedAt, &c.Author.ID, &c.Author.DisplayName); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// InsertPetStudyComment appends a comment (body 1–4000 runes after trim).
func (s *Store) InsertPetStudyComment(ctx context.Context, petStudyID, authorUserID, body string) (PetStudyComment, error) {
	body = strings.TrimSpace(body)
	if body == "" || len([]rune(body)) > 4000 {
		return PetStudyComment{}, ErrValidation
	}
	id := uuid.NewString()
	var c PetStudyComment
	err := s.pool.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO imaging.pet_study_comments (id, pet_study_id, author_user_id, body)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4)
			RETURNING id, body, created_at, author_user_id
		)
		SELECT i.id::text, i.body, i.created_at, i.author_user_id::text, COALESCE(u.full_name, '')
		FROM inserted i
		JOIN identity.users u ON u.id = i.author_user_id
	`, id, petStudyID, authorUserID, body).Scan(
		&c.ID, &c.Body, &c.CreatedAt, &c.Author.ID, &c.Author.DisplayName,
	)
	return c, err
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
