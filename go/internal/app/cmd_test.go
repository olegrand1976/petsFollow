package app

import "testing"

func TestIsSeedNotifyCmd(t *testing.T) {
	if !IsSeedNotifyCmd([]string{"seed-notify"}) {
		t.Fatal("expected seed-notify match")
	}
	if !IsSeedNotifyCmd([]string{"SEED-NOTIFY"}) {
		t.Fatal("expected case-insensitive match")
	}
	if IsSeedNotifyCmd([]string{"seed"}) {
		t.Fatal("seed must not match seed-notify")
	}
	if IsSeedNotifyCmd(nil) {
		t.Fatal("empty args must not match")
	}
}

func TestIsSeedMassCmd(t *testing.T) {
	if !IsSeedMassCmd([]string{"seed-mass"}) {
		t.Fatal("expected seed-mass match")
	}
	if IsSeedMassCmd([]string{"seed"}) {
		t.Fatal("seed must not match seed-mass")
	}
	if IsSeedCmd([]string{"seed-mass"}) {
		t.Fatal("seed-mass must not match seed")
	}
}
