package invoicing

import "context"

// Gateway abstracts Billit (live or mock).
// SendPeppol triggers electronic delivery. country (ISO-2) selects transport hint
// (Peppol for BE/FR/ES; omitted for IT so Billit can route via SDI from Identifiers).
type Gateway interface {
	EnsureParty(ctx context.Context, practice PracticeParty) (partyID string, err error)
	CheckParty(ctx context.Context, partyID, apiKey string) (GatewayStatus, error)
	CreateDocument(ctx context.Context, partyID, apiKey string, doc Document) (externalID string, err error)
	SendPeppol(ctx context.Context, partyID, apiKey, externalID, country string) error
}
