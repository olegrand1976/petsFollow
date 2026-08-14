package eid

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gmb-lib/go-web-eid/exceptions"
)

func TestIdentityFromCertificate_SignatureVerifiedFlag(t *testing.T) {
	web := Identity{
		Firstname:         "A",
		Lastname:          "B",
		NISS:              "96072399886",
		ImportTool:        "web_eid",
		SignatureVerified: true,
	}
	if !web.PublicPrefill().SignatureVerified {
		t.Fatal("PublicPrefill must keep signature_verified=true for Web eID")
	}
	viewer := Identity{Firstname: "A", ImportTool: "eid_viewer_xml", SignatureVerified: false}
	if viewer.PublicPrefill().SignatureVerified {
		t.Fatal("viewer must stay signature_verified=false")
	}
}

func TestResolveChallengeOrigin(t *testing.T) {
	got := ResolveChallengeOrigin("http://localhost:3002", "http://127.0.0.1:3002")
	if got != "http://127.0.0.1:3002" {
		t.Fatalf("loopback alias: got %q", got)
	}
	got = ResolveChallengeOrigin("https://petsfollow.ll-it-sc.be", "http://127.0.0.1:3002")
	if got != "https://petsfollow.ll-it-sc.be" {
		t.Fatalf("must ignore untrusted origin: got %q", got)
	}
	got = ResolveChallengeOrigin("https://petsfollow.ll-it-sc.be", "https://petsfollow.ll-it-sc.be")
	if got != "https://petsfollow.ll-it-sc.be" {
		t.Fatalf("exact match: got %q", got)
	}
}

func TestIsWebEidInfraError(t *testing.T) {
	ocsp := exceptions.Wrap(exceptions.ErrOCSPRequestFailed, errors.New("dial tcp timeout"))
	if !IsWebEidInfraError(fmt.Errorf("eid_token_invalid: %w", ocsp)) {
		t.Fatal("OCSP_REQUEST_FAILED must be infra")
	}
	if !IsWebEidInfraError(context.DeadlineExceeded) {
		t.Fatal("deadline must be infra")
	}
	if IsWebEidInfraError(exceptions.Wrap(exceptions.ErrTokenSignatureInvalid, errors.New("bad"))) {
		t.Fatal("crypto must not be infra")
	}
	if IsWebEidInfraError(errors.New("unexpected mid-file eof marker")) {
		t.Fatal("bare eof substring must not be infra")
	}
	if IsWebEidInfraError(nil) {
		t.Fatal("nil")
	}
}

func TestClassifyWebEidVerifyError(t *testing.T) {
	untrusted := exceptions.Wrap(exceptions.ErrCertificateNotTrusted, errors.New("x509: certificate signed by unknown authority"))
	if got := ClassifyWebEidVerifyError(fmt.Errorf("eid_token_invalid: %w", untrusted)); got != "eid_cert_untrusted" {
		t.Fatalf("typed CERTIFICATE_NOT_TRUSTED: got %q", got)
	}
	revoked := exceptions.Wrap(exceptions.ErrCertificateRevoked, errors.New("revoked"))
	if got := ClassifyWebEidVerifyError(fmt.Errorf("eid_token_invalid: %w", revoked)); got != "eid_cert_untrusted" {
		t.Fatalf("revoked: got %q", got)
	}
	sig := exceptions.Wrap(exceptions.ErrTokenSignatureInvalid, errors.New("bad sig"))
	if got := ClassifyWebEidVerifyError(fmt.Errorf("eid_token_invalid: %w", sig)); got != "eid_token_signature" {
		t.Fatalf("sig: got %q", got)
	}
	if got := ClassifyWebEidVerifyError(fmt.Errorf("eid_token_parse: boom")); got != "eid_token_parse" {
		t.Fatalf("parse: got %q", got)
	}
	ocsp := exceptions.Wrap(exceptions.ErrOCSPRequestFailed, errors.New("timeout"))
	if got := ClassifyWebEidVerifyError(fmt.Errorf("eid_token_invalid: %w", ocsp)); got != "eid_token_infra" {
		t.Fatalf("infra: got %q", got)
	}
	// Legacy string path (no typed exception in chain).
	if got := ClassifyWebEidVerifyError(fmt.Errorf("eid_token_invalid: webeid: CERTIFICATE_NOT_TRUSTED: unknown authority")); got != "eid_cert_untrusted" {
		t.Fatalf("legacy string: got %q", got)
	}
}
