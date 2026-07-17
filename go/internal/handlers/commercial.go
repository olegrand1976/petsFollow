package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

func (a *API) registerCommercialRoutes(r chi.Router) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(a.tokens))
		pr.Use(a.localeFromUserMiddleware)
		pr.Get("/commercial/overview", a.commercialOverview)
		pr.Get("/commercial/vets", a.commercialListVets)
		pr.Post("/commercial/vets", a.commercialEncodeVet)
		pr.Get("/commercial/commissions", a.commercialCommissions)
		pr.Get("/commercial/me/payout-profile", a.commercialGetPayoutProfile)
		pr.Patch("/commercial/me/payout-profile", a.commercialPatchPayoutProfile)
		pr.Get("/commercial/prospects", a.commercialListProspects)
		pr.Post("/commercial/prospects", a.commercialCreateProspect)
		pr.Patch("/commercial/prospects/{id}", a.commercialUpdateProspect)
		pr.Delete("/commercial/prospects/{id}", a.commercialDeleteProspect)
	})
}

func (a *API) requireCommercial(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleCommercial {
		writeErr(w, r, http.StatusForbidden, "forbidden", "commercial_only")
		return authx.Identity{}, false
	}
	return id, true
}

func (a *API) commercialOverview(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	overview, err := a.store.CommercialOverview(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, overview)
}

func (a *API) commercialListVets(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	vets, err := a.store.ListCommercialVets(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, vets)
}

type encodeVetReq struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	FullName     string `json:"fullName"`
	PracticeName string `json:"practiceName"`
	Phone        string `json:"phone"`
	City         string `json:"city"`
	PostalCode   string `json:"postalCode"`
	AddressLine1 string `json:"addressLine1"`
	ContactEmail string `json:"contactEmail"`
}

func (a *API) commercialEncodeVet(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	var req encodeVetReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" || req.FullName == "" || req.PracticeName == "" {
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
	vetID, err := a.store.EncodeVetForCommercial(r.Context(), id.UserID, store.EncodeVetInput{
		Email:            req.Email,
		Password:         req.Password,
		FullName:         req.FullName,
		PracticeName:     req.PracticeName,
		Phone:            req.Phone,
		City:             req.City,
		PostalCode:       req.PostalCode,
		AddressLine1:     req.AddressLine1,
		ContactEmail:     req.ContactEmail,
		PreferredLocale:  localeOf(r),
		AutoReplyDefault: t(r, "defaults.auto_reply_unavailable", nil),
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, map[string]string{"userId": vetID, "email": req.Email})
}

func (a *API) commercialCommissions(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	summary, err := a.store.GetCommercialCommissionSummary(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, summary)
}

func (a *API) commercialGetPayoutProfile(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	profile, err := a.store.GetCommercialPayoutProfile(r.Context(), id.UserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, profile)
}

type payoutProfileReq struct {
	IBAN          string `json:"iban"`
	BIC           string `json:"bic"`
	AccountHolder string `json:"accountHolder"`
}

func (a *API) commercialPatchPayoutProfile(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	var req payoutProfileReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	iban := normalizeIBAN(req.IBAN)
	if iban != "" && !validIBAN(iban) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_iban")
		return
	}
	holder := strings.TrimSpace(req.AccountHolder)
	if len(holder) > 120 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_account_holder")
		return
	}
	bic := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(req.BIC), " ", ""))
	if !validBIC(bic) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_bic")
		return
	}
	profile := store.CommercialPayoutProfile{IBAN: iban, BIC: bic, AccountHolder: holder}
	if err := a.store.UpdateCommercialPayoutProfile(r.Context(), id.UserID, profile); err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, profile)
}

func (a *API) commercialListProspects(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	status := r.URL.Query().Get("status")
	if status != "" && !store.ValidProspectStatus(status) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
		return
	}
	prospects, err := a.store.ListProspects(r.Context(), id.UserID, status)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, prospects)
}

type prospectReq struct {
	PracticeName string `json:"practiceName"`
	ContactName  string `json:"contactName"`
	ContactEmail string `json:"contactEmail"`
	ContactPhone string `json:"contactPhone"`
	City         string `json:"city"`
	Notes        string `json:"notes"`
	Status       string `json:"status"`
}

func (a *API) commercialCreateProspect(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	var req prospectReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if strings.TrimSpace(req.PracticeName) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	if req.Status != "" && !store.ValidProspectStatus(req.Status) {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
		return
	}
	prospect, err := a.store.CreateProspect(r.Context(), id.UserID, store.ProspectInput{
		PracticeName: req.PracticeName,
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		City:         req.City,
		Notes:        req.Notes,
		Status:       req.Status,
		Source:       "commercial",
	})
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, prospect)
}

func (a *API) commercialUpdateProspect(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	prospectID := chi.URLParam(r, "id")
	existing, err := a.store.GetProspect(r.Context(), id.UserID, prospectID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	var req prospectReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	in := store.ProspectInput{
		PracticeName: existing.PracticeName,
		ContactName:  existing.ContactName,
		ContactEmail: existing.ContactEmail,
		ContactPhone: existing.ContactPhone,
		City:         existing.City,
		Notes:        existing.Notes,
		Status:       existing.Status,
	}
	if req.PracticeName != "" {
		in.PracticeName = req.PracticeName
	}
	if req.ContactName != "" {
		in.ContactName = req.ContactName
	}
	if req.ContactEmail != "" {
		in.ContactEmail = req.ContactEmail
	}
	if req.ContactPhone != "" {
		in.ContactPhone = req.ContactPhone
	}
	if req.City != "" {
		in.City = req.City
	}
	if req.Notes != "" {
		in.Notes = req.Notes
	}
	if req.Status != "" {
		if !store.ValidProspectStatus(req.Status) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_status")
			return
		}
		in.Status = req.Status
	}
	prospect, err := a.store.UpdateProspect(r.Context(), id.UserID, prospectID, in)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, prospect)
}

func (a *API) commercialDeleteProspect(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	if err := a.store.DeleteProspect(r.Context(), id.UserID, chi.URLParam(r, "id")); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]string{"status": "deleted"})
}

type vetProspectReq struct {
	PracticeName string `json:"practiceName"`
	ContactName  string `json:"contactName"`
	ContactEmail string `json:"contactEmail"`
	ContactPhone string `json:"contactPhone"`
	City         string `json:"city"`
	Notes        string `json:"notes"`
}

func (a *API) vetCreateProspect(w http.ResponseWriter, r *http.Request) {
	id, err := authx.FromContext(r.Context())
	if err != nil || id.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_only")
		return
	}
	var req vetProspectReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return
	}
	if strings.TrimSpace(req.PracticeName) == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	prospect, err := a.store.CreateVetReferralProspect(r.Context(), id.UserID, store.ProspectInput{
		PracticeName: req.PracticeName,
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		City:         req.City,
		Notes:        req.Notes,
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "no_commercial_assigned")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, prospect)
}
