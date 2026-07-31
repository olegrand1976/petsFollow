package pharmacy

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

// VamregStore is the persistence surface for VAMReg declaration jobs.
type VamregStore interface {
	NextPharmacyJobAttempt(ctx context.Context, jobType, entityID string) (int, error)
	InsertPharmacyJobAudit(ctx context.Context, practiceID, jobType, entityID string, attempt int, status string, req, resp json.RawMessage, errMsg string) error
	UpdateDAFVamregStatus(ctx context.Context, practiceID, dafID, status string) error
	BuildVamregDeclareRequest(ctx context.Context, practiceID, dafID string) (VamregDeclareRequest, error)
	GetDAFVamregGate(ctx context.Context, practiceID, dafID string) (status, vamreg string, hasAB bool, err error)
}

// VamregDeclarer runs one declaration attempt with audit + status update.
type VamregDeclarer struct {
	Client *VamregClient
	Store  VamregStore
}

// NewVamregDeclarer builds the single canonical declarer from config fields.
func NewVamregDeclarer(store VamregStore, baseURL, apiKey string, dryRun bool) *VamregDeclarer {
	return &VamregDeclarer{
		Client: &VamregClient{BaseURL: baseURL, APIKey: apiKey, DryRun: dryRun},
		Store:  store,
	}
}

// ProcessDeclare loads the DAF, calls VAMReg (or dry-run), writes job_audit, updates vamreg_status.
// After a successful Declare, persistence errors are logged only (return nil) so Asynq
// retries never re-POST a declaration that already reached the authority.
// Failure audits use a detached short-timeout context so a cancelled request ctx
// cannot erase the regulatory trail.
// Pre-HTTP errors (build / attempt / audit start) flip vamreg_status to failed so
// the DAF never stays stuck in pending without a recoverable UI path.
func (d *VamregDeclarer) ProcessDeclare(ctx context.Context, practiceID, dafID string) error {
	if d == nil || d.Store == nil {
		return ErrDAFNotFound
	}
	_, vamreg, _, err := d.Store.GetDAFVamregGate(ctx, practiceID, dafID)
	if err != nil {
		return err
	}
	if vamreg == "sent" {
		return nil
	}
	req, err := d.Store.BuildVamregDeclareRequest(ctx, practiceID, dafID)
	if err != nil {
		d.markFailedPreDeclare(ctx, practiceID, dafID, err)
		return err
	}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		d.markFailedPreDeclare(ctx, practiceID, dafID, err)
		return err
	}
	attempt, err := d.Store.NextPharmacyJobAttempt(ctx, "vamreg", dafID)
	if err != nil {
		d.markFailedPreDeclare(ctx, practiceID, dafID, err)
		return err
	}
	if err := d.Store.InsertPharmacyJobAudit(ctx, practiceID, "vamreg", dafID, attempt, "started", reqJSON, nil, ""); err != nil {
		d.markFailedPreDeclare(ctx, practiceID, dafID, err)
		return err
	}

	client := d.Client
	if client == nil {
		client = &VamregClient{DryRun: true}
	}
	res, declErr := client.Declare(ctx, req)
	respJSON := res.ResponseBody
	if len(respJSON) == 0 {
		respJSON = json.RawMessage(`{}`)
	}
	if declErr != nil {
		d.persistFailure(ctx, practiceID, dafID, attempt, reqJSON, respJSON, declErr)
		return declErr
	}
	d.persistSuccess(ctx, practiceID, dafID, attempt, reqJSON, respJSON)
	return nil
}

// markFailedPreDeclare leaves the DAF retryable after a failure before any HTTP call.
func (d *VamregDeclarer) markFailedPreDeclare(parent context.Context, practiceID, dafID string, cause error) {
	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(parent), 5*time.Second)
	defer cancel()
	if err := d.Store.UpdateDAFVamregStatus(auditCtx, practiceID, dafID, "failed"); err != nil {
		log.Printf("vamreg pre-declare status failed daf=%s cause=%v status=%v", dafID, cause, err)
	}
}

func (d *VamregDeclarer) persistSuccess(parent context.Context, practiceID, dafID string, attempt int, reqJSON, respJSON json.RawMessage) {
	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(parent), 5*time.Second)
	defer cancel()
	if err := d.Store.InsertPharmacyJobAudit(auditCtx, practiceID, "vamreg", dafID, attempt, "success", reqJSON, respJSON, ""); err != nil {
		log.Printf("vamreg audit success persist daf=%s: %v", dafID, err)
	}
	if err := d.Store.UpdateDAFVamregStatus(auditCtx, practiceID, dafID, "sent"); err != nil {
		log.Printf("vamreg status sent persist daf=%s: %v", dafID, err)
	}
}

func (d *VamregDeclarer) persistFailure(parent context.Context, practiceID, dafID string, attempt int, reqJSON, respJSON json.RawMessage, declErr error) {
	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(parent), 5*time.Second)
	defer cancel()
	if err := d.Store.InsertPharmacyJobAudit(auditCtx, practiceID, "vamreg", dafID, attempt, "failed", reqJSON, respJSON, declErr.Error()); err != nil {
		log.Printf("vamreg audit failed daf=%s declare=%v audit=%v", dafID, declErr, err)
	}
	if err := d.Store.UpdateDAFVamregStatus(auditCtx, practiceID, dafID, "failed"); err != nil {
		log.Printf("vamreg status failed daf=%s declare=%v status=%v", dafID, declErr, err)
	}
}
