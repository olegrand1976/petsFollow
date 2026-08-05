package handlers

import (
	"net/http"
	"reflect"

	"github.com/olegrand1976/petsFollow/go/internal/seed"
)

// TryClaimInvite exposes soft invite claim outcomes for handlers_test (export_test pattern).
func TryClaimInvite(a *API, r *http.Request, clientUserID, code string) string {
	return a.tryClaimInvite(r, clientUserID, code)
}

// StagingSeedRunnerIsDefault dit si /admin/staging/seed pointe toujours sur seed.Run
// (garde-fou de câblage : les tests stubbent le runner, la prod doit rester sur le vrai seed).
func StagingSeedRunnerIsDefault(a *API) bool {
	if a.stagingSeedRun == nil {
		return false
	}
	return reflect.ValueOf(a.stagingSeedRun).Pointer() == reflect.ValueOf(seed.Run).Pointer()
}
