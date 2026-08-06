package invoicing_test

import (
	"errors"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
)

func TestOpenAPIKeyWrongMaterialReturnsErrSecrets(t *testing.T) {
	sealed, err := invoicing.SealAPIKey("local_enc", "correct-secret-material!!", "my-api-key")
	if err != nil {
		t.Fatal(err)
	}
	_, err = invoicing.OpenAPIKey("local_enc", "wrong-secret-material!!!!!", sealed)
	if err == nil {
		t.Fatal("expected decrypt error")
	}
	if !errors.Is(err, invoicing.ErrSecrets) {
		t.Fatalf("want ErrSecrets got %v", err)
	}
}

func TestOpenAPIKeyUnknownRefReturnsErrSecrets(t *testing.T) {
	_, err := invoicing.OpenAPIKey("local_enc", "enough-secret-material-here", "sm:not-supported")
	if !errors.Is(err, invoicing.ErrSecrets) {
		t.Fatalf("want ErrSecrets got %v", err)
	}
}
