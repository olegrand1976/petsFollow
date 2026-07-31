package pharmacy

import (
	"testing"
	"time"
)

func TestClassifyExpiryBands(t *testing.T) {
	today := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)
	th := BandThresholds{SoonDays: 90, ReturnDays: 60, CriticalDays: 30}

	cases := []struct {
		days   int
		status string
		want   ExpiryBand
	}{
		{120, "active", BandOK},
		{90, "active", BandSoon},
		{61, "active", BandSoon},
		{60, "active", BandReturn},
		{31, "active", BandReturn},
		{30, "active", BandCritical},
		{0, "active", BandCritical},
		{-1, "active", BandExpired},
		{10, "quarantine", BandQuarantine},
	}
	for _, tc := range cases {
		exp := today.AddDate(0, 0, tc.days)
		got := ClassifyExpiry(exp, today, tc.status, th)
		if got != tc.want {
			t.Fatalf("days=%d status=%s got=%s want=%s", tc.days, tc.status, got, tc.want)
		}
	}
}

func TestBrusselsToday(t *testing.T) {
	// 2026-07-26 23:30 UTC = 2026-07-27 01:30 Brussels (CEST)
	now := time.Date(2026, 7, 26, 23, 30, 0, 0, time.UTC)
	d := BrusselsToday(now)
	if d.Format("2006-01-02") != "2026-07-27" {
		t.Fatalf("got %s", d.Format("2006-01-02"))
	}
}
