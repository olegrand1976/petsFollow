// Package eid parses Belgian eID Viewer exports and validates Web eID tokens.
package eid

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"
)

// Identity is the normalized payload returned to the Pro UI for form prefill.
type Identity struct {
	Lastname           string `json:"lastname,omitempty"`
	Firstname          string `json:"firstname,omitempty"`
	Firstnames         string `json:"firstnames,omitempty"`
	BirthDate          string `json:"birth_date,omitempty"`
	BirthPlace         string `json:"birth_place,omitempty"`
	Gender             string `json:"gender,omitempty"`
	Country            string `json:"country,omitempty"`
	NISS               string `json:"niss,omitempty"`
	CardNumber         string `json:"card_number,omitempty"`
	ValidityFrom       string `json:"validity_from,omitempty"`
	ValidityTo         string `json:"validity_to,omitempty"`
	AddressStreet      string `json:"address_street,omitempty"`
	AddressNumber      string `json:"address_number,omitempty"`
	AddressZip         string `json:"address_zip,omitempty"`
	AddressCity        string `json:"address_city,omitempty"`
	SignatureVerified  bool   `json:"signature_verified"`
	ImportTool         string `json:"import_tool,omitempty"`
	PhotoJPEGBase64    string `json:"photo_jpeg_base64,omitempty"`
}

// NonEmptyFieldKeys lists populated identity field names (for audit).
func (id Identity) NonEmptyFieldKeys() []string {
	keys := make([]string, 0, 16)
	add := func(k, v string) {
		if strings.TrimSpace(v) != "" {
			keys = append(keys, k)
		}
	}
	add("lastname", id.Lastname)
	add("firstname", id.Firstname)
	add("firstnames", id.Firstnames)
	add("birth_date", id.BirthDate)
	add("birth_place", id.BirthPlace)
	add("gender", id.Gender)
	add("nationality", id.Country)
	add("niss", id.NISS)
	add("card_number", id.CardNumber)
	add("validity_from", id.ValidityFrom)
	add("validity_to", id.ValidityTo)
	add("address_street", id.AddressStreet)
	add("address_number", id.AddressNumber)
	add("address_zip", id.AddressZip)
	add("address_city", id.AddressCity)
	if strings.TrimSpace(id.PhotoJPEGBase64) != "" {
		keys = append(keys, "photo_jpeg")
	}
	return keys
}

// HasUsefulIdentity reports whether the payload can prefill a client form.
func (id Identity) HasUsefulIdentity() bool {
	return strings.TrimSpace(id.Lastname) != "" ||
		strings.TrimSpace(id.Firstname) != "" ||
		strings.TrimSpace(id.NISS) != ""
}

// PublicPrefill returns a copy safe to send to the browser (no photo bytes).
func (id Identity) PublicPrefill() Identity {
	out := id
	out.PhotoJPEGBase64 = ""
	return out
}

// DigitsOnly strips non-digit characters from a NISS / national number.
func DigitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// BirthDateFromNISS derives YYYY-MM-DD from the Belgian NISS YYMMDD prefix.
func BirthDateFromNISS(niss string) string {
	niss = DigitsOnly(niss)
	if len(niss) < 6 {
		return ""
	}
	year := atoi2(niss[0:2])
	month := atoi2(niss[2:4])
	day := atoi2(niss[4:6])
	currentYY := time.Now().Year() % 100
	fullYear := 1900 + year
	if year <= currentYY {
		fullYear = 2000 + year
	}
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", fullYear, month, day)
}

func atoi2(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		n = n*10 + int(s[i]-'0')
	}
	return n
}

func hashSHA256(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// TitleCase is a light Unicode title-case for names.
func TitleCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	runes := []rune(strings.ToLower(s))
	upperNext := true
	for i, r := range runes {
		if upperNext && unicode.IsLetter(r) {
			runes[i] = unicode.ToTitle(r)
			upperNext = false
			continue
		}
		if r == '-' || r == ' ' || r == '\'' {
			upperNext = true
		}
	}
	return string(runes)
}

// ExtractNISSFromEIDAS strips the PNOBE- prefix from an eIDAS serialNumber.
func ExtractNISSFromEIDAS(idCode string) string {
	idCode = strings.TrimSpace(idCode)
	const prefix = "PNOBE-"
	if strings.HasPrefix(strings.ToUpper(idCode), prefix) {
		return DigitsOnly(idCode[len(prefix):])
	}
	return DigitsOnly(idCode)
}

// HashNISS returns a SHA-256 hex digest of niss + secret (audit, no raw NISS).
func HashNISS(niss, secret string) string {
	return hashSHA256(DigitsOnly(niss) + ":" + secret)
}
