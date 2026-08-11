package pharmacy

import (
	"strings"
	"testing"
)

func TestParseAndValidateAFMPSCSV_FRHeaders(t *testing.T) {
	csv := "Nom;Forme pharmaceutique;Conditionnement;Code CNK;Firme;Numéro d'autorisation;Commercialisé;Substance active;Code ATC;Usage Humain/Vétérinaire\n" +
		"Vetoryl 30 mg;Gélule;30 gélules;1234567;Dechra;BE-V123456;Oui;Trilostane;QH02CA01 - foo;Usage vétérinaire\n" +
		"Skip No CNK;Gélule;10;;;;Oui;;QH02CA01;Usage vétérinaire\n" +
		"Vetoryl 30 mg dup;Gélule;30 gélules;1234567;Dechra;BE-V123456;Non;Trilostane;QH02CA01;Usage vétérinaire\n" +
		"Amoxi Vet;Comprimé;20;7654321;Lab;BE-V999;Oui;Amox;QJ01CA04;Usage vétérinaire\n" +
		"Human Only;cp;10;1111111;X;BE-H;Oui;Y;J01CA04;Usage humain\n"

	rows, report, err := ParseAndValidateAFMPSCSV(strings.NewReader(csv), "pack.csv")
	if err != nil {
		t.Fatal(err)
	}
	if report.Blocked {
		t.Fatalf("blocked: %#v", report)
	}
	if report.SkippedEmptyCNK != 1 {
		t.Fatalf("skipped empty=%d", report.SkippedEmptyCNK)
	}
	if report.DedupDropped != 1 {
		t.Fatalf("dedup=%d", report.DedupDropped)
	}
	// vet-only filter drops human; dedup keeps commercialized Vetoryl
	if report.ReadyCount != 2 {
		t.Fatalf("ready=%d rows=%#v", report.ReadyCount, rows)
	}
	var vetoryl *AFMPSParsedRow
	for i := range rows {
		if rows[i].CNK == "1234567" {
			vetoryl = &rows[i]
		}
		if rows[i].CNK == "7654321" && !rows[i].IsAntibiotic {
			t.Fatalf("expected antibiotic from QJ01: %#v", rows[i])
		}
	}
	if vetoryl == nil {
		t.Fatal("missing vetoryl")
	}
	if vetoryl.ATCCode != "QH02CA01" {
		t.Fatalf("atc strip got %q", vetoryl.ATCCode)
	}
	if vetoryl.PharmaceuticalForm != "Gélule" || vetoryl.AMMNumber != "BE-V123456" {
		t.Fatalf("%#v", vetoryl)
	}
	if vetoryl.Meta["manufacturer"] != "Dechra" {
		t.Fatalf("meta %#v", vetoryl.Meta)
	}
	if !vetoryl.Commercialized {
		t.Fatal("expected commercialized preferred on dedup")
	}
}

func TestParseAndValidateAFMPSCSV_InvalidCNKBlockedRate(t *testing.T) {
	csv := "cnk;name\n" +
		"abc;Bad1\n" +
		"xyz;Bad2\n" +
		"12;Bad3\n" // too short
	_, report, err := ParseAndValidateAFMPSCSV(strings.NewReader(csv), "bad.csv")
	if err == nil {
		t.Fatal("expected error")
	}
	if !report.Blocked || report.BlockReason != "no_ready_rows" {
		t.Fatalf("%#v", report)
	}
}

func TestParseAndValidateAFMPSCSV_HardErrorRate(t *testing.T) {
	// 1 ready + 1 error → 50% > 5%
	csv := "cnk;name\n" +
		"1234567;Good Med\n" +
		"notadigit;Bad Med\n"
	_, report, err := ParseAndValidateAFMPSCSV(strings.NewReader(csv), "mix.csv")
	if err == nil {
		t.Fatal("expected hard error rate block")
	}
	if report.BlockReason != "hard_error_rate_exceeded" {
		t.Fatalf("%#v", report)
	}
}

func TestAFMPSChecksum_StripsBOM(t *testing.T) {
	body := []byte("cnk;name\n1234567;Med\n")
	withBOM := append([]byte{0xEF, 0xBB, 0xBF}, body...)
	if AFMPSChecksum(body) != AFMPSChecksum(withBOM) {
		t.Fatalf("BOM must not change checksum: %s vs %s", AFMPSChecksum(body), AFMPSChecksum(withBOM))
	}
	_, repNoBOM, err := ParseAndValidateAFMPSCSV(strings.NewReader(string(body)), "a.csv")
	if err != nil {
		t.Fatal(err)
	}
	_, repBOM, err := ParseAndValidateAFMPSCSV(strings.NewReader(string(withBOM)), "b.csv")
	if err != nil {
		t.Fatal(err)
	}
	if repNoBOM.ChecksumSHA256 != repBOM.ChecksumSHA256 || repNoBOM.ChecksumSHA256 != AFMPSChecksum(body) {
		t.Fatalf("report checksum mismatch: noBOM=%s bom=%s helper=%s",
			repNoBOM.ChecksumSHA256, repBOM.ChecksumSHA256, AFMPSChecksum(body))
	}
}

func TestStripATCCode(t *testing.T) {
	if got := StripATCCode("QG02AD90 - description"); got != "QG02AD90" {
		t.Fatalf("%q", got)
	}
	if got := StripATCCode("qj01ca04"); got != "QJ01CA04" {
		t.Fatalf("%q", got)
	}
}

func TestConfirmAFMPSPhrase(t *testing.T) {
	if !ConfirmAFMPSPhrase("IMPORT AFMPS") || !ConfirmAFMPSPhrase("IMPORT_AFMPS") {
		t.Fatal("expected accept")
	}
	if ConfirmAFMPSPhrase("yes") {
		t.Fatal("expected reject")
	}
}
