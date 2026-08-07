package headerlinks

import (
	"testing"
)

func TestNormalizeAndValidate_DefaultsEmpty(t *testing.T) {
	p, err := NormalizeAndValidate("BE", EmptyPrefs())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Enabled) == 0 {
		t.Fatal("expected default enabled catalog ids")
	}
	found := false
	for _, id := range p.Enabled {
		if id == "be_vetcompendium" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected be_vetcompendium in defaults, got %#v", p.Enabled)
	}
}

func TestNormalizeAndValidate_RejectsForeignCatalogID(t *testing.T) {
	_, err := NormalizeAndValidate("BE", Prefs{
		Enabled: []string{"fr_ordre"},
	})
	if err == nil || err.Error() != "invalid_catalog_id" {
		t.Fatalf("want invalid_catalog_id, got %v", err)
	}
}

func TestNormalizeAndValidate_CustomHTTPSOnly(t *testing.T) {
	_, err := NormalizeAndValidate("FR", Prefs{
		Custom: []CustomLink{{Label: "Lab", URL: "http://lab.example"}},
	})
	if err == nil || err.Error() != "invalid_custom_url" {
		t.Fatalf("want invalid_custom_url, got %v", err)
	}
	p, err := NormalizeAndValidate("FR", Prefs{
		Custom: []CustomLink{{Label: "Lab", URL: "https://lab.example/path"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Custom) != 1 || !stringsHasPrefix(p.Custom[0].ID, "custom_") {
		t.Fatalf("custom %#v", p.Custom)
	}
}

func TestNormalizeAndValidate_MaxCustom(t *testing.T) {
	customs := make([]CustomLink, MaxCustomLinks+1)
	for i := range customs {
		customs[i] = CustomLink{Label: "L", URL: "https://example.com/" + string(rune('a'+i))}
	}
	_, err := NormalizeAndValidate("ES", Prefs{Custom: customs})
	if err == nil || err.Error() != "too_many_custom_links" {
		t.Fatalf("want too_many_custom_links, got %v", err)
	}
}

func TestResolve_IncludesCustomAndCatalog(t *testing.T) {
	prefs := Prefs{
		Enabled: []string{"be_ordre"},
		Order:   []string{"custom_abc", "be_ordre"},
		Custom:  []CustomLink{{ID: "custom_abc", Label: "Lab", URL: "https://lab.test/"}},
	}
	items := Resolve("BE", "fr", prefs, func(id string) string { return id })
	if len(items) != 2 {
		t.Fatalf("got %#v", items)
	}
	if items[0].Kind != "custom" || items[0].Label != "Lab" {
		t.Fatalf("first %#v", items[0])
	}
	if items[1].ID != "be_ordre" || items[1].Kind != "catalog" {
		t.Fatalf("second %#v", items[1])
	}
}

func TestResolve_DefaultsWhenEmpty(t *testing.T) {
	items := Resolve("FR", "fr", EmptyPrefs(), func(id string) string { return id })
	if len(items) == 0 {
		t.Fatal("expected default FR links")
	}
}

func stringsHasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}
