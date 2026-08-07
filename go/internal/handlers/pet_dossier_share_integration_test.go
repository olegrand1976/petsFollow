package handlers_test

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olegrand1976/petsFollow/go/internal/notifications/email"
	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestPetDossierShareCreateMetaDownloadAndExpiry(t *testing.T) {
	api := newTestAPI(t)
	bundle, err := media.New(config.Config{
		MediaLocalDir: t.TempDir(),
		APIPublicURL:  "http://127.0.0.1:8291",
	})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/dossier-shares", clientTok, map[string]any{
		"email": "pro.externe@petsfollow.test",
	})
	if code != http.StatusCreated {
		t.Fatalf("create share %d %#v", code, env)
	}
	data := dataMap(t, env)
	token, _ := data["token"].(string)
	if token == "" {
		t.Fatalf("expected DEV token in response %#v", data)
	}

	code, env = doJSON(t, api.handler, http.MethodGet, "/api/v1/public/pet-dossier/"+token, nil)
	if code != http.StatusOK {
		t.Fatalf("meta %d %#v", code, env)
	}
	meta := dataMap(t, env)
	if meta["petName"] == "" {
		t.Fatalf("expected petName in meta %#v", meta)
	}
	if phone, _ := meta["commercialPhone"].(string); phone == "" {
		t.Fatalf("expected commercialPhone (seed or fallback) %#v", meta)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/pet-dossier/"+token+"/download", nil)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("download %d %s", rec.Code, rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "zip") {
		t.Fatalf("content-type=%q", ct)
	}
	zipBytes := rec.Body.Bytes()
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip: %v", err)
	}
	foundPDF := false
	for _, f := range zr.File {
		if f.Name == "dossier.pdf" {
			foundPDF = true
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			head, _ := io.ReadAll(io.LimitReader(rc, 5))
			_ = rc.Close()
			if string(head) != "%PDF-" {
				t.Fatalf("dossier.pdf magic %q", head)
			}
		}
	}
	if !foundPDF {
		t.Fatal("dossier.pdf missing from zip")
	}

	// Expire token and expect 410.
	if _, err := api.pool.Exec(context.Background(), `
		UPDATE pets.dossier_share_tokens SET expires_at = $1 WHERE token = $2`,
		time.Now().UTC().Add(-time.Minute), token); err != nil {
		t.Fatalf("expire: %v", err)
	}
	code, env = doJSON(t, api.handler, http.MethodGet, "/api/v1/public/pet-dossier/"+token, nil)
	if code != http.StatusGone {
		t.Fatalf("expired meta want 410 got %d %#v", code, env)
	}
	var n int
	if err := api.pool.QueryRow(context.Background(),
		`SELECT COUNT(*)::int FROM pets.dossier_share_tokens WHERE token = $1`, token).Scan(&n); err != nil {
		t.Fatalf("count after expiry: %v", err)
	}
	if n != 0 {
		t.Fatalf("partage expiré doit être purgé (row+email tiers), %d ligne(s)", n)
	}
	// Token disparu → 404 (plus 410) : cohérent avec purge à l'expiry.
	code, env = doJSON(t, api.handler, http.MethodGet, "/api/v1/public/pet-dossier/"+token, nil)
	if code != http.StatusNotFound {
		t.Fatalf("après purge want 404 got %d %#v", code, env)
	}
}

func TestPetDossierShareOwnerOnly(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	otherTok := loginToken(t, api.handler, "client.marie@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/dossier-shares", otherTok, map[string]any{
		"email": "pro@example.com",
	})
	if code != http.StatusForbidden {
		t.Fatalf("non-owner want 403 got %d %#v", code, env)
	}
}

// Le destinataire est libre : refuser les adresses malformées avant de créer un
// token, et surtout ne pas laisser passer un CR/LF vers la couche SMTP.
func TestPetDossierShareRejectsMalformedEmail(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	for _, bad := range []string{
		"",
		"pas-un-email",
		"pro@example.com\r\nBcc: victime@example.com",
		"pro@example.com, autre@example.com",
	} {
		code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/dossier-shares", clientTok, map[string]any{
			"email": bad,
		})
		if code != http.StatusBadRequest {
			t.Fatalf("email %q : attendu 400, obtenu %d %#v", bad, code, env)
		}
	}
}

func TestPetDossierShareDailyQuota(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	ctx := context.Background()
	var ownerID string
	if err := api.pool.QueryRow(ctx,
		`SELECT id::text FROM identity.users WHERE email='client.demo@petsfollow.test'`).Scan(&ownerID); err != nil {
		t.Fatalf("owner: %v", err)
	}
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(),
			`DELETE FROM pets.dossier_share_tokens WHERE recipient_email='quota@petsfollow.test'`)
	})
	// Saturer le quota du jour sans passer par SMTP.
	for i := range store.MaxDossierSharesPerDay {
		if _, err := api.pool.Exec(ctx, `
			INSERT INTO pets.dossier_share_tokens (id, token, pet_id, owner_user_id, recipient_email, expires_at)
			VALUES ($1::uuid, $2, $3::uuid, $4::uuid, 'quota@petsfollow.test', NOW() + INTERVAL '1 day')`,
			uuid.NewString(), uuid.NewString(), petID, ownerID); err != nil {
			t.Fatalf("insert quota row %d: %v", i, err)
		}
	}

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/dossier-shares", clientTok, map[string]any{
		"email": "pro.quota@petsfollow.test",
	})
	if code != http.StatusTooManyRequests {
		t.Fatalf("quota dépassé : attendu 429, obtenu %d %#v", code, env)
	}
}

// Un partage périmé garde en base l'email d'un tiers qui n'a jamais rien signé :
// le chemin de création purge les partages expirés du propriétaire.
func TestPetDossierSharePurgesExpiredRowsOnCreate(t *testing.T) {
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)
	ctx := context.Background()

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/dossier-shares", clientTok, map[string]any{
		"email": "pro.purge@petsfollow.test",
	})
	if code != http.StatusCreated {
		t.Fatalf("create share %d %#v", code, env)
	}
	stale, _ := dataMap(t, env)["token"].(string)
	if stale == "" {
		t.Fatalf("token DEV attendu %#v", env)
	}
	if _, err := api.pool.Exec(ctx, `
		UPDATE pets.dossier_share_tokens SET expires_at = $1 WHERE token = $2`,
		time.Now().UTC().Add(-time.Minute), stale); err != nil {
		t.Fatalf("expire: %v", err)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/dossier-shares", clientTok, map[string]any{
		"email": "pro.purge2@petsfollow.test",
	})
	if code != http.StatusCreated {
		t.Fatalf("second share %d %#v", code, env)
	}
	var n int
	if err := api.pool.QueryRow(ctx,
		`SELECT COUNT(*)::int FROM pets.dossier_share_tokens WHERE token = $1`, stale).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Fatalf("le partage périmé doit être supprimé, %d ligne(s) restante(s)", n)
	}
}

func TestPublicPetDossierIsRateLimited(t *testing.T) {
	t.Setenv("AUTH_RATE_LIMIT_PER_MIN", "3")
	api := newTestAPI(t)

	throttled := false
	for i := range 8 {
		code, env := doJSON(t, api.handler, http.MethodGet, "/api/v1/public/pet-dossier/inconnu-"+uuid.NewString(), nil)
		if code == http.StatusTooManyRequests {
			throttled = true
			break
		}
		if code != http.StatusNotFound {
			t.Fatalf("essai %d : %d %#v", i, code, env)
		}
	}
	if !throttled {
		t.Fatal("la consultation publique d'un dossier doit être rate-limitée (énumération de tokens)")
	}
}

func TestPetDossierShareRollsBackTokenOnSMTPFailure(t *testing.T) {
	api := newTestAPI(t)
	// Prod-like SMTP (user set, pass empty) → SendCritical échoue (pas soft-fail dev).
	api.api.TestReplaceNotifier(email.NewNotifierAuth(
		"smtp.example.com", 587, "test@petsfollow.test",
		"smtp-user", "", "http://localhost:3002", "https://ll-it-sc.be",
	))

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)
	recipient := "pro.smtp-fail@petsfollow.test"

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/pets/"+petID+"/dossier-shares", clientTok, map[string]any{
		"email": recipient,
	})
	if code != http.StatusBadGateway {
		t.Fatalf("SMTP fail want 502 got %d %#v", code, env)
	}
	var n int
	if err := api.pool.QueryRow(context.Background(),
		`SELECT COUNT(*)::int FROM pets.dossier_share_tokens WHERE recipient_email = $1`, recipient).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Fatalf("token orphelin après échec SMTP: %d ligne(s)", n)
	}
}

func TestUpdateMeContactPhoneCommercial(t *testing.T) {
	api := newTestAPI(t)
	tok := loginToken(t, api.handler, "commercial.demo@petsfollow.test", "CommercialDemo123!")

	code, env := doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/me", tok, map[string]any{
		"contactPhone": "0499 00 11 22",
	})
	if code != http.StatusOK {
		t.Fatalf("patch phone %d %#v", code, env)
	}
	me := dataMap(t, env)
	if me["contactPhone"] != "0499 00 11 22" {
		t.Fatalf("contactPhone=%v", me["contactPhone"])
	}

	// Restore seed phone for other tests / demos.
	_, _ = doAuthJSON(t, api.handler, http.MethodPatch, "/api/v1/me", tok, map[string]any{
		"contactPhone": "0470 12 34 56",
	})
}
