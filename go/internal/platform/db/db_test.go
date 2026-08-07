package db

import (
	"testing"
	"time"
)

const testURL = "postgres://u:p@localhost:5432/petsfollow?sslmode=disable"

// Sans configuration explicite, pgx dérive MaxConns de runtime.NumCPU() : sur
// Cloud Run 1 vCPU cela donne 4 connexions pour 80 requêtes concurrentes.
func TestBuildPoolConfig_AppliesDefaults(t *testing.T) {
	cfg, err := BuildPoolConfig(testURL, DefaultPoolOptions())
	if err != nil {
		t.Fatalf("BuildPoolConfig: %v", err)
	}
	want := DefaultPoolOptions()
	if cfg.MaxConns != want.MaxConns {
		t.Errorf("MaxConns = %d, want %d", cfg.MaxConns, want.MaxConns)
	}
	if cfg.MinConns != want.MinConns {
		t.Errorf("MinConns = %d, want %d", cfg.MinConns, want.MinConns)
	}
	if cfg.MaxConnLifetime != want.MaxConnLifetime {
		t.Errorf("MaxConnLifetime = %s, want %s", cfg.MaxConnLifetime, want.MaxConnLifetime)
	}
	if cfg.HealthCheckPeriod != want.HealthCheckPeriod {
		t.Errorf("HealthCheckPeriod = %s, want %s", cfg.HealthCheckPeriod, want.HealthCheckPeriod)
	}
}

// La chaîne de connexion reste la source de vérité : un ops qui fixe
// pool_max_conns doit être respecté, sinon le réglage disparaît sans bruit.
func TestBuildPoolConfig_URLWins(t *testing.T) {
	cfg, err := BuildPoolConfig(testURL+"&pool_max_conns=7", DefaultPoolOptions())
	if err != nil {
		t.Fatalf("BuildPoolConfig: %v", err)
	}
	if cfg.MaxConns != 7 {
		t.Fatalf("MaxConns = %d, want 7 (valeur de l'URL)", cfg.MaxConns)
	}
}

func TestBuildPoolConfig_MinNeverExceedsMax(t *testing.T) {
	cfg, err := BuildPoolConfig(testURL, PoolOptions{MaxConns: 3, MinConns: 10})
	if err != nil {
		t.Fatalf("BuildPoolConfig: %v", err)
	}
	if cfg.MinConns > cfg.MaxConns {
		t.Fatalf("MinConns %d > MaxConns %d", cfg.MinConns, cfg.MaxConns)
	}
}

func TestBuildPoolConfig_ZeroOptionsKeepPgxDefaults(t *testing.T) {
	cfg, err := BuildPoolConfig(testURL, PoolOptions{})
	if err != nil {
		t.Fatalf("BuildPoolConfig: %v", err)
	}
	if cfg.MaxConns <= 0 {
		t.Fatalf("MaxConns = %d, want the pgx default", cfg.MaxConns)
	}
	if cfg.MaxConnLifetime <= time.Duration(0) {
		t.Fatalf("MaxConnLifetime = %s, want the pgx default", cfg.MaxConnLifetime)
	}
}

func TestBuildPoolConfig_InvalidURL(t *testing.T) {
	if _, err := BuildPoolConfig("://nope", DefaultPoolOptions()); err == nil {
		t.Fatal("expected an error on a malformed connection string")
	}
}
