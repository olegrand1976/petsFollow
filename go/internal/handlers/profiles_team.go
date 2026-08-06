package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) registerProfileRoutes(r chi.Router) {
	r.Get("/me/profiles", a.listMyProfiles)
	r.Post("/me/profiles/switch", a.switchMyProfile)
	r.Get("/me/feature-modules", a.getFeatureModules)
	r.Patch("/me/feature-modules", a.patchFeatureModules)
	r.Get("/vet/team", a.listVetTeam)
	r.Post("/vet/team", a.inviteVetTeam)
	r.Patch("/vet/team/{id}", a.patchVetTeam)
	r.Delete("/vet/team/{id}", a.revokeVetTeam)
	r.Post("/admin/users/{id}/profiles", a.adminAttachProfile)
	r.Post("/commercial/users/{id}/profiles", a.commercialAttachProfile)
}

// requirePracticePerm gates practice-staff endpoints by team permission matrix.
func (a *API) requirePracticePerm(w http.ResponseWriter, r *http.Request, capability string) (authx.Identity, bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return id, false
	}
	if !a.checkPracticePerm(w, r, id, capability) {
		return id, false
	}
	return id, true
}

// checkPracticePerm writes 403/500 and returns false when the identity lacks the capability.
func (a *API) checkPracticePerm(w http.ResponseWriter, r *http.Request, id authx.Identity, capability string) bool {
	if id.PracticeID == "" || !kernel.IsPracticeStaff(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return false
	}
	ok, err := a.store.TeamPermission(r.Context(), id.PracticeID, id.UserID, capability)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return false
	}
	if !ok {
		writeErr(w, r, http.StatusForbidden, "forbidden", "insufficient_permission")
		return false
	}
	return true
}

// allowPracticePerm is a silent capability check (no HTTP write) for mixed-role handlers.
func (a *API) allowPracticePerm(r *http.Request, id authx.Identity, capability string) bool {
	if id.PracticeID == "" || !kernel.IsPracticeStaff(id.Role) {
		return false
	}
	ok, err := a.store.TeamPermission(r.Context(), id.PracticeID, id.UserID, capability)
	return err == nil && ok
}

// requireAnyPracticePerm allows access when the staff member has at least one of the capabilities.
func (a *API) requireAnyPracticePerm(w http.ResponseWriter, r *http.Request, capabilities ...string) (authx.Identity, bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return id, false
	}
	if id.PracticeID == "" || !kernel.IsPracticeStaff(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return id, false
	}
	for _, cap := range capabilities {
		ok, err := a.store.TeamPermission(r.Context(), id.PracticeID, id.UserID, cap)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return id, false
		}
		if ok {
			return id, true
		}
	}
	writeErr(w, r, http.StatusForbidden, "forbidden", "insufficient_permission")
	return id, false
}

// resolvePracticeInviteOwner returns the user ID that owns the practice app-invite code.
// Practice staff need clients.write; reference vet uses self, otherwise practices.reference_vet_user_id.
func (a *API) resolvePracticeInviteOwner(w http.ResponseWriter, r *http.Request, id authx.Identity) (string, bool) {
	if !kernel.IsPracticeStaff(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "invite_role_denied")
		return "", false
	}
	if !a.checkPracticePerm(w, r, id, "clients.write") {
		return "", false
	}
	ok, err := a.store.IsReferenceVet(r.Context(), id.PracticeID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return "", false
	}
	if ok {
		return id.UserID, true
	}
	ref, err := a.store.PracticeReferenceVetUserID(r.Context(), id.PracticeID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return "", false
	}
	if ref == "" {
		return id.UserID, true
	}
	return ref, true
}

func (a *API) listMyProfiles(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	profiles, err := a.store.ListProfiles(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, profiles)
}

func (a *API) switchMyProfile(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var req struct {
		ProfileID string `json:"profileId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.ProfileID) == "" {
		writeErr(w, r, http.StatusBadRequest, "validation", "validation")
		return
	}
	p, err := a.store.SwitchProfile(r.Context(), id.UserID, req.ProfileID)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
		return
	}
	if errors.Is(err, store.ErrForbidden) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "insufficient_permission")
		return
	}
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	u, err := a.store.GetUserByID(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	pair, err := a.tokens.IssueProfile(u.ID, u.Email, u.Role, u.PracticeID, p.ID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"profile":      p,
		"accessToken":  pair.AccessToken,
		"refreshToken": pair.RefreshToken,
		"expiresIn":    pair.ExpiresIn,
	})
}

func (a *API) attachProfileFor(w http.ResponseWriter, r *http.Request, actor authx.Identity) {
	targetID := chi.URLParam(r, "id")
	var req struct {
		Role       string `json:"role"`
		Specialty  string `json:"specialty"`
		PracticeID string `json:"practiceId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "validation", "validation")
		return
	}
	role := kernel.Role(req.Role)
	// Research observatory access is admin-only (not commercial attach).
	if role == kernel.RoleResearch && actor.Role != kernel.RoleAdmin {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	p, err := a.store.AttachProfile(r.Context(), actor.UserID, targetID, store.AttachProfileInput{
		Role:       role,
		Specialty:  req.Specialty,
		PracticeID: req.PracticeID,
	})
	switch {
	case errors.Is(err, store.ErrCannotModifyOwnProfiles):
		writeErr(w, r, http.StatusForbidden, "cannot_modify_own_profiles", "cannot_modify_own_profiles")
	case errors.Is(err, store.ErrProfileExists):
		writeErr(w, r, http.StatusConflict, "profile_exists", "profile_exists")
	case errors.Is(err, store.ErrValidation):
		writeErr(w, r, http.StatusBadRequest, "validation", "validation")
	case err != nil:
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
	default:
		httpx.WriteData(w, http.StatusCreated, p)
	}
}

func (a *API) adminAttachProfile(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	a.attachProfileFor(w, r, id)
}

func (a *API) commercialAttachProfile(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || !kernel.IsSalesForce(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	a.attachProfileFor(w, r, id)
}

func (a *API) listVetTeam(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.PracticeID == "" || !kernel.IsPracticeStaff(id.Role) {
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	members, err := a.store.ListTeamMembers(r.Context(), id.PracticeID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	deskIdle := kernel.DefaultDeskIdleMinutes
	if m, err := a.store.GetPracticeDeskIdleMinutes(r.Context(), id.PracticeID); err == nil {
		deskIdle = m
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"members":         members,
		"deskIdleMinutes": deskIdle,
	})
}

func (a *API) inviteVetTeam(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "team.manage")
	if !ok {
		return
	}
	var req struct {
		Email       string          `json:"email"`
		FullName    string          `json:"fullName"`
		Password    string          `json:"password"`
		TeamRole    store.TeamRole  `json:"teamRole"`
		Permissions map[string]bool `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "validation", "validation")
		return
	}
	m, err := a.store.InviteTeamMember(r.Context(), id.PracticeID, id.UserID, store.InviteTeamMemberInput{
		Email: req.Email, FullName: req.FullName, Password: req.Password,
		TeamRole: req.TeamRole, Permissions: req.Permissions,
	})
	switch {
	case errors.Is(err, store.ErrForbidden):
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
	case errors.Is(err, store.ErrValidation):
		writeErr(w, r, http.StatusBadRequest, "validation", "validation")
	case err != nil:
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
	default:
		httpx.WriteData(w, http.StatusCreated, m)
	}
}

func (a *API) patchVetTeam(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "team.manage")
	if !ok {
		return
	}
	memberID := chi.URLParam(r, "id")
	var req struct {
		TeamRole      *store.TeamRole `json:"teamRole"`
		Permissions   map[string]bool `json:"permissions"`
		DefaultSiteID *string         `json:"defaultSiteId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "validation", "validation")
		return
	}
	m, err := a.store.UpdateTeamMember(r.Context(), id.PracticeID, id.UserID, memberID, req.TeamRole, req.Permissions, req.DefaultSiteID)
	switch {
	case errors.Is(err, store.ErrForbidden):
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
	case errors.Is(err, store.ErrNotFound):
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
	case errors.Is(err, store.ErrValidation):
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_site")
	case err != nil:
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
	default:
		httpx.WriteData(w, http.StatusOK, m)
	}
}

func (a *API) revokeVetTeam(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "team.manage")
	if !ok {
		return
	}
	err := a.store.RevokeTeamMember(r.Context(), id.PracticeID, id.UserID, chi.URLParam(r, "id"))
	switch {
	case errors.Is(err, store.ErrForbidden):
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
	case errors.Is(err, store.ErrNotFound):
		writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
	case err != nil:
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (a *API) getFeatureModules(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	m, err := a.store.GetFeatureModules(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, m)
}

func (a *API) patchFeatureModules(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var req store.FeatureModules
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "validation", "validation")
		return
	}
	req.UserID = id.UserID
	m, err := a.store.UpdateFeatureModules(r.Context(), id.UserID, req)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, m)
}
