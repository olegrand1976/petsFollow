package store

import (
	"context"
	"errors"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
)

func TestInsertDAFStockMovementRequiresTrace(t *testing.T) {
	st := &Store{}
	ctx := context.Background()

	err := st.insertDAFStockMovement(ctx, nil, "00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000002", "", "", "user", "daf", "", -1)
	if !errors.Is(err, pharmacy.ErrDAFTraceRequired) {
		t.Fatalf("want ErrDAFTraceRequired, got %v", err)
	}

	err = st.insertDAFStockMovement(ctx, nil, "00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000003",
		"", "user", "daf", "", -1)
	if !errors.Is(err, pharmacy.ErrDAFTraceRequired) {
		t.Fatalf("want ErrDAFTraceRequired for empty item, got %v", err)
	}

	err = st.insertDAFStockMovement(ctx, nil, "00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000003",
		"00000000-0000-0000-0000-000000000004", "user", "adjust", "", -1)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("want ErrValidation for non-daf reason, got %v", err)
	}
}
