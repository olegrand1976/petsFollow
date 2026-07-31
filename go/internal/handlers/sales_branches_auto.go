package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// internalRunSalesBranchesAuto — job bi-quotidien (10h / 18h Europe/Brussels) :
// crée une branche pour chaque commercial sans branche et non rattaché à un peer commercial,
// puis envoie un email de félicitations.
// Protégé par X-Sales-Branches-Auto-Secret (env SALES_BRANCHES_AUTO_SECRET).
func (a *API) internalRunSalesBranchesAuto(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Sales-Branches-Auto-Secret", a.cfg.SalesBranchesAutoSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	result, err := a.runSalesBranchesAuto(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"created": len(result.Created),
		"skipped": len(result.Skipped),
		"items":   result.Created,
		"skippedItems": result.Skipped,
	})
}

func (a *API) runSalesBranchesAuto(ctx context.Context) (store.AutoBranchRunResult, error) {
	result, err := a.store.RunAutoSalesBranches(ctx)
	if err != nil {
		return result, err
	}
	for _, item := range result.Created {
		if a.notifier == nil {
			continue
		}
		networkURL := strings.TrimRight(a.cfg.ProPublicSiteURL, "/") + "/commercial/network"
		if err := a.notifier.SendSalesBranchCongrats(
			item.Email, item.Locale, item.FullName, item.Branch.Name, item.Branch.Code, networkURL,
		); err != nil {
			fmt.Printf("sales-branches-auto: email to %s failed: %v\n", item.Email, err)
		}
	}
	return result, nil
}
