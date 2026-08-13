package eid

import (
	"embed"
	"errors"
	"testing"
)

//go:embed testdata/sample_valid.eid
var testdataFS embed.FS

func TestParseViewerExport_Sample(t *testing.T) {
	raw, err := testdataFS.ReadFile("testdata/sample_valid.eid")
	if err != nil {
		t.Fatal(err)
	}
	id, err := ParseViewerExport(raw, "sample_valid.eid")
	if err != nil {
		t.Fatal(err)
	}
	if id.Lastname != "Testeur" {
		t.Fatalf("lastname=%q", id.Lastname)
	}
	if id.Firstname != "Camille" {
		t.Fatalf("firstname=%q", id.Firstname)
	}
	if id.NISS != "96072399828" {
		t.Fatalf("niss=%q", id.NISS)
	}
	if id.Country != "BE" {
		t.Fatalf("country=%q", id.Country)
	}
	if id.BirthDate != "1996-07-23" {
		t.Fatalf("birth_date=%q want 1996-07-23 from NISS", id.BirthDate)
	}
	if id.ImportTool != "eid_viewer_xml" {
		t.Fatalf("tool=%q", id.ImportTool)
	}
}

func TestParseViewerExport_PDFRejected(t *testing.T) {
	_, err := ParseViewerExport([]byte("%PDF-1.4"), "card.pdf")
	if !errors.Is(err, ErrPDFUnsupported) {
		t.Fatalf("err=%v", err)
	}
}

func TestExtractNISSFromEIDAS(t *testing.T) {
	if got := ExtractNISSFromEIDAS("PNOBE-96072399828"); got != "96072399828" {
		t.Fatalf("got %q", got)
	}
}
