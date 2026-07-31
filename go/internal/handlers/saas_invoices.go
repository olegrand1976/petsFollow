package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
)

// internalRunSaasInvoices — cron C1 : brouillons Flux A (pas d'envoi Peppol).
// Header X-Saas-Invoices-Secret (env SAAS_INVOICES_SECRET).
//
// Par défaut : mode all (boucle limit/offset jusqu'à scanned < limit).
// Query :
//   - limit (défaut 50, max 100)
//   - offset (départ, défaut 0)
//   - batch=1 : un seul lot (pas de boucle)
func (a *API) internalRunSaasInvoices(w http.ResponseWriter, r *http.Request) {
	if !secretHeaderOK(r, "X-Saas-Invoices-Secret", a.cfg.SaasInvoicesSecret) {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if a.invoicing == nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "invoicing_disabled")
		return
	}
	limit, offset := 50, 0
	all := true
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	if v := strings.TrimSpace(r.URL.Query().Get("batch")); v == "1" || strings.EqualFold(v, "true") {
		all = false
	}
	res, err := a.invoicing.RunMonthlySaasDrafts(r.Context(), limit, offset, all)
	if err != nil {
		a.writeInvoicingErr(w, r, err)
		return
	}
	httpx.WriteData(w, http.StatusOK, res)
}
