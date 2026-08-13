package eid

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

const MaxViewerBytes = 5 << 20 // 5 MiB

var (
	ErrEmptyFile      = errors.New("eid_file_empty")
	ErrFileTooLarge   = errors.New("eid_file_too_large")
	ErrUnrecognized   = errors.New("eid_format_unrecognized")
	ErrInvalidXML     = errors.New("eid_xml_invalid")
	ErrPDFUnsupported = errors.New("eid_pdf_unsupported")
	nissBlobRe        = regexp.MustCompile(`\b(\d{11})\b`)
	dateISORe         = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)
	dateSlashRe       = regexp.MustCompile(`^(\d{2})/(\d{2})/(\d{4})$`)
)

// ParseViewerExport parses an eID Viewer .eid/.xml export (PDF rejected in v1).
func ParseViewerExport(raw []byte, filename string) (Identity, error) {
	if len(raw) == 0 {
		return Identity{}, ErrEmptyFile
	}
	if len(raw) > MaxViewerBytes {
		return Identity{}, ErrFileTooLarge
	}
	ext := strings.ToLower(strings.TrimPrefix(filepathExt(filename), "."))
	trimmed := bytes.TrimSpace(raw)
	if ext == "pdf" || bytes.HasPrefix(trimmed, []byte("%PDF")) {
		return Identity{}, ErrPDFUnsupported
	}
	if ext == "eid" || ext == "xml" || bytes.HasPrefix(trimmed, []byte("<?xml")) {
		return parseXMLExport(raw)
	}
	return Identity{}, ErrUnrecognized
}

func filepathExt(name string) string {
	i := strings.LastIndex(name, ".")
	if i < 0 {
		return ""
	}
	return name[i:]
}

func parseXMLExport(raw []byte) (Identity, error) {
	mapLower, err := collectElementTextByLocalName(raw)
	if err != nil {
		return Identity{}, err
	}
	if len(mapLower) == 0 {
		return Identity{}, ErrInvalidXML
	}
	var textParts []string
	for _, v := range mapLower {
		textParts = append(textParts, v)
	}
	textBlob := strings.Join(textParts, "\n")

	lastname := pickFirst(mapLower, "surname", "achternaam", "nom", "lastname")
	firstname := pickFirst(mapLower, "firstname", "givenname", "voornaam", "prenom", "firstnames")
	firstnames := pickFirst(mapLower, "firstnames", "voornamen", "givennames", "firstname", "voornaam")
	if firstnames == "" {
		firstnames = firstname
	}

	niss := DigitsOnly(pickFirst(mapLower,
		"nationalnumber", "national_number", "rijksregisternummer", "nnin",
		"insz", "niss", "rrn",
	))
	// Named fields: keep digits even if checksum fails (demo / Viewer quirks).
	// Free-text blob fallback: only accept a mod-97 valid NISS (avoid false positives).
	if niss == "" {
		if m := nissBlobRe.FindStringSubmatch(textBlob); len(m) == 2 && ValidBelgianNISS(m[1]) {
			niss = m[1]
		}
	}

	birthRaw := pickFirst(mapLower,
		"dateofbirth", "date_of_birth", "geboortedatum", "birthdate", "birth_date", "naissance",
	)
	birthDate := normalizeDate(birthRaw)
	if birthDate == "" {
		birthDate = BirthDateFromNISS(niss)
	}

	nationality := pickFirst(mapLower, "nationality", "nationaliteit", "nationalite")
	if len(nationality) > 2 {
		nationality = nationalityToISO2(nationality)
	}
	if nationality == "" {
		nationality = "BE"
	} else {
		nationality = strings.ToUpper(nationality)
	}

	photo := pickFirst(mapLower, "photo", "picture", "image", "photograph")
	photo = stripWhitespace(photo)

	street := pickFirst(mapLower,
		"addressstreetandnumber", "address_street", "street", "straat", "adres", "address",
	)
	zip := pickFirst(mapLower, "addresszip", "zip", "postalcode", "postcode", "codepostal", "code_postal")
	city := pickFirst(mapLower, "addressmunicipality", "municipality", "city", "gemeente", "ville")

	id := Identity{
		Lastname:          TitleCase(lastname),
		Firstname:         TitleCase(firstname),
		Firstnames:        TitleCase(firstnames),
		BirthDate:         birthDate,
		BirthPlace:        TitleCase(pickFirst(mapLower, "placeofbirth", "birthplace", "geboorteplaats", "lieunaissance")),
		Gender:            pickFirst(mapLower, "gender", "geslacht", "sexe"),
		Country:           nationality,
		NISS:              niss,
		CardNumber:        pickFirst(mapLower, "cardnumber", "documentnumber", "kaartnummer", "numcarte"),
		ValidityFrom:      normalizeDate(pickFirst(mapLower, "cardvaliditybegin", "validfrom", "valid_from", "begingeldigheid")),
		ValidityTo:        normalizeDate(pickFirst(mapLower, "cardvalidityend", "validto", "valid_to", "eindegeldigheid")),
		AddressStreet:     TitleCase(street),
		AddressZip:        zip,
		AddressCity:       TitleCase(city),
		SignatureVerified: false,
		ImportTool:        "eid_viewer_xml",
	}
	if photo != "" {
		id.PhotoJPEGBase64 = photo
	}
	return id, nil
}

// collectElementTextByLocalName mirrors Ventura: last non-empty text per localName wins.
func collectElementTextByLocalName(raw []byte) (map[string]string, error) {
	dec := xml.NewDecoder(bytes.NewReader(raw))
	dec.CharsetReader = func(_ string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	out := map[string]string{}
	type frame struct {
		name    string
		builder *strings.Builder
	}
	var stack []*frame

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidXML, err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			ln := strings.ToLower(t.Name.Local)
			stack = append(stack, &frame{name: ln, builder: &strings.Builder{}})
		case xml.EndElement:
			if len(stack) == 0 {
				continue
			}
			fr := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if fr.name == "" || strings.Contains(fr.name, "signature") {
				continue
			}
			text := strings.TrimSpace(fr.builder.String())
			if text == "" || len(text) > 20000 {
				continue
			}
			out[fr.name] = text
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			stack[len(stack)-1].builder.Write(t)
		}
	}
	return out, nil
}

func stripWhitespace(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func pickFirst(m map[string]string, keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(m[strings.ToLower(k)]); v != "" {
			return v
		}
	}
	return ""
}

func normalizeDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if dateISORe.MatchString(raw) {
		return raw
	}
	if m := dateSlashRe.FindStringSubmatch(raw); len(m) == 4 {
		return fmt.Sprintf("%s-%s-%s", m[3], m[2], m[1])
	}
	return ""
}

func nationalityToISO2(label string) string {
	l := strings.ToLower(label)
	switch {
	case strings.Contains(l, "belg"):
		return "BE"
	case strings.Contains(l, "fran"):
		return "FR"
	default:
		if len(label) == 2 && unicode.IsLetter(rune(label[0])) {
			return strings.ToUpper(label)
		}
		return ""
	}
}
