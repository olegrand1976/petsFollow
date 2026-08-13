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
	if !cfg.AllowsInMemoryEidNonce() {
		t.Fatal("DEV_SEED must allow memory nonce even in production label")
	}
}
