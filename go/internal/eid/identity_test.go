package eid

import (
	"strings"
	"testing"
)

func TestPublicPrefill_StripsPhoto(t *testing.T) {
	id := Identity{
		Firstname:       "Camille",
		Lastname:        "Testeur",
		NISS:            "96072399828",
		PhotoJPEGBase64: strings.Repeat("A", 200),
	}
	out := id.PublicPrefill()
	if out.PhotoJPEGBase64 != "" {
		t.Fatalf("photo still present")
	}
	if out.Firstname != "Camille" || out.NISS != "96072399828" {
		t.Fatalf("identity fields lost: %#v", out)
	}
	if id.PhotoJPEGBase64 == "" {
		t.Fatal("PublicPrefill must not mutate original")
	}
}

func TestHasUsefulIdentity(t *testing.T) {
	if (Identity{}).HasUsefulIdentity() {
		t.Fatal("empty should be useless")
	}
	if !(Identity{NISS: "96072399828"}).HasUsefulIdentity() {
		t.Fatal("niss alone is useful")
	}
	if !(Identity{Firstname: "A"}).HasUsefulIdentity() {
		t.Fatal("firstname alone is useful")
	}
}

func TestParseViewerExport_WithPhoto(t *testing.T) {
	raw := []byte(`<?xml version="1.0"?>
<Export>
  <surname>Dupont</surname>
  <firstname>Jean</firstname>
  <nationalnumber>85010112345</nationalnumber>
  <photo>aGVsbG8=</photo>
  <addressstreetandnumber>Rue Demo 12</addressstreetandnumber>
  <addresszip>1000</addresszip>
  <addressmunicipality>Bruxelles</addressmunicipality>
</Export>`)
	id, err := ParseViewerExport(raw, "card.eid")
	if err != nil {
		t.Fatal(err)
	}
	if id.PhotoJPEGBase64 == "" {
		t.Fatal("expected photo parsed from XML")
	}
	if id.AddressStreet == "" || id.AddressZip != "1000" {
		t.Fatalf("address %#v", id)
	}
	pub := id.PublicPrefill()
	if pub.PhotoJPEGBase64 != "" {
		t.Fatal("public prefill must strip photo")
	}
}

func TestParseViewerExport_EmptyIdentity(t *testing.T) {
	raw := []byte(`<?xml version="1.0"?><Export><nationality>BE</nationality></Export>`)
	id, err := ParseViewerExport(raw, "empty.eid")
	if err != nil {
		t.Fatal(err)
	}
	if id.HasUsefulIdentity() {
		t.Fatalf("expected empty identity, got %#v", id)
	}
}

func TestBirthDateFromNISS(t *testing.T) {
	if got := BirthDateFromNISS("96072399828"); got != "1996-07-23" {
		t.Fatalf("got %q", got)
	}
	if got := BirthDateFromNISS("850101"); got != "1985-01-01" {
		t.Fatalf("got %q", got)
	}
}

func TestHashNISS_Stable(t *testing.T) {
	a := HashNISS("96.07.23-998.28", "secret")
	b := HashNISS("96072399828", "secret")
	if a != b || a == "" {
		t.Fatalf("hash mismatch %q %q", a, b)
	}
	if HashNISS("96072399828", "other") == a {
		t.Fatal("secret must change hash")
	}
}
