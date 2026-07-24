package spreadsheet

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
)

const (
	MaxFileBytes = 5 << 20 // 5 MiB
	MaxRows      = 2000
	SampleRows   = 15
)

type ParsedSheet struct {
	Format     string              // csv | xlsx
	Headers    []string
	Rows       []map[string]string // header -> value
	SampleRows []map[string]string
}

func DetectFormat(filename, contentType string) string {
	name := strings.ToLower(filename)
	ct := strings.ToLower(contentType)
	switch {
	case strings.HasSuffix(name, ".xlsx"), strings.HasSuffix(name, ".xls"),
		strings.Contains(ct, "spreadsheet"), strings.Contains(ct, "excel"):
		return "xlsx"
	case strings.HasSuffix(name, ".csv"), strings.Contains(ct, "csv"), strings.Contains(ct, "text/plain"):
		return "csv"
	default:
		return ""
	}
}

func Parse(data []byte, format string) (*ParsedSheet, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty_file")
	}
	if len(data) > MaxFileBytes {
		return nil, fmt.Errorf("file_too_large")
	}
	switch format {
	case "csv":
		return parseCSV(data)
	case "xlsx":
		return parseXLSX(data)
	default:
		return nil, fmt.Errorf("unsupported_format")
	}
}

func parseCSV(data []byte) (*ParsedSheet, error) {
	sep := detectCSVSeparator(data)
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = sep
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv_parse: %w", err)
	}
	return sheetFromRecords(records, "csv")
}

func detectCSVSeparator(data []byte) rune {
	line := data
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		line = data[:i]
	}
	semi := bytes.Count(line, []byte{';'})
	comma := bytes.Count(line, []byte{','})
	if semi > comma {
		return ';'
	}
	return ','
}

func parseXLSX(data []byte) (*ParsedSheet, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("xlsx_parse: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("xlsx_empty")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("xlsx_rows: %w", err)
	}
	return sheetFromRecords(rows, "xlsx")
}

func sheetFromRecords(records [][]string, format string) (*ParsedSheet, error) {
	if len(records) < 2 {
		return nil, fmt.Errorf("no_data_rows")
	}
	headers := normalizeHeaders(records[0])
	if len(headers) == 0 {
		return nil, fmt.Errorf("no_headers")
	}
	body := records[1:]
	if len(body) > MaxRows {
		return nil, fmt.Errorf("too_many_rows")
	}
	out := make([]map[string]string, 0, len(body))
	for _, rec := range body {
		if rowEmpty(rec) {
			continue
		}
		m := make(map[string]string, len(headers))
		for i, h := range headers {
			val := ""
			if i < len(rec) {
				val = strings.TrimSpace(rec[i])
			}
			m[h] = val
		}
		out = append(out, m)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no_data_rows")
	}
	sampleN := SampleRows
	if sampleN > len(out) {
		sampleN = len(out)
	}
	return &ParsedSheet{
		Format:     format,
		Headers:    headers,
		Rows:       out,
		SampleRows: out[:sampleN],
	}, nil
}

func normalizeHeaders(raw []string) []string {
	seen := map[string]int{}
	out := make([]string, 0, len(raw))
	for i, h := range raw {
		h = strings.TrimSpace(h)
		if h == "" {
			h = fmt.Sprintf("col_%d", i+1)
		}
		// Ensure valid UTF-8 labels.
		if !utf8.ValidString(h) {
			h = strings.ToValidUTF8(h, "")
		}
		base := h
		if n, ok := seen[base]; ok {
			seen[base] = n + 1
			h = fmt.Sprintf("%s_%d", base, n+1)
		} else {
			seen[base] = 1
		}
		out = append(out, h)
	}
	return out
}

func rowEmpty(rec []string) bool {
	for _, c := range rec {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}
