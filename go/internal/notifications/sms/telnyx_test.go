package sms

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTelnyxSendSuccess(t *testing.T) {
	var gotAuth string
	var gotBody telnyxMessageRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":"msg_123"}}`))
	}))
	defer srv.Close()

	c := &TelnyxClient{APIKey: "key-1", MessagingProfileID: "profile-1", BaseURL: srv.URL}
	res, err := c.Send(context.Background(), "+32470123456", "hello")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if res.ProviderMessageID != "msg_123" || res.DryRun {
		t.Fatalf("unexpected result: %+v", res)
	}
	if gotAuth != "Bearer key-1" {
		t.Fatalf("auth header = %q", gotAuth)
	}
	if gotBody.To != "+32470123456" || gotBody.Text != "hello" || gotBody.MessagingProfileID != "profile-1" {
		t.Fatalf("unexpected request body: %+v", gotBody)
	}
}

func TestTelnyxSendHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"errors":[{"code":"40300"}]}`, http.StatusUnprocessableEntity)
	}))
	defer srv.Close()

	c := &TelnyxClient{APIKey: "key-1", BaseURL: srv.URL}
	_, err := c.Send(context.Background(), "+32470123456", "hello")
	if err == nil || err.Error() != "telnyx_http_422" {
		t.Fatalf("expected telnyx_http_422, got %v", err)
	}
}

func TestTelnyxSendDryRunNoNetwork(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer srv.Close()

	c := &TelnyxClient{APIKey: "key-1", BaseURL: srv.URL, DryRun: true}
	res, err := c.Send(context.Background(), "+32470123456", "hello")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !res.DryRun || res.ProviderMessageID != "dry_run" {
		t.Fatalf("unexpected result: %+v", res)
	}
	if calls != 0 {
		t.Fatalf("dry-run must not hit the network (calls=%d)", calls)
	}
}

func TestTelnyxSendRequiresKeyInLiveMode(t *testing.T) {
	c := &TelnyxClient{}
	if _, err := c.Send(context.Background(), "+32470123456", "hello"); err == nil {
		t.Fatal("expected error without API key")
	}
}
