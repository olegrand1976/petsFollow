package config_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
)

func TestAllowsInMemoryEidNonce(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	cfg := config.Config{}
	if !cfg.AllowsInMemoryEidNonce() {
		t.Fatal("local must allow memory nonce")
	}

	t.Setenv("APP_ENV", "production")
	cfg = config.Config{DevSeedEnabled: false}
	if cfg.AllowsInMemoryEidNonce() {
		t.Fatal("production without DEV_SEED must refuse memory nonce")
	}

	cfg.DevSeedEnabled = true
	if cfg.AllowsInMemoryEidNonce() {
		t.Fatal("production must refuse memory nonce even with DEV_SEED (multi-replica)")
	}

	t.Setenv("APP_ENV", "staging")
	if cfg.AllowsInMemoryEidNonce() {
		t.Fatal("staging must refuse memory nonce")
	}
}

func TestValidateEid_OCSPDisable(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	cfg := config.Config{WebEidDisableOCSP: true}
	if err := cfg.ValidateEid(); err != nil {
		t.Fatalf("local may disable OCSP: %v", err)
	}

	t.Setenv("APP_ENV", "production")
	if err := cfg.ValidateEid(); err == nil {
		t.Fatal("production must refuse WEB_EID_DISABLE_OCSP")
	}

	cfg.WebEidDisableOCSP = false
	if err := cfg.ValidateEid(); err != nil {
		t.Fatalf("OCSP on must always pass: %v", err)
	}
}
