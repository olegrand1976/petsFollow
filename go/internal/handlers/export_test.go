package handlers

import "net/http"

// TryClaimInvite exposes soft invite claim outcomes for handlers_test (export_test pattern).
func TryClaimInvite(a *API, r *http.Request, clientUserID, code string) string {
	return a.tryClaimInvite(r, clientUserID, code)
}
