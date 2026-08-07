package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
	"github.com/olegrand1976/petsFollow/go/internal/platform/media"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

// Les enregistrements de pitch vivent sous pitch-sims/, namespace privé depuis
// le passage en allowlist. L'API ne doit donc plus laisser filtrer ni la clé de
// stockage ni une URL publique, et le flux authentifié doit rester cloisonné au
// propriétaire et à son manager.
func TestPitchAudioIsPrivateAndScopedToOwnerAndManager(t *testing.T) {
	api := newTestAPI(t)
	ctx := context.Background()
	st := store.New(api.pool)

	bundle, err := media.New(config.Config{
		MediaLocalDir: t.TempDir(),
		APIPublicURL:  "http://127.0.0.1:8291",
	})
	if err != nil {
		t.Fatalf("media: %v", err)
	}
	api.api.TestSetMedia(bundle.Store)

	mgrEmail := uniqueEmail("pa-mgr")
	mgrID := insertVerifiedUser(t, api, "commercial_manager", mgrEmail, "CommercialDemo123!", "Pitch Mgr", nil)
	ownerEmail := uniqueEmail("pa-owner")
	ownerID := insertVerifiedUser(t, api, "commercial", ownerEmail, "CommercialDemo123!", "Pitch Owner",
		map[string]any{"manager_user_id": mgrID})
	strangerEmail := uniqueEmail("pa-stranger")
	insertVerifiedUser(t, api, "commercial", strangerEmail, "CommercialDemo123!", "Pitch Stranger", nil)

	var scriptID string
	if err := api.pool.QueryRow(ctx,
		`SELECT id::text FROM sales.pitch_scripts ORDER BY created_at LIMIT 1`).Scan(&scriptID); err != nil {
		t.Skipf("no seeded pitch script (make seed?): %v", err)
	}

	sim, err := st.CreatePitchSimulation(ctx, store.PitchSimulation{
		UserID:        ownerID,
		ScriptID:      scriptID,
		InterestLevel: "neutre",
		Outcome:       "appointment",
	})
	if err != nil {
		t.Fatalf("create sim: %v", err)
	}

	ownerTok := loginToken(t, api.handler, ownerEmail, "CommercialDemo123!")
	mgrTok := loginToken(t, api.handler, mgrEmail, "CommercialDemo123!")
	strangerTok := loginToken(t, api.handler, strangerEmail, "CommercialDemo123!")

	recording := []byte("webm-pitch-recording")
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	hdr := make(textproto.MIMEHeader)
	hdr.Set("Content-Disposition", `form-data; name="file"; filename="call.webm"`)
	hdr.Set("Content-Type", "audio/webm")
	part, err := w.CreatePart(hdr)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(recording); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/commercial/pitch-sims/"+sim.ID+"/audio", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+ownerTok)
	rec := httptest.NewRecorder()
	api.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload pitch audio %d %s", rec.Code, rec.Body.String())
	}
	// La réponse d'upload ne doit plus porter d'URL ni de clé de stockage.
	if raw := rec.Body.String(); jsonHasAnyKey(t, raw, "url", "objectKey", "audioUrl", "audioObjectKey") {
		t.Fatalf("upload response leaks storage location: %s", raw)
	}

	code, env := doAuthJSON(t, api.handler, http.MethodGet, "/api/v1/commercial/pitch-sims", ownerTok, nil)
	if code != http.StatusOK {
		t.Fatalf("list sims %d %#v", code, env)
	}
	listed, _ := json.Marshal(env)
	if jsonHasAnyKey(t, string(listed), "audioUrl", "audioObjectKey") {
		t.Fatalf("history leaks storage location: %s", listed)
	}
	if !bytes.Contains(listed, []byte(`"hasAudio":true`)) {
		t.Fatalf("history must flag playback availability: %s", listed)
	}

	stream := "/api/v1/commercial/pitch-sims/" + sim.ID + "/audio"
	for _, tc := range []struct{ name, token string }{
		{"owner", ownerTok},
		{"manager of the owner", mgrTok},
	} {
		req = httptest.NewRequest(http.MethodGet, stream, nil)
		req.Header.Set("Authorization", "Bearer "+tc.token)
		rec = httptest.NewRecorder()
		api.handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s stream %d %s", tc.name, rec.Code, rec.Body.String())
		}
		got, _ := io.ReadAll(rec.Body)
		if !bytes.Equal(got, recording) {
			t.Fatalf("%s got %d bytes, want the uploaded recording", tc.name, len(got))
		}
		if cc := rec.Header().Get("Cache-Control"); cc != "private, no-store" {
			t.Fatalf("%s Cache-Control=%q — a proxy must not retain the recording", tc.name, cc)
		}
	}

	for _, tc := range []struct{ name, token string }{
		{"anonymous", ""},
		{"commercial of another team", strangerTok},
	} {
		req = httptest.NewRequest(http.MethodGet, stream, nil)
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		rec = httptest.NewRecorder()
		api.handler.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			t.Fatalf("%s must not stream the recording (got 200)", tc.name)
		}
	}
}

// jsonHasAnyKey reports whether the payload carries any of the given JSON keys,
// at any depth — a leak buried in a nested object counts.
func jsonHasAnyKey(t *testing.T, payload string, keys ...string) bool {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(payload), &v); err != nil {
		t.Fatalf("payload is not JSON (%v): %s", err, payload)
	}
	var walk func(any) bool
	walk = func(node any) bool {
		switch n := node.(type) {
		case map[string]any:
			for k, child := range n {
				for _, want := range keys {
					if k == want {
						return true
					}
				}
				if walk(child) {
					return true
				}
			}
		case []any:
			for _, child := range n {
				if walk(child) {
					return true
				}
			}
		}
		return false
	}
	return walk(v)
}
