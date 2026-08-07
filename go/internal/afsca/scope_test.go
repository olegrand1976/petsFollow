package afsca

import "testing"

func TestHeuristicClassifyAndFilter(t *testing.T) {
	items := []Item{
		{Title: "Cas suspect Nil occidental chez un cheval", URL: "u1"},
		{Title: "Rage : vigilance chiens importés", URL: "u2"},
		{Title: "L'AFSCA recrute des CDM", URL: "u3"},
		{Title: "Grippe aviaire H5 en élevage de volailles", URL: "u4"},
	}
	tagged := HeuristicClassify(items)
	if tagged[0].Scope != ScopeLarge {
		t.Fatalf("horse → large, got %q", tagged[0].Scope)
	}
	if tagged[1].Scope != ScopeSmall {
		t.Fatalf("dog → small, got %q", tagged[1].Scope)
	}
	if tagged[2].Scope != ScopeBoth {
		t.Fatalf("recruit → both, got %q", tagged[2].Scope)
	}
	if tagged[3].Scope != ScopeLarge {
		t.Fatalf("poultry → large, got %q", tagged[3].Scope)
	}

	smallOnly := FilterByScope(tagged, ScopeSmall)
	for _, it := range smallOnly {
		if !MatchesScope(ScopeSmall, it.Scope) {
			t.Fatalf("unexpected %q scope %q", it.URL, it.Scope)
		}
	}
	if len(smallOnly) < 2 { // dog + both
		t.Fatalf("small filter got %d", len(smallOnly))
	}

	largeOnly := FilterByScope(tagged, ScopeLarge)
	if len(largeOnly) < 3 { // horse + poultry + both
		t.Fatalf("large filter got %d", len(largeOnly))
	}

	if got := len(FilterByScope(tagged, ScopeBoth)); got != len(tagged) {
		t.Fatalf("both should keep all, got %d", got)
	}
}

func TestHeuristicNoFalsePositiveCatInCatarrhale(t *testing.T) {
	items := HeuristicClassify([]Item{
		{Title: "Fièvre catarrhale ovine : aide financière", URL: "u1"},
		{Title: "Cattle transport rules update", URL: "u2"},
	})
	if items[0].Scope != ScopeLarge {
		t.Fatalf("catarrhale must be large (not both via cat), got %q", items[0].Scope)
	}
	if items[1].Scope != ScopeLarge {
		t.Fatalf("cattle must be large (not both via cat), got %q", items[1].Scope)
	}
}

func TestNormalizeScope(t *testing.T) {
	if NormalizeScope("SMALL") != ScopeSmall || NormalizeScope("x") != ScopeBoth {
		t.Fatal("normalize")
	}
}
