package eid

import (
	"context"
	"errors"
	"fmt"
	"testing"
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
	if !IsWebEidInfraError(fmt.Errorf("eid_token_invalid: ocsp request failed")) {
		t.Fatal("ocsp must be infra")
	}
	if !IsWebEidInfraError(context.DeadlineExceeded) {
		t.Fatal("deadline must be infra")
	}
	if IsWebEidInfraError(errors.New("eid_token_invalid: signature mismatch")) {
		t.Fatal("crypto must not be infra")
	}
	if IsWebEidInfraError(nil) {
		t.Fatal("nil")
	}
}
