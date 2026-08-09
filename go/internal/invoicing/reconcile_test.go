package invoicing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
)

// reconcileStore n'implémente que les méthodes lues par ReconcileSending :
// l'interface embarquée fait paniquer tout appel non prévu, ce qui vaut
// assertion sur le périmètre du job.
type reconcileStore struct {
	Store
	docs    []SendingDoc
	applied map[string]string // orderID -> peppolStatus appliqué
}

func (s *reconcileStore) ListSendingDocuments(context.Context, time.Time, int) ([]SendingDoc, error) {
	return s.docs, nil
}

func (s *reconcileStore) GetConnection(context.Context, string) (Connection, string, error) {
	return Connection{Status: ConnActive, BillitPartyID: "party_1"}, "plain:key", nil
}

func (s *reconcileStore) ApplyBillitWebhookStatus(_ context.Context, orderID string, _ DocStatus, peppolStatus string, _ *time.Time, _ int) error {
	if s.applied == nil {
		s.applied = map[string]string{}
	}
	s.applied[orderID] = peppolStatus
	return nil
}

type reconcileGateway struct {
	Gateway
	statuses map[string]DocumentStatus
	errs     map[string]error
}

func (g *reconcileGateway) FetchStatus(_ context.Context, _, _, externalID string) (DocumentStatus, error) {
	if err := g.errs[externalID]; err != nil {
		return DocumentStatus{}, err
	}
	return g.statuses[externalID], nil
}

func newReconcileService(st Store, gw Gateway) *Service {
	return NewService(st, gw, config.Config{BillitEnabled: true, BillitSecretsBackend: "plain_dev"})
}

func TestReconcileSendingAppliesOnlyTerminalStatuses(t *testing.T) {
	st := &reconcileStore{docs: []SendingDoc{
		{DocumentID: "d1", PracticeID: "p1", BillitOrderID: "ord_delivered"},
		{DocumentID: "d2", PracticeID: "p1", BillitOrderID: "ord_rejected"},
		{DocumentID: "d3", PracticeID: "p1", BillitOrderID: "ord_flying"},
		{DocumentID: "d4", PracticeID: "p1", BillitOrderID: "ord_unknown"},
	}}
	gw := &reconcileGateway{statuses: map[string]DocumentStatus{
		"ord_delivered": {Status: StatusDelivered, PeppolStatus: "delivered", Terminal: true},
		"ord_rejected":  {Status: StatusRejected, PeppolStatus: "rejected", Terminal: true},
		"ord_flying":    {Status: StatusSending, PeppolStatus: "sending"},
		// Statut illisible : ne jamais inventer une confirmation.
		"ord_unknown": {Status: StatusSending, PeppolStatus: "unknown"},
	}}

	res, err := newReconcileService(st, gw).ReconcileSending(context.Background(), time.Hour, 100)
	if err != nil {
		t.Fatal(err)
	}
	if res.Candidates != 4 || res.Reconciled != 2 || res.StillFlying != 2 || res.Failed != 0 {
		t.Fatalf("unexpected result %#v", res)
	}
	if st.applied["ord_delivered"] != "delivered" || st.applied["ord_rejected"] != "rejected" {
		t.Fatalf("terminal statuses not applied: %#v", st.applied)
	}
	if _, ok := st.applied["ord_flying"]; ok {
		t.Fatal("in-flight document must not be rewritten")
	}
	if _, ok := st.applied["ord_unknown"]; ok {
		t.Fatal("unreadable status must not be rewritten")
	}
}

// Une panne Billit ne doit ni écrire de statut, ni interrompre le lot.
func TestReconcileSendingSkipsGatewayFailures(t *testing.T) {
	st := &reconcileStore{docs: []SendingDoc{
		{DocumentID: "d1", PracticeID: "p1", BillitOrderID: "ord_down"},
		{DocumentID: "d2", PracticeID: "p1", BillitOrderID: "ord_delivered"},
	}}
	gw := &reconcileGateway{
		errs: map[string]error{"ord_down": errors.New("http_503")},
		statuses: map[string]DocumentStatus{
			"ord_delivered": {Status: StatusDelivered, PeppolStatus: "delivered", Terminal: true},
		},
	}

	res, err := newReconcileService(st, gw).ReconcileSending(context.Background(), time.Hour, 100)
	if err != nil {
		t.Fatal(err)
	}
	if res.Failed != 1 || res.Reconciled != 1 {
		t.Fatalf("unexpected result %#v", res)
	}
	if _, ok := st.applied["ord_down"]; ok {
		t.Fatal("failed read must leave the document untouched")
	}
	// Le compteur seul ne se diagnostique pas : la cause doit être nommée.
	if len(res.Errors) != 1 || !strings.Contains(res.Errors[0], "d1") || !strings.Contains(res.Errors[0], "fetch_status") {
		t.Fatalf("failure cause not reported: %#v", res.Errors)
	}
}

// Un incident généralisé (clé Billit tournée) ne doit pas gonfler la réponse du
// cron : les causes sont plafonnées et le résumé le signale.
func TestReconcileSendingCapsReportedErrors(t *testing.T) {
	st := &reconcileStore{}
	gw := &reconcileGateway{errs: map[string]error{}}
	for i := 0; i < maxReconcileErrors+5; i++ {
		order := fmt.Sprintf("ord_%d", i)
		st.docs = append(st.docs, SendingDoc{DocumentID: fmt.Sprintf("d%d", i), PracticeID: "p1", BillitOrderID: order})
		gw.errs[order] = errors.New(strings.Repeat("billit body ", 40))
	}

	res, err := newReconcileService(st, gw).ReconcileSending(context.Background(), time.Hour, 100)
	if err != nil {
		t.Fatal(err)
	}
	if res.Failed != maxReconcileErrors+5 || len(res.Errors) != maxReconcileErrors || !res.ErrorsTruncated {
		t.Fatalf("unexpected result %#v", res)
	}
	for _, msg := range res.Errors {
		if len([]rune(msg)) > maxReconcileErrLen+1 {
			t.Fatalf("error message not truncated: %d runes", len([]rune(msg)))
		}
	}
}

func TestReconcileSendingRefusedWhenBillitDisabled(t *testing.T) {
	svc := NewService(&reconcileStore{}, &reconcileGateway{}, config.Config{})
	if _, err := svc.ReconcileSending(context.Background(), time.Hour, 10); !errors.Is(err, ErrDisabled) {
		t.Fatalf("want ErrDisabled, got %v", err)
	}
}
