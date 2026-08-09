package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

type createClientReq struct {
	Email                  string `json:"email"`
	Password               string `json:"password"`
	FullName               string `json:"fullName"`
	FirstName              string `json:"firstName"`
	LastName               string `json:"lastName"`
	ContactPhone           string `json:"contactPhone"`
	Address                string `json:"address"`
	NationalRegistryNumber string `json:"nationalRegistryNumber"`
	BillingVATNumber       string `json:"billingVatNumber"`
	BillingCompanyNumber   string `json:"billingCompanyNumber"`
	BillingStreet          string `json:"billingStreet"`
	BillingCity            string `json:"billingCity"`
	BillingPostal          string `json:"billingPostal"`
	BillingCountry         string `json:"billingCountry"`
	BillingCustomerKind    string `json:"billingCustomerKind"`
	VetUserID              string `json:"vetUserId"`
}

type createVetAdminReq struct {
	Email                string `json:"email"`
	Password             string `json:"password"`
	FullName             string `json:"fullName"`
	PracticeName         string `json:"practiceName"`
	Phone                string `json:"phone"`
	City                 string `json:"city"`
	PostalCode           string `json:"postalCode"`
	AddressLine1         string `json:"addressLine1"`
	ContactEmail         string `json:"contactEmail"`
	AssignedCommercialID string `json:"assignedCommercialId"`
}

func (a *API) createVetClient(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requirePracticePerm(w, r, "clients.write")
	if !ok {
		return
	}
	req, ok := a.decodeCreateClient(w, r)
	if !ok {
		return
	}
	clientID, err := a.store.CreateClientForVet(r.Context(), id.UserID, createClientInputFromReq(req, r))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, map[string]string{"userId": clientID, "email": req.Email})
}

func createClientInputFromReq(req createClientReq, r *http.Request) store.CreateClientInput {
	return store.CreateClientInput{
		Email:                  req.Email,
		Password:               req.Password,
		FullName:               req.FullName,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		ContactPhone:           req.ContactPhone,
		Address:                req.Address,
		NationalRegistryNumber: req.NationalRegistryNumber,
		BillingVATNumber:       req.BillingVATNumber,
		BillingCompanyNumber:   req.BillingCompanyNumber,
		BillingStreet:          req.BillingStreet,
		BillingCity:            req.BillingCity,
		BillingPostal:          req.BillingPostal,
		BillingCountry:         req.BillingCountry,
		BillingCustomerKind:    req.BillingCustomerKind,
		Locale:                 localeOf(r),
	}
}

func (a *API) commercialCreateClient(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireCommercial(w, r)
	if !ok {
		return
	}
	req, ok := a.decodeCreateClient(w, r)
	if !ok {
		return
	}
	in := createClientInputFromReq(req, r)

	// Optional vet link: empty vetUserId → standalone client (no practice).
	if req.VetUserID == "" {
		clientID, err := a.store.CreateClientStandalone(r.Context(), in)
		if err != nil {
			writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
			return
		}
		// Attribute standalone client to the creating commercial (QR/nearby table).
		a.tryLinkCommercialReferral(r, clientID, id.UserID)
		httpx.WriteData(w, http.StatusCreated, map[string]string{"userId": clientID, "email": req.Email})
		return
	}

	owns, err := a.store.CommercialOwnsVet(r.Context(), id.UserID, req.VetUserID)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	if !owns {
		writeErr(w, r, http.StatusForbidden, "forbidden", "vet_not_assigned")
		return
	}
	clientID, err := a.store.CreateClientForVet(r.Context(), req.VetUserID, in)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "not_found")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, map[string]string{"userId": clientID, "email": req.Email})
}

func (a *API) adminCreateClient(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	req, ok := a.decodeCreateClient(w, r)
	if !ok {
		return
	}
	if req.VetUserID == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return
	}
	vet, err := a.store.GetUserByID(r.Context(), req.VetUserID)
	if err != nil || vet.Role != kernel.RoleVet {
		writeErr(w, r, http.StatusNotFound, "not_found", "vet_not_found")
		return
	}
	clientID, err := a.store.CreateClientForVet(r.Context(), req.VetUserID, createClientInputFromReq(req, r))
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, map[string]string{"userId": clientID, "email": req.Email})
}

type createCareProAdminReq struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FullName  string `json:"fullName"`
	Specialty string `json:"specialty"`
}

func (a *API) adminCreateCarePro(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	var req createCareProAdminReq
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
	userID, err := a.store.CreateCareProAsAdmin(r.Context(), req.Email, req.Password, req.FullName, specialty, localeOf(r))
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, map[string]string{
		"userId": userID, "email": req.Email, "role": "care_pro", "specialty": string(specialty),
	})
}

func (a *API) adminCreateVet(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	var req createVetAdminReq
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
	vetID, err := a.store.CreateVetAsAdmin(r.Context(), store.EncodeVetInput{
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
	}, strings.TrimSpace(req.AssignedCommercialID))
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusCreated, map[string]string{"userId": vetID, "email": req.Email})
}

func (a *API) adminListVets(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	rows, err := a.store.ListVetsForAdmin(r.Context())
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return
	}
	httpx.WriteData(w, http.StatusOK, rows)
}

func (a *API) decodeCreateClient(w http.ResponseWriter, r *http.Request) (createClientReq, bool) {
	var req createClientReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_json")
		return req, false
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.FullName = strings.TrimSpace(req.FullName)
	if req.FirstName != "" || req.LastName != "" {
		req.FullName = strings.TrimSpace(req.FirstName + " " + req.LastName)
	}
	phone, phoneCode := normalizeContactPhone(req.ContactPhone, false)
	if phoneCode != "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", phoneCode)
		return req, false
	}
	req.ContactPhone = phone
	addr, addrCode := normalizeClientAddress(req.Address)
	if addrCode != "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", addrCode)
		return req, false
	}
	req.Address = addr
	niss, nissCode := normalizeNationalRegistry(req.NationalRegistryNumber)
	if nissCode != "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", nissCode)
		return req, false
	}
	req.NationalRegistryNumber = niss
	vat, vatCode := normalizeBillingField(req.BillingVATNumber, billingFieldMaxRunes)
	if vatCode != "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", vatCode)
		return req, false
	}
	req.BillingVATNumber = vat
	company, companyCode := normalizeBillingField(req.BillingCompanyNumber, billingFieldMaxRunes)
	if companyCode != "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", companyCode)
		return req, false
	}
	req.BillingCompanyNumber = company
	street, streetCode := normalizeBillingField(req.BillingStreet, billingFieldMaxRunes)
	if streetCode != "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", streetCode)
		return req, false
	}
	req.BillingStreet = street
	city, cityCode := normalizeBillingField(req.BillingCity, billingFieldMaxRunes)
	if cityCode != "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", cityCode)
		return req, false
	}
	req.BillingCity = city
	postal, postalCode := normalizeBillingField(req.BillingPostal, billingFieldMaxRunes)
	if postalCode != "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", postalCode)
		return req, false
	}
	req.BillingPostal = postal
	country, countryCode := normalizeBillingCountry(req.BillingCountry)
	if countryCode != "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", countryCode)
		return req, false
	}
	req.BillingCountry = country
	kind, kindCode := normalizeBillingCustomerKind(req.BillingCustomerKind)
	if kindCode != "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", kindCode)
		return req, false
	}
	req.BillingCustomerKind = kind
	// Password optional: empty → server generates an opaque temp (never shown to the vet).
	if req.Email == "" || req.FullName == "" {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "fields_required")
		return req, false
	}
	if req.Password != "" && len(req.Password) < 8 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "password_too_short")
		return req, false
	}
	if _, err := a.store.GetUserByEmail(r.Context(), req.Email); err == nil {
		details := map[string]any{"exists": true, "email": req.Email}
		if id, idErr := authx.FromContext(r.Context()); idErr == nil && a.allowPracticePerm(r, id, "clients.write") {
			if d, lerr := a.store.LookupClientConflict(r.Context(), req.Email, id.UserID); lerr == nil {
				details = d
			}
		}
		writeErrDetails(w, r, http.StatusConflict, "conflict", "email_already_exists", details)
		return req, false
	} else if !errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return req, false
	}
	return req, true
}
