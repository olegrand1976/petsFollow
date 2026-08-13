package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/eid"
	"github.com/olegrand1976/petsFollow/go/internal/platform/authx"
	"github.com/olegrand1976/petsFollow/go/internal/platform/httpx"
	"github.com/olegrand1976/petsFollow/go/internal/store"
	"github.com/redis/go-redis/v9"
)

const (
	eidNonceTTL       = 5 * time.Minute
	eidMaxUploadBytes = eid.MaxViewerBytes
)

var (
	eidMemNonce sync.Map // userID → eidMemNonceEntry
)

type eidMemNonceEntry struct {
	Nonce     string
	ExpiresAt time.Time
}

func (a *API) requireEidBE(w http.ResponseWriter, r *http.Request) (authx.Identity, bool) {
	if !a.cfg.EidEnabled {
		writeErr(w, r, http.StatusNotFound, "not_found", "eid_disabled")
		return authx.Identity{}, false
	}
	id, ok := a.requirePracticePerm(w, r, "clients.write")
	if !ok {
		return authx.Identity{}, false
	}
	country, err := a.store.GetPracticeCountryCode(r.Context(), id.PracticeID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return authx.Identity{}, false
	}
	if store.NormalizeCountryCode(country) != "BE" {
		writeErr(w, r, http.StatusNotFound, "not_found", "eid_not_available")
		return authx.Identity{}, false
	}
	return id, true
}

func (a *API) eidSiteOrigin() string {
	if o := strings.TrimSpace(a.cfg.EidSiteOrigin); o != "" {
		return eid.SiteOrigin(o)
	}
	return eid.SiteOrigin(a.cfg.ProPublicSiteURL)
}

func (a *API) eidAuditHashSecret() string {
	return a.cfg.JWTSigningKey
}

func (a *API) recordEidReading(ctx context.Context, id authx.Identity, tool string, success bool, identity eid.Identity, errCode string) {
	hash := ""
	if identity.NISS != "" {
		hash = eid.HashNISS(identity.NISS, a.eidAuditHashSecret())
	}
	fields := identity.NonEmptyFieldKeys()
	if err := a.store.InsertEidReading(ctx, store.EidReading{
		PracticeID: id.PracticeID,
		UserID:     id.UserID,
		Tool:       tool,
		Success:    success,
		FieldsRead: fields,
		NISSHash:   hash,
		ErrorCode:  errCode,
	}); err != nil {
		log.Printf("eid_reading insert: %v", err)
	}
}

func (a *API) storeEidNonce(ctx context.Context, userID, nonce string) error {
	key := "eid:webeid:nonce:" + userID
	if a.redis != nil {
		return a.redis.Set(ctx, key, nonce, eidNonceTTL)
	}
	if !a.cfg.AllowsInMemoryEidNonce() {
		return errEidRedisRequired
	}
	eidMemNonce.Store(userID, eidMemNonceEntry{Nonce: nonce, ExpiresAt: time.Now().Add(eidNonceTTL)})
	return nil
}

var errEidRedisRequired = errors.New("eid_redis_required")

func (a *API) takeEidNonce(ctx context.Context, userID string) (string, bool) {
	key := "eid:webeid:nonce:" + userID
	if a.redis != nil {
		v, err := a.redis.GetDel(ctx, key)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				log.Printf("eid nonce getdel: %v", err)
			}
			return "", false
		}
		return v, v != ""
	}
	if !a.cfg.AllowsInMemoryEidNonce() {
		return "", false
	}
	raw, ok := eidMemNonce.LoadAndDelete(userID)
	if !ok {
		return "", false
	}
	entry, ok := raw.(eidMemNonceEntry)
	if !ok || time.Now().After(entry.ExpiresAt) {
		return "", false
	}
	return entry.Nonce, true
}

// importVetEidViewer POST /vet/eid/import — multipart file (.eid/.xml).
func (a *API) importVetEidViewer(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireEidBE(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(eidMaxUploadBytes + (1 << 20)); err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "invalid_multipart")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_required")
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(io.LimitReader(file, eidMaxUploadBytes+1))
	if err != nil {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "file_read_failed")
		return
	}
	filename := ""
	if header != nil {
		filename = filepath.Base(header.Filename)
	}
	identity, err := eid.ParseViewerExport(raw, filename)
	if err != nil {
		code := "eid_parse_failed"
		switch {
		case errors.Is(err, eid.ErrEmptyFile):
			code = "eid_file_empty"
		case errors.Is(err, eid.ErrFileTooLarge):
			code = "eid_file_too_large"
		case errors.Is(err, eid.ErrPDFUnsupported):
			code = "eid_pdf_unsupported"
		case errors.Is(err, eid.ErrUnrecognized):
			code = "eid_format_unrecognized"
		case errors.Is(err, eid.ErrInvalidXML):
			code = "eid_xml_invalid"
		}
		a.recordEidReading(r.Context(), id, "eid_viewer_xml", false, eid.Identity{}, code)
		writeErr(w, r, http.StatusUnprocessableEntity, "validation_error", code)
		return
	}
	if !identity.HasUsefulIdentity() {
		a.recordEidReading(r.Context(), id, identity.ImportTool, false, identity, "eid_identity_empty")
		writeErr(w, r, http.StatusUnprocessableEntity, "validation_error", "eid_identity_empty")
		return
	}
	a.recordEidReading(r.Context(), id, identity.ImportTool, true, identity, "")
	httpx.WriteData(w, http.StatusOK, identity.PublicPrefill())
}

// challengeVetWebEid GET /vet/eid/web-eid/challenge
func (a *API) challengeVetWebEid(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireEidBE(w, r)
	if !ok {
		return
	}
	nonce, err := eid.GenerateChallengeNonce()
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "eid_nonce_failed")
		return
	}
	if err := a.storeEidNonce(r.Context(), id.UserID, nonce); err != nil {
		if errors.Is(err, errEidRedisRequired) {
			writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "eid_redis_required")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "eid_nonce_store_failed")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"nonce":  nonce,
		"origin": a.eidSiteOrigin(),
	})
}

type webEidVerifyReq struct {
	Token json.RawMessage `json:"token"`
}

// verifyVetWebEid POST /vet/eid/web-eid/verify
func (a *API) verifyVetWebEid(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireEidBE(w, r)
	if !ok {
		return
	}
	var req webEidVerifyReq
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil || len(req.Token) == 0 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "token_required")
		return
	}
	nonce, okNonce := a.takeEidNonce(r.Context(), id.UserID)
	if !okNonce {
		a.recordEidReading(r.Context(), id, "web_eid", false, eid.Identity{}, "eid_nonce_expired")
		writeErr(w, r, http.StatusUnprocessableEntity, "validation_error", "eid_nonce_expired")
		return
	}
	cas, err := eid.LoadTrustedCAs()
	if err != nil || len(cas) == 0 {
		a.recordEidReading(r.Context(), id, "web_eid", false, eid.Identity{}, "eid_ca_certs_missing")
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "eid_ca_certs_missing")
		return
	}
	validator, err := eid.NewAuthTokenValidator(a.eidSiteOrigin(), cas, a.cfg.WebEidDisableOCSP)
	if err != nil {
		a.recordEidReading(r.Context(), id, "web_eid", false, eid.Identity{}, "eid_validator_failed")
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "eid_validator_failed")
		return
	}
	identity, err := eid.VerifyAuthToken(r.Context(), validator, req.Token, nonce)
	if err != nil {
		code := "eid_token_invalid"
		msg := err.Error()
		switch {
		case strings.Contains(msg, "eid_token_parse"):
			code = "eid_token_parse"
		case strings.Contains(msg, "eid_identity_empty"):
			code = "eid_identity_empty"
		}
		a.recordEidReading(r.Context(), id, "web_eid", false, eid.Identity{}, code)
		writeErr(w, r, http.StatusUnprocessableEntity, "validation_error", code)
		return
	}
	a.recordEidReading(r.Context(), id, "web_eid", true, identity, "")
	httpx.WriteData(w, http.StatusOK, identity.PublicPrefill())
}
