package i18n

import (
	"sort"
	"strings"
	"testing"
)

// TestCatalogParity vérifie que chaque catalogue supporté expose EXACTEMENT le
// même jeu de clés que le français (source de vérité). Les tests par famille
// (emails / push / sms) ne couvrent qu'une liste de clés énumérée à la main :
// ce test attrape en plus les clés oubliées lors de l'ajout d'une locale et les
// clés mortes (une entrée en trop, invisible car jamais lue).
func TestCatalogParity(t *testing.T) {
	ref, ok := catalogs["fr"]
	if !ok {
		t.Fatal("catalogue fr absent")
	}
	for _, loc := range Supported {
		if loc == "fr" {
			continue
		}
		cat := catalogs[loc]
		var missing, extra []string
		for key := range ref {
			if _, ok := cat[key]; !ok {
				missing = append(missing, key)
			}
		}
		for key := range cat {
			if _, ok := ref[key]; !ok {
				extra = append(extra, key)
			}
		}
		sort.Strings(missing)
		sort.Strings(extra)
		if len(missing) > 0 {
			t.Errorf("%s: %d clé(s) manquante(s) vs fr : %s", loc, len(missing), strings.Join(missing, ", "))
		}
		if len(extra) > 0 {
			t.Errorf("%s: %d clé(s) en trop vs fr : %s", loc, len(extra), strings.Join(extra, ", "))
		}
	}
}

// TestCatalogNoEmptyValues : une valeur vide retomberait silencieusement sur le
// français via T(). Seules les clés vides en français (ex. journey.*.detail
// volontairement absent) sont tolérées.
func TestCatalogNoEmptyValues(t *testing.T) {
	ref := catalogs["fr"]
	for _, loc := range Supported {
		for key, val := range catalogs[loc] {
			if strings.TrimSpace(val) == "" && strings.TrimSpace(ref[key]) != "" {
				t.Errorf("%s: %s est vide alors que fr est renseigné", loc, key)
			}
		}
	}
}

// TestCatalogPlaceholderParity : un placeholder oublié ou renommé lors d'une
// traduction produit un email/SMS avec « {petName} » visible en clair.
func TestCatalogPlaceholderParity(t *testing.T) {
	ref := catalogs["fr"]
	for _, loc := range Supported {
		if loc == "fr" {
			continue
		}
		for key, val := range catalogs[loc] {
			want := placeholders(ref[key])
			got := placeholders(val)
			if strings.Join(want, ",") != strings.Join(got, ",") {
				t.Errorf("%s: %s placeholders %v, fr attend %v", loc, key, got, want)
			}
		}
	}
}

// placeholders extrait les jetons « {nom} » triés d'un message.
func placeholders(s string) []string {
	var out []string
	for {
		open := strings.Index(s, "{")
		if open < 0 {
			break
		}
		close := strings.Index(s[open:], "}")
		if close < 0 {
			break
		}
		out = append(out, s[open:open+close+1])
		s = s[open+close+1:]
	}
	sort.Strings(out)
	return out
}
