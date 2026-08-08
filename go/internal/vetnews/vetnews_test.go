package vetnews

import (
	"strings"
	"testing"
)

func TestStripHTML(t *testing.T) {
	got := StripHTML(`<p>Hello <b>world</b>&nbsp;!</p><script>x()</script>`)
	if !strings.Contains(got, "Hello world") || strings.Contains(got, "<") {
		t.Fatalf("got %q", got)
	}
}

func TestParseFlexibleTime(t *testing.T) {
	cases := []string{"07.08.2026", "07/08/2026", "Mon, 06 Jul 2026 13:06:24 +0000", "Published on 07/08/2026"}
	for _, c := range cases {
		if ParseFlexibleTime(c) == nil {
			t.Fatalf("failed %q", c)
		}
	}
}

func TestClassifyCritical(t *testing.T) {
	cat, imp, _ := Classify(RawItem{
		Title:   "Influenza aviaire H5N1 : nouveau foyer en Belgique",
		Summary: "Alerte sanitaire",
	}, "woah", []string{"epidemio"})
	if imp != ImportanceCritical {
		t.Fatalf("importance %s", imp)
	}
	if cat != CategoryEpidemio {
		t.Fatalf("category %s", cat)
	}
}

func TestClassifyDoesNotOverfireOnSoftWords(t *testing.T) {
	_, imp, _ := Classify(RawItem{
		Title:   "A study of clinical evidence in practice",
		Summary: "Research notes",
	}, "todays_vet_practice", []string{"pratique"})
	if imp == ImportanceCritical {
		t.Fatalf("soft clinical text should not be critical, got %s", imp)
	}
}

func TestSourceCatalog(t *testing.T) {
	cat := SourceCatalog()
	if len(cat) != 6 {
		t.Fatalf("want 6 got %d", len(cat))
	}
	if cat[0].ID != "anses" || cat[0].Name == "" {
		t.Fatalf("%#v", cat[0])
	}
}

func TestParseRSSDedup(t *testing.T) {
	body := []byte(`<?xml version="1.0"?>
<rss version="2.0"><channel>
<item><title>One</title><link>https://example.test/a</link><description>Hi</description><pubDate>Mon, 06 Jul 2026 13:06:24 +0000</pubDate></item>
<item><title>One dup</title><link>https://example.test/a</link><description>Hi</description></item>
<item><title>Two</title><link>https://example.test/b</link><description>Hi</description></item>
</channel></rss>`)
	items, err := ParseRSS(body, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 got %d", len(items))
	}
}

func TestParseAnsesHTML(t *testing.T) {
	raw := `
<article class="actualite actualite-teaser">
  <a href="/fr/content/rage-une-reemergence" class="full-link"></a>
  <img src="/sites/default/files/rage.jpg" />
  <span class="infobar__item infobar__item--date">10/06/2026</span>
  <h3 class="teaser__title card__title"><span>Rage : une réémergence préoccupante</span></h3>
</article>`
	items := ParseAnsesHTML(raw, 5)
	if len(items) != 1 {
		t.Fatalf("got %#v", items)
	}
	if !strings.Contains(items[0].SourceURL, "https://www.anses.fr/fr/content/rage") {
		t.Fatalf("url %q", items[0].SourceURL)
	}
	if items[0].PublishedAt == nil {
		t.Fatal("missing date")
	}
}

func TestParsePointVeterinaireHTML(t *testing.T) {
	raw := `
<div class="row actu_horizontale">
  <div class="col-md-6 thumb"><img src="/images/x.png"/></div>
  <div class="col-md-6">
    <h3><a href="/actualites/actualites-professionnelles/deux-nouveaux-cas.html">Deux nouveaux cas</a></h3>
    <div class="chapo"><a href="/actualites/actualites-professionnelles/deux-nouveaux-cas.html">Teaser text here long enough</a>
      <div class="infos"><strong class="author">Auteur</strong> | 31.07.2026</div>
    </div>
  </div>
</div>
</div>
</div>`
	items := ParsePointVeterinaireHTML(raw, 5)
	if len(items) != 1 {
		t.Fatalf("got %#v", items)
	}
	if items[0].Title != "Deux nouveaux cas" {
		t.Fatalf("title %q", items[0].Title)
	}
}

func TestParseWOAHHTML(t *testing.T) {
	raw := `
<li class="cards__item cards__item--post cards__item--1">
  <h3 class="cards__title">
    <a href="https://www.woah.org/fr/laustralie-signale-le-premier-cas/" title="x">L’Australie signale le premier cas d’influenza aviaire</a>
  </h3>
  <div class="cards__infos-item cards__infos-item--date">Published on 07/08/2026</div>
  <p class="cards__type">Déclaration</p>
</li>`
	items := ParseWOAHHTML(raw, 5)
	if len(items) != 1 {
		t.Fatalf("got %#v", items)
	}
	if items[0].PublishedAt == nil {
		t.Fatal("date")
	}
}

func TestNormalizeURL(t *testing.T) {
	if got := NormalizeURL("https://x.test/a/#frag"); got != "https://x.test/a" {
		t.Fatalf("%q", got)
	}
}

func TestResolveURL(t *testing.T) {
	if got := ResolveURL("https://www.anses.fr", "/fr/content/x"); got != "https://www.anses.fr/fr/content/x" {
		t.Fatalf("%q", got)
	}
}
