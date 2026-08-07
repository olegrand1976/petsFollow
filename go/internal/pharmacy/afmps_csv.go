package pharmacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

const (
	MaxAFMPSCSVBytes     = 100 << 20 // 100 MiB
	AFMPSHardErrorMaxPct = 5.0
	// AFMPSConfirmPhrase is the gate-3 confirmation (CLI: quote or use IMPORT_AFMPS).
	AFMPSConfirmPhrase = "IMPORT AFMPS"
)

// Belgian CNK is typically 7 digits; allow 6–8 to absorb leading-zero variants.
var cnkDigitsRe = regexp.MustCompile(`^\d{6,8}$`)

// ConfirmAFMPSPhrase reports whether s matches the gate-3 confirm token.
func ConfirmAFMPSPhrase(s string) bool {
	s = strings.TrimSpace(s)
	return s == AFMPSConfirmPhrase || s == "IMPORT_AFMPS"
}

// AFMPSParsedRow is one validated/deduped catalogue line ready for staging.
type AFMPSParsedRow struct {
	SourceLine         int               `json:"sourceLine"`
	CNK                string            `json:"cnk"`
	Name               string            `json:"name"`
	ATCCode            string            `json:"atcCode"`
	PharmaceuticalForm string            `json:"pharmaceuticalForm"`
	PackSize           string            `json:"packSize"`
	AMMNumber          string            `json:"ammNumber"`
	IsAntibiotic       bool              `json:"isAntibiotic"`
	Commercialized     bool              `json:"commercialized"`
	Meta               map[string]string `json:"meta,omitempty"`
	Status             string            `json:"status"` // ready | error
	ErrorCode          string            `json:"errorCode,omitempty"`
	ErrorMessage       string            `json:"errorMessage,omitempty"`
}

// AFMPSValidateReport is the gate-1 quality report (no DB write).
type AFMPSValidateReport struct {
	Filename          string           `json:"filename"`
	FileBytes         int              `json:"fileBytes"`
	ChecksumSHA256    string           `json:"checksumSha256"`
	TotalDataLines    int              `json:"totalDataLines"`
	SkippedEmptyCNK   int              `json:"skippedEmptyCnk"`
	DedupDropped      int              `json:"dedupDropped"`
	ReadyCount        int              `json:"readyCount"`
	ErrorCount        int              `json:"errorCount"`
	UniqueCNK         int              `json:"uniqueCnk"`
	CommercializedPct float64          `json:"commercializedPct"`
	HardErrorPct      float64          `json:"hardErrorPct"`
	Blocked           bool             `json:"blocked"`
	BlockReason       string           `json:"blockReason,omitempty"`
	Sample            []AFMPSParsedRow `json:"sample,omitempty"`
}

// ParseAndValidateAFMPSCSV parses AFMPS pack CSV (or simple cnk/name CSV) and builds a gate-1 report.
func ParseAndValidateAFMPSCSV(r io.Reader, filename string) ([]AFMPSParsedRow, AFMPSValidateReport, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, AFMPSValidateReport{}, fmt.Errorf("read csv: %w", err)
	}
	rep := AFMPSValidateReport{
		Filename:       strings.TrimSpace(filename),
		FileBytes:      len(raw),
		ChecksumSHA256: sha256Hex(raw),
	}
	if len(raw) == 0 {
		rep.Blocked = true
		rep.BlockReason = "empty_file"
		return nil, rep, fmt.Errorf("empty_file")
	}
	if len(raw) > MaxAFMPSCSVBytes {
		rep.Blocked = true
		rep.BlockReason = "file_too_large"
		return nil, rep, fmt.Errorf("file_too_large")
	}

	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	firstLine := raw
	if before, _, ok := bytes.Cut(raw, []byte{'\n'}); ok {
		firstLine = before
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
	br.ReuseRecord = false

	headers, err := br.Read()
	if err != nil {
		rep.Blocked = true
		rep.BlockReason = "read_header_failed"
		return nil, rep, fmt.Errorf("read header: %w", err)
	}
	idx := mapAFMPSHeaders(headers)
	if idx["cnk"] < 0 || idx["name"] < 0 {
		rep.Blocked = true
		rep.BlockReason = "missing_cnk_or_name_columns"
		return nil, rep, fmt.Errorf("CSV must include cnk/Code CNK and name/Nom columns (got %v)", headers)
	}

	vetOnly := idx["usage"] >= 0 // filter when Usage column present
	byCNK := map[string]AFMPSParsedRow{}
	cnkOrder := make([]string, 0)
	lineNo := 1 // header

	for {
		rec, err := br.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			rep.Blocked = true
			rep.BlockReason = "csv_parse_error"
			return nil, rep, fmt.Errorf("csv parse line %d: %w", lineNo+1, err)
		}
		lineNo++
		rep.TotalDataLines++

		if vetOnly {
			usage := fieldAt(rec, idx["usage"])
			if usage != "" && !strings.Contains(strings.ToLower(usage), "vétérinaire") &&
				!strings.Contains(strings.ToLower(usage), "veterinaire") &&
				!strings.Contains(strings.ToLower(usage), "veterinary") {
				continue
			}
		}

		row := rowFromAFMPSRecord(rec, idx, lineNo)
		if row.CNK == "" {
			rep.SkippedEmptyCNK++
			continue
		}

		prev, exists := byCNK[row.CNK]
		if !exists {
			byCNK[row.CNK] = row
			cnkOrder = append(cnkOrder, row.CNK)
			continue
		}
		// Prefer commercialized; otherwise keep first.
		if row.Commercialized && !prev.Commercialized {
			byCNK[row.CNK] = row
		}
		rep.DedupDropped++
	}

	out := make([]AFMPSParsedRow, 0, len(cnkOrder))
	commercialized := 0
	for _, cnk := range cnkOrder {
		row := byCNK[cnk]
		if row.Status == "ready" {
			rep.ReadyCount++
			if row.Commercialized {
				commercialized++
			}
		} else {
			rep.ErrorCount++
		}
		out = append(out, row)
	}
	rep.UniqueCNK = len(out)
	if rep.UniqueCNK > 0 {
		rep.CommercializedPct = float64(commercialized) * 100 / float64(rep.UniqueCNK)
		rep.HardErrorPct = float64(rep.ErrorCount) * 100 / float64(rep.UniqueCNK)
	}
	sampleN := min(20, len(out))
	for i := 0; i < sampleN; i++ {
		if out[i].Status == "ready" {
			rep.Sample = append(rep.Sample, out[i])
		}
	}
	if len(rep.Sample) == 0 {
		for i := 0; i < sampleN && i < len(out); i++ {
			rep.Sample = append(rep.Sample, out[i])
		}
	}

	switch {
	case rep.ReadyCount == 0:
		rep.Blocked = true
		rep.BlockReason = "no_ready_rows"
	case rep.HardErrorPct > AFMPSHardErrorMaxPct:
		rep.Blocked = true
		rep.BlockReason = "hard_error_rate_exceeded"
	}

	if rep.Blocked {
		return out, rep, fmt.Errorf("%s", rep.BlockReason)
	}
	return out, rep, nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func mapAFMPSHeaders(headers []string) map[string]int {
	idx := map[string]int{
		"cnk": -1, "name": -1, "atc": -1, "antibiotic": -1, "form": -1, "pack": -1,
		"amm": -1, "firme": -1, "usage": -1, "species": -1, "withdrawal": -1,
		"substance": -1, "delivery": -1, "cti": -1, "commercialized": -1,
	}
	for i, h := range headers {
		key := normalizeHeaderKey(h)
		switch key {
		case "cnk", "code_cnk", "cnk_code":
			idx["cnk"] = i
		case "name", "nom", "product_name", "denomination":
			idx["name"] = i
		case "atc", "atc_code", "code_atc":
			idx["atc"] = i
		case "is_antibiotic", "antibiotic", "antibiotique":
			idx["antibiotic"] = i
		case "pharmaceutical_form", "form", "forme", "forme_pharmaceutique":
			idx["form"] = i
		case "pack_size", "pack", "conditionnement":
			idx["pack"] = i
		case "amm", "amm_number", "numero_dautorisation", "numero_d_autorisation",
			"numero_dauthorisation", "numero_d_authorisation":
			idx["amm"] = i
		case "firme", "manufacturer", "lab", "labo", "titulaire":
			idx["firme"] = i
		case "usage_humainveterinaire", "usage_humain_veterinaire":
			idx["usage"] = i
		case "especes_cibles", "species":
			idx["species"] = i
		case "temps_dattente", "temps_d_attente", "withdrawal":
			idx["withdrawal"] = i
		case "substance_active", "active_substance":
			idx["substance"] = i
		case "mode_de_delivrance":
			idx["delivery"] = i
		case "cti_extended", "cti":
			idx["cti"] = i
		case "commercialise", "commercialized":
			idx["commercialized"] = i
		default:
			if (strings.Contains(key, "autorisation") || strings.Contains(key, "authorisation")) &&
				strings.Contains(key, "numero") {
				idx["amm"] = i
			}
			if strings.Contains(key, "usage") && (strings.Contains(key, "veterinaire") || strings.Contains(key, "humain")) {
				idx["usage"] = i
			}
			if strings.Contains(key, "especes") {
				idx["species"] = i
			}
			if strings.Contains(key, "attente") {
				idx["withdrawal"] = i
			}
			if strings.Contains(key, "substance") {
				idx["substance"] = i
			}
			if strings.Contains(key, "delivrance") {
				idx["delivery"] = i
			}
		}
	}
	return idx
}

// normalizeHeaderKey lowercases, strips diacritics, maps punctuation to underscores.
func normalizeHeaderKey(h string) string {
	h = strings.TrimSpace(h)
	h = strings.TrimPrefix(h, "\ufeff")
	h = strings.ToLower(h)
	var b strings.Builder
	prevUnderscore := false
	for _, r := range h {
		r = unicode.ToLower(stripDiacritic(r))
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			prevUnderscore = false
		default:
			if !prevUnderscore && b.Len() > 0 {
				b.WriteByte('_')
				prevUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}

func stripDiacritic(r rune) rune {
	switch r {
	case 'à', 'á', 'â', 'ä', 'ã':
		return 'a'
	case 'è', 'é', 'ê', 'ë':
		return 'e'
	case 'ì', 'í', 'î', 'ï':
		return 'i'
	case 'ò', 'ó', 'ô', 'ö', 'õ':
		return 'o'
	case 'ù', 'ú', 'û', 'ü':
		return 'u'
	case 'ç':
		return 'c'
	case 'ñ':
		return 'n'
	default:
		return r
	}
}

func fieldAt(rec []string, i int) string {
	if i < 0 || i >= len(rec) {
		return ""
	}
	return strings.TrimSpace(rec[i])
}

func rowFromAFMPSRecord(rec []string, idx map[string]int, lineNo int) AFMPSParsedRow {
	get := func(k string) string { return fieldAt(rec, idx[k]) }
	cnk := strings.TrimSpace(get("cnk"))
	name := strings.TrimSpace(get("name"))
	atcRaw := get("atc")
	atc := StripATCCode(atcRaw)
	form := get("form")
	pack := get("pack")
	amm := get("amm")
	firme := get("firme")
	commercialized := parseOuiNon(get("commercialized"))

	meta := map[string]string{"source": "afmps-pack-csv"}
	if firme != "" {
		meta["manufacturer"] = firme
	}
	if s := get("species"); s != "" {
		meta["target_species"] = s
	}
	if s := get("withdrawal"); s != "" {
		meta["withdrawal_text"] = s
	}
	if s := get("substance"); s != "" {
		meta["active_substance"] = s
	}
	if s := get("delivery"); s != "" {
		meta["delivery_mode"] = s
	}
	if s := get("cti"); s != "" {
		meta["cti_extended"] = s
	}
	if get("commercialized") != "" {
		if commercialized {
			meta["commercialized"] = "Oui"
		} else {
			meta["commercialized"] = "Non"
		}
	}
	if atcRaw != "" && atcRaw != atc {
		meta["atc_raw"] = atcRaw
	}

	ab := parseOuiNon(get("antibiotic"))
	if !ab {
		ab = IsAntibioticATC(atc)
	}

	row := AFMPSParsedRow{
		SourceLine:         lineNo,
		CNK:                cnk,
		Name:               name,
		ATCCode:            atc,
		PharmaceuticalForm: form,
		PackSize:           pack,
		AMMNumber:          amm,
		IsAntibiotic:       ab,
		Commercialized:     commercialized,
		Meta:               meta,
		Status:             "ready",
	}
	if cnk == "" {
		return row
	}
	if name == "" {
		row.Status = "error"
		row.ErrorCode = "missing_name"
		row.ErrorMessage = "name required"
		return row
	}
	if !cnkDigitsRe.MatchString(cnk) {
		row.Status = "error"
		row.ErrorCode = "invalid_cnk"
		row.ErrorMessage = "cnk must be 6–8 digits"
		return row
	}
	return row
}

// StripATCCode keeps the first ATC token (e.g. "QG02AD90 - foo" → "QG02AD90").
func StripATCCode(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Split on whitespace, dash, semicolon, slash
	for _, sep := range []string{"\n", "\r", ";", "/", " - ", " – ", " — ", "\t", " "} {
		if i := strings.Index(s, sep); i > 0 {
			s = s[:i]
			break
		}
	}
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)
	return s
}

// IsAntibioticATC returns true for ATC codes in antibiotic groups (human J01 / vet QJ01).
func IsAntibioticATC(atc string) bool {
	atc = strings.ToUpper(strings.TrimSpace(atc))
	return strings.HasPrefix(atc, "QJ01") || strings.HasPrefix(atc, "J01")
}

func parseOuiNon(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "1", "true", "t", "yes", "y", "oui", "o":
		return true
	}
	return false
}

// AFMPSMetaJSON marshals row meta for storage / upsert.
func AFMPSMetaJSON(meta map[string]string) json.RawMessage {
	if len(meta) == 0 {
		return json.RawMessage(`{"source":"afmps-pack-csv"}`)
	}
	b, err := json.Marshal(meta)
	if err != nil {
		return json.RawMessage(`{"source":"afmps-pack-csv"}`)
	}
	return b
}
