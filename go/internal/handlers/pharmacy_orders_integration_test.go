package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestPharmacyOrdersAndDeliveryNote(t *testing.T) {
	t.Setenv("PHARMACY_ENABLED", "true")
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)
	suffix := uuid.NewString()[:8]
	medID, err := st.UpsertRefMedication(ctx, store.RefMedicationUpsert{
		CNK: "2999" + suffix[:4], Name: "Order Demo Med " + suffix, IsActive: true,
	})
	if err != nil {
		t.Fatalf("med: %v", err)
	}
	tok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")

	supplierEmail := fmt.Sprintf("fournisseur-%s@petsfollow.test", suffix)
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/suppliers", tok, map[string]any{
		"name":  "Alcyon Smoke " + suffix,
		"email": supplierEmail,
	})
	if code != http.StatusOK {
		t.Fatalf("supplier %d %#v", code, env)
	}
	supplierID, _ := dataMap(t, env)["id"].(string)
	if supplierID == "" {
		t.Fatal("missing supplier id")
	}

	// Seuil élevé → alerte même si du stock résiduel existe d'un run précédent.
	code, env = doAuthJSON(t, api.handler, http.MethodPut, "/api/v1/vet/pharmacy/reorder-thresholds", tok, map[string]any{
		"medicationId": medID, "minQty": 10_000,
	})
	if code != http.StatusOK {
		t.Fatalf("threshold %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/orders", tok, map[string]any{
		"fromAlerts": true,
		"supplierId": supplierID,
	})
	if code != http.StatusCreated {
		t.Fatalf("create order %d %#v", code, env)
	}
	orderID, _ := dataMap(t, env)["id"].(string)
	if orderID == "" {
		t.Fatal("missing order id")
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/orders/"+orderID+"/send", tok, map[string]any{
		"toEmail": "random-spam@example.com",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("arbitrary toEmail want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/orders/"+orderID+"/send", tok, map[string]any{
		"toEmail":    supplierEmail,
		"supplierId": supplierID,
	})
	if code != http.StatusOK {
		t.Fatalf("send order %d %#v", code, env)
	}
	if dataMap(t, env)["status"] != "sent" {
		t.Fatalf("status=%v", dataMap(t, env)["status"])
	}

	exp := time.Now().AddDate(0, 0, 180).Format("2006-01-02")
	noteNumber := fmt.Sprintf("BL-%s", suffix)
	lotNumber := fmt.Sprintf("LOT-%s", suffix)
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/delivery-notes", tok, map[string]any{
		"noteNumber":   noteNumber,
		"supplierName": "Alcyon Demo",
		"notify":       true,
		"items": []map[string]any{{
			"medicationId": medID, "lotNumber": lotNumber, "expiresOn": exp, "qty": 10,
		}},
	})
	if code != http.StatusCreated {
		t.Fatalf("delivery note %d %#v", code, env)
	}
	dn := dataMap(t, env)
	if dn["noteNumber"] != noteNumber {
		t.Fatalf("note=%v", dn["noteNumber"])
	}
	items, _ := dn["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("items=%v", items)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/delivery-notes", tok, map[string]any{
		"noteNumber":   noteNumber,
		"supplierName": "Alcyon Demo",
		"items": []map[string]any{{
			"medicationId": medID, "lotNumber": lotNumber + "-b", "expiresOn": exp, "qty": 1,
		}},
	})
	if code != http.StatusConflict {
		t.Fatalf("dup BL want 409 got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj["code"] != "delivery_note_exists" {
		t.Fatalf("dup code %#v", env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/orders", tok, map[string]any{
		"items": []map[string]any{{"medicationId": uuid.NewString(), "qty": 1}},
	})
	if code != http.StatusNotFound {
		t.Fatalf("bad med want 404 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/orders", tok, map[string]any{
		"supplierId": uuid.NewString(),
		"items":      []map[string]any{{"medicationId": medID, "qty": 1}},
	})
	if code != http.StatusBadRequest {
		t.Fatalf("foreign supplier want 400 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/orders", tok, map[string]any{
		"items": []map[string]any{{"medicationId": medID, "qty": 2}},
	})
	if code != http.StatusCreated {
		t.Fatalf("draft for soft-claim %d %#v", code, env)
	}
	claimOrderID, _ := dataMap(t, env)["id"].(string)
	practiceID, _ := dataMap(t, env)["practiceId"].(string)
	if _, err := api.pool.Exec(ctx, `
		UPDATE pharmacy.purchase_orders
		SET to_email = $3, updated_at = now()
		WHERE practice_id = $1 AND id = $2 AND status = 'draft'`, practiceID, claimOrderID, supplierEmail); err != nil {
		t.Fatalf("force soft-claim: %v", err)
	}
	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/vet/pharmacy/orders/"+claimOrderID+"/send", tok, map[string]any{
		"toEmail":    supplierEmail,
		"supplierId": supplierID,
	})
	if code != http.StatusConflict {
		t.Fatalf("fresh soft-claim want 409 order_send_in_progress got %d %#v", code, env)
	}
	if errObj, _ := env["error"].(map[string]any); errObj["code"] != "order_send_in_progress" {
		t.Fatalf("conflict code %#v", env)
	}
	// Stale claim (past TTL): store reclaim + sendFn OK → sent (HTTP path still needs SMTP).
	if _, err := api.pool.Exec(ctx, `
		UPDATE pharmacy.purchase_orders
		SET to_email = $3,
		    updated_at = now() - interval '3 minutes'
		WHERE practice_id = $1 AND id = $2 AND status = 'draft'`, practiceID, claimOrderID, supplierEmail); err != nil {
		t.Fatalf("stale soft-claim: %v", err)
	}
	staleMarked, err := st.SendPurchaseOrderLocked(ctx, practiceID, claimOrderID, supplierEmail, supplierID, func(o store.PurchaseOrder) error {
		return nil
	})
	if err != nil {
		t.Fatalf("stale reclaim send: %v", err)
	}
	if staleMarked.Status != "sent" {
		t.Fatalf("reclaim status=%v", staleMarked.Status)
	}
}
