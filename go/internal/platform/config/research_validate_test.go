package config

import (
	"testing"
)

func TestValidateResearch(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	cfg := Config{ResearchEnabled: true, DevSeedEnabled: false}
	if err := cfg.ValidateResearch(); err == nil {
		t.Fatal("expected error without salt/secret on staging")
	}
	cfg.ResearchAnonSalt = "salt-ok"
	if err := cfg.ValidateResearch(); err == nil {
		t.Fatal("expected error without ETL secret")
	}
	cfg.ResearchEtlSecret = "etl-ok"
	if err := cfg.ValidateResearch(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	t.Setenv("APP_ENV", "local")
	cfgEmpty := Config{ResearchEnabled: true}
	if err := cfgEmpty.ValidateResearch(); err != nil {
		t.Fatalf("local should allow empty secrets: %v", err)
	}
}
