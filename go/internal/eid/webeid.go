package eid

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	webeid "github.com/gmb-lib/go-web-eid"
	"github.com/gmb-lib/go-web-eid/certificate"
)

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

// NewAuthTokenValidator builds a Web eID validator for the given site origin.
func NewAuthTokenValidator(origin string, cas []*x509.Certificate, disableOCSP bool) (webeid.AuthTokenValidator, error) {
	origin = SiteOrigin(origin)
	if origin == "" {
		return nil, fmt.Errorf("eid_site_origin_missing")
	}
	if len(cas) == 0 {
		return nil, fmt.Errorf("eid_ca_certs_missing")
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
		Lastname:   TitleCase(lastname),
		Firstname:  TitleCase(firstname),
		BirthDate:  BirthDateFromNISS(niss),
		Country:    strings.ToUpper(country),
		NISS:       niss,
		ImportTool: "web_eid",
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
