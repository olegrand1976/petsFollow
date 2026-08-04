package store

import "testing"

// Le fallback s'applique quand la table ne connaît pas le couple (espèce, pays) :
// espèce libre héritée, ou pays sans jeu de règles seedé. Il doit reproduire le
// comportement historique de go/pkg/kernel sur la chaîne alimentaire.
func TestFallbackSpeciesRuleMirrorsKernel(t *testing.T) {
	cases := []struct {
		species     string
		foodChain   bool
		dafRequired string
		status      string
	}{
		// Animaux de compagnie : jamais de DAF.
		{"dog", false, DAFRequiredNever, "companion"},
		{"cat", false, DAFRequiredNever, "companion"},
		{"other", false, DAFRequiredNever, "companion"},
		// Espèce inconnue du kernel : fail-closed.
		{"ferret", false, DAFRequiredNever, "companion"},
		{"", false, DAFRequiredNever, "companion"},
		// Espèces de rente : DAF possible selon l'individu.
		{"cattle", true, DAFRequiredPerAnimal, "food_producing"},
		{"poultry", true, DAFRequiredPerAnimal, "food_producing"},
		{"llama", true, DAFRequiredPerAnimal, "food_producing"},
		// Équidés et lapin : chaîne alimentaire mais companion par défaut.
		{"horse", true, DAFRequiredPerAnimal, "companion"},
		{"donkey", true, DAFRequiredPerAnimal, "companion"},
		{"rabbit", true, DAFRequiredPerAnimal, "companion"},
	}
	for _, tc := range cases {
		got := fallbackSpeciesRule(tc.species, "BE")
		if got.IsFoodChain != tc.foodChain {
			t.Errorf("%q: IsFoodChain = %v, want %v", tc.species, got.IsFoodChain, tc.foodChain)
		}
		if got.DAFRequired != tc.dafRequired {
			t.Errorf("%q: DAFRequired = %q, want %q", tc.species, got.DAFRequired, tc.dafRequired)
		}
		if got.DefaultFoodChainStatus != tc.status {
			t.Errorf("%q: DefaultFoodChainStatus = %q, want %q", tc.species, got.DefaultFoodChainStatus, tc.status)
		}
		// Le fallback ne peut pas deviner la taille : jamais gros animal.
		if got.IsLargeAnimal {
			t.Errorf("%q: fallback should never claim large animal", tc.species)
		}
	}
}

func TestValidSpeciesDAFRequired(t *testing.T) {
	for _, ok := range []string{DAFRequiredNever, DAFRequiredAlways, DAFRequiredPerAnimal} {
		if !ValidSpeciesDAFRequired(ok) {
			t.Errorf("%q should be valid", ok)
		}
	}
	for _, bad := range []string{"", "NEVER", "sometimes", "large_animal"} {
		if ValidSpeciesDAFRequired(bad) {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestValidFoodChainStatus(t *testing.T) {
	for _, ok := range []string{"companion", "food_producing", "excluded_from_food_chain"} {
		if !ValidFoodChainStatus(ok) {
			t.Errorf("%q should be valid", ok)
		}
	}
	for _, bad := range []string{"", "COMPANION", "pet", "excluded"} {
		if ValidFoodChainStatus(bad) {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestSpeciesCodeFormat(t *testing.T) {
	for _, ok := range []string{"dog", "guinea_pig", "farmed_fish", "ab"} {
		if !speciesCodeRe.MatchString(ok) {
			t.Errorf("%q should be an accepted code", ok)
		}
	}
	for _, bad := range []string{"", "a", "Dog", "guinea pig", "chien!", "_dog", "9lives"} {
		if speciesCodeRe.MatchString(bad) {
			t.Errorf("%q should be rejected", bad)
		}
	}
}
