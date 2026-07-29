package pharmacy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

func TestVamregClientDryRun(t *testing.T) {
	c := &pharmacy.VamregClient{DryRun: true}
	res, err := c.Declare(context.Background(), pharmacy.VamregDeclareRequest{
		DAFID: "x", PracticeID: "p", Lines: []pharmacy.VamregLine{{MedicationCNK: "1", Qty: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.DryRun || res.HTTPStatus != 200 {
		t.Fatalf("unexpected %#v", res)
	}
}

func TestVamregClientLiveRequiresBaseURL(t *testing.T) {
	c := &pharmacy.VamregClient{DryRun: false, BaseURL: ""}
	_, err := c.Declare(context.Background(), pharmacy.VamregDeclareRequest{DAFID: "d1"})
	if err == nil || !strings.Contains(err.Error(), "vamreg_base_url_required") {
		t.Fatalf("want vamreg_base_url_required, got %v", err)
	}
}

func TestVamregClientIdempotencyKey(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("Idempotency-Key")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := &pharmacy.VamregClient{DryRun: false, BaseURL: srv.URL, HTTPClient: srv.Client()}
	_, err := c.Declare(context.Background(), pharmacy.VamregDeclareRequest{
		DAFID: "daf-uuid-42", PracticeID: "p",
		Lines: []pharmacy.VamregLine{{MedicationCNK: "1", Qty: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotKey != "daf-uuid-42" {
		t.Fatalf("Idempotency-Key=%q", gotKey)
	}
}

func TestValidateVamregPayload(t *testing.T) {
	if err := pharmacy.ValidateVamregPayload(json.RawMessage(`{}`)); err == nil {
		t.Fatal("want incomplete")
	}
	ok := json.RawMessage(`{"species":"dog","indication":"inf","durationDays":5}`)
	if err := pharmacy.ValidateVamregPayload(ok); err != nil {
		t.Fatal(err)
	}
}

type fakeVamregStore struct {
	mu        sync.Mutex
	audits    []string
	status    string
	failAudit bool
}

func (f *fakeVamregStore) NextPharmacyJobAttempt(ctx context.Context, jobType, entityID string) (int, error) {
	return 1, nil
}
func (f *fakeVamregStore) InsertPharmacyJobAudit(ctx context.Context, practiceID, jobType, entityID string, attempt int, status string, req, resp json.RawMessage, errMsg string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failAudit && status != "started" {
		return context.Canceled
	}
	f.audits = append(f.audits, status)
	return nil
}
func (f *fakeVamregStore) UpdateDAFVamregStatus(ctx context.Context, practiceID, dafID, status string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status = status
	return nil
}
func (f *fakeVamregStore) BuildVamregDeclareRequest(ctx context.Context, practiceID, dafID string) (pharmacy.VamregDeclareRequest, error) {
	return pharmacy.VamregDeclareRequest{
		DAFID: dafID, PracticeID: practiceID,
		Lines: []pharmacy.VamregLine{{MedicationCNK: "1", Qty: 1, Payload: json.RawMessage(`{"species":"dog","indication":"x","durationDays":1}`)}},
	}, nil
}
func (f *fakeVamregStore) GetDAFVamregGate(ctx context.Context, practiceID, dafID string) (status, vamreg string, hasAB bool, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.status == "" {
		return "finalized", "pending", true, nil
	}
	return "finalized", f.status, true, nil
}

func TestProcessDeclarePersistsFailureAfterCancel(t *testing.T) {
	st := &fakeVamregStore{}
	decl := &pharmacy.VamregDeclarer{
		Client: &pharmacy.VamregClient{
			DryRun:     false,
			BaseURL:    "http://127.0.0.1:1", // connection refused
			HTTPClient: &http.Client{Timeout: 50 * time.Millisecond},
		},
		Store: st,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled — persistFailure must still write via WithoutCancel
	err := decl.ProcessDeclare(ctx, "p", "daf1")
	if err == nil {
		t.Fatal("want declare error")
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	foundFailed := false
	for _, a := range st.audits {
		if a == "failed" {
			foundFailed = true
		}
	}
	if !foundFailed || st.status != "failed" {
		t.Fatalf("audits=%v status=%q", st.audits, st.status)
	}
}

func TestProcessDeclareSuccessDoesNotFailTaskOnAuditError(t *testing.T) {
	st := &fakeVamregStore{failAudit: true}
	decl := &pharmacy.VamregDeclarer{
		Client: &pharmacy.VamregClient{DryRun: true},
		Store:  st,
	}
	if err := decl.ProcessDeclare(context.Background(), "p", "daf1"); err != nil {
		t.Fatalf("success path must return nil even if audit persist fails: %v", err)
	}
}

func TestProcessDeclareSkipsWhenAlreadySent(t *testing.T) {
	st := &fakeVamregStore{status: "sent"}
	decl := &pharmacy.VamregDeclarer{
		Client: &pharmacy.VamregClient{DryRun: true},
		Store:  st,
	}
	if err := decl.ProcessDeclare(context.Background(), "p", "daf1"); err != nil {
		t.Fatalf("already sent must be no-op: %v", err)
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	if len(st.audits) != 0 {
		t.Fatalf("expected no audits when already sent, got %v", st.audits)
	}
}
