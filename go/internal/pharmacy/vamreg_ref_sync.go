package pharmacy

import (
	"context"
	"fmt"
)

// VamregRefLister is the AFMPS readonly surface used by ref sync (mockable).
type VamregRefLister interface {
	ListTargetSpecies(ctx context.Context) ([]VamregCodeLabel, error)
	ListIndications(ctx context.Context) ([]VamregCodeLabel, error)
	ListPharmaceuticalForms(ctx context.Context) ([]VamregCodeLabel, error)
}

// VamregRefKind constants mirrored by store (avoid import cycle).
const (
	VamregRefKindTargetSpecies      = "target_species"
	VamregRefKindIndication         = "indication"
	VamregRefKindPharmaceuticalForm = "pharmaceutical_form"
)

// FetchVamregRefLists pulls species / indication / pharmaceutical-form from AFMPS.
func FetchVamregRefLists(ctx context.Context, c VamregRefLister) (map[string][]VamregCodeLabel, error) {
	if c == nil {
		return nil, fmt.Errorf("vamreg_afmps_client_nil")
	}
	out := make(map[string][]VamregCodeLabel, 3)

	species, err := c.ListTargetSpecies(ctx)
	if err != nil {
		return nil, fmt.Errorf("target_species: %w", err)
	}
	out[VamregRefKindTargetSpecies] = species

	indications, err := c.ListIndications(ctx)
	if err != nil {
		return nil, fmt.Errorf("indication: %w", err)
	}
	out[VamregRefKindIndication] = indications

	forms, err := c.ListPharmaceuticalForms(ctx)
	if err != nil {
		return nil, fmt.Errorf("pharmaceutical_form: %w", err)
	}
	out[VamregRefKindPharmaceuticalForm] = forms

	return out, nil
}
