package store

import (
	"testing"
	"time"
)

func TestIsoWeekMondayBrussels(t *testing.T) {
	// Sunday 2026-03-29 01:30 UTC = Sunday morning Brussels (DST) → week Mon 2026-03-23
	sun := time.Date(2026, 3, 29, 1, 30, 0, 0, time.UTC)
	got := isoWeekMonday(sun)
	want := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("isoWeekMonday(%v) = %v, want %v", sun, got, want)
	}
	// Thursday Brussels evening should stay in same week
	thu := time.Date(2026, 3, 26, 22, 0, 0, 0, time.UTC) // Fri 00:00 Brussels during CEST? Mar 26 2026 is Thu
	// 22:00 UTC = 23:00 Brussels (CEST from late March) — still Thursday local → Mon 23
	got = isoWeekMonday(thu)
	if !got.Equal(want) {
		t.Fatalf("isoWeekMonday(%v) = %v, want %v", thu, got, want)
	}
}
