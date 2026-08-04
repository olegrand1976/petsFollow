package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// setPetSpecies force l'espèce et le statut chaîne alimentaire d'un patient.
// Le PATCH public ne permet pas de changer l'espèce côté véto : on écrit en base
// pour balayer les trois branches réglementaires sur un même patient.
func setPetSpecies(t *testing.T, api *testAPI, petID, species, foodChainStatus string) {
	t.Helper()
	if _, err := api.pool.Exec(context.Background(), `
		UPDATE pets.pets SET species = $2, food_chain_status = $3 WHERE id = $1`,
		petID, species, foodChainStatus); err != nil {
		t.Fatalf("set species %q: %v", species, err)
	}
}

// makePetDAFEligible établit la précondition métier du DAF : un patient producteur de
// denrées alimentaires. Les tests de plomberie DAF (lien visite, passerelle ordonnance,
// dispensations) piochent le premier patient du seed, souvent un chien — ce qui n'est
// pas un cas d'usage DAF valide. L'espèce d'origine est restaurée en fin de test.
func makePetDAFEligible(t *testing.T, api *testAPI, petID string) {
	t.Helper()
	ctx := context.Background()
	var species, status string
	if err := api.pool.QueryRow(ctx, `
		SELECT COALESCE(species,''), COALESCE(food_chain_status,'companion')
		FROM pets.pets WHERE id = $1`, petID).Scan(&species, &status); err != nil {
		t.Fatalf("read pet species: %v", err)
	}
	// `companion` sur un bovin : le gate espèce passe (always) sans déclencher
	// l'exigence de temps d'attente, hors sujet pour ces tests.
	setPetSpecies(t, api, petID, "cattle", "companion")
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(), `
			UPDATE pets.pets SET species = $2, food_chain_status = $3 WHERE id = $1`,
			petID, species, status)
	})
}

// dafGateFixture prépare un médicament, un patient véto et le corps d'un DAF minimal.
func dafGateFixture(t *testing.T, api *testAPI) (tok, petID, clientID string, body map[string]any) {
	t.Helper()
	ctx := context.Background()
	st := store.New(api.pool)
	suffix := uuid.NewString()[:8]
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2777" + suffix[:4], Name: "Species Gate Med " + suffix, IsActive: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}
	tok = loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pets", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("vet pets %d %#v", code, env)
	}
	items, _ := env["data"].([]any)
	for _, raw := range items {
		p, _ := raw.(map[string]any)
		if p["species"] == "horse" {
			petID, _ = p["id"].(string)
			clientID, _ = p["ownerUserId"].(string)
			break
		}
	}
	if petID == "" {
		t.Skip("no horse pet in seed for DAF species gate test")
	}
	body = map[string]any{
		"petId": petID, "clientUserId": clientID, "notes": "species-gate",
		"items": []map[string]any{{"medicationId": medID, "qty": 1, "ammNumber": "BE-TEST-AMM"}},
	}
	return tok, petID, clientID, body
}

func dafErrCode(env map[string]any) string {
	errObj, _ := env["error"].(map[string]any)
	code, _ := errObj["code"].(string)
	return code
}

// Le DAF ne concerne que les animaux producteurs de denrées alimentaires
// (AR du 21/07/2016). Trois branches : never / always / per_animal.
func TestDAFSpeciesGateOnCreate(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	tok, petID, _, body := dafGateFixture(t, api)
	t.Cleanup(func() { setPetSpecies(t, api, petID, "horse", "companion") })

	t.Run("chien — jamais de DAF", func(t *testing.T) {
		setPetSpecies(t, api, petID, "dog", "companion")
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, body)
		if code != http.StatusConflict {
			t.Fatalf("want 409 got %d %#v", code, env)
		}
		if got := dafErrCode(env); got != "daf_species_not_applicable" {
			t.Fatalf("error code = %q, want daf_species_not_applicable", got)
		}
	})

	t.Run("bovin — toujours producteur de denrées", func(t *testing.T) {
		setPetSpecies(t, api, petID, "cattle", "companion")
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, body)
		if code != http.StatusCreated {
			t.Fatalf("want 201 got %d %#v", code, env)
		}
	})

	t.Run("volaille — pas un gros animal mais DAF requis", func(t *testing.T) {
		setPetSpecies(t, api, petID, "poultry", "companion")
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, body)
		if code != http.StatusCreated {
			t.Fatalf("want 201 got %d %#v", code, env)
		}
	})

	t.Run("cheval non exclu — DAF possible", func(t *testing.T) {
		setPetSpecies(t, api, petID, "horse", "companion")
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, body)
		if code != http.StatusCreated {
			t.Fatalf("want 201 got %d %#v", code, env)
		}
	})

	t.Run("cheval exclu de la chaîne alimentaire — refus", func(t *testing.T) {
		setPetSpecies(t, api, petID, "horse", "excluded_from_food_chain")
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, body)
		if code != http.StatusConflict {
			t.Fatalf("want 409 got %d %#v", code, env)
		}
		if got := dafErrCode(env); got != "daf_species_not_applicable" {
			t.Fatalf("error code = %q, want daf_species_not_applicable", got)
		}
	})
}

// Un DAF sans patient lié n'est pas gaté : l'espèce est indéterminable.
func TestDAFWithoutPetIsNotGated(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	tok, petID, _, body := dafGateFixture(t, api)
	t.Cleanup(func() { setPetSpecies(t, api, petID, "horse", "companion") })

	delete(body, "petId")
	delete(body, "clientUserId")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, body)
	if code != http.StatusCreated {
		t.Fatalf("DAF without pet want 201 got %d %#v", code, env)
	}
}

// Le patient peut changer d'espèce entre la création du brouillon et la finalisation :
// re-vérifier avant d'allouer un numéro séquentiel inviolable.
func TestDAFSpeciesGateOnFinalize(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	tok, petID, _, body := dafGateFixture(t, api)
	t.Cleanup(func() { setPetSpecies(t, api, petID, "horse", "companion") })

	setPetSpecies(t, api, petID, "cattle", "companion")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, body)
	if code != http.StatusCreated {
		t.Fatalf("draft %d %#v", code, env)
	}
	dafID, _ := dataMap(t, env)["id"].(string)

	// Le patient devient un animal de compagnie → le brouillon n'est plus finalisable.
	setPetSpecies(t, api, petID, "dog", "companion")
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/finalize", tok, map[string]any{})
	if code != http.StatusConflict {
		t.Fatalf("finalize want 409 got %d %#v", code, env)
	}
	if got := dafErrCode(env); got != "daf_species_not_applicable" {
		t.Fatalf("error code = %q, want daf_species_not_applicable", got)
	}
}

// Le pays de la clinique pilote la règle : basculer la règle BE du chien sur `always`
// doit ouvrir le DAF sans redéploiement.
func TestDAFSpeciesGateFollowsAdminRule(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	tok, petID, _, body := dafGateFixture(t, api)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	restore := func() {
		doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/species/dog", adminTok, map[string]any{
			"rule": map[string]any{
				"countryCode": "BE", "isLargeAnimal": false, "dafRequired": "never",
				"isFoodChain": false, "defaultFoodChainStatus": "companion",
			},
		})
		setPetSpecies(t, api, petID, "horse", "companion")
	}
	t.Cleanup(restore)

	setPetSpecies(t, api, petID, "dog", "companion")
	code, env := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/admin/species/dog", adminTok, map[string]any{
		"rule": map[string]any{
			"countryCode": "BE", "isLargeAnimal": false, "dafRequired": "always",
			"isFoodChain": true, "defaultFoodChainStatus": "companion",
		},
	})
	if code != http.StatusOK {
		t.Fatalf("patch dog rule %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, body)
	if code != http.StatusCreated {
		t.Fatalf("after rule change want 201 got %d %#v", code, env)
	}
}
