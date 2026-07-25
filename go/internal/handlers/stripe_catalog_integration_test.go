package handlers_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminStripeCatalogGetACL(t *testing.T) {
	api := newTestAPI(t)

	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/stripe-catalog", adminTok, nil)
	if code != http.StatusOK {
		t.Fatalf("admin get %d %#v", code, env)
	}
	data, ok := env["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected catalog object, got %#v", env["data"])
	}
	if _, ok := data["products"]; !ok {
		t.Fatalf("missing products: %#v", data)
	}
	if _, ok := data["prices"]; !ok {
		t.Fatalf("missing prices: %#v", data)
	}

	vetTok := loginToken(t, api.handler, "vet.demo@petsfollow.test", "VetDemo123!")
	code, env = doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/admin/stripe-catalog", vetTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("vet should be forbidden, got %d %#v", code, env)
	}
}

func TestAdminStripeCatalogImportProducts(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	csv := `"id","Name","Description","Tax Code","plan_slug (metadata)"
"prod_test_e2e_hr","E2E Test Product","import test",,"annual"
`
	code, env := doAuthMultipart(t, api.handler, "/api/v1/admin/stripe-catalog/import", adminTok, "products", "products.csv", csv)
	if code != http.StatusOK {
		t.Fatalf("import products %d %#v", code, env)
	}
	data := dataMap(t, env)
	inserted, _ := data["inserted"].(float64)
	updated, _ := data["updated"].(float64)
	if inserted+updated < 1 {
		t.Fatalf("expected insert/update, got %#v", data)
	}
}

func TestAdminStripeCatalogImportErrors(t *testing.T) {
	api := newTestAPI(t)
	adminTok := loginToken(t, api.handler, "admin.demo@petsfollow.test", "AdminDemo123!")

	code, env := doAuthMultipart(t, api.handler, "/api/v1/admin/stripe-catalog/import", adminTok, "", "x.csv", "a,b\n")
	if code != http.StatusBadRequest || errCode(env) != "bad_request" {
		t.Fatalf("kind_required: got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/admin/stripe-catalog/import?kind=products", adminTok, nil)
	if code == http.StatusOK {
		t.Fatal("expected error without file")
	}

	code, env = doAuthMultipart(t, api.handler, "/api/v1/admin/stripe-catalog/import", adminTok, "products", "empty.csv", `"id","Name"`+"\n")
	if code != http.StatusBadRequest || errCode(env) != "bad_request" {
		t.Fatalf("empty_csv: got %d %#v", code, env)
	}

	badPrices := `"Price ID","Product ID","Amount"
"price_x","prod_x","nope"
`
	code, env = doAuthMultipart(t, api.handler, "/api/v1/admin/stripe-catalog/import", adminTok, "prices", "bad.csv", badPrices)
	if code != http.StatusBadRequest || errCode(env) != "bad_request" {
		t.Fatalf("invalid_csv: got %d %#v", code, env)
	}
}

func doAuthMultipart(t *testing.T, h http.Handler, path, token, kind, filename, content string) (int, map[string]any) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if kind != "" {
		if err := w.WriteField("kind", kind); err != nil {
			t.Fatal(err)
		}
	}
	if filename != "" {
		part, err := w.CreateFormFile("file", filename)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Accept-Language", "fr")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	return rec.Code, envelope
}
