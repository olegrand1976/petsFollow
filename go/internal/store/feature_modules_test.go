package store

import "testing"

func TestFeatureModulesDefaults(t *testing.T) {
	m := FeatureModules{UserID: "x"}
	if m.ModuleCarePlus || m.ModuleHorse || m.ModuleKennel || m.ModuleFamily {
		t.Fatal("defaults must be false")
	}
}
