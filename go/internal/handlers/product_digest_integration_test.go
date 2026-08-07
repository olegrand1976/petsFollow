package handlers_test

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/notifications/email"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func TestProductDigestIngestAndRunWithBranch(t *testing.T) {
	t.Setenv("PRODUCT_DIGEST_SECRET", "test-product-digest-secret")
	api := newTestAPI(t)
	st := store.New(api.pool)

	digestDate := time.Date(2099, 1, 15, 0, 0, 0, 0, time.UTC)
	dateStr := digestDate.Format("2006-01-02")
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(),
			`DELETE FROM ops.product_digest_sends WHERE digest_date = $1::date`, digestDate)
		_, _ = api.pool.Exec(context.Background(),
			`DELETE FROM ops.product_digests WHERE digest_date = $1::date`, digestDate)
	})

	// Bypass Gemini: upsert a ready digest with branch metadata.
	now := time.Now().UTC()
	meta, _ := json.Marshal(map[string]any{
		"branch":      "staging",
		"environment": "staging",
		"source":      "staging",
		"commitCount": 1,
	})
	if err := st.UpsertProductDigest(context.Background(), store.ProductDigest{
		DigestDate: digestDate,
		Headline:   "Rappels de soins plus visibles",
		BodyText:   "• Les rappels Care apparaissent sur la timeline",
		HeadlineByLocale: map[string]string{
			"fr": "Rappels de soins plus visibles",
			"en": "Care reminders more visible",
		},
		BodyByLocale: map[string]string{
			"fr": "• Les rappels Care apparaissent sur la timeline",
			"en": "• Care reminders show on the timeline",
		},
		CommitsJSON: []byte(`[{"sha":"abc","subject":"feat: care reminders"}]`),
		Status:      "ready",
		GeneratedAt: &now,
		Meta:        meta,
	}); err != nil {
		t.Fatalf("upsert digest: %v", err)
	}

	smtpHost, smtpPort, apiBase := resolveMailhog(t)
	mailhogUp := smtpPort > 0
	if mailhogUp {
		api.api.TestReplaceNotifier(email.NewNotifier(
			smtpHost, smtpPort, "digest@petsfollow.test",
			"http://localhost:3002", "https://ll-it-sc.be",
		))
		req, _ := http.NewRequest(http.MethodDelete, apiBase+"/api/v1/messages", nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
		}
	} else {
		// Default test notifier is isDevSMTP (soft-fail) — force hard-fail to assert retry path.
		api.api.TestReplaceNotifier(email.NewNotifierAuth(
			"smtp.invalid.petsfollow", 587, "digest@petsfollow.test",
			"smtp-user", "", "http://localhost:3002", "https://ll-it-sc.be",
		))
	}

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/internal/product-digest/run", map[string]any{
		"date": dateStr,
	})
	if code != http.StatusUnauthorized {
		t.Fatalf("missing secret want 401 got %d %#v", code, env)
	}

	code, env = doJSONWithHeaders(t, api.handler, http.MethodPost, "/api/v1/internal/product-digest/run", map[string]any{
		"date": dateStr,
	}, map[string]string{
		"X-Product-Digest-Secret": "test-product-digest-secret",
	})
	if code != http.StatusOK {
		t.Fatalf("run %d %#v", code, env)
	}
	data := dataMap(t, env)
	if branch, _ := data["branch"].(string); !strings.Contains(branch, "staging") || !strings.Contains(branch, "TEST") {
		t.Fatalf("branch %#v want staging + TEST tag", data["branch"])
	}

	if !mailhogUp {
		if data["status"] != "failed" {
			t.Fatalf("without MailHog want status=failed got %#v", data)
		}
		t.Log("MailHog SMTP not reachable — skipped live SMTP assertion (branch + failed status checked)")
		return
	}
	if data["status"] != "sent" && data["status"] != "partial" {
		t.Fatalf("status %#v", data)
	}
	sent, _ := data["sent"].(float64)
	if sent < 1 {
		t.Fatalf("expected at least 1 email via MailHog, got sent=%v failed=%v", data["sent"], data["failed"])
	}
	deadline := time.Now().Add(5 * time.Second)
	found := false
	for time.Now().Before(deadline) {
		resp, err := http.Get(apiBase + "/api/v2/messages")
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			raw := string(body)
			if strings.Contains(raw, "environnement de TEST") || strings.Contains(raw, "[staging") {
				found = true
				break
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !found {
		t.Fatal("MailHog: no message mentioning staging branch")
	}
}

func TestProductDigestListRequiresAuthAndRole(t *testing.T) {
	t.Setenv("PRODUCT_DIGEST_SECRET", "test-product-digest-secret")
	api := newTestAPI(t)
	st := store.New(api.pool)

	digestDate := time.Date(2099, 3, 10, 0, 0, 0, 0, time.UTC)
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(),
			`DELETE FROM ops.product_digests WHERE digest_date = $1::date`, digestDate)
	})
	now := time.Now().UTC()
	meta, _ := json.Marshal(map[string]any{"environment": "staging"})
	if err := st.UpsertProductDigest(context.Background(), store.ProductDigest{
		DigestDate:       digestDate,
		Headline:         "Liste test",
		BodyText:         "• Item",
		HeadlineByLocale: map[string]string{"fr": "Liste test"},
		BodyByLocale:     map[string]string{"fr": "• Item"},
		CommitsJSON:      []byte(`[]`),
		Status:           "ready",
		GeneratedAt:      &now,
		Meta:             meta,
	}); err != nil {
		t.Fatal(err)
	}

	code, _ := doJSON(t, api.handler, http.MethodGet, "/api/v1/product-digests", nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", code)
	}

	code, env := doJSON(t, api.handler, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": "vet.demo@petsfollow.test", "password": "VetDemo123!",
	})
	if code != http.StatusOK {
		t.Fatalf("login %d %#v", code, env)
	}
	tok, _ := dataMap(t, env)["accessToken"].(string)
	code, env = doJSONWithHeaders(t, api.handler, http.MethodGet, "/api/v1/product-digests?limit=5", nil, map[string]string{
		"Authorization": "Bearer " + tok,
	})
	if code != http.StatusOK {
		t.Fatalf("list %d %#v", code, env)
	}
	data := dataMap(t, env)
	items, _ := data["items"].([]any)
	if len(items) < 1 {
		t.Fatalf("expected items, got %#v", data)
	}
}

func TestProductDigestWeeklyRunEmptyAndSend(t *testing.T) {
	t.Setenv("PRODUCT_DIGEST_SECRET", "test-product-digest-secret")
	t.Setenv("APP_ENV", "staging")
	api := newTestAPI(t)
	st := store.New(api.pool)

	weekDay := time.Date(2099, 4, 11, 0, 0, 0, 0, time.UTC) // Saturday
	weekStart := time.Date(2099, 4, 6, 0, 0, 0, 0, time.UTC) // Monday
	digestDate := time.Date(2099, 4, 10, 0, 0, 0, 0, time.UTC)
	staffEmail := strings.Replace(uniqueEmail("weekly-digest"), "@petsfollow.test", "@example.com", 1)

	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(),
			`DELETE FROM ops.product_digest_weekly_sends WHERE week_start = $1::date`, weekStart)
		_, _ = api.pool.Exec(context.Background(),
			`DELETE FROM ops.product_digests WHERE digest_date = $1::date`, digestDate)
		_, _ = api.pool.Exec(context.Background(),
			`DELETE FROM identity.users WHERE email = $1`, staffEmail)
	})

	code, env := doJSONWithHeaders(t, api.handler, http.MethodPost, "/api/v1/internal/product-digest/weekly-run", map[string]any{
		"date": weekDay.Format("2006-01-02"),
	}, map[string]string{
		"X-Product-Digest-Secret": "test-product-digest-secret",
	})
	if code != http.StatusOK {
		t.Fatalf("empty week %d %#v", code, env)
	}
	if dataMap(t, env)["reason"] != "empty_week" {
		t.Fatalf("want empty_week got %#v", env)
	}

	now := time.Now().UTC()
	meta, _ := json.Marshal(map[string]any{"environment": "staging", "branch": "staging"})
	if err := st.UpsertProductDigest(context.Background(), store.ProductDigest{
		DigestDate:       digestDate,
		Headline:         "Hebdo item",
		BodyText:         "• Point A",
		HeadlineByLocale: map[string]string{"fr": "Hebdo item"},
		BodyByLocale:     map[string]string{"fr": "• Point A"},
		CommitsJSON:      []byte(`[]`),
		Status:           "ready",
		GeneratedAt:      &now,
		Meta:             meta,
	}); err != nil {
		t.Fatal(err)
	}

	staffID := insertVerifiedUser(t, api, "admin", staffEmail, "AdminDemo123!", "Weekly Staff", nil)

	smtpHost, smtpPort, _ := resolveMailhog(t)
	if smtpPort > 0 {
		api.api.TestReplaceNotifier(email.NewNotifier(
			smtpHost, smtpPort, "digest@petsfollow.test",
			"http://localhost:3002", "https://ll-it-sc.be",
		))
	} else {
		api.api.TestReplaceNotifier(email.NewNotifierAuth(
			"smtp.invalid.petsfollow", 587, "digest@petsfollow.test",
			"smtp-user", "", "http://localhost:3002", "https://ll-it-sc.be",
		))
	}

	code, env = doJSONWithHeaders(t, api.handler, http.MethodPost, "/api/v1/internal/product-digest/weekly-run", map[string]any{
		"date": weekDay.Format("2006-01-02"),
	}, map[string]string{
		"X-Product-Digest-Secret": "test-product-digest-secret",
	})
	if code != http.StatusOK {
		t.Fatalf("weekly run %d %#v", code, env)
	}
	data := dataMap(t, env)
	if data["weekStart"] != weekStart.Format("2006-01-02") {
		t.Fatalf("weekStart %#v", data["weekStart"])
	}
	branch, _ := data["branch"].(string)
	if !strings.Contains(branch, "TEST") {
		t.Fatalf("branch want TEST tag got %#v", branch)
	}
	if smtpPort == 0 {
		if data["status"] != "failed" {
			t.Fatalf("without MailHog want failed got %#v", data)
		}
		return
	}
	if data["status"] != "sent" && data["status"] != "partial" {
		t.Fatalf("status %#v", data)
	}
	sent, _ := data["sent"].(float64)
	if sent < 1 {
		t.Fatalf("expected send to staff %s, got %#v", staffID, data)
	}
}

func TestProductDigestIngestStoresBranchMeta(t *testing.T) {
	t.Setenv("PRODUCT_DIGEST_SECRET", "test-product-digest-secret")
	api := newTestAPI(t)
	st := store.New(api.pool)

	digestDate := time.Date(2099, 2, 1, 0, 0, 0, 0, time.UTC)
	dateStr := digestDate.Format("2006-01-02")
	t.Cleanup(func() {
		_, _ = api.pool.Exec(context.Background(),
			`DELETE FROM ops.product_digests WHERE digest_date = $1::date`, digestDate)
	})

	code, env := doJSONWithHeaders(t, api.handler, http.MethodPost, "/api/v1/internal/product-digest/ingest", map[string]any{
		"date":        dateStr,
		"branch":      "staging",
		"environment": "staging",
		"commits":     []any{},
	}, map[string]string{
		"X-Product-Digest-Secret": "test-product-digest-secret",
	})
	if code != http.StatusOK {
		t.Fatalf("ingest %d %#v", code, env)
	}
	data := dataMap(t, env)
	if data["status"] != "empty" || data["branch"] != "staging" {
		t.Fatalf("ingest data %#v", data)
	}

	d, err := st.GetProductDigest(context.Background(), digestDate)
	if err != nil || d == nil {
		t.Fatalf("get digest: %v %#v", err, d)
	}
	var meta map[string]any
	if err := json.Unmarshal(d.Meta, &meta); err != nil {
		t.Fatal(err)
	}
	if meta["branch"] != "staging" || meta["environment"] != "staging" {
		t.Fatalf("meta %#v", meta)
	}
}

func dialTCP(t *testing.T, addr string) bool {
	t.Helper()
	c, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}

// resolveMailhog returns host, SMTP port and HTTP API base for local MailHog.
// Host ports vary (compose may map 1025→1027, 8025→8027).
func resolveMailhog(t *testing.T) (host string, smtpPort int, apiBase string) {
	t.Helper()
	host = "127.0.0.1"
	candidates := []struct {
		smtp int
		api  int
	}{
		{1025, 8025},
		{1027, 8027},
		{1025, 8026},
		{1027, 8026},
	}
	for _, c := range candidates {
		smtpAddr := net.JoinHostPort(host, strconv.Itoa(c.smtp))
		apiAddr := net.JoinHostPort(host, strconv.Itoa(c.api))
		if dialTCP(t, smtpAddr) && dialTCP(t, apiAddr) {
			return host, c.smtp, "http://" + apiAddr
		}
	}
	return host, 0, ""
}
