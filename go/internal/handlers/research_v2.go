package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) registerResearchV2Routes(r chi.Router) {
	r.Get("/research/groups", a.listResearchGroups)
	r.Post("/research/groups", a.createResearchGroup)
	r.Get("/research/groups/{groupID}", a.getResearchGroup)
	r.Patch("/research/groups/{groupID}", a.patchResearchGroup)
	r.Delete("/research/groups/{groupID}", a.deleteResearchGroup)
	r.Get("/research/groups/{groupID}/members", a.listResearchGroupMembers)
	r.Post("/research/groups/{groupID}/members", a.addResearchGroupMember)
	r.Delete("/research/groups/{groupID}/members/{userID}", a.removeResearchGroupMember)
	r.Get("/research/dataroom/events", a.researchDataRoomEvents)
	r.Get("/admin/research/groups", a.adminListResearchGroups)
	r.Patch("/admin/research/groups/{groupID}/dataroom", a.adminSetResearchGroupDataroom)
}

func researchGroupID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := chi.URLParam(r, "groupID")
	if !isUUID(id) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_id")
		return "", false
	}
	return id, true
}

func (a *API) listResearchGroups(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireResearchRole(w, r)
	if !ok {
		return
	}
	items, err := a.store.ListResearchGroupsForUser(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) createResearchGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireResearchRole(w, r)
	if !ok {
		return
	}
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	g, err := a.store.CreateResearchGroup(r.Context(), id.UserID, body.Name, body.Description)
	if err != nil {
		switch err.Error() {
		case "name_required", "name_too_long", "description_too_long":
			writeErr(w, r, http.StatusBadRequest, "bad_request", err.Error())
		default:
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		}
		return
	}
	httpx.WriteData(w, http.StatusCreated, g)
}

func (a *API) getResearchGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireResearchRole(w, r)
	if !ok {
		return
	}
	groupID, ok := researchGroupID(w, r)
	if !ok {
		return
	}
	g, err := a.store.GetResearchGroupForUser(r.Context(), id.UserID, groupID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, g)
}

func (a *API) patchResearchGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireResearchRole(w, r)
	if !ok {
		return
	}
	groupID, ok := researchGroupID(w, r)
	if !ok {
		return
	}
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	g, err := a.store.UpdateResearchGroup(r.Context(), id.UserID, groupID, body.Name, body.Description)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		case err.Error() == "forbidden":
			writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		case err.Error() == "name_required",
			err.Error() == "name_too_long",
			err.Error() == "description_too_long":
			writeErr(w, r, http.StatusBadRequest, "bad_request", err.Error())
		default:
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		}
		return
	}
	httpx.WriteData(w, http.StatusOK, g)
}

func (a *API) deleteResearchGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireResearchRole(w, r)
	if !ok {
		return
	}
	groupID, ok := researchGroupID(w, r)
	if !ok {
		return
	}
	err := a.store.DeleteResearchGroup(r.Context(), id.UserID, groupID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		case err.Error() == "forbidden":
			writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		default:
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		}
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *API) listResearchGroupMembers(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireResearchRole(w, r)
	if !ok {
		return
	}
	groupID, ok := researchGroupID(w, r)
	if !ok {
		return
	}
	items, err := a.store.ListResearchGroupMembers(r.Context(), id.UserID, groupID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) addResearchGroupMember(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireResearchRole(w, r)
	if !ok {
		return
	}
	groupID, ok := researchGroupID(w, r)
	if !ok {
		return
	}
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	_, err := a.store.TryAddResearchGroupMemberByEmail(r.Context(), id.UserID, groupID, body.Email)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		case "email_required":
			writeErr(w, r, http.StatusBadRequest, "bad_request", "email_required")
		default:
			if errors.Is(err, store.ErrNotFound) {
				writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
				return
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		}
		return
	}
	// Uniform response — do not reveal whether the email maps to a research user.
	httpx.WriteData(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *API) removeResearchGroupMember(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireResearchRole(w, r)
	if !ok {
		return
	}
	groupID, ok := researchGroupID(w, r)
	if !ok {
		return
	}
	targetID := chi.URLParam(r, "userID")
	if !isUUID(targetID) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_id")
		return
	}
	err := a.store.RemoveResearchGroupMember(r.Context(), id.UserID, groupID, targetID)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		case "last_owner":
			writeErr(w, r, http.StatusBadRequest, "bad_request", "last_owner")
		default:
			if errors.Is(err, store.ErrNotFound) {
				writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
				return
			}
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		}
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"ok": true})
}

// researchDataRoomEvents requires research role + membership in a dataroom-enabled group.
func (a *API) researchDataRoomEvents(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireResearchRole(w, r)
	if !ok {
		return
	}
	allowed, err := a.store.UserHasResearchDataroomAccess(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !allowed {
		writeErr(w, r, http.StatusForbidden, "research_dataroom_forbidden", "research_dataroom_forbidden")
		return
	}
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	items, err := a.store.ResearchDataRoomEvents(r.Context(),
		strings.TrimSpace(q.Get("species")),
		strings.TrimSpace(q.Get("signal")),
		strings.TrimSpace(q.Get("country")),
		strings.TrimSpace(q.Get("postalCode")),
		strings.TrimSpace(q.Get("from")),
		strings.TrimSpace(q.Get("to")),
		limit,
	)
	if err != nil {
		if strings.Contains(err.Error(), "invalid_") {
			writeErr(w, r, http.StatusBadRequest, "bad_request", err.Error())
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"items":      items,
		"kAnonymity": store.ResearchKAnonymity,
	})
}

func (a *API) adminListResearchGroups(w http.ResponseWriter, r *http.Request) {
	if !a.requireResearchEnabled(w, r) {
		return
	}
	if _, ok := a.requireAdminOrDev(w, r); !ok {
		return
	}
	items, err := a.store.ListAllResearchGroups(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) adminSetResearchGroupDataroom(w http.ResponseWriter, r *http.Request) {
	if !a.requireResearchEnabled(w, r) {
		return
	}
	if _, ok := a.requireAdminOrDev(w, r); !ok {
		return
	}
	groupID, ok := researchGroupID(w, r)
	if !ok {
		return
	}
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	g, err := a.store.SetResearchGroupDataroomEnabled(r.Context(), groupID, body.Enabled)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, g)
}
