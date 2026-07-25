package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PetDocument struct {
	ID               string    `json:"id"`
	PetID            string    `json:"petId"`
	UploadedByUserID string    `json:"uploadedByUserId"`
	UploaderName     string    `json:"uploaderName,omitempty"`
	Title            string    `json:"title"`
	FileName         string    `json:"fileName"`
	ContentType      string    `json:"contentType"`
	FileURL          string    `json:"fileUrl"`
	ObjectKey        string    `json:"-"`
	SizeBytes        int64     `json:"sizeBytes"`
	CreatedAt        time.Time `json:"createdAt"`
}

type CreatePetDocumentInput struct {
	PetID            string
	UploadedByUserID string
	Title            string
	FileName         string
	ContentType      string
	FileURL          string
	ObjectKey        string
	SizeBytes        int64
}

func (s *Store) ListPetDocuments(ctx context.Context, petID string) ([]PetDocument, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.id::text, d.pet_id::text, d.uploaded_by_user_id::text,
			COALESCE(u.full_name, ''), COALESCE(d.title, ''), d.file_name,
			d.content_type, d.file_url, COALESCE(d.object_key, ''), d.size_bytes, d.created_at
		FROM pets.documents d
		LEFT JOIN identity.users u ON u.id = d.uploaded_by_user_id
		WHERE d.pet_id = $1
		ORDER BY d.created_at DESC`, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PetDocument
	for rows.Next() {
		var d PetDocument
		if err := rows.Scan(
			&d.ID, &d.PetID, &d.UploadedByUserID, &d.UploaderName,
			&d.Title, &d.FileName, &d.ContentType, &d.FileURL, &d.ObjectKey, &d.SizeBytes, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []PetDocument{}
	}
	return out, nil
}

func (s *Store) CreatePetDocument(ctx context.Context, in CreatePetDocumentInput) (PetDocument, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = strings.TrimSpace(in.FileName)
	}
	d := PetDocument{
		ID:               uuid.NewString(),
		PetID:            in.PetID,
		UploadedByUserID: in.UploadedByUserID,
		Title:            title,
		FileName:         in.FileName,
		ContentType:      in.ContentType,
		FileURL:          in.FileURL,
		ObjectKey:        in.ObjectKey,
		SizeBytes:        in.SizeBytes,
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO pets.documents (
			id, pet_id, uploaded_by_user_id, title, file_name, content_type, file_url, object_key, size_bytes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING created_at`,
		d.ID, d.PetID, d.UploadedByUserID, d.Title, d.FileName, d.ContentType, d.FileURL, d.ObjectKey, d.SizeBytes,
	).Scan(&d.CreatedAt)
	return d, err
}

func (s *Store) GetPetDocument(ctx context.Context, id string) (PetDocument, error) {
	var d PetDocument
	err := s.pool.QueryRow(ctx, `
		SELECT d.id::text, d.pet_id::text, d.uploaded_by_user_id::text,
			COALESCE(u.full_name, ''), COALESCE(d.title, ''), d.file_name,
			d.content_type, d.file_url, COALESCE(d.object_key, ''), d.size_bytes, d.created_at
		FROM pets.documents d
		LEFT JOIN identity.users u ON u.id = d.uploaded_by_user_id
		WHERE d.id = $1`, id).Scan(
		&d.ID, &d.PetID, &d.UploadedByUserID, &d.UploaderName,
		&d.Title, &d.FileName, &d.ContentType, &d.FileURL, &d.ObjectKey, &d.SizeBytes, &d.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return PetDocument{}, ErrNotFound
	}
	return d, err
}

func (s *Store) DeletePetDocument(ctx context.Context, id string) (PetDocument, error) {
	doc, err := s.GetPetDocument(ctx, id)
	if err != nil {
		return PetDocument{}, err
	}
	tag, err := s.pool.Exec(ctx, `DELETE FROM pets.documents WHERE id = $1`, id)
	if err != nil {
		return PetDocument{}, err
	}
	if tag.RowsAffected() == 0 {
		return PetDocument{}, ErrNotFound
	}
	return doc, nil
}
