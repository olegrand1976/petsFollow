package app_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Garde-fou anti-régression : le gain Go 1.26 (Green Tea GC) et les durcissements
// build/runtime disparaissent silencieusement si un refactor retire -trimpath,
// GOMEMLIMIT ou la version golang:1.26. On lit les sources de vérité du repo.

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// go/internal/app/<this file> → repo root is ../../..
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

func TestDockerfileAPIHardening(t *testing.T) {
	content := readRepoFile(t, "deploy/Dockerfile.api")
	for _, want := range []string{
		"golang:1.26",
		"-trimpath",
		`-ldflags="-s -w"`,
		"USER petsfollow",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("deploy/Dockerfile.api missing %q", want)
		}
	}
}

func TestCloudRunAPIRuntimeEnv(t *testing.T) {
	content := readRepoFile(t, "infra/gcp/lib/deploy-run-args.sh")
	for _, want := range []string{
		"GOMEMLIMIT",
		"800MiB",
		"TRUSTED_PROXY_HOPS",
		"BFF_PROXY_SECRET",
		"pf_nuxt_secrets",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("infra/gcp/lib/deploy-run-args.sh missing %q", want)
		}
	}
}

func TestGoModTargets126(t *testing.T) {
	content := readRepoFile(t, "go/go.mod")
	// Go ≥1.21 : `go 1.26.5` (tidy moderne) ou `go 1.26` + `toolchain go1.26.x`.
	hasLanguage := strings.Contains(content, "\ngo 1.26\n") || strings.Contains(content, "\ngo 1.26.")
	if !hasLanguage {
		t.Error("go/go.mod must declare go 1.26 or go 1.26.x")
	}
	hasToolchain := strings.Contains(content, "toolchain go1.26") || strings.Contains(content, "\ngo 1.26.")
	if !hasToolchain {
		t.Error("go/go.mod must pin go 1.26.x (language line or toolchain directive)")
	}
}

// TestCrewAIStagingDeployGuard locks Phase-2 wiring: staging URL + IAM ID token
// + shared secret mount (fail-closed path for the shared orchestrator).
func TestCrewAIStagingDeployGuard(t *testing.T) {
	content := readRepoFile(t, "infra/gcp/lib/deploy-run-args.sh")
	for _, want := range []string{
		"CREWAI_BASE_URL",
		"CREWAI_USE_ID_TOKEN",
		"crewai_use_id_token",
		"CREWAI_USE_ID_TOKEN:-true",
		"CREWAI_SHARED_SECRET",
		"CREWAI_WEBHOOK_SECRET_STAGING",
		"crewai-orchestrator-staging",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("infra/gcp/lib/deploy-run-args.sh missing %q", want)
		}
	}
}
