package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) registerAdminRoutes(r chi.Router) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/admin/metrics/overview", a.adminMetricsOverview)
		pr.Get("/admin/users", a.adminListUsers)
		pr.Get("/admin/payments", a.adminListPayments)
		pr.Get("/admin/commercials", a.adminListCommercials)
		pr.Get("/admin/filiation", a.adminListFiliation)
		pr.Get("/admin/filiation/events", a.adminListFiliationEvents)
		pr.Get("/admin/commercial-managers", a.adminListCommercialManagers)
		pr.Get("/admin/sales-branches", a.adminListSalesBranches)
		pr.Post("/admin/sales-branches", a.adminCreateSalesBranch)
		pr.Patch("/admin/commercials/{id}/branch", a.adminSetCommercialBranch)
		pr.Post("/admin/commercials", a.adminCreateCommercial)
		pr.Patch("/admin/commercials/{id}/assign", a.adminAssignVet)
		pr.Patch("/admin/commercials/{id}/manager", a.adminSetCommercialManager)
		pr.Patch("/admin/commercials/{id}/base-location", a.adminPatchCommercialBaseLocation)
		pr.Get("/admin/vets", a.adminListVets)
		pr.Get("/admin/vets/unassigned", a.adminListUnassignedVets)
		pr.Patch("/admin/vets/{id}/unassign", a.adminUnassignVet)
		pr.Get("/admin/vets/{id}/assign-suggestions", a.adminVetAssignSuggestions)
		pr.Post("/admin/vets", a.adminCreateVet)
		pr.Post("/admin/clients", a.adminCreateClient)
		pr.Post("/admin/care-pros", a.adminCreateCarePro)
		pr.Get("/admin/commercials/{id}/commissions", a.adminCommercialCommissions)
		pr.Get("/admin/prospects", a.adminListProspects)
		pr.Get("/admin/staging/seed", a.adminStagingSeedStatus)
		pr.Post("/admin/staging/seed", a.adminStagingSeed)
		a.registerClientImportRoutes(pr)
		a.registerStripeCatalogRoutes(pr)
	})
}

func (a *API) adminListCommercials(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	rows, err := a.store.ListAllCommercialsAdmin(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) adminListFiliation(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	f := store.FiliationFilter{}
	applyFiliationQuery(&f, r)
	if raw := strings.TrimSpace(r.URL.Query().Get("branchId")); raw != "" {
		if !isUUID(raw) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "bad_request")
			return
		}
		f.BranchID = raw
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("commercialId")); raw != "" {
		if !isUUID(raw) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "bad_request")
			return
		}
		f.CommercialIDs = []string{raw}
	}
	page, err := a.store.ListFiliation(r.Context(), f)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	writeFiliationList(w, r, page)
}

func (a *API) adminListFiliationEvents(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	f := store.FiliationEventFilter{}
	applyFiliationEventQuery(&f, r)
	if raw := strings.TrimSpace(r.URL.Query().Get("commercialId")); raw != "" {
		if !isUUID(raw) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "bad_request")
			return
		}
		f.CommercialIDs = []string{raw}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("clientUserId")); raw != "" {
		if !isUUID(raw) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "bad_request")
			return
		}
		f.ClientUserID = raw
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("vetUserId")); raw != "" {
		if !isUUID(raw) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "bad_request")
			return
		}
		f.VetUserID = raw
	}
	page, err := a.store.ListFiliationEvents(r.Context(), f)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	writeFiliationEvents(w, r, page)
}

func (a *API) adminListSalesBranches(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	rows, err := a.store.ListSalesBranches(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

type createBranchReq struct {
	Name          string `json:"name"`
	Code          string `json:"code"`
	ExternalMLMID string `json:"externalMlmId"`
}

func (a *API) adminCreateSalesBranch(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	var req createBranchReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	b, err := a.store.CreateSalesBranch(r.Context(), req.Name, req.Code, req.ExternalMLMID)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
			return
		}
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "conflict", "code_already_exists")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, b)
}

type setBranchReq struct {
	BranchID string `json:"branchId"`
}

func (a *API) adminSetCommercialBranch(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	commercialID := chi.URLParam(r, "id")
	var req setBranchReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err := a.store.SetUserBranch(r.Context(), commercialID, strings.TrimSpace(req.BranchID)); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) adminListCommercialManagers(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	rows, err := a.store.ListCommercialManagers(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

type createCommercialReq struct {
	Email          string   `json:"email"`
	Password       string   `json:"password"`
	FullName       string   `json:"fullName"`
	ManagerUserID  string   `json:"managerUserId"`
	Role           string   `json:"role"` // commercial (default) | commercial_manager
	BaseLat        *float64 `json:"baseLat"`
	BaseLng        *float64 `json:"baseLng"`
	BaseCity       string   `json:"baseCity"`
	BasePostalCode string   `json:"basePostalCode"`
}

func (a *API) adminCreateCommercial(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	var req createCommercialReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" || req.FullName == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	if len(req.Password) < 8 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "password_too_short")
		return
	}
	if _, err := a.store.GetUserByEmail(r.Context(), req.Email); err == nil {
		writeErr(w, r, http.StatusConflict, "conflict", "email_already_exists")
		return
	} else if !errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "commercial"
	}
	var userID string
	var err error
	switch role {
	case "commercial_manager":
		userID, err = a.store.CreateCommercialManagerUser(r.Context(), req.Email, req.Password, req.FullName)
	case "commercial":
		userID, err = a.store.CreateCommercialUserWithManager(r.Context(), req.Email, req.Password, req.FullName, strings.TrimSpace(req.ManagerUserID))
	default:
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_role")
		return
	}
	if err != nil {
		if err.Error() == "invalid_manager" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_manager")
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_manager")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if req.BaseLat != nil || req.BaseLng != nil || strings.TrimSpace(req.BaseCity) != "" || strings.TrimSpace(req.BasePostalCode) != "" {
		if err := a.store.UpdateCommercialBaseLocation(r.Context(), userID, store.CommercialBaseLocation{
			Lat: req.BaseLat, Lng: req.BaseLng,
			City: req.BaseCity, PostalCode: req.BasePostalCode,
		}); err != nil {
			if errors.Is(err, store.ErrValidation) {
				writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_base_location")
				return
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
	}
	httpx.WriteData(w, http.StatusCreated, map[string]string{"userId": userID, "email": req.Email, "role": role})
}

type assignVetReq struct {
	VetUserID string `json:"vetUserId"`
}

type setManagerReq struct {
	ManagerUserID string `json:"managerUserId"`
}

func (a *API) adminSetCommercialManager(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	commercialID := chi.URLParam(r, "id")
	var req setManagerReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err := a.store.SetCommercialManager(r.Context(), commercialID, strings.TrimSpace(req.ManagerUserID)); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		if err.Error() == "invalid_manager" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_manager")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
}

type baseLocationReq struct {
	BaseLat        *float64 `json:"baseLat"`
	BaseLng        *float64 `json:"baseLng"`
	BaseCity       string   `json:"baseCity"`
	BasePostalCode string   `json:"basePostalCode"`
}

func (a *API) adminPatchCommercialBaseLocation(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	commercialID := chi.URLParam(r, "id")
	var req baseLocationReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if err := a.store.UpdateCommercialBaseLocation(r.Context(), commercialID, store.CommercialBaseLocation{
		Lat: req.BaseLat, Lng: req.BaseLng,
		City: req.BaseCity, PostalCode: req.BasePostalCode,
	}); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_base_location")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	loc, err := a.store.GetCommercialBaseLocation(r.Context(), commercialID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, loc)
}

func (a *API) adminAssignVet(w http.ResponseWriter, r *http.Request) {
	adminID, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	commercialID := chi.URLParam(r, "id")
	var req assignVetReq
	if err := httpx.DecodeJSON(r, &req); err != nil || req.VetUserID == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	if err := a.store.AssignVetToCommercial(r.Context(), req.VetUserID, commercialID, adminID.UserID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_commercial")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "assigned", "vetUserId": req.VetUserID, "commercialId": commercialID})
}

func (a *API) adminListUnassignedVets(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	rows, err := a.store.ListUnassignedVets(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if rows == nil {
		rows = []store.UnassignedVetRow{}
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) adminUnassignVet(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	vetID := chi.URLParam(r, "id")
	if strings.TrimSpace(vetID) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	if err := a.store.UnassignVetFromCommercial(r.Context(), vetID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "unassigned", "vetUserId": vetID})
}

func (a *API) adminVetAssignSuggestions(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	vetID := chi.URLParam(r, "id")
	if strings.TrimSpace(vetID) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	rows, err := a.store.SuggestCommercialsForVet(r.Context(), vetID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if rows == nil {
		rows = []store.CommercialSuggestion{}
	}
	a.refineCommercialSuggestionNotes(r, rows)
	httpx.WriteData(w, http.StatusOK, rows)
}

// refineCommercialSuggestionNotes optionally asks Gemini to polish notes; failures keep deterministic notes.
func (a *API) refineCommercialSuggestionNotes(r *http.Request, rows []store.CommercialSuggestion) {
	if a.gemini == nil || !a.gemini.Configured() || len(rows) == 0 {
		return
	}
	type noteIn struct {
		UserID string `json:"userId"`
		Note   string `json:"note"`
		Score  int    `json:"score"`
	}
	in := make([]noteIn, 0, len(rows))
	for _, s := range rows {
		in = append(in, noteIn{UserID: s.UserID, Note: s.Note, Score: s.Score})
	}
	rawIn, err := json.Marshal(in)
	if err != nil {
		return
	}
	system := `Tu aides un admin petsFollow à assigner un cabinet à un commercial.
Pour chaque suggestion, refine la note en 1 phrase courte, professionnelle, en français.
Réponds UNIQUEMENT un JSON array: [{"userId":"...","note":"..."}]`
	user := "Suggestions:\n" + string(rawIn)
	out, err := a.gemini.GenerateJSONLite(r.Context(), system, user, 0.2)
	if err != nil || strings.TrimSpace(out) == "" {
		return
	}
	var refined []struct {
		UserID string `json:"userId"`
		Note   string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &refined); err != nil {
		// Strip markdown fences if present
		trimmed := strings.TrimSpace(out)
		if i := strings.Index(trimmed, "["); i >= 0 {
			if j := strings.LastIndex(trimmed, "]"); j > i {
				_ = json.Unmarshal([]byte(trimmed[i:j+1]), &refined)
			}
		}
	}
	if len(refined) == 0 {
		return
	}
	byID := make(map[string]string, len(refined))
	for _, n := range refined {
		if n.UserID != "" && strings.TrimSpace(n.Note) != "" {
			byID[n.UserID] = strings.TrimSpace(n.Note)
		}
	}
	for i := range rows {
		if note, ok := byID[rows[i].UserID]; ok {
			rows[i].Note = note
		}
	}
}

func (a *API) adminCommercialCommissions(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_limit")
			return
		}
		limit = n
	}
	summary, err := a.store.GetCommercialCommissionSummary(r.Context(), chi.URLParam(r, "id"), limit)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, summary)
}

func (a *API) adminListProspects(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	status := r.URL.Query().Get("status")
	if status != "" && !store.ValidProspectStatus(status) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
		return
	}
	rows, err := a.store.ListAllProspects(r.Context(), status)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) requireAdmin(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleAdmin {
		writeErr(w, r, http.StatusForbidden, "forbidden", "admin_only")
		return authx.Identity{}, false
	}
	return id, true
}

func (a *API) adminMetricsOverview(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	from, to := parseAdminRange(r)
	m, err := a.store.AdminMetricsOverview(r.Context(), from, to)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, m)
}

func (a *API) adminListUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	from, to := parseAdminRange(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 50
	offset := (page - 1) * limit
	rows, err := a.store.ListAdminUsers(r.Context(), r.URL.Query().Get("role"), from, to, limit, offset)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) adminListPayments(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	from, to := parseAdminRange(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 50
	offset := (page - 1) * limit
	rows, err := a.store.ListAdminPayments(r.Context(), from, to, r.URL.Query().Get("status"), limit, offset)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}
