package app

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/db"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// ImportCNKOnly loads a CSV into pharmacy.ref_medications.
// Usage: petsfollow-api import-cnk --file=path.csv [--deactivate-missing] [--dry-run]
func ImportCNKOnly(ctx context.Context, cfg config.Config, args []string) error {
	filePath, deactivateMissing, dryRun, err := parseImportCNKArgs(args)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	rows, err := ParseCNKCSV(bytes.NewReader(raw))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return fmt.Errorf("no data rows in CSV (need cnk + name columns)")
	}
	if dryRun {
		fmt.Printf("dry-run: %d rows would be upserted (deactivate-missing=%v)\n", len(rows), deactivateMissing)
		return nil
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	st := store.New(pool)

	var upserted int
	keep := make([]string, 0, len(rows))
	for _, row := range rows {
		if _, err := st.UpsertRefMedication(ctx, row); err != nil {
			return fmt.Errorf("upsert cnk=%s: %w", row.CNK, err)
		}
		upserted++
		keep = append(keep, row.CNK)
	}
	var deactivated int64
	if deactivateMissing {
		deactivated, err = st.DeactivateMissingRefMedications(ctx, keep)
		if err != nil {
			return fmt.Errorf("deactivate-missing: %w", err)
		}
	}
	fmt.Printf("import-cnk OK: upserted=%d deactivated=%d\n", upserted, deactivated)
	return nil
}

func parseImportCNKArgs(args []string) (file string, deactivate, dryRun bool, err error) {
	for _, a := range args[1:] {
		switch {
		case strings.HasPrefix(a, "--file="):
			file = strings.TrimPrefix(a, "--file=")
		case a == "--deactivate-missing":
			deactivate = true
		case a == "--dry-run":
			dryRun = true
		case a == "-h" || a == "--help":
			err = fmt.Errorf("usage: import-cnk --file=path.csv [--deactivate-missing] [--dry-run]")
			return
		default:
			err = fmt.Errorf("unknown arg %q (usage: import-cnk --file=path.csv [--deactivate-missing] [--dry-run])", a)
			return
		}
	}
	if strings.TrimSpace(file) == "" {
		err = fmt.Errorf("missing --file= (usage: import-cnk --file=path.csv [--deactivate-missing] [--dry-run])")
	}
	return
}

// ParseCNKCSV reads a semicolon- or comma-separated AFMPS-like CSV.
// Required headers (case-insensitive): cnk, name.
// Optional: atc_code|atc, is_antibiotic|antibiotic, pharmaceutical_form|form, pack_size|pack.
func ParseCNKCSV(r io.Reader) ([]store.RefMedicationUpsert, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	firstLine := raw
	if i := bytes.IndexByte(raw, '\n'); i >= 0 {
		firstLine = raw[:i]
	}
	comma := ','
	if bytes.Count(firstLine, []byte(";")) > bytes.Count(firstLine, []byte(",")) {
		comma = ';'
	}

	br := csv.NewReader(bytes.NewReader(raw))
	br.Comma = comma
	br.LazyQuotes = true
	br.TrimLeadingSpace = true
	br.FieldsPerRecord = -1

	headers, err := br.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	idx := mapCNKHeaders(headers)
	if idx["cnk"] < 0 || idx["name"] < 0 {
		return nil, fmt.Errorf("CSV must include cnk and name columns (got %v)", headers)
	}

	var out []store.RefMedicationUpsert
	for {
		rec, err := br.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		row, ok := rowFromCNKRecord(rec, idx)
		if !ok {
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

func mapCNKHeaders(headers []string) map[string]int {
	idx := map[string]int{"cnk": -1, "name": -1, "atc": -1, "antibiotic": -1, "form": -1, "pack": -1}
	for i, h := range headers {
		key := strings.ToLower(strings.TrimSpace(h))
		key = strings.ReplaceAll(key, " ", "_")
		switch key {
		case "cnk", "code_cnk", "cnk_code":
			idx["cnk"] = i
		case "name", "nom", "product_name", "denomination":
			idx["name"] = i
		case "atc", "atc_code", "code_atc":
			idx["atc"] = i
		case "is_antibiotic", "antibiotic", "antibiotique":
			idx["antibiotic"] = i
		case "pharmaceutical_form", "form", "forme":
			idx["form"] = i
		case "pack_size", "pack", "conditionnement":
			idx["pack"] = i
		}
	}
	return idx
}

func rowFromCNKRecord(rec []string, idx map[string]int) (store.RefMedicationUpsert, bool) {
	get := func(key string) string {
		i := idx[key]
		if i < 0 || i >= len(rec) {
			return ""
		}
		return strings.TrimSpace(rec[i])
	}
	cnk := get("cnk")
	name := get("name")
	if cnk == "" || name == "" {
		return store.RefMedicationUpsert{}, false
	}
	meta, _ := json.Marshal(map[string]string{"source": "import-cnk"})
	return store.RefMedicationUpsert{
		CNK:                cnk,
		Name:               name,
		ATCCode:            get("atc"),
		PharmaceuticalForm: get("form"),
		PackSize:           get("pack"),
		IsAntibiotic:       parseBoolish(get("antibiotic")),
		IsActive:           true,
		AFMPSMeta:          meta,
	}, true
}

func parseBoolish(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "1", "true", "t", "yes", "y", "oui", "o":
		return true
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n != 0
	}
	return false
}
