package handlers

import (
	"strings"
	"testing"
	"time"
)

func TestDigestAudienceLabelAddsTestTag(t *testing.T) {
	got := digestAudienceLabel("staging", "fr")
	if !strings.Contains(got, "staging") || !strings.Contains(got, "TEST") {
		t.Fatalf("got %q", got)
	}
	gotEN := digestAudienceLabel("staging", "en")
	if !strings.Contains(gotEN, "TEST") {
		t.Fatalf("en tag %q", gotEN)
	}
	prod := digestAudienceLabel("production", "fr")
	if prod != "production" {
		t.Fatalf("prod label %q", prod)
	}
}

func TestIsoWeekStartMonday(t *testing.T) {
	// Saturday 2026-08-08 Brussels → Monday 2026-08-03
	sat := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	got := isoWeekStartMonday(sat)
	want := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("week start %v want %v", got, want)
	}
}
