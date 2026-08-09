package invoicing

import "context"

// Gateway abstracts Billit (live or mock).
// Send triggers delivery on the transport resolved from the counterparty:
// Peppol (B2B), SDI (Italy), SMTP (particulier — no e-invoice network address).
type Gateway interface {
	EnsureParty(ctx context.Context, practice PracticeParty) (partyID string, err error)
	CheckParty(ctx context.Context, partyID, apiKey string) (GatewayStatus, error)
	CreateDocument(ctx context.Context, partyID, apiKey string, doc Document) (externalID string, err error)
	Send(ctx context.Context, partyID, apiKey, externalID string, transport Transport) error
	FetchStatus(ctx context.Context, partyID, apiKey, externalID string) (DocumentStatus, error)
}

// DocumentStatus est l'état d'un ordre lu chez Billit (réconciliation).
// Terminal distingue une réponse exploitable d'un statut encore en vol ou
// illisible : on ne réécrit jamais un document sur une lecture incertaine.
type DocumentStatus struct {
	Status       DocStatus
	PeppolStatus string
	Terminal     bool
}
