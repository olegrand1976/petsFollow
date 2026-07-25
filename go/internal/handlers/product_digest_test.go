package handlers

import (
	"testing"
	"time"
)

func TestBrusselsDate(t *testing.T) {
	// 2026-07-22 15:30 UTC → 17:30 Brussels (CEST)
	got := brusselsDate(time.Date(2026, 7, 22, 15, 30, 0, 0, time.UTC))
	if got.Format("2006-01-02") != "2026-07-22" {
		t.Fatalf("got %s want 2026-07-22", got.Format("2006-01-02"))
	}
	// 22:30 UTC in summer = 00:30 next day Brussels
	got2 := brusselsDate(time.Date(2026, 7, 22, 22, 30, 0, 0, time.UTC))
	if got2.Format("2006-01-02") != "2026-07-23" {
		t.Fatalf("got %s want 2026-07-23", got2.Format("2006-01-02"))
	}
}

func TestParseDigestDate(t *testing.T) {
	d, err := parseDigestDate("2026-07-22")
	if err != nil {
		t.Fatal(err)
	}
	if d.Format("2006-01-02") != "2026-07-22" {
		t.Fatalf("got %s", d.Format("2006-01-02"))
	}
	if _, err := parseDigestDate("not-a-date"); err == nil {
		t.Fatal("expected error")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	m := map[string]string{"en": "Hello", "fr": "Bonjour"}
	if got := firstNonEmpty(m, "fr"); got != "Bonjour" {
		t.Fatalf("got %q", got)
	}
	if got := firstNonEmpty(map[string]string{"en": "Hi"}, "fr"); got != "Hi" {
		t.Fatalf("got %q", got)
	}
}
