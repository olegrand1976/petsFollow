package store_test

import (
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestValidateVamregRefListsRejectsEmpty(t *testing.T) {
	lists := map[string][]pharmacy.VamregCodeLabel{
		pharmacy.VamregRefKindTargetSpecies:      {{Code: "CAN"}},
		pharmacy.VamregRefKindIndication:         {{Code: "INF"}},
		pharmacy.VamregRefKindPharmaceuticalForm: {},
	}
	if err := store.ValidateVamregRefLists(lists); err != store.ErrVamregRefEmptyList {
		t.Fatalf("got %v", err)
	}
}

func TestValidateVamregRefListsAcceptsDedupe(t *testing.T) {
	lists := map[string][]pharmacy.VamregCodeLabel{
		pharmacy.VamregRefKindTargetSpecies: {
			{Code: " CAN ", Fr: "Chien"},
			{Code: "CAN", Fr: "Dog dup"},
			{Code: ""},
		},
		pharmacy.VamregRefKindIndication:         {{Code: "INF"}},
		pharmacy.VamregRefKindPharmaceuticalForm: {{Code: "TAB"}},
	}
	if err := store.ValidateVamregRefLists(lists); err != nil {
		t.Fatal(err)
	}
}
