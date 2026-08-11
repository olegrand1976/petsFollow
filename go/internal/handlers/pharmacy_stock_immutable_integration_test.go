package handlers_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestPharmacyStockMovementsImmutableAndRetentionStats(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	suffix := uuid.NewString()[:8]
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2777" + suffix[:4], Name: "Immutable Demo " + suffix, IsActive: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}

	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	exp := time.Now().AddDate(0, 0, 90).Format("2006-01-02")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
		"medicationId": medID,
		"lotNumber":    "LOT-IMM-" + suffix,
		"expiresOn":    exp,
		"qty":          2,
		"unit":         "box",
	})
	if code != http.StatusCreated {
		t.Fatalf("receipt %d %#v", code, env)
	}

	var movID, createdBy string
	if err := api.pool.QueryRow(ctx, `
		SELECT id::text, COALESCE(created_by::text,'')
		FROM pharmacy.stock_movements
		ORDER BY created_at DESC LIMIT 1`).Scan(&movID, &createdBy); err != nil {
		t.Fatal(err)
	}
	if movID == "" {
		t.Fatal("no movement")
	}

	_, err = api.pool.Exec(ctx, `UPDATE pharmacy.stock_movements SET reason_detail = 'tamper' WHERE id = $1::uuid`, movID)
	if err == nil {
		t.Fatal("expected immutable block on UPDATE")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "stock_movements_immutable") && !strings.Contains(msg, "permission") {
		t.Fatalf("want immutability/permission error, got %v", err)
	}

	_, err = api.pool.Exec(ctx, `DELETE FROM pharmacy.stock_movements WHERE id = $1::uuid`, movID)
	if err == nil {
		t.Fatal("expected immutable block on DELETE")
	}

	if createdBy != "" {
		var n int
		if err := api.pool.QueryRow(ctx, `SELECT pharmacy.rgpd_null_stock_movement_created_by($1::uuid)`, createdBy).Scan(&n); err != nil {
			t.Fatalf("rgpd nullify: %v", err)
		}
		var after string
		if err := api.pool.QueryRow(ctx, `SELECT COALESCE(created_by::text,'') FROM pharmacy.stock_movements WHERE id = $1::uuid`, movID).Scan(&after); err != nil {
			t.Fatal(err)
		}
		if after != "" {
			t.Fatalf("created_by still %q after rgpd nullify", after)
		}
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/movements/retention-stats", tok, nil)
	if code != http.StatusOK {
		t.Fatalf("retention-stats %d %#v", code, env)
	}
	got := dataMap(t, env)
	if got["immutableAppRole"] != true {
		t.Fatalf("%#v", got)
	}
	if got["retentionYears"].(float64) != float64(store.StockMovementsRetentionYears) {
		t.Fatalf("%#v", got)
	}
	if got["total"].(float64) < 1 {
		t.Fatalf("expected movements %#v", got)
	}
}
