package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

// AFMPSImportJob tracks CSV validate → review → commit.
type AFMPSImportJob struct {
	ID                string          `json:"id"`
	CreatedByAdminID  string          `json:"createdByAdminId,omitempty"`
	Filename          string          `json:"filename"`
	ContentType       string          `json:"contentType"`
	FileBytes         int             `json:"fileBytes"`
	ChecksumSHA256    string          `json:"checksumSha256"`
	Status            string          `json:"status"`
	Report            json.RawMessage `json:"report,omitempty"`
	RowCount          int             `json:"rowCount"`
	ReadyCount        int             `json:"readyCount"`
	ErrorCount        int             `json:"errorCount"`
	InsertCount       int             `json:"insertCount"`
	UpdateCount       int             `json:"updateCount"`
	UnchangedCount    int             `json:"unchangedCount"`
	DeactivatePreview int             `json:"deactivatePreview"`
	UpsertedCount     int             `json:"upsertedCount"`
	DeactivatedCount  int             `json:"deactivatedCount"`
	ErrorMessage      string          `json:"errorMessage,omitempty"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

// AFMPSImportRow is one staged CNK line.
type AFMPSImportRow struct {
	ID                 string          `json:"id"`
	JobID              string          `json:"jobId"`
	RowNumber          int             `json:"rowNumber"`
	SourceLine         int             `json:"sourceLine"`
	CNK                string          `json:"cnk"`
	Name               string          `json:"name"`
	ATCCode            string          `json:"atcCode"`
	PharmaceuticalForm string          `json:"pharmaceuticalForm"`
	PackSize           string          `json:"packSize"`
	AMMNumber          string          `json:"ammNumber"`
	IsAntibiotic       bool            `json:"isAntibiotic"`
	Collision          string          `json:"collision"`
	AFMPSMeta          json.RawMessage `json:"afmpsMeta,omitempty"`
	Status             string          `json:"status"`
	ErrorCode          string          `json:"errorCode,omitempty"`
	ErrorMessage       string          `json:"errorMessage,omitempty"`
}

type AFMPSImportDetail struct {
	Job        AFMPSImportJob   `json:"job"`
	Rows       []AFMPSImportRow `json:"rows"`
	RowsTotal  int              `json:"rowsTotal"`
	RowsLimit  int              `json:"rowsLimit"`
	RowsOffset int              `json:"rowsOffset"`
}

type AFMPSCommitResult struct {
	Upserted    int   `json:"upserted"`
	Deactivated int64 `json:"deactivated"`
}

type existingRefMed struct {
	name, atc, form, pack, amm string
	ab                         bool
}

const afmpsJobCols = `
	id::text, COALESCE(created_by_admin_id::text, ''), filename, content_type,
	file_bytes, checksum_sha256, status, report_json,
	row_count, ready_count, error_count,
	insert_count, update_count, unchanged_count, deactivate_preview,
	upserted_count, deactivated_count,
	error_message, created_at, updated_at`

func scanAFMPSJob(row pgx.Row) (AFMPSImportJob, error) {
	var j AFMPSImportJob
	var errMsg *string
	var report []byte
	err := row.Scan(
		&j.ID, &j.CreatedByAdminID, &j.Filename, &j.ContentType,
		&j.FileBytes, &j.ChecksumSHA256, &j.Status, &report,
		&j.RowCount, &j.ReadyCount, &j.ErrorCount,
		&j.InsertCount, &j.UpdateCount, &j.UnchangedCount, &j.DeactivatePreview,
		&j.UpsertedCount, &j.DeactivatedCount,
		&errMsg, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return j, err
	}
	if len(report) > 0 {
		j.Report = json.RawMessage(report)
	}
	if errMsg != nil {
		j.ErrorMessage = *errMsg
	}
	return j, nil
}

// CreateAFMPSImportFromParsed persists gate-1 result (validated or blocked).
func (s *Store) CreateAFMPSImportFromParsed(
	ctx context.Context,
	adminID string,
	filename string,
	rows []pharmacy.AFMPSParsedRow,
	report pharmacy.AFMPSValidateReport,
) (AFMPSImportJob, error) {
	id := uuid.NewString()
	status := "validated"
	if report.Blocked {
		status = "blocked"
	}
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return AFMPSImportJob{}, err
	}

	var adminPtr *string
	if a := strings.TrimSpace(adminID); a != "" {
		adminPtr = &a
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return AFMPSImportJob{}, err
	}
	defer tx.Rollback(ctx)

	existing, err := loadExistingRefMeds(ctx, tx, readyCNKs(rows))
	if err != nil {
		return AFMPSImportJob{}, err
	}
	insertCount, updateCount, unchangedCount := classifyCollisionsFromMap(rows, existing)
	deactivatePreview, err := countDeactivatePreview(ctx, tx, readyCNKs(rows))
	if err != nil {
		return AFMPSImportJob{}, err
	}

	ready, errs := 0, 0
	for _, r := range rows {
		if r.Status == "ready" {
			ready++
		} else {
			errs++
		}
	}

	j, err := scanAFMPSJob(tx.QueryRow(ctx, `
		INSERT INTO pharmacy.afmps_import_jobs (
			id, created_by_admin_id, filename, content_type, file_bytes, checksum_sha256,
			status, report_json, row_count, ready_count, error_count,
			insert_count, update_count, unchanged_count, deactivate_preview, updated_at
		) VALUES (
			$1,$2::uuid,$3,'text/csv',$4,$5,$6,$7::jsonb,$8,$9,$10,$11,$12,$13,$14,now()
		)
		RETURNING `+afmpsJobCols,
		id, adminPtr, filename, report.FileBytes, report.ChecksumSHA256,
		status, string(reportJSON), len(rows), ready, errs,
		insertCount, updateCount, unchangedCount, deactivatePreview,
	))
	if err != nil {
		return AFMPSImportJob{}, err
	}

	if err := copyAFMPSImportRows(ctx, tx, id, rows, existing); err != nil {
		return AFMPSImportJob{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AFMPSImportJob{}, err
	}
	return j, nil
}

func copyAFMPSImportRows(
	ctx context.Context,
	tx pgx.Tx,
	jobID string,
	rows []pharmacy.AFMPSParsedRow,
	existing map[string]existingRefMed,
) error {
	if len(rows) == 0 {
		return nil
	}
	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"pharmacy", "afmps_import_rows"},
		[]string{
			"job_id", "row_number", "source_line", "cnk", "name", "atc_code",
			"pharmaceutical_form", "pack_size", "amm_number", "is_antibiotic",
			"collision", "afmps_meta", "status", "error_code", "error_message",
		},
		pgx.CopyFromSlice(len(rows), func(i int) ([]any, error) {
			r := rows[i]
			meta := pharmacy.AFMPSMetaJSON(r.Meta)
			collision := "error"
			st := r.Status
			if st == "" {
				st = "ready"
			}
			if st == "ready" {
				collision = collisionFromExisting(r, existing)
			}
			var errCode, errMsg any
			if r.ErrorCode != "" {
				errCode = r.ErrorCode
			}
			if r.ErrorMessage != "" {
				errMsg = r.ErrorMessage
			}
			return []any{
				jobID, i + 1, r.SourceLine, r.CNK, r.Name, r.ATCCode,
				r.PharmaceuticalForm, r.PackSize, r.AMMNumber, r.IsAntibiotic,
				collision, string(meta), st, errCode, errMsg,
			}, nil
		}),
	)
	return err
}

func readyCNKs(rows []pharmacy.AFMPSParsedRow) []string {
	keep := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.Status == "ready" && r.CNK != "" {
			keep = append(keep, r.CNK)
		}
	}
	return keep
}

func loadExistingRefMeds(ctx context.Context, q interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, cnks []string) (map[string]existingRefMed, error) {
	out := make(map[string]existingRefMed, len(cnks))
	if len(cnks) == 0 {
		return out, nil
	}
	rows, err := q.Query(ctx, `
		SELECT cnk, name, COALESCE(atc_code,''), COALESCE(pharmaceutical_form,''),
		       COALESCE(pack_size,''), COALESCE(amm_number,''), is_antibiotic
		FROM pharmacy.ref_medications WHERE cnk = ANY($1)`, cnks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cnk string
		var e existingRefMed
		if err := rows.Scan(&cnk, &e.name, &e.atc, &e.form, &e.pack, &e.amm, &e.ab); err != nil {
			return nil, err
		}
		out[cnk] = e
	}
	return out, rows.Err()
}

func classifyCollisionsFromMap(rows []pharmacy.AFMPSParsedRow, existing map[string]existingRefMed) (insert, update, unchanged int) {
	for _, r := range rows {
		if r.Status != "ready" {
			continue
		}
		switch collisionFromExisting(r, existing) {
		case "insert":
			insert++
		case "update":
			update++
		case "unchanged":
			unchanged++
		}
	}
	return insert, update, unchanged
}

func collisionFromExisting(r pharmacy.AFMPSParsedRow, existing map[string]existingRefMed) string {
	e, ok := existing[r.CNK]
	if !ok {
		return "insert"
	}
	if e.name == r.Name && e.atc == r.ATCCode && e.form == r.PharmaceuticalForm &&
		e.pack == r.PackSize && e.amm == r.AMMNumber && e.ab == r.IsAntibiotic {
		return "unchanged"
	}
	return "update"
}

func countDeactivatePreview(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, keep []string) (int, error) {
	if len(keep) == 0 {
		return 0, nil
	}
	var n int
	err := q.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM pharmacy.ref_medications
		WHERE is_active AND NOT (cnk = ANY($1))`, keep).Scan(&n)
	return n, err
}

func (s *Store) GetAFMPSImportJob(ctx context.Context, id string) (AFMPSImportJob, error) {
	j, err := scanAFMPSJob(s.pool.QueryRow(ctx, `
		SELECT `+afmpsJobCols+` FROM pharmacy.afmps_import_jobs WHERE id = $1`, id))
	if err == pgx.ErrNoRows {
		return j, ErrNotFound
	}
	return j, err
}

func (s *Store) ListAFMPSImportJobs(ctx context.Context, limit int) ([]AFMPSImportJob, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT `+afmpsJobCols+`
		FROM pharmacy.afmps_import_jobs
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AFMPSImportJob, 0)
	for rows.Next() {
		j, err := scanAFMPSJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func afmpsCollisionFilterOK(collision string) bool {
	switch collision {
	case "", "all", "insert", "update", "unchanged", "error":
		return true
	default:
		return false
	}
}

func (s *Store) CountAFMPSImportRows(ctx context.Context, jobID, collision string) (int, error) {
	var n int
	var err error
	if collision == "" || collision == "all" {
		err = s.pool.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM pharmacy.afmps_import_rows WHERE job_id = $1`, jobID).Scan(&n)
	} else {
		err = s.pool.QueryRow(ctx, `
			SELECT COUNT(*)::int FROM pharmacy.afmps_import_rows
			WHERE job_id = $1 AND collision = $2`, jobID, collision).Scan(&n)
	}
	return n, err
}

func (s *Store) ListAFMPSImportRows(ctx context.Context, jobID string, limit, offset int, collision string) ([]AFMPSImportRow, error) {
	if limit <= 0 {
		limit = 200
	}
	if limit > 2000 {
		limit = 2000
	}
	if offset < 0 {
		offset = 0
	}
	var rows pgx.Rows
	var err error
	if collision == "" || collision == "all" {
		rows, err = s.pool.Query(ctx, `
			SELECT id::text, job_id::text, row_number, source_line,
			       cnk, name, atc_code, pharmaceutical_form, pack_size, amm_number,
			       is_antibiotic, collision, afmps_meta, status,
			       COALESCE(error_code,''), COALESCE(error_message,'')
			FROM pharmacy.afmps_import_rows
			WHERE job_id = $1
			ORDER BY
			  CASE status WHEN 'error' THEN 0 WHEN 'ready' THEN 1 ELSE 2 END,
			  row_number
			LIMIT $2 OFFSET $3`, jobID, limit, offset)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT id::text, job_id::text, row_number, source_line,
			       cnk, name, atc_code, pharmaceutical_form, pack_size, amm_number,
			       is_antibiotic, collision, afmps_meta, status,
			       COALESCE(error_code,''), COALESCE(error_message,'')
			FROM pharmacy.afmps_import_rows
			WHERE job_id = $1 AND collision = $4
			ORDER BY
			  CASE status WHEN 'error' THEN 0 WHEN 'ready' THEN 1 ELSE 2 END,
			  row_number
			LIMIT $2 OFFSET $3`, jobID, limit, offset, collision)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AFMPSImportRow, 0)
	for rows.Next() {
		var r AFMPSImportRow
		var meta []byte
		if err := rows.Scan(
			&r.ID, &r.JobID, &r.RowNumber, &r.SourceLine,
			&r.CNK, &r.Name, &r.ATCCode, &r.PharmaceuticalForm, &r.PackSize, &r.AMMNumber,
			&r.IsAntibiotic, &r.Collision, &meta, &r.Status,
			&r.ErrorCode, &r.ErrorMessage,
		); err != nil {
			return nil, err
		}
		if len(meta) > 0 {
			r.AFMPSMeta = json.RawMessage(meta)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) GetAFMPSImportDetail(ctx context.Context, id string, limit, offset int, collision string) (AFMPSImportDetail, error) {
	job, err := s.GetAFMPSImportJob(ctx, id)
	if err != nil {
		return AFMPSImportDetail{}, err
	}
	if limit <= 0 {
		limit = 200
	}
	if !afmpsCollisionFilterOK(collision) {
		collision = "all"
	}
	total, err := s.CountAFMPSImportRows(ctx, id, collision)
	if err != nil {
		return AFMPSImportDetail{}, err
	}
	rows, err := s.ListAFMPSImportRows(ctx, id, limit, offset, collision)
	if err != nil {
		return AFMPSImportDetail{}, err
	}
	return AFMPSImportDetail{
		Job: job, Rows: rows,
		RowsTotal: total, RowsLimit: limit, RowsOffset: offset,
	}, nil
}

// MarkAFMPSImportReviewed is gate 2 (human OK).
func (s *Store) MarkAFMPSImportReviewed(ctx context.Context, id string) (AFMPSImportJob, error) {
	j, err := s.GetAFMPSImportJob(ctx, id)
	if err != nil {
		return j, err
	}
	if j.Status != "validated" {
		return j, ErrConflict
	}
	if j.ReadyCount == 0 {
		return j, ErrConflict
	}
	return scanAFMPSJob(s.pool.QueryRow(ctx, `
		UPDATE pharmacy.afmps_import_jobs
		SET status = 'reviewed', updated_at = now()
		WHERE id = $1 AND status = 'validated'
		RETURNING `+afmpsJobCols, id))
}

// PatchAFMPSImportRow excludes or re-includes a staged row before commit.
func (s *Store) PatchAFMPSImportRow(ctx context.Context, jobID, rowID string, exclude *bool, status string) (AFMPSImportRow, error) {
	job, err := s.GetAFMPSImportJob(ctx, jobID)
	if err != nil {
		return AFMPSImportRow{}, err
	}
	if job.Status != "validated" && job.Status != "reviewed" {
		return AFMPSImportRow{}, ErrConflict
	}
	var row AFMPSImportRow
	err = s.pool.QueryRow(ctx, `
		SELECT id::text, job_id::text, row_number, source_line,
		       cnk, name, atc_code, pharmaceutical_form, pack_size, amm_number,
		       is_antibiotic, collision, afmps_meta, status,
		       COALESCE(error_code,''), COALESCE(error_message,'')
		FROM pharmacy.afmps_import_rows WHERE id = $1 AND job_id = $2`, rowID, jobID).Scan(
		&row.ID, &row.JobID, &row.RowNumber, &row.SourceLine,
		&row.CNK, &row.Name, &row.ATCCode, &row.PharmaceuticalForm, &row.PackSize, &row.AMMNumber,
		&row.IsAntibiotic, &row.Collision, &row.AFMPSMeta, &row.Status,
		&row.ErrorCode, &row.ErrorMessage,
	)
	if err == pgx.ErrNoRows {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if row.Status == "upserted" || row.Status == "error" {
		return row, ErrConflict
	}
	newStatus := row.Status
	if exclude != nil {
		if *exclude {
			newStatus = "excluded"
		} else if row.Status == "excluded" && row.ErrorCode == "" {
			newStatus = "ready"
		}
	}
	if status == "ready" || status == "excluded" {
		newStatus = status
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE pharmacy.afmps_import_rows SET status = $3
		WHERE id = $1 AND job_id = $2`, rowID, jobID, newStatus)
	if err != nil {
		return row, err
	}
	row.Status = newStatus
	if err := s.refreshAFMPSJobCounts(ctx, jobID); err != nil {
		return row, err
	}
	return row, nil
}

func (s *Store) refreshAFMPSJobCounts(ctx context.Context, jobID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := refreshAFMPSJobCountsTx(ctx, tx, jobID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func refreshAFMPSJobCountsTx(ctx context.Context, tx pgx.Tx, jobID string) error {
	type staged struct {
		id, cnk, name, atc, form, pack, amm string
		ab                                 bool
	}
	rows, err := tx.Query(ctx, `
		SELECT id::text, cnk, name, atc_code, pharmaceutical_form, pack_size, amm_number, is_antibiotic
		FROM pharmacy.afmps_import_rows
		WHERE job_id = $1 AND status = 'ready'`, jobID)
	if err != nil {
		return err
	}
	stagedRows := make([]staged, 0)
	readyCNKs := make([]string, 0)
	for rows.Next() {
		var r staged
		if err := rows.Scan(&r.id, &r.cnk, &r.name, &r.atc, &r.form, &r.pack, &r.amm, &r.ab); err != nil {
			rows.Close()
			return err
		}
		stagedRows = append(stagedRows, r)
		readyCNKs = append(readyCNKs, r.cnk)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	existing, err := loadExistingRefMeds(ctx, tx, readyCNKs)
	if err != nil {
		return err
	}

	insert, update, unchanged := 0, 0, 0
	ids := make([]string, 0, len(stagedRows))
	collisions := make([]string, 0, len(stagedRows))
	for _, r := range stagedRows {
		parsed := pharmacy.AFMPSParsedRow{
			CNK: r.cnk, Name: r.name, ATCCode: r.atc, PharmaceuticalForm: r.form,
			PackSize: r.pack, AMMNumber: r.amm, IsAntibiotic: r.ab, Status: "ready",
		}
		c := collisionFromExisting(parsed, existing)
		ids = append(ids, r.id)
		collisions = append(collisions, c)
		switch c {
		case "insert":
			insert++
		case "update":
			update++
		case "unchanged":
			unchanged++
		}
	}
	if len(ids) > 0 {
		if _, err := tx.Exec(ctx, `
			UPDATE pharmacy.afmps_import_rows AS r
			SET collision = v.collision
			FROM (
				SELECT UNNEST($1::uuid[]) AS id, UNNEST($2::text[]) AS collision
			) AS v
			WHERE r.id = v.id`, ids, collisions); err != nil {
			return err
		}
	}

	deactivatePreview, err := countDeactivatePreview(ctx, tx, readyCNKs)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE pharmacy.afmps_import_jobs j SET
			ready_count = (SELECT COUNT(*)::int FROM pharmacy.afmps_import_rows r WHERE r.job_id = j.id AND r.status = 'ready'),
			error_count = (SELECT COUNT(*)::int FROM pharmacy.afmps_import_rows r WHERE r.job_id = j.id AND r.status = 'error'),
			insert_count = $2,
			update_count = $3,
			unchanged_count = $4,
			deactivate_preview = $5,
			updated_at = now()
		WHERE j.id = $1`, jobID, insert, update, unchanged, deactivatePreview)
	return err
}

// CommitAFMPSImport is gate 3 — transactional; only when status=reviewed.
func (s *Store) CommitAFMPSImport(ctx context.Context, id string, deactivateMissing bool) (AFMPSCommitResult, error) {
	var result AFMPSCommitResult
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `
		SELECT status FROM pharmacy.afmps_import_jobs WHERE id = $1 FOR UPDATE`, id).Scan(&status)
	if err == pgx.ErrNoRows {
		return result, ErrNotFound
	}
	if err != nil {
		return result, err
	}
	if status != "reviewed" {
		return result, ErrConflict
	}
	if _, err := tx.Exec(ctx, `
		UPDATE pharmacy.afmps_import_jobs
		SET status = 'committing', updated_at = now()
		WHERE id = $1`, id); err != nil {
		return result, err
	}

	rows, err := tx.Query(ctx, `
		SELECT id::text, cnk, name, atc_code, pharmaceutical_form, pack_size, amm_number,
		       is_antibiotic, afmps_meta
		FROM pharmacy.afmps_import_rows
		WHERE job_id = $1 AND status = 'ready'
		ORDER BY row_number`, id)
	if err != nil {
		return result, err
	}
	type readyRow struct {
		id, cnk, name, atc, form, pack, amm string
		ab                                 bool
		meta                               []byte
	}
	var ready []readyRow
	for rows.Next() {
		var rr readyRow
		if err := rows.Scan(&rr.id, &rr.cnk, &rr.name, &rr.atc, &rr.form, &rr.pack, &rr.amm, &rr.ab, &rr.meta); err != nil {
			rows.Close()
			return result, err
		}
		ready = append(ready, rr)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return result, err
	}

	keep := make([]string, 0, len(ready))
	for _, rr := range ready {
		norm := NormalizeMedicationName(rr.name)
		newID := uuid.NewString()
		meta := rr.meta
		if len(meta) == 0 {
			meta = []byte(`{}`)
		}
		var outID string
		err := tx.QueryRow(ctx, `
			INSERT INTO pharmacy.ref_medications (
				id, cnk, name, name_normalized, atc_code, pharmaceutical_form, pack_size,
				amm_number, is_antibiotic, is_active, afmps_meta, updated_at
			) VALUES (
				$1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''),
				COALESCE($8, ''), $9, true, $10::jsonb, now()
			)
			ON CONFLICT (cnk) DO UPDATE SET
				name = EXCLUDED.name,
				name_normalized = EXCLUDED.name_normalized,
				atc_code = EXCLUDED.atc_code,
				pharmaceutical_form = EXCLUDED.pharmaceutical_form,
				pack_size = EXCLUDED.pack_size,
				amm_number = CASE
					WHEN EXCLUDED.amm_number <> '' THEN EXCLUDED.amm_number
					ELSE pharmacy.ref_medications.amm_number
				END,
				is_antibiotic = EXCLUDED.is_antibiotic,
				is_active = true,
				afmps_meta = pharmacy.ref_medications.afmps_meta || EXCLUDED.afmps_meta,
				updated_at = now()
			RETURNING id::text`,
			newID, rr.cnk, rr.name, norm, rr.atc, rr.form, rr.pack, rr.amm, rr.ab, string(meta),
		).Scan(&outID)
		if err != nil {
			return result, fmt.Errorf("upsert cnk=%s: %w", rr.cnk, err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE pharmacy.afmps_import_rows SET status = 'upserted' WHERE id = $1`, rr.id); err != nil {
			return result, err
		}
		result.Upserted++
		keep = append(keep, rr.cnk)
	}

	if deactivateMissing && len(keep) > 0 {
		tag, err := tx.Exec(ctx, `
			UPDATE pharmacy.ref_medications
			SET is_active = false, updated_at = now()
			WHERE is_active AND NOT (cnk = ANY($1))`, keep)
		if err != nil {
			return result, err
		}
		result.Deactivated = tag.RowsAffected()
	}

	if _, err := tx.Exec(ctx, `
		UPDATE pharmacy.afmps_import_jobs
		SET status = 'completed', upserted_count = $2, deactivated_count = $3, updated_at = now()
		WHERE id = $1`, id, result.Upserted, result.Deactivated); err != nil {
		return result, err
	}
	if err := tx.Commit(ctx); err != nil {
		return result, err
	}
	return result, nil
}

// DeleteAFMPSImportJob removes staging (not ref_medications). Blocked while committing.
func (s *Store) DeleteAFMPSImportJob(ctx context.Context, id string) error {
	j, err := s.GetAFMPSImportJob(ctx, id)
	if err != nil {
		return err
	}
	if j.Status == "committing" {
		return ErrConflict
	}
	tag, err := s.pool.Exec(ctx, `DELETE FROM pharmacy.afmps_import_jobs WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
