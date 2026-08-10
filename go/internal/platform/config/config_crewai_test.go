package config_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
)

func TestLoadCrewAIEnv(t *testing.T) {
	t.Setenv("CREWAI_BASE_URL", "https://crewai.example/run")
	t.Setenv("CREWAI_SHARED_SECRET", "shared-sekrit")
	t.Setenv("CREWAI_USE_ID_TOKEN", "true")

	cfg := config.Load()
	if cfg.CrewAIBaseURL != "https://crewai.example/run" {
		t.Fatalf("base URL %q", cfg.CrewAIBaseURL)
	}
	if cfg.CrewAISharedSecret != "shared-sekrit" {
		t.Fatalf("secret %q", cfg.CrewAISharedSecret)
	}
	if !cfg.CrewAIUseIDToken {
		t.Fatal("expected CrewAIUseIDToken")
	}
}

func TestLoadCrewAIUseIDTokenOffByDefault(t *testing.T) {
	t.Setenv("CREWAI_BASE_URL", "")
	t.Setenv("CREWAI_SHARED_SECRET", "")
	t.Setenv("CREWAI_USE_ID_TOKEN", "")

	cfg := config.Load()
	if cfg.CrewAIUseIDToken {
		t.Fatal("expected CrewAIUseIDToken false when unset")
	}
}
