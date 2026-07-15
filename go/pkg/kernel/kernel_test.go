package kernel

import "testing"

func TestCalculateBPM(t *testing.T) {
	if got := CalculateBPM(60, 60); got != 60 {
		t.Fatalf("expected 60 got %d", got)
	}
	if got := CalculateBPM(30, 60); got != 30 {
		t.Fatalf("expected 30 got %d", got)
	}
}

func TestIsHeartRateAlert(t *testing.T) {
	if !IsHeartRateAlert(50, 60, 140) {
		t.Fatal("expected alert for low bpm")
	}
	if !IsHeartRateAlert(150, 60, 140) {
		t.Fatal("expected alert for high bpm")
	}
	if IsHeartRateAlert(80, 60, 140) {
		t.Fatal("expected no alert")
	}
}
