package headerlinks

import "strings"

// CatalogEntry is a system-defined external link for a country.
type CatalogEntry struct {
	ID             string
	DefaultEnabled bool
	// LabelKey is the i18n key suffix under headerLinks.catalog.<id>
	LabelKey string
	URLFor   func(locale string) string
}

const (
	MaxCustomLinks = 10
	MaxLabelLen    = 40
	MaxURLLen      = 2048
)

func CatalogForCountry(countryCode string) []CatalogEntry {
	switch strings.ToUpper(strings.TrimSpace(countryCode)) {
	case "BE":
		return catalogBE
	case "FR":
		return catalogFR
	case "ES":
		return catalogES
	default:
		return nil
	}
}

func CatalogByID(countryCode, id string) (CatalogEntry, bool) {
	for _, e := range CatalogForCountry(countryCode) {
		if e.ID == id {
			return e, true
		}
	}
	return CatalogEntry{}, false
}

// LabelFR is the French display label (fallback when locale catalogs are incomplete).
func LabelFR(id string) string {
	if s, ok := labelsFR[id]; ok {
		return s
	}
	return id
}

var labelsFR = map[string]string{
	"be_afsca":             "AFSCA — vétérinaires",
	"be_afsca_news":        "AFSCA — newsletters",
	"be_vetcompendium":     "Vetcompendium",
	"be_afmps":             "AFMPS",
	"be_sanitel_med":       "Sanitel-Med",
	"be_amcra":             "AMCRA",
	"be_ordre":             "Ordre des vétérinaires",
	"fr_ordre":             "Ordre national des vétérinaires",
	"fr_ircp":              "Index RCP (ANMV)",
	"fr_calypso":           "CalypsoVet",
	"fr_pharmacovigilance": "Pharmacovigilance ANMV",
	"fr_icad":              "I-CAD",
	"fr_anses":             "ANSES",
	"es_cimavet":           "CIMA Vet",
	"es_aemps":             "AEMPS — médicaments vétérinaires",
	"es_presvet":           "PRESVET",
	"es_notificavet":       "NotificaVET",
	"es_ocv":               "Organización Colegial Veterinaria",
}

func DefaultEnabledIDs(countryCode string) []string {
	var out []string
	for _, e := range CatalogForCountry(countryCode) {
		if e.DefaultEnabled {
			out = append(out, e.ID)
		}
	}
	return out
}

func localeNL(locale string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(locale)), "nl")
}

var catalogBE = []CatalogEntry{
	{
		ID: "be_afsca", DefaultEnabled: false, LabelKey: "be_afsca",
		URLFor: func(locale string) string {
			if localeNL(locale) {
				return "https://favv-afsca.be/nl/themas/zelfstandige-beroepen/zelfstandige-dierenartsen"
			}
			return "https://favv-afsca.be/fr/themes/metiers-independants/veterinaires-independants"
		},
	},
	{
		ID: "be_afsca_news", DefaultEnabled: true, LabelKey: "be_afsca_news",
		URLFor: func(locale string) string {
			if localeNL(locale) {
				return "https://favv-afsca.be/nl/themas/zelfstandige-beroepen/zelfstandige-dierenartsen/nieuwsbrief-voor-dierenartsen"
			}
			return "https://favv-afsca.be/fr/themes/metiers-independants/veterinaires-independants/newsletters-pour-les-veterinaires"
		},
	},
	{
		ID: "be_vetcompendium", DefaultEnabled: true, LabelKey: "be_vetcompendium",
		URLFor: func(locale string) string {
			if localeNL(locale) {
				return "https://www.vetcompendium.be/nl"
			}
			return "https://www.vetcompendium.be/fr"
		},
	},
	{
		ID: "be_afmps", DefaultEnabled: true, LabelKey: "be_afmps",
		URLFor: func(locale string) string {
			if localeNL(locale) {
				return "https://www.afmps.be/nl"
			}
			return "https://www.afmps.be/fr"
		},
	},
	{
		ID: "be_sanitel_med", DefaultEnabled: false, LabelKey: "be_sanitel_med",
		URLFor: func(locale string) string {
			if localeNL(locale) {
				return "https://www.afmps.be/nl/SANITEL-MED"
			}
			return "https://www.afmps.be/fr/SANITEL-MED"
		},
	},
	{
		ID: "be_amcra", DefaultEnabled: false, LabelKey: "be_amcra",
		URLFor: func(string) string { return "https://www.amcra.be/" },
	},
	{
		ID: "be_ordre", DefaultEnabled: true, LabelKey: "be_ordre",
		URLFor: func(string) string { return "https://www.ordre-veterinaires.be/" },
	},
}

var catalogFR = []CatalogEntry{
	{
		ID: "fr_ordre", DefaultEnabled: true, LabelKey: "fr_ordre",
		URLFor: func(string) string { return "https://www.veterinaire.fr/" },
	},
	{
		ID: "fr_ircp", DefaultEnabled: true, LabelKey: "fr_ircp",
		URLFor: func(string) string { return "https://www.ircp.anmv.anses.fr/" },
	},
	{
		ID: "fr_calypso", DefaultEnabled: true, LabelKey: "fr_calypso",
		URLFor: func(string) string { return "https://calypsovet.fr/" },
	},
	{
		ID: "fr_pharmacovigilance", DefaultEnabled: false, LabelKey: "fr_pharmacovigilance",
		URLFor: func(string) string { return "https://pharmacovigilance-anmv.anses.fr/" },
	},
	{
		ID: "fr_icad", DefaultEnabled: true, LabelKey: "fr_icad",
		URLFor: func(string) string { return "https://www.i-cad.fr/" },
	},
	{
		ID: "fr_anses", DefaultEnabled: false, LabelKey: "fr_anses",
		URLFor: func(string) string { return "https://www.anses.fr/" },
	},
}

var catalogES = []CatalogEntry{
	{
		ID: "es_cimavet", DefaultEnabled: true, LabelKey: "es_cimavet",
		URLFor: func(string) string { return "https://cimavet.aemps.es/" },
	},
	{
		ID: "es_aemps", DefaultEnabled: true, LabelKey: "es_aemps",
		URLFor: func(string) string { return "https://www.aemps.gob.es/medicamentos-veterinarios/" },
	},
	{
		ID: "es_presvet", DefaultEnabled: true, LabelKey: "es_presvet",
		URLFor: func(string) string { return "https://servicio.mapama.gob.es/presvet" },
	},
	{
		ID: "es_notificavet", DefaultEnabled: false, LabelKey: "es_notificavet",
		URLFor: func(string) string { return "https://sinaem.aemps.es/FVVET/notificavet" },
	},
	{
		ID: "es_ocv", DefaultEnabled: true, LabelKey: "es_ocv",
		URLFor: func(string) string { return "https://www.colvet.es/" },
	},
}
