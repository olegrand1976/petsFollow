package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func (a *API) listVetSiteRooms(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	includeInactive := r.URL.Query().Get("includeInactive") == "1" || r.URL.Query().Get("includeInactive") == "true"
	items, err := a.store.ListRooms(r.Context(), id.PracticeID, chi.URLParam(r, "siteID"), includeInactive)
	if err != nil {
		if errors.Is(err, store.ErrValidation) || errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_site")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, items)
}

type createRoomReq struct {
	Name      string `json:"name"`
	SortOrder *int   `json:"sortOrder"`
}

func (a *API) createVetSiteRoom(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	var req createRoomReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	room, err := a.store.CreateRoom(r.Context(), id.PracticeID, chi.URLParam(r, "siteID"), store.CreateRoomInput{
		Name: req.Name, SortOrder: req.SortOrder,
	})
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			code := "invalid_room"
			msg := err.Error()
			switch {
			case strings.Contains(msg, "name_required"):
				code = "name_required"
			case strings.Contains(msg, "room_name_taken"):
				code = "room_name_taken"
			case strings.Contains(msg, "site_inactive"), strings.Contains(msg, "invalid_site"):
				code = "invalid_site"
			}
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_site")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, room)
}

type patchRoomReq struct {
	Name      *string `json:"name"`
	Active    *bool   `json:"active"`
	SortOrder *int    `json:"sortOrder"`
}

func (a *API) patchVetSiteRoom(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	var req patchRoomReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	room, err := a.store.PatchRoom(r.Context(), id.PracticeID, chi.URLParam(r, "siteID"), chi.URLParam(r, "roomID"), store.PatchRoomInput{
		Name: req.Name, Active: req.Active, SortOrder: req.SortOrder,
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "room_not_found")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			code := "invalid_room"
			msg := err.Error()
			switch {
			case strings.Contains(msg, "name_required"):
				code = "name_required"
			case strings.Contains(msg, "room_name_taken"):
				code = "room_name_taken"
			}
			writeErr(w, r, http.StatusBadRequest, "bad_request", code)
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, room)
}

func (a *API) deactivateVetSiteRoom(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "calendar.manage")
	if !ok {
		return
	}
	room, err := a.store.DeactivateRoom(r.Context(), id.PracticeID, chi.URLParam(r, "siteID"), chi.URLParam(r, "roomID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "room_not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, room)
}
