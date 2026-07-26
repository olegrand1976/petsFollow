package store

import (
	"testing"
	"time"
)

func TestScheduleLocationNeverNil(t *testing.T) {
	cases := []string{"", "Europe/Brussels", "Invalid/Zone", "UTC"}
	for _, tz := range cases {
		loc := scheduleLocation(tz)
		if loc == nil {
			t.Fatalf("scheduleLocation(%q) returned nil", tz)
		}
		// Must not panic.
		_ = time.Now().In(loc).Format(time.RFC3339)
	}
}

func TestScheduleLocationBrusselsOffset(t *testing.T) {
	loc := scheduleLocation("Europe/Brussels")
	// Mid-winter: CET = UTC+1 (no DST).
	winter := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	got := winter.In(loc)
	if _, offset := got.Zone(); offset != 3600 {
		t.Fatalf("expected CET +1h, got offset=%d name=%s", offset, got.Location().String())
	}
}
