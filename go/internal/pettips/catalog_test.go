package pettips

import (
	"testing"
	"time"
)

func TestSelectFiltersBySpeciesAndMonth(t *testing.T) {
	july := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	got := Select([]string{"dog", "cat"}, july, 10)
	ids := map[string]bool{}
	for _, d := range got {
		ids[d.ID] = true
	}
	if !ids["heat_safety"] || !ids["parasite_season"] || !ids["hydration"] {
		t.Fatalf("summer dog/cat tips missing: %#v", ids)
	}
	if ids["horse_pasture"] {
		t.Fatalf("horse tip should not match dog/cat household")
	}
	if ids["cold_paw_care"] {
		t.Fatalf("winter tip should not match July")
	}
}

func TestSelectHorseOnly(t *testing.T) {
	may := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	got := Select([]string{"horse"}, may, 5)
	if len(got) == 0 {
		t.Fatal("expected horse + general tips")
	}
	var hasHorse, hasHeat bool
	for _, d := range got {
		if d.ID == "horse_pasture" {
			hasHorse = true
		}
		if d.ID == "heat_safety" {
			hasHeat = true
		}
	}
	if !hasHorse {
		t.Fatalf("missing horse_pasture: %#v", got)
	}
	if hasHeat {
		t.Fatal("heat_safety is dog/cat only")
	}
}

func TestSelectEmptyHousehold(t *testing.T) {
	got := Select(nil, time.Now(), 3)
	if len(got) != 0 {
		t.Fatalf("want empty got %#v", got)
	}
}

func TestSelectLimitAndPriority(t *testing.T) {
	july := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	got := Select([]string{"dog"}, july, 2)
	if len(got) != 2 {
		t.Fatalf("limit 2 got %d", len(got))
	}
	if got[0].Priority > got[1].Priority {
		t.Fatalf("not sorted by priority: %#v", got)
	}
}
