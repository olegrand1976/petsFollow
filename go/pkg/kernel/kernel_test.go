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


func TestSupportsHeartRateControl(t *testing.T) {
	if !SupportsHeartRateControl("dog") || !SupportsHeartRateControl("cat") || !SupportsHeartRateControl("horse") {
		t.Fatal("expected dog/cat/horse supported")
	}
	if SupportsHeartRateControl("other") || SupportsHeartRateControl("") || SupportsHeartRateControl("cattle") {
		t.Fatal("expected other/empty/cattle unsupported")
	}
}

func TestIsFoodChainSpecies(t *testing.T) {
	for _, sp := range []string{"horse", "donkey", "cattle", "sheep", "goat", "pig", "poultry", "rabbit", "alpaca", "llama"} {
		if !IsFoodChainSpecies(sp) {
			t.Fatalf("expected %s food-chain", sp)
		}
	}
	if IsFoodChainSpecies("dog") || IsFoodChainSpecies("cat") || IsFoodChainSpecies("other") {
		t.Fatal("expected companion species not food-chain UI")
	}
}

func TestDefaultFoodChainStatus(t *testing.T) {
	if got := DefaultFoodChainStatus("cattle"); got != "food_producing" {
		t.Fatalf("cattle: got %q", got)
	}
	if got := DefaultFoodChainStatus("alpaca"); got != "food_producing" {
		t.Fatalf("alpaca: got %q", got)
	}
	if got := DefaultFoodChainStatus("horse"); got != "companion" {
		t.Fatalf("horse: got %q", got)
	}
	if got := DefaultFoodChainStatus("rabbit"); got != "companion" {
		t.Fatalf("rabbit: got %q", got)
	}
	if got := DefaultFoodChainStatus("dog"); got != "companion" {
		t.Fatalf("dog: got %q", got)
	}
}

func TestIsHeartRateDeltaAlert(t *testing.T) {
	prev := 80
	if !IsHeartRateDeltaAlert(110, &prev, 30) {
		t.Fatal("expected alert for +30")
	}
	if IsHeartRateDeltaAlert(109, &prev, 30) {
		t.Fatal("expected no alert for +29")
	}
	if IsHeartRateDeltaAlert(110, nil, 30) {
		t.Fatal("expected no alert without previous")
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
