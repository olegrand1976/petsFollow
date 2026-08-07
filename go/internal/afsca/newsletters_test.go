package afsca

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseIndexHTML(t *testing.T) {
	html := `
<div class="paragraph__body field--text-formatted"><p>07/08/2026 - <a href="https://www.static.favv.be/newsletters-vt-da/newsletter559_fr.asp" target="_blank">Newsletter N°559</a><br>
Cas suspect d'infection par le virus du Nil occidental confirmé chez un cheval en Belgique</p>

<hr>
<p>09/07/2026 - <a href="https://www.static.favv.be/newsletters-vt-da/newsletter558_fr.asp" target="_blank">Newsletter N°558</a><br>
Fièvre catarrhale ovine : aide financière</p>

<hr>
<p>09/07/2026 - <a href="https://www.static.favv.be/newsletters-vt-da/newsletter558_fr.asp" target="_blank">Newsletter N°558</a><br>
Fièvre catarrhale ovine : aide financière</p>

<hr>
<p>08/07/2026 - <a href="https://www.static.favv.be/newsletters-vt-da/newsletter557_fr.asp" target="_blank">Newsletter N°557</a><br>
Période de vacances et risque de rage</p>
`
	items := ParseIndexHTML(html, 5)
	if len(items) != 3 {
		t.Fatalf("want 3 unique items, got %d %#v", len(items), items)
	}
	if items[0].Date != "2026-08-07" {
		t.Fatalf("date %q", items[0].Date)
	}
	if items[0].Label != "Newsletter N°559" {
		t.Fatalf("label %q", items[0].Label)
	}
	if !strings.Contains(items[0].Title, "Nil occidental") {
		t.Fatalf("title %q", items[0].Title)
	}
	if !strings.HasSuffix(items[0].URL, "newsletter559_fr.asp") {
		t.Fatalf("url %q", items[0].URL)
	}
}

func TestParseIndexHTMLRealSnippet(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(file), "testdata", "index_fr_snippet.html"))
	if err != nil {
		t.Fatal(err)
	}
	items := ParseIndexHTML(string(raw), 5)
	if len(items) < 3 {
		t.Fatalf("expected ≥3 from real snippet, got %d", len(items))
	}
	if items[0].Date != "2026-08-07" {
		t.Fatalf("date %q", items[0].Date)
	}
	if !strings.Contains(items[0].Title, "Nil occidental") {
		t.Fatalf("title %q", items[0].Title)
	}
	// Snippet contains a duplicated newsletter558 entry — ensure URL dedupe.
	seen := map[string]int{}
	for _, it := range items {
		seen[it.URL]++
		if seen[it.URL] > 1 {
			t.Fatalf("duplicate url %q", it.URL)
		}
	}
}

func TestParseIndexHTMLLimit(t *testing.T) {
	html := `
<p>07/08/2026 - <a href="https://example.test/a">Newsletter N°1</a><br>A</p>
<p>06/08/2026 - <a href="https://example.test/b">Newsletter N°2</a><br>B</p>
<p>05/08/2026 - <a href="https://example.test/c">Newsletter N°3</a><br>C</p>
`
	items := ParseIndexHTML(html, 2)
	if len(items) != 2 {
		t.Fatalf("got %d", len(items))
	}
	if got := ClampLimit(999); got != MaxLimit {
		t.Fatalf("clamp %d", got)
	}
}

func TestResolveLocale(t *testing.T) {
	if ResolveLocale("nl") != "nl" || ResolveLocale("NL") != "nl" {
		t.Fatal("nl")
	}
	if ResolveLocale("fr") != "fr" || ResolveLocale("en") != "fr" {
		t.Fatal("fallback fr")
	}
}

func TestClientIndexURL(t *testing.T) {
	c := NewClient()
	fr := c.IndexURL("fr")
	nl := c.IndexURL("nl")
	if !strings.Contains(fr, "/fr/") || !strings.Contains(nl, "/nl/") {
		t.Fatalf("fr=%q nl=%q", fr, nl)
	}
	c.BaseURL = "http://example.test"
	if !strings.HasPrefix(c.IndexURL("fr"), "http://example.test/fr/") {
		t.Fatalf("custom base %q", c.IndexURL("fr"))
	}
}

func TestResolveAbsoluteURL(t *testing.T) {
	got := resolveAbsoluteURL("https://favv-afsca.be", "/fr/newsletters/x")
	if got != "https://favv-afsca.be/fr/newsletters/x" {
		t.Fatalf("%q", got)
	}
	got = resolveAbsoluteURL("https://favv-afsca.be/fr", "https://www.static.favv.be/a.asp")
	if got != "https://www.static.favv.be/a.asp" {
		t.Fatalf("%q", got)
	}
}
