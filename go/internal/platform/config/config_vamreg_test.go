package config_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
)

func TestValidateVamregDryRunOK(t *testing.T) {
	cfg := config.Config{VamregDryRun: true}
	if err := cfg.ValidateVamreg(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateVamregLiveRequiresURL(t *testing.T) {
	cfg := config.Config{VamregDryRun: false, VamregBaseURL: "", VamregAPIKey: "k"}
	if err := cfg.ValidateVamreg(); err == nil {
		t.Fatal("want error")
	}
	cfg.VamregBaseURL = "https://vamreg.example"
	if err := cfg.ValidateVamreg(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateVamregLiveRequiresAPIKey(t *testing.T) {
	cfg := config.Config{VamregDryRun: false, VamregBaseURL: "https://vamreg.example", VamregAPIKey: ""}
	if err := cfg.ValidateVamreg(); err == nil {
		t.Fatal("want API key error")
	}
	cfg.VamregAPIKey = "secret"
	if err := cfg.ValidateVamreg(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateVamregLiveRejectsAFMPSListsBase(t *testing.T) {
	cfg := config.Config{
		VamregDryRun:  false,
		VamregBaseURL: "https://app.fagg-afmps.be/vamreg/api",
		VamregAPIKey:  "secret",
	}
	if err := cfg.ValidateVamreg(); err == nil {
		t.Fatal("want AFMPS lists collision error")
	}
	cfg.VamregBaseURL = "https://app.fagg-afmps.be/vamreg/api/"
	if err := cfg.ValidateVamreg(); err == nil {
		t.Fatal("want trailing-slash collision error")
	}
	cfg.VamregBaseURL = "https://declare.example/v1"
	if err := cfg.ValidateVamreg(); err != nil {
		t.Fatal(err)
	}
}
