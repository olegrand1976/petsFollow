package eid

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	webeid "github.com/gmb-lib/go-web-eid"
	"github.com/gmb-lib/go-web-eid/certificate"
	"github.com/gmb-lib/go-web-eid/exceptions"
)

// ErrCACertsMissing is returned when no Belgian eID CA could be loaded.
var ErrCACertsMissing = errors.New("eid_ca_certs_missing")

// SiteOrigin cleans EID_SITE_ORIGIN / Pro public URL for Web eID binding.
func SiteOrigin(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimRight(raw, "/")
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return raw
	}
	return u.Scheme + "://" + u.Host
}

// IsLocalhostHTTP reports whether origin is loopback http (dev allowance).
func IsLocalhostHTTP(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if u.Scheme != "http" {
		return false
	}
	h := strings.ToLower(u.Hostname())
	return h == "localhost" || h == "127.0.0.1" || h == "::1"
}

// ResolveChallengeOrigin picks the Web eID site origin for a challenge.
// Prefer the browser Origin when it matches the configured site, or when both
// are local http loopback aliases (localhost ↔ 127.0.0.1) — Web eID binds the
// tab origin strictly.
func ResolveChallengeOrigin(configured, requestOrigin string) string {
	configured = SiteOrigin(configured)
	req := SiteOrigin(requestOrigin)
	if req == "" {
		return configured
	}
	if req == configured {
		return req
	}
	if IsLocalhostHTTP(configured) && IsLocalhostHTTP(req) {
		return req
	}
	return configured
}

// NewAuthTokenValidator builds a Web eID validator for the given site origin.
func NewAuthTokenValidator(origin string, cas []*x509.Certificate, disableOCSP bool) (webeid.AuthTokenValidator, error) {
	origin = SiteOrigin(origin)
	if origin == "" {
		return nil, fmt.Errorf("eid_site_origin_missing")
	}
	if len(cas) == 0 {
		return nil, ErrCACertsMissing
	}
	b := webeid.NewAuthTokenValidatorBuilder().
		WithSiteOrigin(origin).
		WithTrustedCertificateAuthorities(cas...)
	if IsLocalhostHTTP(origin) {
		b = b.WithAllowInsecureLocalhostOrigin()
	}
	if disableOCSP {
		b = b.WithoutUserCertificateRevocationCheckWithOcsp()
	}
	return b.Build()
}

type validatorCacheKey struct {
	origin      string
	disableOCSP bool
}

var (
	validatorMu    sync.Mutex
	validatorCache = map[validatorCacheKey]webeid.AuthTokenValidator{}
)

// CachedAuthTokenValidator returns a process-wide validator for origin+OCSP mode
// (embedded BE CAs). Safe for concurrent use after Build.
func CachedAuthTokenValidator(origin string, disableOCSP bool) (webeid.AuthTokenValidator, error) {
	origin = SiteOrigin(origin)
	key := validatorCacheKey{origin: origin, disableOCSP: disableOCSP}
	validatorMu.Lock()
	defer validatorMu.Unlock()
	if v, ok := validatorCache[key]; ok {
		return v, nil
	}
	cas, err := LoadTrustedCAs()
	if err != nil || len(cas) == 0 {
		if err == nil {
			err = ErrCACertsMissing
		}
		return nil, err
	}
	v, err := NewAuthTokenValidator(origin, cas, disableOCSP)
	if err != nil {
		return nil, err
	}
	validatorCache[key] = v
	return v, nil
}

// VerifyAuthToken validates a Web eID auth token JSON and extracts identity.
func VerifyAuthToken(ctx context.Context, validator webeid.AuthTokenValidator, tokenJSON []byte, challengeNonce string) (Identity, error) {
	if validator == nil {
		return Identity{}, fmt.Errorf("eid_webeid_unavailable")
	}
	token, err := webeid.Parse(tokenJSON)
	if err != nil {
		return Identity{}, fmt.Errorf("eid_token_parse: %w", err)
	}
	cert, err := validator.Validate(ctx, token, challengeNonce)
	if err != nil {
		return Identity{}, fmt.Errorf("eid_token_invalid: %w", err)
	}
	return IdentityFromCertificate(cert)
}

// webEidExceptionCode returns go-web-eid's stable Code when present (Unwrap-safe).
// Prefer this over Error() substrings — Wrap clones sentinels so errors.Is is unreliable.
func webEidExceptionCode(err error) string {
	var xerr *exceptions.Error
	if errors.As(err, &xerr) && xerr != nil && xerr.Code != "" {
		return xerr.Code
	}
	return ""
}

// ClassifyWebEidVerifyError maps VerifyAuthToken / Validate failures to API msgKeys.
func ClassifyWebEidVerifyError(err error) string {
	if err == nil {
		return ""
	}
	if IsWebEidInfraError(err) {
		return "eid_token_infra"
	}
	switch webEidExceptionCode(err) {
	case exceptions.ErrTokenParse.Code, exceptions.ErrTokenUnsupportedFormat.Code:
		return "eid_token_parse"
	case exceptions.ErrCertificateNotTrusted.Code,
		exceptions.ErrCertificateRevoked.Code,
		exceptions.ErrCertificateExpired.Code,
		exceptions.ErrCertificateNotYetValid.Code,
		exceptions.ErrCertificateDisallowedPolicy.Code,
		exceptions.ErrUserCertificateWrongPurpose.Code:
		return "eid_cert_untrusted"
	case exceptions.ErrTokenSignatureInvalid.Code,
		exceptions.ErrSignatureValueInvalid.Code,
		exceptions.ErrIdentityBindingMismatch.Code:
		return "eid_token_signature"
	}
	// Local wraps (Parse / IdentityFromCertificate) and legacy string fallbacks.
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "eid_token_parse"):
		return "eid_token_parse"
	case strings.Contains(msg, "eid_identity_empty"):
		return "eid_identity_empty"
	case strings.Contains(msg, "certificate_not_trusted"),
		strings.Contains(msg, "certificate not trusted"),
		strings.Contains(msg, "unknown authority"):
		return "eid_cert_untrusted"
	case strings.Contains(msg, "token_signature_invalid"),
		strings.Contains(msg, "signature_invalid"),
		strings.Contains(msg, "invalid signature"):
		return "eid_token_signature"
	default:
		return "eid_token_invalid"
	}
}

// IsWebEidInfraError reports OCSP/network/timeout failures (retryable; do not treat as bad PIN/token).
func IsWebEidInfraError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	if webEidExceptionCode(err) == exceptions.ErrOCSPRequestFailed.Code {
		return true
	}
	msg := strings.ToLower(err.Error())
	for _, needle := range []string{
		"ocsp_request_failed",
		"ocsp request failed",
		"i/o timeout",
		"connection refused",
		"connection reset",
		"temporary failure",
		"no such host",
		"tls handshake",
		"unexpected eof",
		"network is unreachable",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	// Broad "timeout" only when not a user/library action timeout (those are client-side).
	if strings.Contains(msg, "timeout") && !strings.Contains(msg, "user_timeout") && !strings.Contains(msg, "action_timeout") {
		return true
	}
	return false
}

// IdentityFromCertificate maps a validated auth certificate to Identity.
func IdentityFromCertificate(cert *x509.Certificate) (Identity, error) {
	if cert == nil {
		return Identity{}, fmt.Errorf("eid_cert_missing")
	}
	lastname, _ := certificate.SubjectSurname(cert)
	firstname, _ := certificate.SubjectGivenName(cert)
	idCode, _ := certificate.SubjectIDCode(cert)
	country, _ := certificate.SubjectCountryCode(cert)
	niss := ExtractNISSFromEIDAS(idCode)
	if country == "" {
		country = "BE"
	}
	id := Identity{
		Lastname:          TitleCase(lastname),
		Firstname:         TitleCase(firstname),
		BirthDate:         BirthDateFromNISS(niss),
		Country:           strings.ToUpper(country),
		NISS:              niss,
		ImportTool:        "web_eid",
		SignatureVerified: true, // caller must only invoke after Validate succeeded
	}
	if id.Lastname == "" && id.Firstname == "" && id.NISS == "" {
		return Identity{}, fmt.Errorf("eid_identity_empty")
	}
	return id, nil
}

// GenerateChallengeNonce creates a base64 nonce (>=256 bits of entropy).
func GenerateChallengeNonce() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}
