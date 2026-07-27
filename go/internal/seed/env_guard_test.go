package seed_test

import (
	"context"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/seed"
)

// La garde APP_ENV court avant toute utilisation du pool : un pool nil suffit donc
// à prouver qu'aucune requête n'est émise quand l'environnement n'est pas seedable.
func TestSeedRefusedOutsideSeedableEnv(t *testing.T) {
	for _, env := range []string{"production", "prod", "prd", "live", ""} {
		t.Setenv("APP_ENV", env)
		if err := seed.Run(context.Background(), nil); err == nil {
			t.Fatalf("seed.Run must refuse APP_ENV=%q", env)
		}
		if err := seed.RunMass(context.Background(), nil); err == nil {
			t.Fatalf("seed.RunMass must refuse APP_ENV=%q", env)
		}
	}
}
