package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func applyFiliationQuery(f *store.FiliationFilter, r *http.Request) {
	f.Query = strings.TrimSpace(r.URL.Query().Get("q"))
	f.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	f.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
}

// writeFiliationList returns the paginated page object by default.
// format=items keeps the previous flat-array contract for legacy clients.
func writeFiliationList(w http.ResponseWriter, r *http.Request, page store.FiliationPage) {
	if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "items") {
		httpx.WriteData(w, http.StatusOK, page.Items)
		return
	}
	httpx.WriteData(w, http.StatusOK, page)
}

func writeFiliationEvents(w http.ResponseWriter, r *http.Request, page store.FiliationEventPage) {
	if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("format")), "items") {
		httpx.WriteData(w, http.StatusOK, page.Items)
		return
	}
	httpx.WriteData(w, http.StatusOK, page)
}

func applyFiliationEventQuery(f *store.FiliationEventFilter, r *http.Request) {
	f.EventType = strings.TrimSpace(r.URL.Query().Get("eventType"))
	f.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	f.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
}
