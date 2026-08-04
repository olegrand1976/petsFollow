package handlers_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// speciesByCode indexe la réponse /species ou /admin/species par code.
func speciesByCode(t *testing.T, env map[string]any) map[string]map[string]any {
	t.Helper()
	data := dataMap(t, env)
	rows, _ := data["species"].([]any)
	out := map[string]map[string]any{}
	for _, raw := range rows {
		m, _ := raw.(map[string]any)
		code, _ := m["code"].(string)
		if code != "" {
			out[code] = m
		}
	}
	return out
}

func speciesRule(t *testing.T, row map[string]any) map[string]any {
	t.Helper()
	rule, _ := row["rule"].(map[string]any)
	if rule == nil {
		t.Fatalf("species row without rule: %#v", row)
	}
	return rule
}

// Le seed BE encode la réglementation belge (AR du 21/07/2016) : le DAF suit les
// animaux producteurs de denrées, pas la taille de l'animal.
func TestAdminSpeciesBelgianDefaults(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/species", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin list %d %#v", code, env)
	}
	if got, _ := dataMap(t, env)["countryCode"].(string); got != "BE" {
		t.Fatalf("countryCode = %q, want BE", got)
	}
	byCode := speciesByCode(t, env)

	cases := []struct {
		species     string
		largeAnimal bool
		dafRequired string
	}{
		// Animaux de compagnie — jamais de DAF.
		{"dog", false, "never"},
		{"cat", false, "never"},
		// Gros animaux producteurs de denrées.
		{"cattle", true, "always"},
		{"sheep", true, "always"},
		{"goat", true, "always"},
		{"pig", true, "always"},
		{"alpaca", true, "always"},
		// Équidés — producteurs de denrées sauf exclusion par passeport.
		{"horse", true, "per_animal"},
		{"donkey", true, "per_animal"},
		// Le point qui distingue « gros animal » de la règle légale : la volaille et
		// le lapin relèvent du DAF sans être des gros animaux.
		{"poultry", false, "always"},
		{"rabbit", false, "always"},
	}
	for _, tc := range cases {
		row, ok := byCode[tc.species]
		if !ok {
			t.Fatalf("species %q missing from catalogue", tc.species)
		}
		rule := speciesRule(t, row)
		if rule["isLargeAnimal"] != tc.largeAnimal {
			t.Errorf("%s: isLargeAnimal = %v, want %v", tc.species, rule["isLargeAnimal"], tc.largeAnimal)
		}
		if rule["dafRequired"] != tc.dafRequired {
			t.Errorf("%s: dafRequired = %v, want %q", tc.species, rule["dafRequired"], tc.dafRequired)
		}
	}

	// Les abeilles produisent une denrée (miel) : DAF applicable, mais espèce inactive
	// tant que le correctif Flutter n'est pas déployé.
	bee, ok := byCode["bee"]
	if !ok {
		t.Fatalf("bee missing from catalogue")
	}
	if speciesRule(t, bee)["dafRequired"] != "always" {
		t.Errorf("bee dafRequired = %v, want always", speciesRule(t, bee)["dafRequired"])
	}
	if bee["isActive"] != false {
		t.Errorf("bee should ship inactive, got %v", bee["isActive"])
	}
}

// `other` doit rester le dernier choix des sélecteurs (Nuxt et Flutter s'appuient dessus).
func TestAdminSpeciesOtherSortsLast(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/species", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin list %d %#v", code, env)
	}
	rows, _ := dataMap(t, env)["species"].([]any)
	if len(rows) == 0 {
		t.Fatalf("empty catalogue")
	}
	last, _ := rows[len(rows)-1].(map[string]any)
	if last["code"] != "other" {
		t.Fatalf("last species = %v, want other", last["code"])
	}
}

// /species ne sert que les espèces actives — l'admin voit tout via /admin/species.
func TestSpeciesPublicListOnlyActive(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/species", vetTok, nil)
	if code != http.StatusOK {
		t.Fatalf("public list %d %#v", code, env)
	}
	byCode := speciesByCode(t, env)
	if _, ok := byCode["dog"]; !ok {
		t.Fatalf("dog missing from active catalogue")
	}
	if _, ok := byCode["bee"]; ok {
		t.Fatalf("inactive species bee leaked into /species")
	}
	for c, row := range byCode {
		if row["isActive"] != true {
			t.Fatalf("inactive species %q served by /species", c)
		}
	}
}

func TestAdminSpeciesRequiresAdmin(t *testing.T) {
	api := newTestAPI(t)
	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/species", vetTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("vet GET admin species: want 403 got %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/species", vetTok, map[string]any{
		"code": "vetmade", "labels": map[string]string{"fr": "Test"},
	})
	if code != http.StatusForbidden {
		t.Fatalf("vet POST admin species: want 403 got %d %#v", code, env)
	}
}

func TestAdminSpeciesCreateAndPatch(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	newCode := "testsp_" + uuid.NewString()[:8]

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/species", adminTok, map[string]any{
		"code":      newCode,
		"labels":    map[string]string{"fr": "Espèce de test", "nl": "Testsoort", "en": "Test species"},
		"sortOrder": 500,
		"isActive":  true,
		"rule": map[string]any{
			"countryCode": "BE", "isLargeAnimal": true, "dafRequired": "always",
			"isFoodChain": true, "defaultFoodChainStatus": "food_producing",
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create species %d %#v", code, env)
	}

	// Le code est la clé : recréer doit être refusé, pas écraser silencieusement.
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/species", adminTok, map[string]any{
		"code": newCode, "labels": map[string]string{"fr": "Doublon"},
	})
	if code != http.StatusConflict {
		t.Fatalf("duplicate species: want 409 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/species", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, env)
	}
	row, ok := speciesByCode(t, env)[newCode]
	if !ok {
		t.Fatalf("created species %q missing from list", newCode)
	}
	rule := speciesRule(t, row)
	if rule["dafRequired"] != "always" || rule["isLargeAnimal"] != true {
		t.Fatalf("rule not persisted: %#v", rule)
	}

	// Désactivation + bascule réglementaire.
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/species/"+newCode, adminTok, map[string]any{
		"isActive": false,
		"rule": map[string]any{
			"countryCode": "BE", "isLargeAnimal": false, "dafRequired": "never",
			"isFoodChain": false, "defaultFoodChainStatus": "companion",
		},
	})
	if code != http.StatusOK {
		t.Fatalf("patch species %d %#v", code, env)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/species", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list after patch %d %#v", code, env)
	}
	row = speciesByCode(t, env)[newCode]
	if row["isActive"] != false {
		t.Fatalf("isActive not persisted: %#v", row)
	}
	if speciesRule(t, row)["dafRequired"] != "never" {
		t.Fatalf("rule not updated: %#v", speciesRule(t, row))
	}
}

func TestAdminSpeciesValidation(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	bad := []struct {
		name string
		body map[string]any
	}{
		{"code avec espace", map[string]any{"code": "bad code", "labels": map[string]string{"fr": "x"}}},
		{"code avec ponctuation", map[string]any{"code": "chien!", "labels": map[string]string{"fr": "x"}}},
		{"code trop court", map[string]any{"code": "x", "labels": map[string]string{"fr": "x"}}},
		{"code vide", map[string]any{"code": "", "labels": map[string]string{"fr": "x"}}},
		{"sans libellé", map[string]any{"code": "nolabel"}},
	}
	for _, tc := range bad {
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/species", adminTok, tc.body)
		if code != http.StatusBadRequest {
			t.Errorf("%s: want 400 got %d %#v", tc.name, code, env)
		}
	}

	// La casse est normalisée, pas rejetée : un code saisi en majuscules reste utilisable.
	upper := "TestCase" + uuid.NewString()[:6]
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/species", adminTok, map[string]any{
		"code": upper, "labels": map[string]string{"fr": "Normalisation"}, "isActive": false,
	})
	if code != http.StatusCreated {
		t.Fatalf("uppercase code should normalize, got %d %#v", code, env)
	}
	if got, _ := dataMap(t, env)["code"].(string); got != strings.ToLower(upper) {
		t.Fatalf("code = %q, want %q", got, strings.ToLower(upper))
	}

	// daf_required hors énum : refus, sinon la valeur casserait le gate silencieusement.
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/species/dog", adminTok, map[string]any{
		"rule": map[string]any{
			"countryCode": "BE", "dafRequired": "sometimes", "defaultFoodChainStatus": "companion",
		},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("invalid dafRequired: want 400 got %d %#v", code, env)
	}

	// Pays hors allowlist : refus plutôt que repli silencieux sur BE.
	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/species/dog", adminTok, map[string]any{
		"rule": map[string]any{
			"countryCode": "ZZ", "dafRequired": "never", "defaultFoodChainStatus": "companion",
		},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("invalid country: want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/species/does_not_exist", adminTok, map[string]any{
		"isActive": false,
	})
	if code != http.StatusNotFound {
		t.Fatalf("unknown species: want 404 got %d %#v", code, env)
	}
}
