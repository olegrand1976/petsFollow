package spreadsheet

import (
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestParseCSVComma(t *testing.T) {
	data := []byte("Email,Nom\nclient@test.com,Alice\n")
	got, err := Parse(data, "csv")
	if err != nil {
		t.Fatal(err)
	}
	if got.Format != "csv" || len(got.Headers) != 2 || len(got.Rows) != 1 {
		t.Fatalf("unexpected %#v", got)
	}
	if got.Rows[0]["Email"] != "client@test.com" || got.Rows[0]["Nom"] != "Alice" {
		t.Fatalf("row %#v", got.Rows[0])
	}
}

func TestParseCSVSemicolon(t *testing.T) {
	data := []byte("Courriel;Nom client\na@b.co;Bob\n")
	got, err := Parse(data, "csv")
	if err != nil {
		t.Fatal(err)
	}
	if got.Rows[0]["Courriel"] != "a@b.co" {
		t.Fatalf("got %#v", got.Rows[0])
	}
}

func TestParseXLSX(t *testing.T) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetCellValue(sheet, "A1", "email")
	_ = f.SetCellValue(sheet, "B1", "fullName")
	_ = f.SetCellValue(sheet, "A2", "x@y.z")
	_ = f.SetCellValue(sheet, "B2", "X Y")
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(buf.Bytes(), "xlsx")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) != 1 || got.Rows[0]["email"] != "x@y.z" {
		t.Fatalf("got %#v", got)
	}
}

func TestDetectFormat(t *testing.T) {
	if DetectFormat("clients.csv", "text/csv") != "csv" {
		t.Fatal("csv")
	}
	if DetectFormat("clients.xlsx", "") != "xlsx" {
		t.Fatal("xlsx")
	}
	if DetectFormat("notes.txt", "text/plain") != "csv" {
		t.Fatal("plain as csv")
	}
}
