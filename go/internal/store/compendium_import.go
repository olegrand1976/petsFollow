package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

// CompendiumImportJob tracks PDF → AI extract → human review → upsert.
type CompendiumImportJob struct {
	ID               string    `json:"id"`
	CreatedByAdminID string    `json:"createdByAdminId"`
	Filename         string    `json:"filename"`
	ContentType      string    `json:"contentType"`
	PageStart        int       `json:"pageStart"`
	PageEnd          int       `json:"pageEnd"`
	Status           string    `json:"status"`
	PDFObjectKey     string    `json:"pdfObjectKey,omitempty"`
	ExtractDone      int       `json:"extractDone"`
	ExtractTotal     int       `json:"extractTotal"`
	ExtractPct       int       `json:"extractPct"`
	RowCount         int       `json:"rowCount"`
	ReadyCount       int       `json:"readyCount"`
	ErrorCount       int       `json:"errorCount"`
	ReviewedCount    int       `json:"reviewedCount"`
	ReviewPct        int       `json:"reviewPct"`
	UpsertedCount    int       `json:"upsertedCount"`
	ErrorMessage     string    `json:"errorMessage,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// CompendiumImportRow is one staged medication.
type CompendiumImportRow struct {
	ID                 string          `json:"id"`
	JobID              string          `json:"jobId"`
	RowNumber          int             `json:"rowNumber"`
	SourcePage         *int            `json:"sourcePage,omitempty"`
	CNK                string          `json:"cnk"`
	Name               string          `json:"name"`
	Manufacturer       string          `json:"manufacturer,omitempty"`
	ActiveSubstance    string          `json:"activeSubstance,omitempty"`
	Strength           string          `json:"strength,omitempty"`
	ATCCode            string          `json:"atcCode"`
	PharmaceuticalForm string          `json:"pharmaceuticalForm"`
	PackSize           string          `json:"packSize"`
	IsAntibiotic       bool            `json:"isAntibiotic"`
	SuggestedCNK       string          `json:"suggestedCnk,omitempty"`
	MatchScore         *float64        `json:"matchScore,omitempty"`
	MatchCandidates    json.RawMessage `json:"matchCandidates,omitempty"`
	RawJSON            json.RawMessage `json:"rawJson,omitempty"`
	Status             string          `json:"status"`
	ErrorCode          string          `json:"errorCode,omitempty"`
	ErrorMessage       string          `json:"errorMessage,omitempty"`
}

type CompendiumImportDetail struct {
	Job             CompendiumImportJob   `json:"job"`
	Rows            []CompendiumImportRow `json:"rows"`
	RefCatalogCount int                   `json:"refCatalogCount"`
}

type CreateCompendiumImportInput struct {
	ID               string // optional; generated when empty
	CreatedByAdminID string
	Filename         string
	ContentType      string
	PageStart        int
	PageEnd          int
	PDFObjectKey     string
}

func pct(done, total int) int {
	if total <= 0 {
		return 0
	}
	if done < 0 {
		done = 0
	}
	if done > total {
		done = total
	}
	return (done * 100) / total
}

func (j *CompendiumImportJob) withPercents() {
	j.ExtractPct = pct(j.ExtractDone, j.ExtractTotal)
	j.ReviewPct = pct(j.ReviewedCount, j.RowCount)
}

func scanCompendiumJob(row pgx.Row) (CompendiumImportJob, error) {
	var j CompendiumImportJob
	var errMsg *string
	err := row.Scan(
		&j.ID, &j.CreatedByAdminID, &j.Filename, &j.ContentType,
		&j.PageStart, &j.PageEnd, &j.Status, &j.PDFObjectKey,
		&j.ExtractDone, &j.ExtractTotal,
		&j.RowCount, &j.ReadyCount, &j.ErrorCount, &j.ReviewedCount, &j.UpsertedCount,
		&errMsg, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return j, err
	}
	if errMsg != nil {
		j.ErrorMessage = *errMsg
	}
	j.withPercents()
	return j, nil
}

const compendiumJobCols = `
	id::text, created_by_admin_id::text, filename, content_type,
	page_start, page_end, status, COALESCE(pdf_object_key, ''),
	extract_done, extract_total,
	row_count, ready_count, error_count, reviewed_count, upserted_count,
	error_message, created_at, updated_at`

// CreateCompendiumImportJob inserts an uploaded job (extract not started).
func (s *Store) CreateCompendiumImportJob(ctx context.Context, in CreateCompendiumImportInput) (CompendiumImportJob, error) {
	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = uuid.NewString()
	}
	j, err := scanCompendiumJob(s.pool.QueryRow(ctx, `
		INSERT INTO pharmacy.compendium_import_jobs (
			id, created_by_admin_id, filename, content_type, page_start, page_end,
			status, pdf_object_key, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,'uploaded',$7,now())
		RETURNING `+compendiumJobCols, id, in.CreatedByAdminID, in.Filename, in.ContentType,
		in.PageStart, in.PageEnd, in.PDFObjectKey))
	return j, err
}

func (s *Store) GetCompendiumImportJob(ctx context.Context, id string) (CompendiumImportJob, error) {
	j, err := scanCompendiumJob(s.pool.QueryRow(ctx, `
		SELECT `+compendiumJobCols+` FROM pharmacy.compendium_import_jobs WHERE id = $1`, id))
	if err == pgx.ErrNoRows {
		return j, ErrNotFound
	}
	return j, err
}

func (s *Store) ListCompendiumImportJobs(ctx context.Context, limit int) ([]CompendiumImportJob, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT `+compendiumJobCols+`
		FROM pharmacy.compendium_import_jobs
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CompendiumImportJob, 0)
	for rows.Next() {
		j, err := scanCompendiumJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (s *Store) GetCompendiumImportDetail(ctx context.Context, id string) (CompendiumImportDetail, error) {
	job, err := s.GetCompendiumImportJob(ctx, id)
	if err != nil {
		return CompendiumImportDetail{}, err
	}
	rows, err := s.ListCompendiumImportRows(ctx, id)
	if err != nil {
		return CompendiumImportDetail{}, err
	}
	n, _ := s.CountActiveRefMedications(ctx)
	return CompendiumImportDetail{Job: job, Rows: rows, RefCatalogCount: n}, nil
}

func (s *Store) CountActiveRefMedications(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM pharmacy.ref_medications WHERE is_active`).Scan(&n)
	return n, err
}

func (s *Store) ListCompendiumImportRows(ctx context.Context, jobID string) ([]CompendiumImportRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, job_id::text, row_number, source_page,
		       cnk, name, COALESCE(manufacturer,''), COALESCE(active_substance,''), COALESCE(strength,''),
		       atc_code, pharmaceutical_form, pack_size, is_antibiotic,
		       COALESCE(suggested_cnk,''), match_score, COALESCE(match_candidates, '[]'::jsonb),
		       raw_json, status, COALESCE(error_code,''), COALESCE(error_message,'')
		FROM pharmacy.compendium_import_rows
		WHERE job_id = $1
		ORDER BY row_number ASC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CompendiumImportRow, 0)
	for rows.Next() {
		var r CompendiumImportRow
		var raw, cands []byte
		if err := rows.Scan(
			&r.ID, &r.JobID, &r.RowNumber, &r.SourcePage,
			&r.CNK, &r.Name, &r.Manufacturer, &r.ActiveSubstance, &r.Strength,
			&r.ATCCode, &r.PharmaceuticalForm, &r.PackSize, &r.IsAntibiotic,
			&r.SuggestedCNK, &r.MatchScore, &cands,
			&raw, &r.Status, &r.ErrorCode, &r.ErrorMessage,
		); err != nil {
			return nil, err
		}
		r.RawJSON = raw
		r.MatchCandidates = cands
		out = append(out, r)
	}
	return out, rows.Err()
}

// BeginCompendiumExtractResult describes how the next extract run should proceed.
type BeginCompendiumExtractResult struct {
	ResumeFromChunk int // 0-based index into ChunkPageRanges
}

const compendiumExtractStaleAfter = 15 * time.Minute

// BeginCompendiumExtract claims the job for extraction.
// forceRestart clears progress + rows. Otherwise a failed/stale job resumes from extract_done.
// Active extracting jobs updated within 15 minutes refuse a second claim (anti double-run).
func (s *Store) BeginCompendiumExtract(ctx context.Context, id string, extractTotal int, forceRestart bool) (BeginCompendiumExtractResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return BeginCompendiumExtractResult{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	var extractDone int
	var updatedAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT status, extract_done, updated_at
		FROM pharmacy.compendium_import_jobs
		WHERE id = $1
		FOR UPDATE`, id).Scan(&status, &extractDone, &updatedAt)
	if err == pgx.ErrNoRows {
		return BeginCompendiumExtractResult{}, ErrNotFound
	}
	if err != nil {
		return BeginCompendiumExtractResult{}, err
	}
	switch status {
	case "uploaded", "failed", "extracting":
	default:
		return BeginCompendiumExtractResult{}, ErrConflict
	}
	if status == "extracting" && time.Since(updatedAt) < compendiumExtractStaleAfter {
		return BeginCompendiumExtractResult{}, ErrConflict
	}

	resumeFrom := 0
	canResume := !forceRestart && status != "uploaded" && extractDone > 0 && extractDone < extractTotal
	if canResume {
		resumeFrom = extractDone
	} else {
		if _, err := tx.Exec(ctx, `DELETE FROM pharmacy.compendium_import_rows WHERE job_id = $1`, id); err != nil {
			return BeginCompendiumExtractResult{}, err
		}
		extractDone = 0
		resumeFrom = 0
	}

	_, err = tx.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs
		SET status = 'extracting', extract_total = $2, extract_done = $3,
		    error_message = NULL, updated_at = now()
		WHERE id = $1`, id, extractTotal, extractDone)
	if err != nil {
		return BeginCompendiumExtractResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return BeginCompendiumExtractResult{}, err
	}
	return BeginCompendiumExtractResult{ResumeFromChunk: resumeFrom}, nil
}

// MarkCompendiumExtracting is kept for tests that force an extracting status without a run.
func (s *Store) MarkCompendiumExtracting(ctx context.Context, id string, extractTotal int) error {
	_, err := s.BeginCompendiumExtract(ctx, id, extractTotal, true)
	return err
}

func (s *Store) SetCompendiumExtractProgress(ctx context.Context, id string, done int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs
		SET extract_done = $2, updated_at = now()
		WHERE id = $1`, id, done)
	return err
}

type CompendiumRowInsert struct {
	SourcePage         *int
	CNK                string
	Name               string
	Manufacturer       string
	ActiveSubstance    string
	Strength           string
	ATCCode            string
	PharmaceuticalForm string
	PackSize           string
	IsAntibiotic       bool
	SuggestedCNK       string
	MatchScore         *float64
	MatchCandidates    json.RawMessage
	RawJSON            json.RawMessage
	Status             string
	ErrorCode          string
	ErrorMessage       string
}

// AppendCompendiumExtractRows inserts rows for one completed chunk (keeps prior chunks on failure).
func (s *Store) AppendCompendiumExtractRows(ctx context.Context, jobID string, rows []CompendiumRowInsert) error {
	if len(rows) == 0 {
		return nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var base int
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(row_number), 0) FROM pharmacy.compendium_import_rows WHERE job_id = $1`, jobID).Scan(&base); err != nil {
		return err
	}
	for i, row := range rows {
		raw := row.RawJSON
		if len(raw) == 0 {
			raw = json.RawMessage(`{}`)
		}
		cands := row.MatchCandidates
		if len(cands) == 0 {
			cands = json.RawMessage(`[]`)
		}
		status := row.Status
		if status == "" {
			status = "pending"
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO pharmacy.compendium_import_rows (
				id, job_id, row_number, source_page, cnk, name,
				manufacturer, active_substance, strength, atc_code,
				pharmaceutical_form, pack_size, is_antibiotic,
				suggested_cnk, match_score, match_candidates,
				raw_json, status, error_code, error_message
			) VALUES (
				$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16::jsonb,$17::jsonb,$18,NULLIF($19,''),NULLIF($20,'')
			)`,
			uuid.NewString(), jobID, base+i+1, row.SourcePage,
			strings.TrimSpace(row.CNK), strings.TrimSpace(row.Name),
			strings.TrimSpace(row.Manufacturer), strings.TrimSpace(row.ActiveSubstance),
			strings.TrimSpace(row.Strength), strings.TrimSpace(row.ATCCode),
			strings.TrimSpace(row.PharmaceuticalForm),
			strings.TrimSpace(row.PackSize), row.IsAntibiotic,
			strings.TrimSpace(row.SuggestedCNK), row.MatchScore, string(cands),
			string(raw), status, row.ErrorCode, row.ErrorMessage,
		); err != nil {
			return err
		}
	}
	if err := refreshCompendiumJobCountsTx(ctx, tx, jobID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// FinalizeCompendiumExtract marks the job extracted after all chunks succeeded.
func (s *Store) FinalizeCompendiumExtract(ctx context.Context, jobID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := refreshCompendiumJobCountsTx(ctx, tx, jobID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs
		SET status = 'extracted',
		    extract_done = extract_total,
		    error_message = NULL,
		    updated_at = now()
		WHERE id = $1`, jobID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ReplaceCompendiumExtractRows clears previous rows and inserts a new extract batch; marks extracted.
func (s *Store) ReplaceCompendiumExtractRows(ctx context.Context, jobID string, rows []CompendiumRowInsert) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM pharmacy.compendium_import_rows WHERE job_id = $1`, jobID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	if err := s.AppendCompendiumExtractRows(ctx, jobID, rows); err != nil {
		return err
	}
	return s.FinalizeCompendiumExtract(ctx, jobID)
}

func (s *Store) FailCompendiumImportJob(ctx context.Context, id, msg string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs
		SET status = 'failed', error_message = $2, updated_at = now()
		WHERE id = $1`, id, msg)
	return err
}

// DeleteCompendiumImportJob removes a staging job (+ CASCADE rows).
// Refuses extracting/committing. Does not touch pharmacy.ref_medications.
// Returns the PDF object key for media cleanup.
func (s *Store) DeleteCompendiumImportJob(ctx context.Context, id string) (pdfObjectKey string, err error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var status, key string
	err = tx.QueryRow(ctx, `
		SELECT status, pdf_object_key
		FROM pharmacy.compendium_import_jobs
		WHERE id = $1
		FOR UPDATE`, id).Scan(&status, &key)
	if err == pgx.ErrNoRows {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if status == "extracting" || status == "committing" {
		return "", ErrConflict
	}
	if _, err := tx.Exec(ctx, `DELETE FROM pharmacy.compendium_import_jobs WHERE id = $1`, id); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return key, nil
}

type PatchCompendiumRowInput struct {
	CNK                *string `json:"cnk"`
	Name               *string `json:"name"`
	Manufacturer       *string `json:"manufacturer"`
	ActiveSubstance    *string `json:"activeSubstance"`
	Strength           *string `json:"strength"`
	ATCCode            *string `json:"atcCode"`
	PharmaceuticalForm *string `json:"pharmaceuticalForm"`
	PackSize           *string `json:"packSize"`
	IsAntibiotic       *bool   `json:"isAntibiotic"`
	SourcePage         *int    `json:"sourcePage"`
	Excluded           *bool   `json:"excluded"`
}

func (s *Store) PatchCompendiumImportRow(ctx context.Context, jobID, rowID string, in PatchCompendiumRowInput) (CompendiumImportRow, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CompendiumImportRow{}, err
	}
	defer tx.Rollback(ctx)

	var r CompendiumImportRow
	var raw, cands []byte
	err = tx.QueryRow(ctx, `
		SELECT id::text, job_id::text, row_number, source_page,
		       cnk, name, COALESCE(manufacturer,''), COALESCE(active_substance,''), COALESCE(strength,''),
		       atc_code, pharmaceutical_form, pack_size, is_antibiotic,
		       COALESCE(suggested_cnk,''), match_score, COALESCE(match_candidates, '[]'::jsonb),
		       raw_json, status, COALESCE(error_code,''), COALESCE(error_message,'')
		FROM pharmacy.compendium_import_rows
		WHERE id = $1 AND job_id = $2
		FOR UPDATE`, rowID, jobID).Scan(
		&r.ID, &r.JobID, &r.RowNumber, &r.SourcePage,
		&r.CNK, &r.Name, &r.Manufacturer, &r.ActiveSubstance, &r.Strength,
		&r.ATCCode, &r.PharmaceuticalForm, &r.PackSize, &r.IsAntibiotic,
		&r.SuggestedCNK, &r.MatchScore, &cands,
		&raw, &r.Status, &r.ErrorCode, &r.ErrorMessage,
	)
	if err == pgx.ErrNoRows {
		return r, ErrNotFound
	}
	if err != nil {
		return r, err
	}
	r.RawJSON = raw
	r.MatchCandidates = cands
	if r.Status == "upserted" {
		return r, ErrConflict
	}

	if in.CNK != nil {
		r.CNK = strings.TrimSpace(*in.CNK)
	}
	if in.Name != nil {
		r.Name = strings.TrimSpace(*in.Name)
	}
	if in.Manufacturer != nil {
		r.Manufacturer = strings.TrimSpace(*in.Manufacturer)
	}
	if in.ActiveSubstance != nil {
		r.ActiveSubstance = strings.TrimSpace(*in.ActiveSubstance)
	}
	if in.Strength != nil {
		r.Strength = strings.TrimSpace(*in.Strength)
	}
	if in.ATCCode != nil {
		r.ATCCode = strings.TrimSpace(*in.ATCCode)
	}
	if in.PharmaceuticalForm != nil {
		r.PharmaceuticalForm = strings.TrimSpace(*in.PharmaceuticalForm)
	}
	if in.PackSize != nil {
		r.PackSize = strings.TrimSpace(*in.PackSize)
	}
	if in.IsAntibiotic != nil {
		r.IsAntibiotic = *in.IsAntibiotic
	}
	if in.SourcePage != nil {
		if *in.SourcePage < 1 {
			r.SourcePage = nil
		} else {
			p := *in.SourcePage
			r.SourcePage = &p
		}
	}
	if in.Excluded != nil {
		if *in.Excluded {
			r.Status = "excluded"
			r.ErrorCode = ""
			r.ErrorMessage = ""
		} else if r.Status == "excluded" {
			r.Status = "pending"
		}
	}
	if r.Status != "excluded" {
		// Any PATCH is a human control action → confirm valid rows to ready.
		st, code, msg := pharmacy.ClassifyExtractedRow(pharmacy.ExtractedMedication{
			CNK:  r.CNK,
			Name: r.Name,
		}, true)
		r.Status = st
		r.ErrorCode = code
		r.ErrorMessage = msg
	}

	_, err = tx.Exec(ctx, `
		UPDATE pharmacy.compendium_import_rows
		SET cnk = $3, name = $4, manufacturer = $5, active_substance = $6, strength = $7, atc_code = $8,
		    pharmaceutical_form = $9, pack_size = $10, is_antibiotic = $11, source_page = $12,
		    status = $13, error_code = NULLIF($14,''), error_message = NULLIF($15,'')
		WHERE id = $1 AND job_id = $2`,
		rowID, jobID, r.CNK, r.Name, r.Manufacturer, r.ActiveSubstance, r.Strength, r.ATCCode,
		r.PharmaceuticalForm, r.PackSize, r.IsAntibiotic, r.SourcePage,
		r.Status, r.ErrorCode, r.ErrorMessage,
	)
	if err != nil {
		return r, err
	}
	if err := refreshCompendiumJobCountsTx(ctx, tx, jobID); err != nil {
		return r, err
	}
	if err := tx.Commit(ctx); err != nil {
		return r, err
	}
	return r, nil
}

// GetCompendiumImportRow loads one staging row (job + row id).
func (s *Store) GetCompendiumImportRow(ctx context.Context, jobID, rowID string) (CompendiumImportRow, error) {
	var r CompendiumImportRow
	var raw, cands []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, job_id::text, row_number, source_page,
		       cnk, name, COALESCE(manufacturer,''), COALESCE(active_substance,''), COALESCE(strength,''),
		       atc_code, pharmaceutical_form, pack_size, is_antibiotic,
		       COALESCE(suggested_cnk,''), match_score, COALESCE(match_candidates, '[]'::jsonb),
		       raw_json, status, COALESCE(error_code,''), COALESCE(error_message,'')
		FROM pharmacy.compendium_import_rows
		WHERE id = $1 AND job_id = $2`, rowID, jobID).Scan(
		&r.ID, &r.JobID, &r.RowNumber, &r.SourcePage,
		&r.CNK, &r.Name, &r.Manufacturer, &r.ActiveSubstance, &r.Strength,
		&r.ATCCode, &r.PharmaceuticalForm, &r.PackSize, &r.IsAntibiotic,
		&r.SuggestedCNK, &r.MatchScore, &cands,
		&raw, &r.Status, &r.ErrorCode, &r.ErrorMessage,
	)
	if err == pgx.ErrNoRows {
		return r, ErrNotFound
	}
	if err != nil {
		return r, err
	}
	r.RawJSON = raw
	r.MatchCandidates = cands
	return r, nil
}

// LookupCompendiumCNKResult is the outcome of a manual AFMPS-catalogue CNK lookup.
type LookupCompendiumCNKResult struct {
	Row      CompendiumImportRow
	CNKFound *bool // set only when verifyCNK was non-empty
}

// LookupCompendiumImportRowCNK locks the row, rematches against pharmacy.ref_medications,
// and updates suggestions without promoting to ready.
// Refuses ready/upserted/excluded (ErrConflict). Catalogue search uses locked row fields.
func (s *Store) LookupCompendiumImportRowCNK(ctx context.Context, jobID, rowID, verifyCNK string) (LookupCompendiumCNKResult, error) {
	verifyCNK = strings.TrimSpace(verifyCNK)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return LookupCompendiumCNKResult{}, err
	}
	defer tx.Rollback(ctx)

	var r CompendiumImportRow
	var raw, cands []byte
	err = tx.QueryRow(ctx, `
		SELECT id::text, job_id::text, row_number, source_page,
		       cnk, name, COALESCE(manufacturer,''), COALESCE(active_substance,''), COALESCE(strength,''),
		       atc_code, pharmaceutical_form, pack_size, is_antibiotic,
		       COALESCE(suggested_cnk,''), match_score, COALESCE(match_candidates, '[]'::jsonb),
		       raw_json, status, COALESCE(error_code,''), COALESCE(error_message,'')
		FROM pharmacy.compendium_import_rows
		WHERE id = $1 AND job_id = $2
		FOR UPDATE`, rowID, jobID).Scan(
		&r.ID, &r.JobID, &r.RowNumber, &r.SourcePage,
		&r.CNK, &r.Name, &r.Manufacturer, &r.ActiveSubstance, &r.Strength,
		&r.ATCCode, &r.PharmaceuticalForm, &r.PackSize, &r.IsAntibiotic,
		&r.SuggestedCNK, &r.MatchScore, &cands,
		&raw, &r.Status, &r.ErrorCode, &r.ErrorMessage,
	)
	if err == pgx.ErrNoRows {
		return LookupCompendiumCNKResult{}, ErrNotFound
	}
	if err != nil {
		return LookupCompendiumCNKResult{}, err
	}
	r.RawJSON = raw
	r.MatchCandidates = cands
	if r.Status == "ready" || r.Status == "upserted" || r.Status == "excluded" {
		return LookupCompendiumCNKResult{}, ErrConflict
	}

	var cnkFound *bool
	match := pharmacy.CNKMatchResult{}
	preserveExisting := false

	if verifyCNK != "" {
		ref, gerr := s.GetRefMedicationByCNK(ctx, verifyCNK)
		if gerr != nil && !errors.Is(gerr, ErrNotFound) {
			return LookupCompendiumCNKResult{}, gerr
		}
		found := gerr == nil
		cnkFound = &found
		if found {
			match = pharmacy.CNKMatchResult{
				SuggestedCNK: ref.CNK,
				Score:        1,
				Candidates: []pharmacy.CNKMatchCandidate{{
					CNK:                ref.CNK,
					Name:               ref.Name,
					PharmaceuticalForm: ref.PharmaceuticalForm,
					PackSize:           ref.PackSize,
					Score:              1,
				}},
				AutoFill: strings.TrimSpace(r.CNK) == "",
			}
		}
	}

	if verifyCNK == "" || (cnkFound != nil && !*cnkFound) {
		if strings.TrimSpace(r.Name) == "" {
			if verifyCNK != "" && cnkFound != nil && !*cnkFound {
				preserveExisting = true
			}
		} else {
			hits, searchErr := s.SearchRefMedications(ctx, r.Name, 10)
			if searchErr != nil {
				return LookupCompendiumCNKResult{}, searchErr
			}
			refs := make([]pharmacy.RefMedMatchInput, 0, len(hits))
			for _, h := range hits {
				refs = append(refs, pharmacy.RefMedMatchInput{
					CNK:                h.CNK,
					Name:               h.Name,
					PharmaceuticalForm: h.PharmaceuticalForm,
					PackSize:           h.PackSize,
				})
			}
			nameMatch := pharmacy.SuggestCNK(pharmacy.ExtractedMedication{
				CNK:                r.CNK,
				Name:               r.Name,
				Manufacturer:       r.Manufacturer,
				PharmaceuticalForm: r.PharmaceuticalForm,
				PackSize:           r.PackSize,
			}, refs)
			if verifyCNK != "" && cnkFound != nil && !*cnkFound {
				nameMatch.AutoFill = false
				if len(nameMatch.Candidates) == 0 {
					preserveExisting = true
				}
			}
			if !preserveExisting {
				match = nameMatch
			}
		}
	}

	if !preserveExisting {
		m := pharmacy.ExtractedMedication{
			CNK:                r.CNK,
			Name:               r.Name,
			Manufacturer:       r.Manufacturer,
			ActiveSubstance:    r.ActiveSubstance,
			ATCCode:            r.ATCCode,
			PharmaceuticalForm: r.PharmaceuticalForm,
			PackSize:           r.PackSize,
			IsAntibiotic:       r.IsAntibiotic,
			SourcePage:         r.SourcePage,
		}
		classified, status, code, msg := pharmacy.ClassifyAfterMatch(m, match)
		r.CNK = classified.CNK
		r.Status = status
		r.ErrorCode = code
		r.ErrorMessage = msg
		r.SuggestedCNK = match.SuggestedCNK
		candsRaw, _ := json.Marshal(match.Candidates)
		if len(match.Candidates) == 0 {
			candsRaw = []byte("[]")
		}
		r.MatchCandidates = candsRaw
		var scorePtr *float64
		if match.Score > 0 || match.SuggestedCNK != "" {
			sc := match.Score
			scorePtr = &sc
		}
		r.MatchScore = scorePtr

		_, err = tx.Exec(ctx, `
			UPDATE pharmacy.compendium_import_rows
			SET cnk = $3, suggested_cnk = NULLIF($4,''), match_score = $5,
			    match_candidates = $6::jsonb,
			    status = $7, error_code = NULLIF($8,''), error_message = NULLIF($9,'')
			WHERE id = $1 AND job_id = $2`,
			rowID, jobID, r.CNK, r.SuggestedCNK, scorePtr, string(candsRaw),
			r.Status, r.ErrorCode, r.ErrorMessage,
		)
		if err != nil {
			return LookupCompendiumCNKResult{}, err
		}
		if err := refreshCompendiumJobCountsTx(ctx, tx, jobID); err != nil {
			return LookupCompendiumCNKResult{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return LookupCompendiumCNKResult{}, err
	}
	return LookupCompendiumCNKResult{Row: r, CNKFound: cnkFound}, nil
}

func refreshCompendiumJobCountsTx(ctx context.Context, tx pgx.Tx, jobID string) error {
	_, err := tx.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs j
		SET row_count = COALESCE(s.total, 0),
		    ready_count = COALESCE(s.ready, 0),
		    error_count = COALESCE(s.err, 0),
		    reviewed_count = COALESCE(s.reviewed, 0),
		    upserted_count = COALESCE(s.upserted, 0),
		    updated_at = now()
		FROM (
			SELECT
				COUNT(*)::int AS total,
				COUNT(*) FILTER (WHERE status = 'ready')::int AS ready,
				COUNT(*) FILTER (WHERE status = 'error')::int AS err,
				COUNT(*) FILTER (WHERE status IN ('ready','excluded','upserted'))::int AS reviewed,
				COUNT(*) FILTER (WHERE status = 'upserted')::int AS upserted
			FROM pharmacy.compendium_import_rows
			WHERE job_id = $1
		) s
		WHERE j.id = $1`, jobID)
	return err
}

// ConfirmCompendiumPendingRows promotes pending/error rows with valid CNK+name to ready.
func (s *Store) ConfirmCompendiumPendingRows(ctx context.Context, jobID string) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `
		SELECT status FROM pharmacy.compendium_import_jobs WHERE id = $1 FOR UPDATE`, jobID).Scan(&status)
	if err == pgx.ErrNoRows {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	if status != "extracted" {
		return 0, ErrConflict
	}

	tag, err := tx.Exec(ctx, `
		UPDATE pharmacy.compendium_import_rows
		SET status = 'ready', error_code = NULL, error_message = NULL
		WHERE job_id = $1
		  AND status IN ('pending', 'error')
		  AND TRIM(cnk) <> ''
		  AND TRIM(name) <> ''`, jobID)
	if err != nil {
		return 0, err
	}
	if err := refreshCompendiumJobCountsTx(ctx, tx, jobID); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

type CompendiumCommitResult struct {
	Upserted int `json:"upserted"`
	Skipped  int `json:"skipped"`
}

// CommitCompendiumImport upserts ready rows into pharmacy.ref_medications.
func (s *Store) CommitCompendiumImport(ctx context.Context, jobID string) (CompendiumCommitResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CompendiumCommitResult{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `
		SELECT status FROM pharmacy.compendium_import_jobs WHERE id = $1 FOR UPDATE`, jobID).Scan(&status)
	if err == pgx.ErrNoRows {
		return CompendiumCommitResult{}, ErrNotFound
	}
	if err != nil {
		return CompendiumCommitResult{}, err
	}
	if status != "extracted" {
		return CompendiumCommitResult{}, ErrConflict
	}
	if _, err := tx.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs SET status = 'committing', updated_at = now() WHERE id = $1`, jobID); err != nil {
		return CompendiumCommitResult{}, err
	}

	rows, err := tx.Query(ctx, `
		SELECT id::text, cnk, name, COALESCE(manufacturer,''), COALESCE(active_substance,''),
		       COALESCE(strength,''), atc_code, pharmaceutical_form, pack_size, is_antibiotic
		FROM pharmacy.compendium_import_rows
		WHERE job_id = $1 AND status = 'ready'
		ORDER BY row_number`, jobID)
	if err != nil {
		return CompendiumCommitResult{}, err
	}
	type readyRow struct {
		id, cnk, name, manufacturer, substance, strength, atc, form, pack string
		ab                                                                bool
	}
	var ready []readyRow
	for rows.Next() {
		var rr readyRow
		if err := rows.Scan(&rr.id, &rr.cnk, &rr.name, &rr.manufacturer, &rr.substance,
			&rr.strength, &rr.atc, &rr.form, &rr.pack, &rr.ab); err != nil {
			rows.Close()
			return CompendiumCommitResult{}, err
		}
		ready = append(ready, rr)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return CompendiumCommitResult{}, err
	}

	upserted := 0
	for _, rr := range ready {
		metaMap := map[string]any{
			"source": "compendium-pdf",
			"jobId":  jobID,
			"rowId":  rr.id,
		}
		if strings.TrimSpace(rr.manufacturer) != "" {
			metaMap["manufacturer"] = strings.TrimSpace(rr.manufacturer)
		}
		if strings.TrimSpace(rr.substance) != "" {
			metaMap["activeSubstance"] = strings.TrimSpace(rr.substance)
		}
		if strings.TrimSpace(rr.strength) != "" {
			metaMap["strength"] = strings.TrimSpace(rr.strength)
		}
		meta, _ := json.Marshal(metaMap)
		norm := NormalizeMedicationName(rr.name)
		newID := uuid.NewString()
		var id string
		err := tx.QueryRow(ctx, `
			INSERT INTO pharmacy.ref_medications (
				id, cnk, name, name_normalized, atc_code, pharmaceutical_form, pack_size,
				is_antibiotic, is_active, afmps_meta, updated_at
			) VALUES (
				$1, $2, $3, $4, NULLIF($5,''), NULLIF($6,''), NULLIF($7,''),
				$8, true, $9::jsonb, now()
			)
			ON CONFLICT (cnk) DO UPDATE SET
				name = EXCLUDED.name,
				name_normalized = EXCLUDED.name_normalized,
				atc_code = EXCLUDED.atc_code,
				pharmaceutical_form = EXCLUDED.pharmaceutical_form,
				pack_size = EXCLUDED.pack_size,
				is_antibiotic = EXCLUDED.is_antibiotic,
				is_active = true,
				afmps_meta = pharmacy.ref_medications.afmps_meta || EXCLUDED.afmps_meta,
				updated_at = now()
			RETURNING id::text`,
			newID, rr.cnk, rr.name, norm, rr.atc, rr.form, rr.pack, rr.ab, string(meta),
		).Scan(&id)
		if err != nil {
			return CompendiumCommitResult{}, fmt.Errorf("upsert cnk=%s: %w", rr.cnk, err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE pharmacy.compendium_import_rows SET status = 'upserted' WHERE id = $1`, rr.id); err != nil {
			return CompendiumCommitResult{}, err
		}
		upserted++
	}

	if err := refreshCompendiumJobCountsTx(ctx, tx, jobID); err != nil {
		return CompendiumCommitResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs
		SET status = 'completed', updated_at = now() WHERE id = $1`, jobID); err != nil {
		return CompendiumCommitResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CompendiumCommitResult{}, err
	}
	skipped := 0
	detail, _ := s.GetCompendiumImportJob(ctx, jobID)
	if detail.RowCount > upserted {
		skipped = detail.RowCount - upserted
	}
	return CompendiumCommitResult{Upserted: upserted, Skipped: skipped}, nil
}
