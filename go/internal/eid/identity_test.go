package eid

import (
	"strings"
	"testing"
)

func TestPublicPrefill_StripsUnusedPHI(t *testing.T) {
	id := Identity{
		Firstname:       "Camille",
		Lastname:        "Testeur",
		NISS:            "96072399886",
		BirthDate:       "1996-07-23",
		Gender:          "F",
		CardNumber:      "000",
		PhotoJPEGBase64: strings.Repeat("A", 200),
		AddressCity:     "Bruxelles",
		ImportTool:      "eid_viewer_xml",
	}
	out := id.PublicPrefill()
	if out.PhotoJPEGBase64 != "" || out.BirthDate != "" || out.Gender != "" || out.CardNumber != "" {
		t.Fatalf("unused PHI still present: %#v", out)
	}
	if out.Firstname != "Camille" || out.NISS != "96072399886" || out.AddressCity != "Bruxelles" {
		t.Fatalf("identity fields lost: %#v", out)
	}
	if id.PhotoJPEGBase64 == "" || id.BirthDate == "" {
		t.Fatal("PublicPrefill must not mutate original")
	}
}

func TestValidBelgianNISS(t *testing.T) {
	if !ValidBelgianNISS("96072399886") {
		t.Fatal("expected valid")
	}
	if !ValidBelgianNISS("85.01.01-123.87") {
		t.Fatal("formatted valid niss")
	}
	if ValidBelgianNISS("96072399828") {
		t.Fatal("old fixture checksum must fail")
	}
	if ValidBelgianNISS("123") || ValidBelgianNISS("") {
		t.Fatal("short/empty must fail")
	}
}

func TestHasUsefulIdentity(t *testing.T) {
	if (Identity{}).HasUsefulIdentity() {
		t.Fatal("empty should be useless")
	}
	if !(Identity{NISS: "96072399886"}).HasUsefulIdentity() {
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
  <nationalnumber>85010112387</nationalnumber>
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
	if pub.PhotoJPEGBase64 != "" || pub.BirthDate != "" {
		t.Fatal("public prefill must strip photo and unused PHI")
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

func TestParseViewerExport_BlobNISSRequiresChecksum(t *testing.T) {
	// No named nationalnumber — only an 11-digit run in free text (invalid checksum → ignored).
	raw := []byte(`<?xml version="1.0"?><Export><surname>X</surname><note>ref 96072399828 end</note></Export>`)
	id, err := ParseViewerExport(raw, "blob.eid")
	if err != nil {
		t.Fatal(err)
	}
	if id.NISS != "" {
		t.Fatalf("invalid blob niss accepted: %q", id.NISS)
	}
	rawOK := []byte(`<?xml version="1.0"?><Export><surname>X</surname><note>ref 96072399886 end</note></Export>`)
	id, err = ParseViewerExport(rawOK, "blob.eid")
	if err != nil {
		t.Fatal(err)
	}
	if id.NISS != "96072399886" {
		t.Fatalf("valid blob niss missing: %#v", id)
	}
}

func TestBirthDateFromNISS(t *testing.T) {
	if got := BirthDateFromNISS("96072399886"); got != "1996-07-23" {
		t.Fatalf("got %q", got)
	}
	if got := BirthDateFromNISS("850101"); got != "1985-01-01" {
		t.Fatalf("got %q", got)
	}
}

func TestHashNISS_Stable(t *testing.T) {
	a := HashNISS("96.07.23-998.86", "secret")
	b := HashNISS("96072399886", "secret")
	if a != b || a == "" {
		t.Fatalf("hash mismatch %q %q", a, b)
	}
	if HashNISS("96072399886", "other") == a {
		t.Fatal("secret must change hash")
	}
}
