package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type createLabPanelReq struct {
	CollectedAt *time.Time                  `json:"collectedAt"`
	LabName     string                      `json:"labName"`
	Notes       string                      `json:"notes"`
	DocumentID  *string                     `json:"documentId"`
	Results     []store.LabPanelResultInput `json:"results"`
}

type patchLabPanelReq struct {
	Results *[]store.LabPanelResultInput `json:"results"`
}

// requireLabWriteAccess — practice staff with pets.write_clinical only (no care_pro).
func (a *API) requireLabWriteAccess(w http.ResponseWriter, r *http.Request, petID string) (store.Pet, authx.Identity, bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return store.Pet{}, authx.Identity{}, false
	}
	if id.Role == kernel.RoleClient || id.Role == kernel.RoleCarePro {
		writeErr(w, r, http.StatusForbidden, "forbidden", "pro_only")
		return store.Pet{}, authx.Identity{}, false
	}
	if !kernel.IsPracticeStaff(id.Role) && id.Role != kernel.RoleAdmin {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return store.Pet{}, authx.Identity{}, false
	}
	pet, ok := a.requirePetAccess(w, r, petID, id, store.PermWriteNotes)
	if !ok {
		return store.Pet{}, authx.Identity{}, false
	}
	if kernel.IsPracticeStaff(id.Role) && !a.allowPracticePerm(r, id, "pets.write_clinical") {
		writeErr(w, r, http.StatusForbidden, "forbidden", "insufficient_permission")
		return store.Pet{}, authx.Identity{}, false
	}
	return pet, id, true
}

func (a *API) createLabPanel(w http.ResponseWriter, r *http.Request) {
	pet, id, ok := a.requireLabWriteAccess(w, r, chi.URLParam(r, "petID"))
	if !ok {
		return
	}
	var req createLabPanelReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	var collected time.Time
	if req.CollectedAt != nil {
		collected = *req.CollectedAt
	}
	if req.DocumentID != nil && strings.TrimSpace(*req.DocumentID) != "" {
		doc, err := a.store.GetPetDocument(r.Context(), strings.TrimSpace(*req.DocumentID))
		if err != nil || doc.PetID != pet.ID {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_document")
			return
		}
	}
	panel, err := a.store.CreateLabPanel(
		r.Context(), pet.ID, pet.PracticeID, id.UserID,
		collected, req.LabName, req.Notes, req.DocumentID, req.Results,
	)
	a.writeLabPanelErr(w, r, err)
	if err != nil {
		return
	}
	httpx.WriteData(w, http.StatusCreated, panel)
}

func (a *API) listLabPanels(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	if _, ok := a.requirePetAccess(w, r, petID, id, store.PermRead); !ok {
		return
	}
	panels, err := a.store.ListLabPanels(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, panels)
}

func (a *API) getLabPanel(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	panelID := chi.URLParam(r, "panelID")
	if !isUUID(panelID) {
		writeErr(w, r, http.StatusNotFound, "not_found", "lab_panel_not_found")
		return
	}
	if _, ok := a.requirePetAccess(w, r, petID, id, store.PermRead); !ok {
		return
	}
	panel, err := a.store.GetLabPanel(r.Context(), panelID)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "lab_panel_not_found")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if panel.PetID != petID {
		writeErr(w, r, http.StatusNotFound, "not_found", "lab_panel_not_found")
		return
	}
	httpx.WriteData(w, http.StatusOK, panel)
}

func (a *API) patchLabPanel(w http.ResponseWriter, r *http.Request) {
	petID := chi.URLParam(r, "petID")
	panelID := chi.URLParam(r, "panelID")
	if !isUUID(panelID) {
		writeErr(w, r, http.StatusNotFound, "not_found", "lab_panel_not_found")
		return
	}
	if _, _, ok := a.requireLabWriteAccess(w, r, petID); !ok {
		return
	}
	existing, err := a.store.GetLabPanel(r.Context(), panelID)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "lab_panel_not_found")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if existing.PetID != petID {
		writeErr(w, r, http.StatusNotFound, "not_found", "lab_panel_not_found")
		return
	}
	var req patchLabPanelReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if req.Results == nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "results_required")
		return
	}
	if len(*req.Results) == 0 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_lab_panel")
		return
	}
	panel, err := a.store.ReplaceLabPanelResults(r.Context(), panelID, *req.Results)
	a.writeLabPanelErr(w, r, err)
	if err != nil {
		return
	}
	httpx.WriteData(w, http.StatusOK, panel)
}

func (a *API) deleteLabPanel(w http.ResponseWriter, r *http.Request) {
	petID := chi.URLParam(r, "petID")
	panelID := chi.URLParam(r, "panelID")
	if !isUUID(panelID) {
		writeErr(w, r, http.StatusNotFound, "not_found", "lab_panel_not_found")
		return
	}
	if _, _, ok := a.requireLabWriteAccess(w, r, petID); !ok {
		return
	}
	existing, err := a.store.GetLabPanel(r.Context(), panelID)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "lab_panel_not_found")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if existing.PetID != petID {
		writeErr(w, r, http.StatusNotFound, "not_found", "lab_panel_not_found")
		return
	}
	if err := a.store.DeleteLabPanel(r.Context(), panelID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "lab_panel_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"deleted": true})
}

func (a *API) labAnalyteTrend(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	if _, ok := a.requirePetAccess(w, r, petID, id, store.PermRead); !ok {
		return
	}
	analyte := chi.URLParam(r, "analyteCode")
	points, err := a.store.ListLabAnalyteTrend(r.Context(), petID, analyte)
	if errors.Is(err, store.ErrUnknownAnalyte) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "unknown_analyte")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, points)
}

func (a *API) writeLabPanelErr(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, store.ErrUnknownAnalyte) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "unknown_analyte")
		return
	}
	if errors.Is(err, store.ErrDuplicateAnalyte) || errors.Is(err, store.ErrInvalidLabPanel) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_lab_panel")
		return
	}
	writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
}
