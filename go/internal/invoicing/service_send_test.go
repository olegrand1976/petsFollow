package invoicing

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
)

// sendStore n'implémente que ce que SendDocument touche ; toute autre méthode
// panique, ce qui vaut assertion sur le périmètre de l'envoi.
type sendStore struct {
	Store
	doc     Document
	written []string // peppol_status successifs
}

func (s *sendStore) GetConnection(context.Context, string) (Connection, string, error) {
	return Connection{Status: ConnActive, BillitPartyID: "party_1"}, "plain:key", nil
}

func (s *sendStore) GetDocument(context.Context, string, string) (Document, error) {
	return s.doc, nil
}

func (s *sendStore) ClaimDocumentForSend(context.Context, string, string, int, int) (Document, DocStatus, error) {
	return s.doc, StatusDraft, nil
}

func (s *sendStore) UpdateDocumentExternal(_ context.Context, _, _, _ string, status DocStatus, peppolStatus string, _ *time.Time) error {
	s.doc.Status = status
	s.written = append(s.written, peppolStatus)
	return nil
}

type sendGateway struct {
	Gateway
	err error
}

func (g *sendGateway) Send(context.Context, string, string, string, Transport) error {
	return g.err
}

func individualInvoice() Document {
	return Document{
		ID:            "doc_1",
		PracticeID:    "p1",
		Type:          DocInvoice,
		Status:        StatusDraft,
		BillitOrderID: "ord_1",
		Counterparty:  Counterparty{Name: "Particulier", CustomerKind: KindIndividual, Country: "BE", Email: "client@example.test"},
	}
}

// Le refus d'identité de Billit ne doit pas se noyer dans ErrGateway : c'est la
// seule panne d'envoi que le cabinet peut lever lui-même, et l'API doit pouvoir
// le lui dire au lieu d'un 502 « service indisponible ».
func TestSendDocumentSurfacesAccountUnverified(t *testing.T) {
	st := &sendStore{doc: individualInvoice()}
	// Même forme que le client Billit : sentinelle + corps brut pour le support.
	gwErr := fmt.Errorf("%w: billit send: http_400 YouMustConfirmYourIdentity…", ErrAccountUnverified)
	svc := NewService(st, &sendGateway{err: gwErr},
		config.Config{BillitEnabled: true, BillitSecretsBackend: "plain_dev"})

	_, err := svc.SendDocument(context.Background(), "p1", "doc_1")
	if !errors.Is(err, ErrAccountUnverified) {
		t.Fatalf("want ErrAccountUnverified, got %v", err)
	}
	if errors.Is(err, ErrGateway) {
		t.Fatalf("account gate must not be reported as a gateway outage: %v", err)
	}
	// Trace d'audit : le canal reste dit (mail au particulier, pas Peppol) et la
	// cause est nommée, sinon le support relit « échec » sans savoir quoi faire.
	if len(st.written) != 1 || st.written[0] != "email_account_unverified" {
		t.Fatalf("peppol_status écrits = %#v", st.written)
	}
}

// Une vraie panne de passerelle garde son 502 : ne pas transformer toute erreur
// d'envoi en « validez votre compte ».
func TestSendDocumentKeepsGatewayFailureDistinct(t *testing.T) {
	st := &sendStore{doc: individualInvoice()}
	svc := NewService(st, &sendGateway{err: errors.New("billit send: http_503 upstream")},
		config.Config{BillitEnabled: true, BillitSecretsBackend: "plain_dev"})

	_, err := svc.SendDocument(context.Background(), "p1", "doc_1")
	if !errors.Is(err, ErrGateway) {
		t.Fatalf("want ErrGateway, got %v", err)
	}
	if errors.Is(err, ErrAccountUnverified) {
		t.Fatalf("upstream outage must not claim the account is unverified: %v", err)
	}
	if len(st.written) != 1 || st.written[0] != "email_send_failed" {
		t.Fatalf("peppol_status écrits = %#v", st.written)
	}
}
