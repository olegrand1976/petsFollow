package pharmacy_test

import (
	"context"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

type stubVamregRefs struct {
	species, indications, forms []pharmacy.VamregCodeLabel
	errSpecies                  error
}

func (s stubVamregRefs) ListTargetSpecies(ctx context.Context) ([]pharmacy.VamregCodeLabel, error) {
	return s.species, s.errSpecies
}
func (s stubVamregRefs) ListIndications(ctx context.Context) ([]pharmacy.VamregCodeLabel, error) {
	return s.indications, nil
}
func (s stubVamregRefs) ListPharmaceuticalForms(ctx context.Context) ([]pharmacy.VamregCodeLabel, error) {
	return s.forms, nil
}

func TestFetchVamregRefLists(t *testing.T) {
	lists, err := pharmacy.FetchVamregRefLists(context.Background(), stubVamregRefs{
		species:     []pharmacy.VamregCodeLabel{{Code: "CAN", Fr: "Chien"}},
		indications: []pharmacy.VamregCodeLabel{{Code: "INF", Fr: "Infection"}},
		forms:       []pharmacy.VamregCodeLabel{{Code: "TAB", Fr: "Comprimé"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(lists[pharmacy.VamregRefKindTargetSpecies]) != 1 || lists[pharmacy.VamregRefKindTargetSpecies][0].Code != "CAN" {
		t.Fatalf("%#v", lists)
	}
	if len(lists[pharmacy.VamregRefKindIndication]) != 1 || len(lists[pharmacy.VamregRefKindPharmaceuticalForm]) != 1 {
		t.Fatalf("%#v", lists)
	}
}

func TestFetchVamregRefListsNilClient(t *testing.T) {
	if _, err := pharmacy.FetchVamregRefLists(context.Background(), nil); err == nil {
		t.Fatal("expected error")
	}
}
