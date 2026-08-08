package vetnews

import (
	"strings"
	"unicode"
)

// Classify assigns category + importance from title/summary/source heuristics.
func Classify(item RawItem, sourceID string, defaultTags []string) (category, importance string, tags []string) {
	blob := strings.ToLower(item.Title + " " + item.Summary + " " + strings.Join(item.Tags, " "))
	blob = normalizeLetters(blob)

	importance = ImportanceLow
	category = CategoryGeneral

	switch sourceID {
	case "woah", "anses":
		importance = ImportanceMedium
		category = CategoryEpidemio
	case "veterinary_evidence":
		category = CategoryScience
		importance = ImportanceLow
	case "clinicians_brief":
		category = CategoryClinical
		importance = ImportanceMedium
	case "todays_vet_practice":
		category = CategoryPractice
		importance = ImportanceMedium
	case "point_veterinaire":
		category = CategoryPractice
		importance = ImportanceMedium
	}

	criticalHints := []string{
		"outbreak", "épidémie", "epidemie", "epizoot", "foyer", "rappel de lot", "product recall",
		"alerte sanitaire", "high pathogenicity", "haute pathogénicité",
		"influenza aviaire", "avian influenza", "fièvre aphteuse", "foot-and-mouth", "peste porcine",
		"african swine", "rage ", "rabies", "dermatose nodulaire", "lumpy skin",
	}
	highHints := []string{
		"surveillance", "réglement", "reglement", "décret", "arrete", "arrêté", "legislation",
		"législation", "cvmp", "withdrawal", "résistance antimicrobien", "antimicrobial resistance",
		"one health", "zoonose", "zoonosis", "west nile", "nil occidental", "bluetongue", "fièvre catarrhale",
	}
	clinicalHints := []string{
		"chirurgie", "surgery", "diagnostic", "traitement", "treatment", "clinique", "clinical",
		"canin", "félin", "feline", "canine", "anesth",
	}
	scienceHints := []string{
		"peer-reviewed", "evidence-based", "ebvm", "essai clinique", "clinical trial",
	}
	regulatoryHints := []string{
		"réglement", "reglement", "décret", "arrêté", "legislation", "législation",
		"autorisation de mise sur le marché", "peppol",
	}

	if containsAny(blob, criticalHints) {
		importance = ImportanceCritical
		category = CategoryEpidemio
	} else if containsAny(blob, highHints) {
		if importance != ImportanceCritical {
			importance = ImportanceHigh
		}
		if category == CategoryGeneral || category == CategoryPractice {
			category = CategoryEpidemio
		}
	}

	if containsAny(blob, regulatoryHints) && importance != ImportanceCritical {
		category = CategoryRegulatory
		if importance == ImportanceLow {
			importance = ImportanceHigh
		}
	}
	if containsAny(blob, clinicalHints) && category == CategoryGeneral {
		category = CategoryClinical
	}
	if containsAny(blob, scienceHints) && (category == CategoryGeneral || category == CategoryPractice) {
		category = CategoryScience
	}

	tags = uniqueNonEmpty(append(append([]string{}, defaultTags...), item.Tags...))
	tags = uniqueNonEmpty(append(tags, category, importance))
	return category, importance, tags
}

func normalizeLetters(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' || r == '-' || r == '\'' {
			return unicode.ToLower(r)
		}
		return ' '
	}, s)
}

func containsAny(blob string, phrases []string) bool {
	for _, p := range phrases {
		if p != "" && strings.Contains(blob, strings.ToLower(p)) {
			return true
		}
	}
	return false
}
