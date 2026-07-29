package handlers

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestShouldSendPharmacyExpiryDigest(t *testing.T) {
	sumAlert := store.ExpirySummary{Soon: 1}
	sumEmpty := store.ExpirySummary{}

	cases := []struct {
		name       string
		enabled    bool
		force      bool
		weekday    int
		configured int
		sum        store.ExpirySummary
		want       bool
	}{
		{"disabled", false, true, 1, 1, sumAlert, false},
		{"empty skip", true, true, 1, 1, sumEmpty, false},
		{"monday ok", true, false, 1, 1, sumAlert, true},
		{"tuesday skip", true, false, 2, 1, sumAlert, false},
		{"force midweek", true, true, 3, 1, sumAlert, true},
		{"force empty still skip", true, true, 3, 1, sumEmpty, false},
		{"quarantine only", true, false, 1, 1, store.ExpirySummary{Quarantine: 2}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldSendPharmacyExpiryDigest(tc.enabled, tc.force, tc.weekday, tc.configured, tc.sum)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
