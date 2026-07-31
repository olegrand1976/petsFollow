package pharmacy

import "errors"

// ErrInvoicesConnectResellerPending — DAF→Billit (BIL-9 / S5) bloqué jusqu'accès reseller.
var ErrInvoicesConnectResellerPending = errors.New("invoices_connect_reseller_pending")

// EnqueueInvoicesConnect is the future Asynq entrypoint for DAF→Billit export.
// Intentionally a no-go until Billit reseller credentials are available (roadmap P0-2).
func EnqueueInvoicesConnect(_ string /* dafID */) error {
	return ErrInvoicesConnectResellerPending
}
