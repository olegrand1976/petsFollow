package mock

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
)

// Gateway is an in-memory Billit stand-in for local/CI.
type Gateway struct {
	seq  atomic.Int64
	mu   sync.Mutex
	docs map[string]invoicing.Document
	sent map[string]invoicing.Transport
}

func New() *Gateway {
	return &Gateway{
		docs: map[string]invoicing.Document{},
		sent: map[string]invoicing.Transport{},
	}
}

func (g *Gateway) EnsureParty(_ context.Context, practice invoicing.PracticeParty) (string, error) {
	short := practice.PracticeID
	if len(short) > 8 {
		short = short[:8]
	}
	return fmt.Sprintf("party_mock_%s", short), nil
}

func (g *Gateway) CheckParty(_ context.Context, partyID, apiKey string) (invoicing.GatewayStatus, error) {
	if partyID == "" || apiKey == "" {
		return invoicing.GatewayStatus{Complete: false, Message: "missing_credentials"}, nil
	}
	return invoicing.GatewayStatus{Complete: true, Message: "ok"}, nil
}

func (g *Gateway) CreateDocument(_ context.Context, _, _ string, doc invoicing.Document) (string, error) {
	// Globally unique — sequential ord_N collided across tests and poisoned webhook apply/usage.
	id := fmt.Sprintf("ord_%d_%s", g.seq.Add(1), uuid.NewString())
	g.mu.Lock()
	defer g.mu.Unlock()
	doc.BillitOrderID = id
	g.docs[id] = doc
	return id, nil
}

func (g *Gateway) Send(_ context.Context, _, _, externalID string, transport invoicing.Transport) error {
	if transport == "" {
		return fmt.Errorf("mock send: transport required (order %s)", externalID)
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	doc, ok := g.docs[externalID]
	if !ok {
		// Retry after DB-only rejected state (order known to PF, not this process memory).
		doc = invoicing.Document{BillitOrderID: externalID}
	}
	doc.Status = invoicing.StatusDelivered
	doc.PeppolStatus = "delivered"
	g.docs[externalID] = doc
	g.sent[externalID] = transport
	return nil
}

// FetchStatus rejoue la lecture d'ordre côté Billit : le mock livre en synchrone,
// donc un ordre envoyé répond `delivered`. Un ordre inconnu reste non terminal —
// la réconciliation ne doit rien réécrire sur une lecture vide.
func (g *Gateway) FetchStatus(_ context.Context, _, _, externalID string) (invoicing.DocumentStatus, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	doc, ok := g.docs[externalID]
	if !ok || doc.Status != invoicing.StatusDelivered {
		return invoicing.DocumentStatus{Status: invoicing.StatusSending, PeppolStatus: "unknown"}, nil
	}
	return invoicing.DocumentStatus{
		Status:       invoicing.StatusDelivered,
		PeppolStatus: "delivered",
		Terminal:     true,
	}, nil
}

// LastTransport exposes the channel used for an order (tests / smoke).
func (g *Gateway) LastTransport(externalID string) invoicing.Transport {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.sent[externalID]
}
