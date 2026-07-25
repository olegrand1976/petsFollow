package seed

import (
	"context"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// seedStripeCatalog upserts demo/staging Stripe product+price IDs used for local checkout.
// Production/staging Live should prefer admin CSV import (/admin/stripe-catalog) over this list.
func seedStripeCatalog(ctx context.Context, st *store.Store) error {
	_, err := st.UpsertDefaultStripeCatalog(ctx)
	return err
}
