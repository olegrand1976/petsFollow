package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/notifications/email"
	"github.com/olegrand1976/petsFollow/go/internal/pharmacy"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// TestPharmacyPlanCoverage covers roadmap items already ✅ that lacked dedicated assertions
// (2.A GET price, 2.C notifier gate + suppliers, 2.D MANUAL BL + deposit scope,
// 2.E CSV export, S4 retry-in-flight, 4.D ban, 4.F FK retention).
func TestPharmacyPlanCoverage(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	t.Setenv("VAMREG_DRY_RUN", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	suffix := uuid.NewString()[:8]
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	t.Run("2A_prices_get_after_put", func(t *testing.T) {
		medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
			CNK: "2555" + suffix[:4], Name: "PriceGet " + suffix, IsActive: true,
		})
		if err != nil {
			t.Fatalf("med: %v", err)
		}
		code, env := doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/pharmacy/prices/"+medID, tok, map[string]any{
			"purchasePriceCents": 500, "sellPriceCents": 1200, "vatPercent": 6,
		})
		if code != http.StatusOK {
			t.Fatalf("put %d %#v", code, env)
		}
		code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/prices/"+medID, tok, nil)
		if code != http.StatusOK {
			t.Fatalf("get %d %#v", code, env)
		}
		p := dataMap(t, env)
		if int(p["purchasePriceCents"].(float64)) != 500 || int(p["sellPriceCents"].(float64)) != 1200 {
			t.Fatalf("price mismatch %#v", p)
		}
	})

	t.Run("2C_send_requires_notifier", func(t *testing.T) {
		medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
			CNK: "2556" + suffix[:4], Name: "OrderNil " + suffix, IsActive: true,
		})
		if err != nil {
			t.Fatalf("med: %v", err)
		}
		code, env := doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/pharmacy/reorder-thresholds", tok, map[string]any{
			"medicationId": medID, "minQty": 10_000,
		})
		if code != http.StatusOK {
			t.Fatalf("threshold %d %#v", code, env)
		}
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/orders", tok, map[string]any{
			"fromAlerts": true,
		})
		if code != http.StatusCreated {
			t.Fatalf("order %d %#v", code, env)
		}
		orderID, _ := dataMap(t, env)["id"].(string)
		api.api.TestReplaceNotifier(nil)
		t.Cleanup(func() {
			api.api.TestReplaceNotifier(email.NewNotifier("127.0.0.1", 9, "test@petsfollow.test", "http://localhost:3002", "https://ll-it-sc.be"))
		})

		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/orders/"+orderID+"/send", tok, map[string]any{
			"toEmail": "fournisseur@petsfollow.test",
		})
		if code != http.StatusBadGateway {
			t.Fatalf("send without notifier want 502 got %d %#v", code, env)
		}
		if errObj, _ := env["error"].(map[string]any); errObj["code"] != "email_send_failed" {
			t.Fatalf("code %#v", env)
		}
	})

	t.Run("2C_suppliers_upsert_list", func(t *testing.T) {
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/suppliers", tok, map[string]any{
			"name": "Alcyon " + suffix, "email": "alcyon-" + suffix + "@petsfollow.test",
		})
		if code != http.StatusOK {
			t.Fatalf("create supplier %d %#v", code, env)
		}
		code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pharmacy/suppliers", tok, nil)
		if code != http.StatusOK {
			t.Fatalf("list suppliers %d %#v", code, env)
		}
		rows, _ := env["data"].([]any)
		found := false
		for _, raw := range rows {
			s, _ := raw.(map[string]any)
			name, _ := s["name"].(string)
			if strings.Contains(name, suffix) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("supplier not listed %#v", env)
		}
	})

	t.Run("2D_manual_delivery_note_number", func(t *testing.T) {
		medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
			CNK: "2557" + suffix[:4], Name: "ManualBL " + suffix, IsActive: true,
		})
		if err != nil {
			t.Fatalf("med: %v", err)
		}
		exp := time.Now().AddDate(0, 0, 180).Format("2006-01-02")
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/delivery-notes", tok, map[string]any{
			"noteNumber":   "",
			"supplierName": "Manual",
			"notify":       false,
			"items": []map[string]any{{
				"medicationId": medID, "lotNumber": "MAN-" + suffix, "expiresOn": exp, "qty": 2,
			}},
		})
		if code != http.StatusCreated {
			t.Fatalf("manual BL %d %#v", code, env)
		}
		note, _ := dataMap(t, env)["noteNumber"].(string)
		if !strings.HasPrefix(note, "MANUAL-") {
			t.Fatalf("want MANUAL-* got %q", note)
		}
	})

	t.Run("2D_foreign_deposit_rejected", func(t *testing.T) {
		medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
			CNK: "2558" + suffix[:4], Name: "DepX " + suffix, IsActive: true,
		})
		if err != nil {
			t.Fatalf("med: %v", err)
		}
		exp := time.Now().AddDate(0, 0, 120).Format("2006-01-02")
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
			"medicationId": medID,
			"depositId":    uuid.NewString(),
			"lotNumber":    "DEP-" + suffix,
			"expiresOn":    exp,
			"qty":          1,
		})
		if code != http.StatusBadRequest {
			t.Fatalf("foreign deposit want 400 got %d %#v", code, env)
		}
	})

	t.Run("2E_inventory_csv_export", func(t *testing.T) {
		cancelOpenInventorySessions(t, api, tok)
		medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
			CNK: "2559" + suffix[:4], Name: "InvCSV " + suffix, IsActive: true,
		})
		if err != nil {
			t.Fatalf("med: %v", err)
		}
		exp := time.Now().AddDate(0, 0, 200).Format("2006-01-02")
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
			"medicationId": medID, "lotNumber": "CSV-" + suffix, "expiresOn": exp, "qty": 4,
		})
		if code != http.StatusCreated {
			t.Fatalf("receipt %d %#v", code, env)
		}
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/inventory/sessions", tok, map[string]any{})
		if code != http.StatusCreated {
			t.Fatalf("start %d %#v", code, env)
		}
		sessionID, _ := dataMap(t, env)["id"].(string)
		lines, _ := dataMap(t, env)["lines"].([]any)
		counts := make([]map[string]any, 0, len(lines))
		for _, raw := range lines {
			ln, _ := raw.(map[string]any)
			counts = append(counts, map[string]any{
				"lineId": ln["id"], "countedQty": ln["systemQty"],
			})
		}
		code, env = doAuthJSON(t, api.handler, http.MethodPost,
			"/api/v1/vet/pharmacy/inventory/sessions/"+sessionID+"/close", tok, map[string]any{"counts": counts})
		if code != http.StatusOK {
			t.Fatalf("close %d %#v", code, env)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/vet/pharmacy/inventory/sessions/"+sessionID+"/export.csv", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rec := httptest.NewRecorder()
		api.handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("csv %d %s", rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		if !strings.Contains(body, "cnk;name;lot") || !strings.Contains(body, "CSV-"+suffix) {
			snip := body
			if len(snip) > 240 {
				snip = snip[:240]
			}
			t.Fatalf("csv missing header/lot: %s", snip)
		}
	})

	t.Run("S4_vamreg_retry_blocked_while_pending", func(t *testing.T) {
		abID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
			CNK: "2560" + suffix[:4], Name: "RetryPend " + suffix, IsActive: true, IsAntibiotic: true,
		})
		if err != nil {
			t.Fatalf("med: %v", err)
		}
		exp := time.Now().AddDate(0, 0, 120).Format("2006-01-02")
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
			"medicationId": abID, "lotNumber": "RP-" + suffix, "expiresOn": exp, "qty": 2,
		})
		if code != http.StatusCreated {
			t.Fatalf("receipt %d %#v", code, env)
		}
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
			"items": []map[string]any{{
				"medicationId": abID, "qty": 1, "ammNumber": "BE-RP-1",
				"vamregPayload": map[string]any{"species": "dog", "indication": "x", "durationDays": 3},
			}},
		})
		if code != http.StatusCreated {
			t.Fatalf("draft %d %#v", code, env)
		}
		dafID, _ := dataMap(t, env)["id"].(string)
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/finalize", tok, nil)
		if code != http.StatusOK {
			t.Fatalf("finalize %d %#v", code, env)
		}
		doc := dataMap(t, env)
		if nested, ok := doc["daf"].(map[string]any); ok {
			doc = nested
		}
		practiceID, _ := doc["practiceId"].(string)
		if _, err := api.pool.Exec(ctx, `
			UPDATE pharmacy.daf_documents SET vamreg_status = 'pending'
			WHERE practice_id = $1 AND id = $2`, practiceID, dafID); err != nil {
			t.Fatalf("force pending: %v", err)
		}
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/vamreg/retry", tok, nil)
		if code != http.StatusConflict {
			t.Fatalf("retry pending want 409 got %d %#v", code, env)
		}
		if errObj, _ := env["error"].(map[string]any); errObj["code"] != "daf_vamreg_in_flight" {
			t.Fatalf("want daf_vamreg_in_flight %#v", env)
		}
	})

	t.Run("4D_food_chain_banned_blocks_finalize", func(t *testing.T) {
		medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
			CNK: "2561" + suffix[:4], Name: "BannedFC " + suffix, IsActive: true,
		})
		if err != nil {
			t.Fatalf("med: %v", err)
		}
		code, env := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pharmacy/medications/"+medID+"/withdrawal", tok, map[string]any{
			"withdrawalMeatDays": 28, "withdrawalMilkDays": 7, "withdrawalEggsDays": 0,
			"foodChainBanned": true,
		})
		if code != http.StatusOK {
			t.Fatalf("ban overlay %d %#v", code, env)
		}
		petsCode, petsEnv := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/vet/pets", tok, nil)
		if petsCode != http.StatusOK {
			t.Fatalf("pets %d", petsCode)
		}
		var petID, clientID string
		items, _ := petsEnv["data"].([]any)
		for _, raw := range items {
			p, _ := raw.(map[string]any)
			if p["name"] == "Spirit" || p["species"] == "horse" {
				petID, _ = p["id"].(string)
				clientID, _ = p["ownerUserId"].(string)
				break
			}
		}
		if petID == "" {
			t.Skip("no horse in seed")
		}
		doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pets/"+petID+"/food-chain", tok, map[string]any{
			"foodChainStatus": "food_producing",
		})
		exp := time.Now().AddDate(0, 0, 200).Format("2006-01-02")
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
			"medicationId": medID, "lotNumber": "BAN-" + suffix, "expiresOn": exp, "qty": 2,
		})
		if code != http.StatusCreated {
			t.Fatalf("receipt %d %#v", code, env)
		}
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
			"petId": petID, "clientUserId": clientID,
			"items": []map[string]any{{"medicationId": medID, "qty": 1, "ammNumber": "BE-BAN"}},
		})
		if code != http.StatusCreated {
			t.Fatalf("draft %d %#v", code, env)
		}
		dafID, _ := dataMap(t, env)["id"].(string)
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/finalize", tok, nil)
		if code != http.StatusConflict {
			t.Fatalf("banned finalize want 409 got %d %#v", code, env)
		}
		if errObj, _ := env["error"].(map[string]any); errObj["code"] != "food_chain_banned_medication" {
			t.Fatalf("want food_chain_banned_medication %#v", env)
		}
	})

	t.Run("4F_pharmacy_user_fks_not_cascade_delete", func(t *testing.T) {
		rows, err := api.pool.Query(ctx, `
			SELECT c.conrelid::regclass::text AS tbl, c.confdeltype::text AS del
			FROM pg_constraint c
			JOIN pg_class ref ON ref.oid = c.confrelid
			JOIN pg_namespace nref ON nref.oid = ref.relnamespace
			WHERE c.contype = 'f'
			  AND nref.nspname = 'identity' AND ref.relname = 'users'
			  AND c.conrelid::regclass::text LIKE 'pharmacy.%'`)
		if err != nil {
			t.Fatalf("query fks: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var tbl, del string
			if err := rows.Scan(&tbl, &del); err != nil {
				t.Fatal(err)
			}
			if del == "c" {
				t.Fatalf("pharmacy FK %s ON DELETE CASCADE from identity.users — breaks 5y retention", tbl)
			}
		}
	})

	t.Run("block_expired_on_daf_false_allows_expired_lot", func(t *testing.T) {
		medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
			CNK: "2562" + suffix[:4], Name: "ExpDAF " + suffix, IsActive: true,
		})
		if err != nil {
			t.Fatalf("med: %v", err)
		}
		code, env := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pharmacy/settings", tok, map[string]any{
			"allowExpiredReceipt": true,
			"blockExpiredOnDaf":   false,
		})
		if code != http.StatusOK {
			t.Fatalf("settings %d %#v", code, env)
		}
		t.Cleanup(func() {
			doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pharmacy/settings", tok, map[string]any{
				"allowExpiredReceipt": false,
				"blockExpiredOnDaf":   true,
			})
		})
		expired := time.Now().AddDate(0, 0, -3).Format("2006-01-02")
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
			"medicationId": medID, "lotNumber": "EXP-" + suffix, "expiresOn": expired, "qty": 3,
		})
		if code != http.StatusCreated {
			t.Fatalf("expired receipt %d %#v", code, env)
		}
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
			"items": []map[string]any{{"medicationId": medID, "qty": 1, "ammNumber": "BE-EXP-1"}},
		})
		if code != http.StatusCreated {
			t.Fatalf("draft %d %#v", code, env)
		}
		dafID, _ := dataMap(t, env)["id"].(string)
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/"+dafID+"/finalize", tok, nil)
		if code != http.StatusOK {
			t.Fatalf("finalize expired lot want 200 got %d %#v", code, env)
		}
	})

	t.Run("daf_foreign_deposit_rejected", func(t *testing.T) {
		medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
			CNK: "2563" + suffix[:4], Name: "DepDAF " + suffix, IsActive: true,
		})
		if err != nil {
			t.Fatalf("med: %v", err)
		}
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
			"items": []map[string]any{{
				"medicationId": medID, "qty": 1, "ammNumber": "BE-DEP-X",
				"depositId": uuid.NewString(),
			}},
		})
		if code != http.StatusBadRequest {
			t.Fatalf("foreign deposit DAF want 400 got %d %#v", code, env)
		}
	})

	t.Run("daf_invalid_deposit_uuid_rejected", func(t *testing.T) {
		medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
			CNK: "2564" + suffix[:4], Name: "BadUUID " + suffix, IsActive: true,
		})
		if err != nil {
			t.Fatalf("med: %v", err)
		}
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf", tok, map[string]any{
			"items": []map[string]any{{
				"medicationId": medID, "qty": 1, "ammNumber": "BE-BAD-UUID",
				"depositId": "not-a-uuid",
			}},
		})
		if code != http.StatusBadRequest {
			t.Fatalf("invalid deposit UUID want 400 got %d %#v", code, env)
		}
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/daf/preview-fefo", tok, map[string]any{
			"items": []map[string]any{{
				"medicationId": medID, "qty": 1, "depositId": uuid.NewString(),
			}},
		})
		if code != http.StatusBadRequest {
			t.Fatalf("preview foreign deposit want 400 got %d %#v", code, env)
		}
	})

	t.Run("block_expired_on_adjust_out_false", func(t *testing.T) {
		medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
			CNK: "2565" + suffix[:4], Name: "AdjExp " + suffix, IsActive: true,
		})
		if err != nil {
			t.Fatalf("med: %v", err)
		}
		code, env := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pharmacy/settings", tok, map[string]any{
			"allowExpiredReceipt":     true,
			"blockExpiredOnAdjustOut": false,
		})
		if code != http.StatusOK {
			t.Fatalf("settings %d %#v", code, env)
		}
		t.Cleanup(func() {
			doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/vet/pharmacy/settings", tok, map[string]any{
				"allowExpiredReceipt":     false,
				"blockExpiredOnAdjustOut": true,
			})
		})
		expired := time.Now().AddDate(0, 0, -5).Format("2006-01-02")
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches", tok, map[string]any{
			"medicationId": medID, "lotNumber": "ADJ-" + suffix, "expiresOn": expired, "qty": 5,
		})
		if code != http.StatusCreated {
			t.Fatalf("receipt %d %#v", code, env)
		}
		batchObj, _ := dataMap(t, env)["batch"].(map[string]any)
		batchID, _ := batchObj["id"].(string)
		if batchID == "" {
			t.Fatalf("missing batch id %#v", env)
		}
		code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/batches/"+batchID+"/adjust", tok, map[string]any{
			"delta": -1, "detail": "allow expired adjust",
		})
		if code != http.StatusOK {
			t.Fatalf("adjust expired want 200 got %d %#v", code, env)
		}
	})

	t.Run("export_includes_pharmacy_daf_keys", func(t *testing.T) {
		code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/me/export", tok, nil)
		if code != http.StatusOK {
			t.Fatalf("export %d %#v", code, env)
		}
		data := dataMap(t, env)
		for _, key := range []string{"pharmacyDafAsClient", "pharmacyDafAsPetOwner", "pharmacyDafAsPrescriber"} {
			if _, ok := data[key]; !ok {
				t.Fatalf("export missing %s keys=%v", key, keysOf(data))
			}
		}
	})
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestEnqueueInvoicesConnectResellerGate(t *testing.T) {
	err := pharmacy.EnqueueInvoicesConnect("any-daf")
	if !errors.Is(err, pharmacy.ErrInvoicesConnectResellerPending) {
		t.Fatalf("want ErrInvoicesConnectResellerPending got %v", err)
	}
}
