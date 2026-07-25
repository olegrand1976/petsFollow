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

func TestNormalizeHeartRateDurations(t *testing.T) {
	cases := []struct {
		in   []int
		want []int
	}{
		{nil, []int{60}},
		{[]int{}, []int{60}},
		{[]int{15, 30}, []int{15, 30}},
		{[]int{60, 15, 15, 99}, []int{15, 60}},
		{[]int{30, 60, 15}, []int{15, 30, 60}},
	}
	for _, tc := range cases {
		got := NormalizeHeartRateDurations(tc.in)
		if len(got) != len(tc.want) {
			t.Fatalf("in %#v: got %#v want %#v", tc.in, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("in %#v: got %#v want %#v", tc.in, got, tc.want)
			}
		}
	}
}
