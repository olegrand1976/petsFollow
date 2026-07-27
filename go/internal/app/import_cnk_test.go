package app

import (
	"strings"
	"testing"
)

func TestParseCNKCSV(t *testing.T) {
	csv := "cnk;name;atc_code;is_antibiotic;form;pack\n" +
		"1234567;Amoxicilline 500mg;J01CA04;true;cp;20\n" +
		"7654321;Vaccin rage;;;inj;1\n" +
		";skip empty cnk;;;;\n"
	rows, err := ParseCNKCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows %#v", len(rows), rows)
	}
	if rows[0].CNK != "1234567" || !rows[0].IsAntibiotic || rows[0].ATCCode != "J01CA04" {
		t.Fatalf("row0 %#v", rows[0])
	}
	if rows[1].IsAntibiotic || rows[1].Name != "Vaccin rage" {
		t.Fatalf("row1 %#v", rows[1])
	}
}

func TestParseCNKCSVComma(t *testing.T) {
	csv := "cnk,name,is_antibiotic\n999,Test Med,0\n"
	rows, err := ParseCNKCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].CNK != "999" || rows[0].IsAntibiotic {
		t.Fatalf("%#v", rows)
	}
}
