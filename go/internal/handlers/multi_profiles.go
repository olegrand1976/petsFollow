package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type registerClientReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"fullName"`
}

type registerCareProReq struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FullName  string `json:"fullName"`
	Specialty string `json:"specialty"`
}

func (a *API) registerClient(w http.ResponseWriter, r *http.Request) {
	var req registerClientReq
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
	locale := localeOf(r)
	result, err := a.store.RegisterClient(r.Context(), store.RegisterClientInput{
		Email: req.Email, Password: req.Password, FullName: req.FullName, PreferredLocale: locale,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	confirmURL := fmt.Sprintf("%s/confirm-email?token=%s", strings.TrimRight(a.cfg.ProPublicSiteURL, "/"), result.Token)
	_ = a.notifier.SendConfirmRegistration(req.Email, locale, req.FullName, confirmURL)
	httpx.WriteData(w, http.StatusCreated, map[string]any{
		"message":     t(r, "success.confirm_email_sent", nil),
		"confirmPath": "/confirm-email?token=" + result.Token,
	})
}

func (a *API) registerCarePro(w http.ResponseWriter, r *http.Request) {
	if !a.cfg.CareProPublicRegister {
		writeErr(w, r, http.StatusForbidden, "forbidden", "registration_disabled")
		return
	}
	var req registerCareProReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	specialty := kernel.ProfessionalSpecialty(strings.TrimSpace(req.Specialty))
	if req.Email == "" || req.Password == "" || req.FullName == "" || !kernel.ValidSpecialty(specialty) {
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
	locale := localeOf(r)
	result, err := a.store.RegisterCarePro(r.Context(), store.RegisterCareProInput{
		Email: req.Email, Password: req.Password, FullName: req.FullName,
		Specialty: specialty, PreferredLocale: locale,
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	confirmURL := fmt.Sprintf("%s/confirm-email?token=%s", strings.TrimRight(a.cfg.ProPublicSiteURL, "/"), result.Token)
	_ = a.notifier.SendConfirmRegistration(req.Email, locale, req.FullName, confirmURL)
	httpx.WriteData(w, http.StatusCreated, map[string]any{
		"message":     t(r, "success.confirm_email_sent", nil),
		"confirmPath": "/confirm-email?token=" + result.Token,
	})
}

func (a *API) linkExistingVetClient(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	err = a.store.LinkExistingClientToVet(r.Context(), id.UserID, clientID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		if errors.Is(err, store.ErrConflict) {
			writeErr(w, r, http.StatusConflict, "conflict", "already_linked")
			return
		}
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "not_a_client")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{"linked": true, "userId": clientID})
}

type shareAccessReq struct {
	Email         string  `json:"email"`
	GranteeUserID string  `json:"granteeUserId"`
	Permission    string  `json:"permission"`
	ExpiresAt     *string `json:"expiresAt"` // RFC3339 ; omit = no expiry
}

func (a *API) listPetShares(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, ok := a.requirePetShareManager(w, r, petID, id)
	if !ok {
		return
	}
	_ = pet
	rows, err := a.store.ListPetAccess(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) createPetShare(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	pet, ok := a.requirePetShareManager(w, r, petID, id)
	if !ok {
		return
	}
	var req shareAccessReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	granteeID, ok := a.resolveGrantee(w, r, id, req)
	if !ok {
		return
	}
	if granteeID == pet.OwnerUserID {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "self_grant")
		return
	}
	perm := strings.TrimSpace(req.Permission)
	if perm == "" {
		perm = string(store.PermWriteNotes)
	}
	if !store.ValidAccessPermission(perm) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_permission")
		return
	}
	expiresAt, ok := a.parseShareExpiresAt(w, r, req.ExpiresAt)
	if !ok {
		return
	}
	grant, err := a.store.GrantPetAccess(r.Context(), petID, granteeID, id.UserID, perm, expiresAt)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, grant)
}

func (a *API) deletePetShare(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	petID := chi.URLParam(r, "petID")
	granteeID := chi.URLParam(r, "granteeID")
	if _, ok := a.requirePetShareManager(w, r, petID, id); !ok {
		return
	}
	if err := a.store.RevokePetAccess(r.Context(), petID, granteeID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]bool{"revoked": true})
}

func (a *API) listClientShares(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	clientID := chi.URLParam(r, "clientID")
	if !a.requireClientShareManager(w, r, clientID, id) {
		return
	}
	rows, err := a.store.ListClientAccess(r.Context(), clientID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) createClientShare(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	clientID := chi.URLParam(r, "clientID")
	if !a.requireClientShareManager(w, r, clientID, id) {
		return
	}
	var req shareAccessReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	granteeID, ok := a.resolveGrantee(w, r, id, req)
	if !ok {
		return
	}
	if granteeID == clientID {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "self_grant")
		return
	}
	perm := strings.TrimSpace(req.Permission)
	if perm == "" {
		perm = string(store.PermWriteNotes)
	}
	if !store.ValidAccessPermission(perm) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_permission")
		return
	}
	expiresAt, ok := a.parseShareExpiresAt(w, r, req.ExpiresAt)
	if !ok {
		return
	}
	grant, err := a.store.GrantClientAccess(r.Context(), clientID, granteeID, id.UserID, perm, expiresAt)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, grant)
}

func (a *API) deleteClientShare(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusUnauthorized, "unauthorized", "login_required")
		return
	}
	clientID := chi.URLParam(r, "clientID")
	granteeID := chi.URLParam(r, "granteeID")
	if !a.requireClientShareManager(w, r, clientID, id) {
		return
	}
	if err := a.store.RevokeClientAccess(r.Context(), clientID, granteeID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]bool{"revoked": true})
}

func (a *API) listPracticeColleagues(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet || id.PracticeID == "" {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	rows, err := a.store.ListPracticeColleagueVets(r.Context(), id.PracticeID, id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) parseShareExpiresAt(w http.ResponseWriter, r *http.Request, raw *string) (*time.Time, bool) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, true
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(*raw))
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_expires_at")
		return nil, false
	}
	if !t.After(time.Now()) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_expires_at")
		return nil, false
	}
	return &t, true
}

func (a *API) resolveGrantee(w http.ResponseWriter, r *http.Request, actor authx.Identity, req shareAccessReq) (string, bool) {
	var u store.User
	var err error
	if req.GranteeUserID != "" {
		u, err = a.store.GetUserByID(r.Context(), req.GranteeUserID)
		if err != nil {
			writeErr(w, r, http.StatusNotFound, "not_found", "user_not_found")
			return "", false
		}
	} else {
		email := strings.TrimSpace(strings.ToLower(req.Email))
		if email == "" {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
			return "", false
		}
		u, err = a.store.GetUserByEmail(r.Context(), email)
		if err != nil {
			writeErr(w, r, http.StatusNotFound, "not_found", "user_not_found")
			return "", false
		}
	}
	if u.ID == actor.UserID {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "self_grant")
		return "", false
	}
	switch u.Role {
	case kernel.RoleCarePro, kernel.RoleVet, kernel.RoleClient:
		return u.ID, true
	default:
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_grantee_role")
		return "", false
	}
}

func (a *API) requirePetShareManager(w http.ResponseWriter, r *http.Request, petID string, id authx.Identity) (store.Pet, bool) {
	pet, err := a.store.GetPet(r.Context(), petID)
	if err != nil {
		writeErr(w, r, http.StatusNotFound, "not_found", "pet_not_found")
		return store.Pet{}, false
	}
	switch id.Role {
	case kernel.RoleClient:
		if pet.OwnerUserID != id.UserID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "not_your_pet")
			return store.Pet{}, false
		}
	case kernel.RoleVet:
		if pet.PracticeID != id.PracticeID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "wrong_practice")
			return store.Pet{}, false
		}
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return store.Pet{}, false
	}
	return pet, true
}

func (a *API) requireClientShareManager(w http.ResponseWriter, r *http.Request, clientID string, id authx.Identity) bool {
	switch id.Role {
	case kernel.RoleClient:
		if clientID != id.UserID {
			writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
			return false
		}
	case kernel.RoleVet:
		// vet must have practice_clients link
		clients, err := a.store.ListClientsByPractice(r.Context(), id.PracticeID)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return false
		}
		found := false
		for _, c := range clients {
			if c.UserID == clientID {
				found = true
				break
			}
		}
		if !found {
			writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
			return false
		}
	default:
		writeErr(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return false
	}
	return true
}

func (a *API) listCareProClients(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCarePro(w, r)
	if !ok {
		return
	}
	rows, err := a.store.ListCareProClients(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) listCareProPets(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCarePro(w, r)
	if !ok {
		return
	}
	rows, err := a.store.ListCareProAccessiblePets(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) listCareProVisits(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCarePro(w, r)
	if !ok {
		return
	}
	rows, err := a.store.ListCareProVisits(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) requireCarePro(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleCarePro {
		writeErr(w, r, http.StatusForbidden, "forbidden", "care_pro_only")
		return authx.Identity{}, false
	}
	return id, true
}
