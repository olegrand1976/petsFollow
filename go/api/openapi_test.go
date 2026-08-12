package api_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gopkg.in/yaml.v3"
)

// Gate CI léger : la spec OpenAPI reste parseable et documente les paths critiques.
// Pas de sync parfaite routes↔spec en V1 — seulement anti-dérive grossière.
func TestOpenAPISpecCriticalPaths(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	specPath := filepath.Join(filepath.Dir(thisFile), "openapi.yaml")
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read openapi.yaml: %v", err)
	}

	var doc struct {
		OpenAPI string `yaml:"openapi"`
		Paths   map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse openapi.yaml: %v", err)
	}
	if doc.OpenAPI == "" {
		t.Fatal("openapi version missing")
	}
	if len(doc.Paths) == 0 {
		t.Fatal("paths empty")
	}

	required := []string{
		"/health",
		"/api/v1/auth/login",
		"/api/v1/pets",
		"/api/v1/clients",
	}
	for _, p := range required {
		if _, ok := doc.Paths[p]; !ok {
			t.Errorf("critical path missing from openapi.yaml: %s", p)
		}
	}
}
