package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const compendiumCommitBatchSize = 50

// Stale committing jobs (crash / killed process) can be reclaimed after this age.
const compendiumCommittingStale = 5 * time.Minute

// testFailNextCompendiumBatch forces the next commitCompendiumBatch to error (integration tests).
var testFailNextCompendiumBatch atomic.Bool

// TestFailNextCompendiumCommitBatch arms a one-shot failure for the next commit batch.
func (s *Store) TestFailNextCompendiumCommitBatch() {
	testFailNextCompendiumBatch.Store(true)
}

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
	ATCCode            string          `json:"atcCode"`
	PharmaceuticalForm string          `json:"pharmaceuticalForm"`
	PackSize           string          `json:"packSize"`
	IsAntibiotic       bool            `json:"isAntibiotic"`
	RawJSON            json.RawMessage `json:"rawJson,omitempty"`
	Status             string          `json:"status"`
	ErrorCode          string          `json:"errorCode,omitempty"`
	ErrorMessage       string          `json:"errorMessage,omitempty"`
	HumanReviewed      bool            `json:"humanReviewed"`
}

type CompendiumImportDetail struct {
	Job  CompendiumImportJob   `json:"job"`
	Rows []CompendiumImportRow `json:"rows"`
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
	return CompendiumImportDetail{Job: job, Rows: rows}, nil
}

func (s *Store) ListCompendiumImportRows(ctx context.Context, jobID string) ([]CompendiumImportRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, job_id::text, row_number, source_page,
		       cnk, name, atc_code, pharmaceutical_form, pack_size, is_antibiotic,
		       raw_json, status, COALESCE(error_code,''), COALESCE(error_message,''),
		       human_reviewed
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
		var raw []byte
		if err := rows.Scan(
			&r.ID, &r.JobID, &r.RowNumber, &r.SourcePage,
			&r.CNK, &r.Name, &r.ATCCode, &r.PharmaceuticalForm, &r.PackSize, &r.IsAntibiotic,
			&raw, &r.Status, &r.ErrorCode, &r.ErrorMessage, &r.HumanReviewed,
		); err != nil {
			return nil, err
		}
		r.RawJSON = raw
		out = append(out, r)
	}
	return out, rows.Err()
}

// MarkCompendiumExtracting sets status + extract_total and clears any previous staging rows.
func (s *Store) MarkCompendiumExtracting(ctx context.Context, id string, extractTotal int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs
		SET status = 'extracting', extract_total = $2, extract_done = 0,
		    row_count = 0, ready_count = 0, error_count = 0,
		    reviewed_count = 0, upserted_count = 0,
		    error_message = NULL, updated_at = now()
		WHERE id = $1 AND status IN ('uploaded', 'failed')`, id, extractTotal)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrConflict
	}
	if _, err := tx.Exec(ctx, `DELETE FROM pharmacy.compendium_import_rows WHERE job_id = $1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
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
	ATCCode            string
	PharmaceuticalForm string
	PackSize           string
	IsAntibiotic       bool
	RawJSON            json.RawMessage
	Status             string
	ErrorCode          string
	ErrorMessage       string
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
	ready, errs := 0, 0
	for i, row := range rows {
		raw := row.RawJSON
		if len(raw) == 0 {
			raw = json.RawMessage(`{}`)
		}
		status := row.Status
		if status == "" {
			status = "pending"
		}
		switch status {
		case "ready":
			ready++
		case "error":
			errs++
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO pharmacy.compendium_import_rows (
				id, job_id, row_number, source_page, cnk, name, atc_code,
				pharmaceutical_form, pack_size, is_antibiotic, raw_json,
				status, error_code, error_message, human_reviewed
			) VALUES (
				$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12,NULLIF($13,''),NULLIF($14,''),false
			)`,
			uuid.NewString(), jobID, i+1, row.SourcePage,
			strings.TrimSpace(row.CNK), strings.TrimSpace(row.Name),
			strings.TrimSpace(row.ATCCode), strings.TrimSpace(row.PharmaceuticalForm),
			strings.TrimSpace(row.PackSize), row.IsAntibiotic, string(raw),
			status, row.ErrorCode, row.ErrorMessage,
		); err != nil {
			return err
		}
	}
	// AI classification is not human review — reviewPct stays 0 until patch/exclude.
	_, err = tx.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs
		SET status = 'extracted',
		    extract_done = extract_total,
		    row_count = $2, ready_count = $3, error_count = $4,
		    reviewed_count = 0, error_message = NULL, updated_at = now()
		WHERE id = $1`, jobID, len(rows), ready, errs)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) FailCompendiumImportJob(ctx context.Context, id, msg string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs
		SET status = 'failed', error_message = $2, updated_at = now()
		WHERE id = $1`, id, msg)
	return err
}

type PatchCompendiumRowInput struct {
	CNK                *string `json:"cnk"`
	Name               *string `json:"name"`
	ATCCode            *string `json:"atcCode"`
	PharmaceuticalForm *string `json:"pharmaceuticalForm"`
	PackSize           *string `json:"packSize"`
	IsAntibiotic       *bool   `json:"isAntibiotic"`
	Excluded           *bool   `json:"excluded"`
}

func (s *Store) PatchCompendiumImportRow(ctx context.Context, jobID, rowID string, in PatchCompendiumRowInput) (CompendiumImportRow, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CompendiumImportRow{}, err
	}
	defer tx.Rollback(ctx)

	var jobStatus string
	err = tx.QueryRow(ctx, `
		SELECT status FROM pharmacy.compendium_import_jobs WHERE id = $1 FOR UPDATE`, jobID).Scan(&jobStatus)
	if err == pgx.ErrNoRows {
		return CompendiumImportRow{}, ErrNotFound
	}
	if err != nil {
		return CompendiumImportRow{}, err
	}
	if jobStatus != "extracted" {
		return CompendiumImportRow{}, ErrConflict
	}

	var r CompendiumImportRow
	var raw []byte
	err = tx.QueryRow(ctx, `
		SELECT id::text, job_id::text, row_number, source_page,
		       cnk, name, atc_code, pharmaceutical_form, pack_size, is_antibiotic,
		       raw_json, status, COALESCE(error_code,''), COALESCE(error_message,''),
		       human_reviewed
		FROM pharmacy.compendium_import_rows
		WHERE id = $1 AND job_id = $2
		FOR UPDATE`, rowID, jobID).Scan(
		&r.ID, &r.JobID, &r.RowNumber, &r.SourcePage,
		&r.CNK, &r.Name, &r.ATCCode, &r.PharmaceuticalForm, &r.PackSize, &r.IsAntibiotic,
		&raw, &r.Status, &r.ErrorCode, &r.ErrorMessage, &r.HumanReviewed,
	)
	if err == pgx.ErrNoRows {
		return r, ErrNotFound
	}
	if err != nil {
		return r, err
	}
	r.RawJSON = raw
	if r.Status == "upserted" {
		return r, ErrConflict
	}

	if in.CNK != nil {
		r.CNK = strings.TrimSpace(*in.CNK)
	}
	if in.Name != nil {
		r.Name = strings.TrimSpace(*in.Name)
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
		if r.Name == "" {
			r.Status = "error"
			r.ErrorCode = "missing_name"
			r.ErrorMessage = "name required"
		} else if r.CNK == "" {
			r.Status = "error"
			r.ErrorCode = "missing_cnk"
			r.ErrorMessage = "cnk required for national dictionary"
		} else {
			r.Status = "ready"
			r.ErrorCode = ""
			r.ErrorMessage = ""
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE pharmacy.compendium_import_rows
		SET cnk = $3, name = $4, atc_code = $5, pharmaceutical_form = $6, pack_size = $7,
		    is_antibiotic = $8, status = $9,
		    error_code = NULLIF($10,''), error_message = NULLIF($11,''),
		    human_reviewed = true
		WHERE id = $1 AND job_id = $2`,
		rowID, jobID, r.CNK, r.Name, r.ATCCode, r.PharmaceuticalForm, r.PackSize,
		r.IsAntibiotic, r.Status, r.ErrorCode, r.ErrorMessage,
	)
	if err != nil {
		return r, err
	}
	r.HumanReviewed = true
	if err := refreshCompendiumJobCountsTx(ctx, tx, jobID); err != nil {
		return r, err
	}
	if err := tx.Commit(ctx); err != nil {
		return r, err
	}
	return r, nil
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
				COUNT(*) FILTER (WHERE human_reviewed)::int AS reviewed,
				COUNT(*) FILTER (WHERE status = 'upserted')::int AS upserted
			FROM pharmacy.compendium_import_rows
			WHERE job_id = $1
		) s
		WHERE j.id = $1`, jobID)
	return err
}

type CompendiumCommitResult struct {
	Upserted int `json:"upserted"`
	Skipped  int `json:"skipped"`
}

type compendiumReadyRow struct {
	id, cnk, name, atc, form, pack string
	ab                             bool
}

// CommitCompendiumImport upserts ready rows into pharmacy.ref_medications in batches of 50.
// Each batch uses its own short TX after status=committing. Partial upserts are kept on mid-fail;
// the job is reverted to extracted so commit can be retried.
// Concurrent commits are rejected: only status=extracted (or stale committing) can be claimed.
func (s *Store) CommitCompendiumImport(ctx context.Context, jobID string) (CompendiumCommitResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CompendiumCommitResult{}, err
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM pharmacy.compendium_import_jobs WHERE id = $1)`, jobID).Scan(&exists)
	if err != nil {
		return CompendiumCommitResult{}, err
	}
	if !exists {
		return CompendiumCommitResult{}, ErrNotFound
	}

	staleBefore := time.Now().UTC().Add(-compendiumCommittingStale)
	tag, err := tx.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs
		SET status = 'committing', error_message = NULL, updated_at = now()
		WHERE id = $1 AND (
			status = 'extracted'
			OR (status = 'committing' AND updated_at < $2)
		)`, jobID, staleBefore)
	if err != nil {
		return CompendiumCommitResult{}, err
	}
	if tag.RowsAffected() == 0 {
		return CompendiumCommitResult{}, ErrConflict
	}

	rows, err := tx.Query(ctx, `
		SELECT id::text, cnk, name, atc_code, pharmaceutical_form, pack_size, is_antibiotic
		FROM pharmacy.compendium_import_rows
		WHERE job_id = $1 AND status = 'ready'
		ORDER BY row_number`, jobID)
	if err != nil {
		return CompendiumCommitResult{}, err
	}
	var ready []compendiumReadyRow
	for rows.Next() {
		var rr compendiumReadyRow
		if err := rows.Scan(&rr.id, &rr.cnk, &rr.name, &rr.atc, &rr.form, &rr.pack, &rr.ab); err != nil {
			rows.Close()
			return CompendiumCommitResult{}, err
		}
		ready = append(ready, rr)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return CompendiumCommitResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CompendiumCommitResult{}, err
	}

	upserted := 0
	seenCNK := make(map[string]struct{}, len(ready))
	for i := 0; i < len(ready); i += compendiumCommitBatchSize {
		end := i + compendiumCommitBatchSize
		if end > len(ready) {
			end = len(ready)
		}
		n, err := s.commitCompendiumBatch(ctx, jobID, ready[i:end], seenCNK)
		upserted += n
		if err != nil {
			_ = s.refreshCompendiumJobCounts(ctx, jobID)
			_ = s.revertCompendiumCommitToExtracted(ctx, jobID, fmt.Sprintf("commit_batch:%v", err))
			skipped := len(ready) - upserted
			if skipped < 0 {
				skipped = 0
			}
			return CompendiumCommitResult{Upserted: upserted, Skipped: skipped}, err
		}
	}

	tx2, err := s.pool.Begin(ctx)
	if err != nil {
		_ = s.revertCompendiumCommitToExtracted(ctx, jobID, fmt.Sprintf("finalize_begin:%v", err))
		return CompendiumCommitResult{Upserted: upserted}, err
	}
	defer tx2.Rollback(ctx)
	if err := refreshCompendiumJobCountsTx(ctx, tx2, jobID); err != nil {
		_ = s.revertCompendiumCommitToExtracted(ctx, jobID, fmt.Sprintf("finalize_counts:%v", err))
		return CompendiumCommitResult{Upserted: upserted}, err
	}
	if _, err := tx2.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs
		SET status = 'completed', error_message = NULL, updated_at = now() WHERE id = $1`, jobID); err != nil {
		_ = s.revertCompendiumCommitToExtracted(ctx, jobID, fmt.Sprintf("finalize_status:%v", err))
		return CompendiumCommitResult{Upserted: upserted}, err
	}
	if err := tx2.Commit(ctx); err != nil {
		_ = s.revertCompendiumCommitToExtracted(ctx, jobID, fmt.Sprintf("finalize_commit:%v", err))
		return CompendiumCommitResult{Upserted: upserted}, err
	}

	skipped := len(ready) - upserted
	if skipped < 0 {
		skipped = 0
	}
	return CompendiumCommitResult{Upserted: upserted, Skipped: skipped}, nil
}

// revertCompendiumCommitToExtracted unlocks a job stuck in committing so the admin can retry commit.
func (s *Store) revertCompendiumCommitToExtracted(ctx context.Context, jobID, msg string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE pharmacy.compendium_import_jobs
		SET status = 'extracted', error_message = $2, updated_at = now()
		WHERE id = $1 AND status = 'committing'`, jobID, msg)
	return err
}

func (s *Store) refreshCompendiumJobCounts(ctx context.Context, jobID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := refreshCompendiumJobCountsTx(ctx, tx, jobID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) commitCompendiumBatch(ctx context.Context, jobID string, batch []compendiumReadyRow, seenCNK map[string]struct{}) (int, error) {
	if testFailNextCompendiumBatch.CompareAndSwap(true, false) {
		return 0, fmt.Errorf("forced_test_fail")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	upserted := 0
	for _, rr := range batch {
		cnkKey := strings.ToLower(strings.TrimSpace(rr.cnk))
		if cnkKey == "" {
			continue
		}
		if _, ok := seenCNK[cnkKey]; ok {
			if _, err := tx.Exec(ctx, `
				UPDATE pharmacy.compendium_import_rows
				SET status = 'excluded', error_code = 'duplicate_cnk', error_message = 'duplicate cnk in batch'
				WHERE id = $1`, rr.id); err != nil {
				return 0, err
			}
			continue
		}
		seenCNK[cnkKey] = struct{}{}
		meta, _ := json.Marshal(map[string]string{
			"source": "compendium-pdf",
			"jobId":  jobID,
			"rowId":  rr.id,
		})
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
			return 0, fmt.Errorf("upsert cnk=%s: %w", rr.cnk, err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE pharmacy.compendium_import_rows SET status = 'upserted' WHERE id = $1`, rr.id); err != nil {
			return 0, err
		}
		upserted++
	}
	if err := refreshCompendiumJobCountsTx(ctx, tx, jobID); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return upserted, nil
}
