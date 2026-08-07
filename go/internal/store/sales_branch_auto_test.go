package store

import "testing"

func TestDeriveBranchNameAndCode(t *testing.T) {
	tests := []struct {
		full     string
		wantName string
		wantCode string
	}{
		{"Camille Dupont", "Dupont C", "DUPONTC"},
		{"Jean-Pierre Martin", "Martin J", "MARTINJ"},
		{"Élodie Léger", "Léger É", "LEGERE"},
		{"Dupont", "Dupont", "DUPONT"},
		{"  Marie   VanBerg  ", "VanBerg M", "VANBERGM"},
		{"", "", ""},
	}
	for _, tt := range tests {
		name, code := DeriveBranchNameAndCode(tt.full)
		if name != tt.wantName || code != tt.wantCode {
			t.Errorf("DeriveBranchNameAndCode(%q) = (%q, %q), want (%q, %q)",
				tt.full, name, code, tt.wantName, tt.wantCode)
		}
	}
}
