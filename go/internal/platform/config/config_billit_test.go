package config_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
)

func TestValidateBillitDisabledOK(t *testing.T) {
	cfg := config.Config{}
	if err := cfg.ValidateBillit(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateBillitLiveRefused(t *testing.T) {
	cfg := config.Config{BillitEnabled: true, BillitMockEnabled: false}
	if err := cfg.ValidateBillit(); err == nil {
		t.Fatal("expected error for live without client")
	}
}

func TestValidateBillitPlainRefusedOutsideDev(t *testing.T) {
	cfg := config.Config{
		BillitEnabled: true, BillitMockEnabled: true,
		BillitSecretsBackend: "plain_dev", DevSeedEnabled: false,
	}
	if err := cfg.ValidateBillit(); err == nil {
		t.Fatal("expected plain_dev refused")
	}
}

func TestValidateBillitMockDevOK(t *testing.T) {
	cfg := config.Config{
		BillitEnabled: true, BillitMockEnabled: true,
		BillitSecretsBackend: "plain_dev", DevSeedEnabled: true,
	}
	if err := cfg.ValidateBillit(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateBillitLocalEncNeedsKey(t *testing.T) {
	cfg := config.Config{
		BillitEnabled: true, BillitMockEnabled: true,
		BillitSecretsBackend: "local_enc", DevSeedEnabled: false,
	}
	if err := cfg.ValidateBillit(); err == nil {
		t.Fatal("expected key required")
	}
	cfg.BillitSecretsKey = "enough-secret-material-here"
	if err := cfg.ValidateBillit(); err != nil {
		t.Fatal(err)
	}
}
