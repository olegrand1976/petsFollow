package pharmacy

import "testing"

func TestIsPredictableAFMPSImportObjectKey(t *testing.T) {
	cases := []struct {
		key  string
		want bool
	}{
		{"", true},
		{"afmps-imports/latest.csv", true},
		{"AFMPS-IMPORTS/LATEST.CSV", true},
		{"afmps-imports/other/latest.csv", true},
		{"afmps-imports/a1b2c3d4-e5f6.csv", false},
		{"afmps-imports/pack-2026-08.csv", false},
	}
	for _, tc := range cases {
		if got := IsPredictableAFMPSImportObjectKey(tc.key); got != tc.want {
			t.Fatalf("%q: got %v want %v", tc.key, got, tc.want)
		}
	}
}
