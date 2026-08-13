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
	Origin    string
	ExpiresAt time.Time
}

type eidChallenge struct {
	Nonce  string
	Origin string
}

func (a *API) requireEidBE(w http.ResponseWriter, r *http.Request, rlKind string) (authx.Identity, bool) {
	if !a.cfg.EidEnabled {
		writeErr(w, r, http.StatusNotFound, "not_found", "eid_disabled")
		return authx.Identity{}, false
	}
	id, ok := a.requirePracticePerm(w, r, "clients.write")
	if !ok {
		return authx.Identity{}, false
	}
	country, err := a.store.GetPracticeCountryCode(r.Context(), id.PracticeID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, r, http.StatusNotFound, "not_found", "eid_not_available")
			return authx.Identity{}, false
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "internal")
		return authx.Identity{}, false
	}
	if store.NormalizeCountryCode(country) != "BE" {
		writeErr(w, r, http.StatusNotFound, "not_found", "eid_not_available")
		return authx.Identity{}, false
	}
	kind := strings.TrimSpace(rlKind)
	if kind == "" {
		kind = "any"
	}
	if a.eidRL != nil && !a.eidRL.Allow("eid:"+kind+":"+id.UserID) {
		writeErr(w, r, http.StatusTooManyRequests, "rate_limited", "too_many_requests")
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

func (a *API) resolveEidOrigin(r *http.Request) string {
	// BFF forwards browser Origin as X-PF-Web-Eid-Origin (Origin is often stripped on
	// proxy). Trust that header only when X-PF-Proxy-Secret matches — same gate as
	// X-PF-Client-IP — so a direct API caller cannot spoof the challenge origin.
	reqOrigin := ""
	if secretHeaderOK(r, httpx.HeaderPFProxySecret, a.cfg.BFFProxySecret) {
		reqOrigin = r.Header.Get("X-PF-Web-Eid-Origin")
	}
	if reqOrigin == "" {
		reqOrigin = r.Header.Get("Origin")
	}
	return eid.ResolveChallengeOrigin(a.eidSiteOrigin(), reqOrigin)
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

func (a *API) storeEidChallenge(ctx context.Context, userID string, ch eidChallenge) error {
	key := "eid:webeid:nonce:" + userID
	payload, err := json.Marshal(ch)
	if err != nil {
		return err
	}
	if a.redis != nil {
		return a.redis.Set(ctx, key, string(payload), eidNonceTTL)
	}
	if !a.cfg.AllowsInMemoryEidNonce() {
		return errEidRedisRequired
	}
	eidMemNonce.Store(userID, eidMemNonceEntry{
		Nonce:     ch.Nonce,
		Origin:    ch.Origin,
		ExpiresAt: time.Now().Add(eidNonceTTL),
	})
	return nil
}

var (
	errEidRedisRequired    = errors.New("eid_redis_required")
	errEidNonceExpired     = errors.New("eid_nonce_expired")
	errEidNonceStoreFailed = errors.New("eid_nonce_store_failed")
)

func parseEidChallengePayload(raw string) (eidChallenge, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return eidChallenge{}, false
	}
	var ch eidChallenge
	if err := json.Unmarshal([]byte(raw), &ch); err == nil && ch.Nonce != "" {
		return ch, true
	}
	// Legacy: bare nonce string (pre-origin binding).
	return eidChallenge{Nonce: raw}, true
}

// takeEidChallenge atomically consumes the pending Web eID challenge.
func (a *API) takeEidChallenge(ctx context.Context, userID string) (eidChallenge, error) {
	key := "eid:webeid:nonce:" + userID
	if a.redis != nil {
		v, err := a.redis.GetDel(ctx, key)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return eidChallenge{}, errEidNonceExpired
			}
			log.Printf("eid nonce getdel: %v", err)
			return eidChallenge{}, errEidNonceStoreFailed
		}
		ch, ok := parseEidChallengePayload(v)
		if !ok {
			return eidChallenge{}, errEidNonceExpired
		}
		return ch, nil
	}
	if !a.cfg.AllowsInMemoryEidNonce() {
		return eidChallenge{}, errEidRedisRequired
	}
	raw, ok := eidMemNonce.LoadAndDelete(userID)
	if !ok {
		return eidChallenge{}, errEidNonceExpired
	}
	entry, ok := raw.(eidMemNonceEntry)
	if !ok || time.Now().After(entry.ExpiresAt) || entry.Nonce == "" {
		return eidChallenge{}, errEidNonceExpired
	}
	return eidChallenge{Nonce: entry.Nonce, Origin: entry.Origin}, nil
}

// importVetEidViewer POST /vet/eid/import — multipart file (.eid/.xml).
func (a *API) importVetEidViewer(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireEidBE(w, r, "import")
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
	id, ok := a.requireEidBE(w, r, "webeid")
	if !ok {
		return
	}
	nonce, err := eid.GenerateChallengeNonce()
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "internal", "eid_nonce_failed")
		return
	}
	origin := a.resolveEidOrigin(r)
	if origin == "" {
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "eid_site_origin_missing")
		return
	}
	ch := eidChallenge{Nonce: nonce, Origin: origin}
	if err := a.storeEidChallenge(r.Context(), id.UserID, ch); err != nil {
		if errors.Is(err, errEidRedisRequired) {
			writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "eid_redis_required")
			return
		}
		writeErr(w, r, http.StatusInternalServerError, "internal", "eid_nonce_store_failed")
		return
	}
	httpx.WriteData(w, http.StatusOK, map[string]any{
		"nonce":  nonce,
		"origin": origin,
	})
}

type webEidVerifyReq struct {
	Token json.RawMessage `json:"token"`
}

// verifyVetWebEid POST /vet/eid/web-eid/verify
func (a *API) verifyVetWebEid(w http.ResponseWriter, r *http.Request) {
	id, ok := a.requireEidBE(w, r, "webeid")
	if !ok {
		return
	}
	var req webEidVerifyReq
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil || len(req.Token) == 0 {
		writeErr(w, r, http.StatusBadRequest, "bad_request", "token_required")
		return
	}
	ch, err := a.takeEidChallenge(r.Context(), id.UserID)
	if err != nil {
		switch {
		case errors.Is(err, errEidNonceStoreFailed), errors.Is(err, errEidRedisRequired):
			a.recordEidReading(r.Context(), id, "web_eid", false, eid.Identity{}, "eid_nonce_store_failed")
			writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "eid_nonce_store_failed")
		default:
			a.recordEidReading(r.Context(), id, "web_eid", false, eid.Identity{}, "eid_nonce_expired")
			writeErr(w, r, http.StatusUnprocessableEntity, "validation_error", "eid_nonce_expired")
		}
		return
	}
	// Validator must match the origin stored with the challenge (localhost ↔ 127.0.0.1).
	origin := ch.Origin
	if origin == "" {
		origin = a.resolveEidOrigin(r)
		if origin == "" {
			origin = a.eidSiteOrigin()
		}
	}
	validator, err := eid.CachedAuthTokenValidator(origin, a.cfg.WebEidDisableOCSP)
	if err != nil {
		_ = a.storeEidChallenge(r.Context(), id.UserID, ch)
		code := "eid_validator_failed"
		if errors.Is(err, eid.ErrCACertsMissing) {
			code = "eid_ca_certs_missing"
		}
		a.recordEidReading(r.Context(), id, "web_eid", false, eid.Identity{}, code)
		writeErr(w, r, http.StatusServiceUnavailable, "unavailable", code)
		return
	}
	identity, err := eid.VerifyAuthToken(r.Context(), validator, req.Token, ch.Nonce)
	if err != nil {
		if eid.IsWebEidInfraError(err) {
			if storeErr := a.storeEidChallenge(r.Context(), id.UserID, ch); storeErr != nil {
				log.Printf("eid nonce restore after infra error: %v", storeErr)
			}
			a.recordEidReading(r.Context(), id, "web_eid", false, eid.Identity{}, "eid_token_infra")
			writeErr(w, r, http.StatusServiceUnavailable, "unavailable", "eid_token_infra")
			return
		}
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
