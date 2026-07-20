package store

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/i18n"
)

type ColumnMapping struct {
	Email    *string `json:"email"`
	FullName *string `json:"fullName"`
	Locale   *string `json:"locale"`
}

type ClientImportJob struct {
	ID                   string          `json:"id"`
	VetUserID            string          `json:"vetUserId"`
	PracticeID           string          `json:"practiceId"`
	CreatedByAdminID     string          `json:"createdByAdminId"`
	Filename             string          `json:"filename"`
	ContentType          string          `json:"contentType"`
	SourceFormat         string          `json:"sourceFormat"`
	Status               string          `json:"status"`
	Headers              []string        `json:"headers"`
	SampleRows           json.RawMessage `json:"sampleRows"`
	ColumnMapping        *ColumnMapping  `json:"columnMapping,omitempty"`
	GeminiRaw            json.RawMessage `json:"geminiRaw,omitempty"`
	RowCount             int             `json:"rowCount"`
	OkCount              int             `json:"okCount"`
	ErrorCount           int             `json:"errorCount"`
	CreatedCount         int             `json:"createdCount"`
	CredentialsAvailable bool            `json:"credentialsAvailable"`
	ErrorMessage         string          `json:"errorMessage,omitempty"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            time.Time       `json:"updatedAt"`
	VetFullName          string          `json:"vetFullName,omitempty"`
	PracticeName         string          `json:"practiceName,omitempty"`
}

type ClientImportRow struct {
	ID            string            `json:"id"`
	JobID         string            `json:"jobId"`
	RowNumber     int               `json:"rowNumber"`
	Raw           map[string]string `json:"raw"`
	Email         string            `json:"email,omitempty"`
	FullName      string            `json:"fullName,omitempty"`
	Locale        string            `json:"locale,omitempty"`
	Status        string            `json:"status"`
	ErrorCode     string            `json:"errorCode,omitempty"`
	ErrorMessage  string            `json:"errorMessage,omitempty"`
	CreatedUserID string            `json:"createdUserId,omitempty"`
}

type ClientImportJobDetail struct {
	Job  ClientImportJob   `json:"job"`
	Rows []ClientImportRow `json:"rows"`
}

type CreateClientImportInput struct {
	VetUserID        string
	CreatedByAdminID string
	Filename         string
	ContentType      string
	SourceFormat     string
	Headers          []string
	SampleRows       []map[string]string
	Rows             []map[string]string
}

func (s *Store) CreateClientImportJob(ctx context.Context, in CreateClientImportInput) (ClientImportJobDetail, error) {
	var practiceID string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(practice_id::text,'') FROM identity.users
		WHERE id=$1 AND role='vet'`, in.VetUserID).Scan(&practiceID)
	if errors.Is(err, pgx.ErrNoRows) || practiceID == "" {
		return ClientImportJobDetail{}, ErrNotFound
	}
	if err != nil {
		return ClientImportJobDetail{}, err
	}

	jobID := uuid.NewString()
	headersJSON, _ := json.Marshal(in.Headers)
	sampleJSON, _ := json.Marshal(in.SampleRows)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ClientImportJobDetail{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO practice.client_import_jobs (
			id, vet_user_id, practice_id, created_by_admin_id,
			filename, content_type, source_format, status,
			headers, sample_rows, row_count
		) VALUES ($1,$2,$3,$4,$5,$6,$7,'uploaded',$8,$9,$10)`,
		jobID, in.VetUserID, practiceID, in.CreatedByAdminID,
		in.Filename, in.ContentType, in.SourceFormat,
		headersJSON, sampleJSON, len(in.Rows))
	if err != nil {
		return ClientImportJobDetail{}, err
	}

	for i, raw := range in.Rows {
		rawJSON, _ := json.Marshal(raw)
		_, err = tx.Exec(ctx, `
			INSERT INTO practice.client_import_rows (id, job_id, row_number, raw, status)
			VALUES ($1,$2,$3,$4,'pending')`,
			uuid.NewString(), jobID, i+1, rawJSON)
		if err != nil {
			return ClientImportJobDetail{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return ClientImportJobDetail{}, err
	}
	return s.GetClientImportJobDetail(ctx, jobID)
}

func (s *Store) ListClientImportJobs(ctx context.Context, limit int) ([]ClientImportJob, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT j.id::text, j.vet_user_id::text, j.practice_id::text, j.created_by_admin_id::text,
			j.filename, j.content_type, j.source_format, j.status,
			j.headers, j.sample_rows, j.column_mapping, j.gemini_raw,
			j.row_count, j.ok_count, j.error_count, j.created_count,
			j.credentials_cipher IS NOT NULL AND j.credentials_expires_at > NOW() AND j.credentials_downloaded_at IS NULL,
			COALESCE(j.error_message,''), j.created_at, j.updated_at,
			COALESCE(u.full_name,''), COALESCE(p.name,'')
		FROM practice.client_import_jobs j
		LEFT JOIN identity.users u ON u.id = j.vet_user_id
		LEFT JOIN practice.practices p ON p.id = j.practice_id
		ORDER BY j.created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ClientImportJob, 0)
	for rows.Next() {
		j, err := scanClientImportJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanClientImportJob(row scannable) (ClientImportJob, error) {
	var j ClientImportJob
	var headersJSON, sampleJSON, mappingJSON, geminiJSON []byte
	var credsAvail bool
	err := row.Scan(
		&j.ID, &j.VetUserID, &j.PracticeID, &j.CreatedByAdminID,
		&j.Filename, &j.ContentType, &j.SourceFormat, &j.Status,
		&headersJSON, &sampleJSON, &mappingJSON, &geminiJSON,
		&j.RowCount, &j.OkCount, &j.ErrorCount, &j.CreatedCount,
		&credsAvail, &j.ErrorMessage, &j.CreatedAt, &j.UpdatedAt,
		&j.VetFullName, &j.PracticeName,
	)
	if err != nil {
		return j, err
	}
	_ = json.Unmarshal(headersJSON, &j.Headers)
	j.SampleRows = json.RawMessage(sampleJSON)
	if len(mappingJSON) > 0 && string(mappingJSON) != "null" {
		var m ColumnMapping
		if err := json.Unmarshal(mappingJSON, &m); err == nil {
			j.ColumnMapping = &m
		}
	}
	if len(geminiJSON) > 0 && string(geminiJSON) != "null" {
		j.GeminiRaw = json.RawMessage(geminiJSON)
	}
	j.CredentialsAvailable = credsAvail
	return j, nil
}

func (s *Store) GetClientImportJob(ctx context.Context, jobID string) (ClientImportJob, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT j.id::text, j.vet_user_id::text, j.practice_id::text, j.created_by_admin_id::text,
			j.filename, j.content_type, j.source_format, j.status,
			j.headers, j.sample_rows, j.column_mapping, j.gemini_raw,
			j.row_count, j.ok_count, j.error_count, j.created_count,
			j.credentials_cipher IS NOT NULL AND j.credentials_expires_at > NOW() AND j.credentials_downloaded_at IS NULL,
			COALESCE(j.error_message,''), j.created_at, j.updated_at,
			COALESCE(u.full_name,''), COALESCE(p.name,'')
		FROM practice.client_import_jobs j
		LEFT JOIN identity.users u ON u.id = j.vet_user_id
		LEFT JOIN practice.practices p ON p.id = j.practice_id
		WHERE j.id=$1`, jobID)
	j, err := scanClientImportJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return j, ErrNotFound
	}
	return j, err
}

func (s *Store) GetClientImportJobDetail(ctx context.Context, jobID string) (ClientImportJobDetail, error) {
	job, err := s.GetClientImportJob(ctx, jobID)
	if err != nil {
		return ClientImportJobDetail{}, err
	}
	rows, err := s.listClientImportRows(ctx, jobID)
	if err != nil {
		return ClientImportJobDetail{}, err
	}
	return ClientImportJobDetail{Job: job, Rows: rows}, nil
}

func (s *Store) listClientImportRows(ctx context.Context, jobID string) ([]ClientImportRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, job_id::text, row_number, raw,
			COALESCE(email,''), COALESCE(full_name,''), COALESCE(locale,''),
			status, COALESCE(error_code,''), COALESCE(error_message,''),
			COALESCE(created_user_id::text,'')
		FROM practice.client_import_rows
		WHERE job_id=$1
		ORDER BY row_number`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ClientImportRow, 0)
	for rows.Next() {
		var r ClientImportRow
		var rawJSON []byte
		if err := rows.Scan(
			&r.ID, &r.JobID, &r.RowNumber, &rawJSON,
			&r.Email, &r.FullName, &r.Locale,
			&r.Status, &r.ErrorCode, &r.ErrorMessage, &r.CreatedUserID,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(rawJSON, &r.Raw)
		if r.Raw == nil {
			r.Raw = map[string]string{}
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) SaveClientImportMappingSuggestion(ctx context.Context, jobID string, mapping ColumnMapping, geminiRaw json.RawMessage) error {
	mappingJSON, _ := json.Marshal(mapping)
	var gemini any
	if len(geminiRaw) > 0 {
		gemini = []byte(geminiRaw)
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE practice.client_import_jobs
		SET column_mapping=$2, gemini_raw=COALESCE($3, gemini_raw), status='mapping_ready', updated_at=NOW()
		WHERE id=$1 AND status IN ('uploaded','mapping_ready','preview_ready')`,
		jobID, mappingJSON, gemini)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrConflict
	}
	return nil
}

func (s *Store) ApplyClientImportMapping(ctx context.Context, jobID string, mapping ColumnMapping) (ClientImportJobDetail, error) {
	job, err := s.GetClientImportJob(ctx, jobID)
	if err != nil {
		return ClientImportJobDetail{}, err
	}
	switch job.Status {
	case "uploaded", "mapping_ready", "preview_ready":
	default:
		return ClientImportJobDetail{}, ErrConflict
	}

	rows, err := s.listClientImportRows(ctx, jobID)
	if err != nil {
		return ClientImportJobDetail{}, err
	}

	mappingJSON, _ := json.Marshal(mapping)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ClientImportJobDetail{}, err
	}
	defer tx.Rollback(ctx)

	okCount, errCount := 0, 0
	for _, row := range rows {
		if row.Status == "excluded" || row.Status == "created" {
			continue
		}
		email := mappedValue(row.Raw, mapping.Email)
		fullName := mappedValue(row.Raw, mapping.FullName)
		locale := mappedValue(row.Raw, mapping.Locale)
		email = strings.TrimSpace(strings.ToLower(email))
		fullName = strings.TrimSpace(fullName)
		if locale != "" {
			locale = i18n.NormalizeLocale(locale)
		}

		status := "ready"
		errCode, errMsg := "", ""
		switch {
		case email == "" || !looksLikeEmail(email):
			status = "error"
			errCode = "invalid_email"
			errMsg = "invalid email"
		case fullName == "":
			status = "error"
			errCode = "missing_full_name"
			errMsg = "missing full name"
		default:
			var exists bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM identity.users WHERE lower(email)=$1)`, email).Scan(&exists); err != nil {
				return ClientImportJobDetail{}, err
			}
			if exists {
				status = "error"
				errCode = "email_already_exists"
				errMsg = "email already exists"
			}
		}
		if status == "ready" {
			okCount++
		} else {
			errCount++
		}
		_, err = tx.Exec(ctx, `
			UPDATE practice.client_import_rows
			SET email=$2, full_name=$3, locale=$4, status=$5, error_code=$6, error_message=$7
			WHERE id=$1 AND status NOT IN ('excluded','created')`,
			row.ID, nullIfEmpty(email), nullIfEmpty(fullName), nullIfEmpty(locale), status, nullIfEmpty(errCode), nullIfEmpty(errMsg))
		if err != nil {
			return ClientImportJobDetail{}, err
		}
	}

	// Recount excluded so counters stay coherent.
	var excluded int
	_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM practice.client_import_rows WHERE job_id=$1 AND status='excluded'`, jobID).Scan(&excluded)

	_, err = tx.Exec(ctx, `
		UPDATE practice.client_import_jobs
		SET column_mapping=$2, status='preview_ready', ok_count=$3, error_count=$4, updated_at=NOW()
		WHERE id=$1`, jobID, mappingJSON, okCount, errCount)
	if err != nil {
		return ClientImportJobDetail{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ClientImportJobDetail{}, err
	}
	return s.GetClientImportJobDetail(ctx, jobID)
}

func mappedValue(raw map[string]string, header *string) string {
	if header == nil || *header == "" {
		return ""
	}
	return strings.TrimSpace(raw[*header])
}

func looksLikeEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return false
	}
	dot := strings.LastIndexByte(s, '.')
	return dot > at+1 && dot < len(s)-1
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

type PatchClientImportRowInput struct {
	Excluded *bool
	Email    *string
	FullName *string
	Locale   *string
}

func (s *Store) PatchClientImportRow(ctx context.Context, jobID, rowID string, in PatchClientImportRowInput) (ClientImportRow, error) {
	job, err := s.GetClientImportJob(ctx, jobID)
	if err != nil {
		return ClientImportRow{}, err
	}
	if job.Status != "preview_ready" && job.Status != "mapping_ready" {
		return ClientImportRow{}, ErrConflict
	}

	rows, err := s.listClientImportRows(ctx, jobID)
	if err != nil {
		return ClientImportRow{}, err
	}
	var target *ClientImportRow
	for i := range rows {
		if rows[i].ID == rowID {
			target = &rows[i]
			break
		}
	}
	if target == nil {
		return ClientImportRow{}, ErrNotFound
	}
	if target.Status == "created" {
		return ClientImportRow{}, ErrConflict
	}

	email := target.Email
	fullName := target.FullName
	locale := target.Locale
	status := target.Status

	if in.Excluded != nil && *in.Excluded {
		status = "excluded"
	} else if in.Excluded != nil && !*in.Excluded && status == "excluded" {
		status = "pending"
	}
	if in.Email != nil {
		email = strings.TrimSpace(strings.ToLower(*in.Email))
	}
	if in.FullName != nil {
		fullName = strings.TrimSpace(*in.FullName)
	}
	if in.Locale != nil {
		locale = strings.TrimSpace(*in.Locale)
		if locale != "" {
			locale = i18n.NormalizeLocale(locale)
		}
	}

	errCode, errMsg := "", ""
	if status != "excluded" {
		status = "ready"
		switch {
		case email == "" || !looksLikeEmail(email):
			status = "error"
			errCode = "invalid_email"
			errMsg = "invalid email"
		case fullName == "":
			status = "error"
			errCode = "missing_full_name"
			errMsg = "missing full name"
		default:
			var exists bool
			if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM identity.users WHERE lower(email)=$1)`, email).Scan(&exists); err != nil {
				return ClientImportRow{}, err
			}
			if exists {
				status = "error"
				errCode = "email_already_exists"
				errMsg = "email already exists"
			}
		}
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE practice.client_import_rows
		SET email=$2, full_name=$3, locale=$4, status=$5, error_code=$6, error_message=$7
		WHERE id=$1 AND job_id=$8`,
		rowID, nullIfEmpty(email), nullIfEmpty(fullName), nullIfEmpty(locale),
		status, nullIfEmpty(errCode), nullIfEmpty(errMsg), jobID)
	if err != nil {
		return ClientImportRow{}, err
	}

	// Refresh job counters.
	_, _ = s.pool.Exec(ctx, `
		UPDATE practice.client_import_jobs j SET
			ok_count = (SELECT COUNT(*) FROM practice.client_import_rows r WHERE r.job_id=j.id AND r.status='ready'),
			error_count = (SELECT COUNT(*) FROM practice.client_import_rows r WHERE r.job_id=j.id AND r.status='error'),
			status = CASE WHEN j.status IN ('uploaded','mapping_ready') THEN 'preview_ready' ELSE j.status END,
			updated_at = NOW()
		WHERE j.id=$1`, jobID)

	detail, err := s.GetClientImportJobDetail(ctx, jobID)
	if err != nil {
		return ClientImportRow{}, err
	}
	for _, r := range detail.Rows {
		if r.ID == rowID {
			return r, nil
		}
	}
	return ClientImportRow{}, ErrNotFound
}

type CommitClientImportResult struct {
	Job              ClientImportJob `json:"job"`
	CreatedCount     int             `json:"createdCount"`
	ErrorCount       int             `json:"errorCount"`
	CredentialsToken string          `json:"credentialsToken,omitempty"`
}

func (s *Store) CommitClientImport(ctx context.Context, jobID, signingKey string) (CommitClientImportResult, error) {
	job, err := s.GetClientImportJob(ctx, jobID)
	if err != nil {
		return CommitClientImportResult{}, err
	}
	// Only preview_ready can be committed — re-commit of completed would wipe credentials.
	if job.Status != "preview_ready" {
		return CommitClientImportResult{}, ErrConflict
	}

	tag, err := s.pool.Exec(ctx, `
		UPDATE practice.client_import_jobs SET status='importing', updated_at=NOW()
		WHERE id=$1 AND status='preview_ready'`, jobID)
	if err != nil {
		return CommitClientImportResult{}, err
	}
	if tag.RowsAffected() == 0 {
		return CommitClientImportResult{}, ErrConflict
	}

	rows, err := s.listClientImportRows(ctx, jobID)
	if err != nil {
		return CommitClientImportResult{}, err
	}

	creds := make([]importCredential, 0)
	created, failed := 0, 0

	for _, row := range rows {
		if row.Status != "ready" {
			continue
		}
		password, err := randomPassword(16)
		if err != nil {
			return CommitClientImportResult{}, err
		}
		userID, err := s.CreateClientForVet(ctx, job.VetUserID, CreateClientInput{
			Email:       row.Email,
			Password:    password,
			FullName:    row.FullName,
			Locale:      row.Locale,
			SkipJourney: true,
		})
		if err != nil {
			failed++
			code := "create_failed"
			msg := err.Error()
			if strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique") {
				code = "email_already_exists"
				msg = "email already exists"
			}
			_, _ = s.pool.Exec(ctx, `
				UPDATE practice.client_import_rows
				SET status='error', error_code=$2, error_message=$3
				WHERE id=$1`, row.ID, code, truncateMsg(msg, 200))
			continue
		}
		created++
		creds = append(creds, importCredential{Email: row.Email, Password: password, FullName: row.FullName})
		_, _ = s.pool.Exec(ctx, `
			UPDATE practice.client_import_rows
			SET status='created', created_user_id=$2, error_code=NULL, error_message=NULL
			WHERE id=$1`, row.ID, userID)
	}

	token := ""
	var cipherBlob []byte
	var tokenHash any
	var expires any
	if len(creds) > 0 {
		csvBytes, err := buildCredentialsCSV(creds)
		if err != nil {
			return CommitClientImportResult{}, err
		}
		cipherBlob, err = encryptBytes(signingKey, csvBytes)
		if err != nil {
			return CommitClientImportResult{}, err
		}
		rawToken, err := randomToken(32)
		if err != nil {
			return CommitClientImportResult{}, err
		}
		token = rawToken
		h := sha256.Sum256([]byte(rawToken))
		tokenHash = hex.EncodeToString(h[:])
		expires = time.Now().UTC().Add(24 * time.Hour)
	}

	status := "completed"
	_, err = s.pool.Exec(ctx, `
		UPDATE practice.client_import_jobs SET
			status=$2,
			created_count=created_count+$3,
			error_count=(SELECT COUNT(*) FROM practice.client_import_rows WHERE job_id=$1 AND status='error'),
			ok_count=(SELECT COUNT(*) FROM practice.client_import_rows WHERE job_id=$1 AND status='created'),
			credentials_cipher=$4,
			credentials_token_hash=$5,
			credentials_expires_at=$6,
			credentials_downloaded_at=NULL,
			updated_at=NOW()
		WHERE id=$1`, jobID, status, created, cipherBlob, tokenHash, expires)
	if err != nil {
		return CommitClientImportResult{}, err
	}

	job, err = s.GetClientImportJob(ctx, jobID)
	if err != nil {
		return CommitClientImportResult{}, err
	}
	return CommitClientImportResult{
		Job:              job,
		CreatedCount:     created,
		ErrorCount:       failed,
		CredentialsToken: token,
	}, nil
}

func (s *Store) DownloadClientImportCredentials(ctx context.Context, jobID, token, signingKey string) ([]byte, error) {
	var cipherBlob []byte
	var tokenHash string
	var expires *time.Time
	var downloaded *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT credentials_cipher, COALESCE(credentials_token_hash,''), credentials_expires_at, credentials_downloaded_at
		FROM practice.client_import_jobs WHERE id=$1`, jobID).Scan(&cipherBlob, &tokenHash, &expires, &downloaded)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if len(cipherBlob) == 0 || tokenHash == "" || expires == nil || time.Now().UTC().After(*expires) || downloaded != nil {
		return nil, ErrNotFound
	}
	h := sha256.Sum256([]byte(token))
	if hex.EncodeToString(h[:]) != tokenHash {
		return nil, ErrNotFound
	}
	plain, err := decryptBytes(signingKey, cipherBlob)
	if err != nil {
		return nil, err
	}
	_, _ = s.pool.Exec(ctx, `
		UPDATE practice.client_import_jobs
		SET credentials_downloaded_at=NOW(), credentials_cipher=NULL, credentials_token_hash=NULL, updated_at=NOW()
		WHERE id=$1`, jobID)
	return plain, nil
}

type importCredential struct {
	Email    string
	Password string
	FullName string
}

func buildCredentialsCSV(creds []importCredential) ([]byte, error) {
	var b strings.Builder
	w := csv.NewWriter(&b)
	w.Comma = ';'
	if err := w.Write([]string{"email", "fullName", "password"}); err != nil {
		return nil, err
	}
	for _, c := range creds {
		if err := w.Write([]string{c.Email, c.FullName, c.Password}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return []byte(b.String()), w.Error()
}

func randomPassword(n int) (string, error) {
	const alphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$"
	buf := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = alphabet[int(buf[i])%len(alphabet)]
	}
	return string(buf), nil
}

func randomToken(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func encryptBytes(keyMaterial string, plain []byte) ([]byte, error) {
	key := sha256.Sum256([]byte(keyMaterial))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plain, nil), nil
}

func decryptBytes(keyMaterial string, blob []byte) ([]byte, error) {
	key := sha256.Sum256([]byte(keyMaterial))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(blob) < gcm.NonceSize() {
		return nil, fmt.Errorf("cipher_too_short")
	}
	nonce, ciphertext := blob[:gcm.NonceSize()], blob[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func truncateMsg(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
