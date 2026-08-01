package store

import "testing"

func TestComputeLabFlag(t *testing.T) {
	low, high := 0.5, 1.5
	vLow, vOk, vHigh := 0.2, 1.0, 2.0
	if got := ComputeLabFlag(&vLow, &low, &high); got != LabFlagLow {
		t.Fatalf("low: got %s", got)
	}
	if got := ComputeLabFlag(&vOk, &low, &high); got != LabFlagNormal {
		t.Fatalf("normal: got %s", got)
	}
	if got := ComputeLabFlag(&vHigh, &low, &high); got != LabFlagHigh {
		t.Fatalf("high: got %s", got)
	}
	if got := ComputeLabFlag(&vOk, nil, nil); got != LabFlagUnknown {
		t.Fatalf("no refs: got %s", got)
	}
	if got := ComputeLabFlag(nil, &low, &high); got != LabFlagUnknown {
		t.Fatalf("nil value: got %s", got)
	}
}
