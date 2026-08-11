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
	lot := "LOT-IMM-" + suffix
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
		"lotNumber":    lot,
		"expiresOn":    exp,
		"qty":          2,
		"unit":         "box",
	})
	if code != http.StatusCreated {
		t.Fatalf("receipt %d %#v", code, env)
	}

	var movID, createdBy string
	if err := api.pool.QueryRow(ctx, `
		SELECT m.id::text, COALESCE(m.created_by::text,'')
		FROM pharmacy.stock_movements m
		JOIN pharmacy.medication_batches b ON b.id = m.batch_id
		WHERE b.lot_number = $1
		ORDER BY m.created_at DESC LIMIT 1`, lot).Scan(&movID, &createdBy); err != nil {
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

	// FK-style nullify of delivery_note_id must be allowed (column already null → no-op path).
	_, err = api.pool.Exec(ctx, `
		UPDATE pharmacy.stock_movements SET delivery_note_id = NULL WHERE id = $1::uuid`, movID)
	if err != nil {
		t.Fatalf("nullify delivery_note_id should be allowed: %v", err)
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
	if got["immutableEnforced"] != true {
		t.Fatalf("%#v", got)
	}
	if got["retentionYears"].(float64) != float64(store.StockMovementsRetentionYears) {
		t.Fatalf("%#v", got)
	}
	if got["total"].(float64) < 1 {
		t.Fatalf("expected movements %#v", got)
	}
}
