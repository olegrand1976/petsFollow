package billit_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/invoicing"
	"github.com/olegrand1976/petsFollow/go/internal/invoicing/billit"
)

func TestBaseURLConstants(t *testing.T) {
	if billit.SandboxBaseURL != "https://api.sandbox.billit.be" {
		t.Fatalf("sandbox=%s", billit.SandboxBaseURL)
	}
	if billit.DefaultBaseURL != "https://api.billit.be" {
		t.Fatalf("default=%s", billit.DefaultBaseURL)
	}
	c := billit.NewClient("")
	if c == nil {
		t.Fatal("nil client")
	}
}

func TestVerifyAndParseWebhook(t *testing.T) {
	body := []byte(`{"OrderID":12345,"EventType":"OrderDelivered","Status":"delivered"}`)
	secret := "test-secret"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	if !billit.VerifyWebhookSignature(secret, body, sig) {
		t.Fatal("sig should verify")
	}
	if billit.VerifyWebhookSignature(secret, body, "deadbeef") {
		t.Fatal("bad sig")
	}
	if billit.VerifyWebhookSignature(secret, body, "ab") {
		t.Fatal("short hex must not panic / must refuse")
	}
	orderID, eventType, status, peppol, err := billit.ParseWebhook(body)
	if err != nil {
		t.Fatal(err)
	}
	if orderID != "12345" || eventType != "OrderDelivered" {
		t.Fatalf("%s %s", orderID, eventType)
	}
	if status != invoicing.StatusDelivered || peppol != "delivered" {
		t.Fatalf("%s %s", status, peppol)
	}
}

func TestWebhookDedupeKey(t *testing.T) {
	a := []byte(`{"OrderID":1,"Status":"sending","EventType":"OrderSent"}`)
	b := []byte(`{"OrderID":1,"Status":"delivered","EventType":"OrderDelivered"}`)
	ka := billit.WebhookDedupeKey(a)
	kb := billit.WebhookDedupeKey(b)
	if ka == kb {
		t.Fatal("distinct status payloads must not share dedupe key")
	}
	if billit.WebhookDedupeKey(a) != ka {
		t.Fatal("exact redelivery must be stable")
	}
	withEvt := []byte(`{"EventID":"evt-9","OrderID":1,"Status":"delivered"}`)
	if got := billit.WebhookDedupeKey(withEvt); got != "evt:evt-9" {
		t.Fatalf("got %s", got)
	}
	// Generic "Id" is often the order id — must NOT collapse status progression.
	withOrderIdOnly := []byte(`{"Id":"99","OrderID":99,"Status":"sending","EventType":"OrderSent"}`)
	withOrderIdDelivered := []byte(`{"Id":"99","OrderID":99,"Status":"delivered","EventType":"OrderDelivered"}`)
	if billit.WebhookDedupeKey(withOrderIdOnly) == billit.WebhookDedupeKey(withOrderIdDelivered) {
		t.Fatal("Id field must not be used as event dedupe key")
	}
	if !strings.HasPrefix(billit.WebhookDedupeKey(withOrderIdOnly), "sha256:") {
		t.Fatal("expected body hash fallback when EventID absent")
	}
}

func TestClientCreateAndSend(t *testing.T) {
	var gotParty, gotKey, gotStrict string
	var created bool
	var sendBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotParty = r.Header.Get("PartyID")
		gotKey = r.Header.Get("ApiKey")
		if r.URL.Path == "/v1/orders/commands/send" {
			gotStrict = r.Header.Get("StrictTransportType")
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/parties/party1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"PartyID":1,"Name":"Vet","VATNumber":"BE1000000021"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/orders":
			created = true
			var ord billit.OrderDTO
			_ = json.NewDecoder(r.Body).Decode(&ord)
			if ord.Customer.CountryCode != "BE" {
				t.Errorf("country=%s", ord.Customer.CountryCode)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("99"))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/orders/commands/send":
			_ = json.NewDecoder(r.Body).Decode(&sendBody)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := billit.NewClient(srv.URL)
	ctx := context.Background()
	st, err := c.CheckParty(ctx, "party1", "key1")
	if err != nil || !st.Complete {
		t.Fatalf("%v %+v", err, st)
	}
	doc := invoicing.Document{
		Type: invoicing.DocInvoice,
		Counterparty: invoicing.Counterparty{
			Name: "Vet", Country: "BE", VATNumber: "BE1000000021",
			Street: "Rue Test 1", City: "Bruxelles", Postal: "1000",
		},
		Lines: []invoicing.Line{{Description: "A", Quantity: 1, UnitPriceExclCents: 1000, VATPercent: 21}},
	}
	id, err := c.CreateDocument(ctx, "party1", "key1", doc)
	if err != nil || id != "99" || !created {
		t.Fatalf("id=%s err=%v created=%v", id, err, created)
	}
	if err := c.Send(ctx, "party1", "key1", id, invoicing.TransportPeppol); err != nil {
		t.Fatal(err)
	}
	if gotParty != "party1" || gotKey != "key1" {
		t.Fatalf("auth %s %s", gotParty, gotKey)
	}
	// Billit lit `Transporttype` : toute autre orthographe est ignorée et
	// l'ordre repart sur le canal par défaut du cabinet.
	if sendBody["Transporttype"] != "Peppol" {
		t.Fatalf("expected Peppol transport %#v", sendBody)
	}
	if _, legacy := sendBody["Transport"]; legacy {
		t.Fatalf("legacy Transport key must not be sent %#v", sendBody)
	}
	if gotStrict != "true" {
		t.Fatalf("Peppol must be strict (no silent email fallback), got %q", gotStrict)
	}

	sendBody, gotStrict = nil, ""
	if err := c.Send(ctx, "party1", "key1", id, invoicing.TransportSDI); err != nil {
		t.Fatal(err)
	}
	if sendBody["Transporttype"] != "SDI" {
		t.Fatalf("IT must route over SDI %#v", sendBody)
	}
	if gotStrict != "true" {
		t.Fatalf("SDI must be strict, got %q", gotStrict)
	}

	// Particulier : pas d'adresse réseau → email, et surtout pas de mode strict.
	sendBody, gotStrict = nil, ""
	if err := c.Send(ctx, "party1", "key1", id, invoicing.TransportSMTP); err != nil {
		t.Fatal(err)
	}
	if sendBody["Transporttype"] != "SMTP" {
		t.Fatalf("individual must route over SMTP %#v", sendBody)
	}
	if gotStrict != "" {
		t.Fatalf("SMTP must not set StrictTransportType, got %q", gotStrict)
	}

	if err := c.Send(ctx, "party1", "key1", id, ""); err == nil {
		t.Fatal("empty transport must fail closed")
	}
}

func TestParsePartyStatus(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		complete bool
	}{
		{name: "empty", body: "", complete: false},
		{name: "unknown_shape", body: `{"Foo":"bar"}`, complete: false},
		{name: "flag_true", body: `{"Complete":true}`, complete: true},
		{name: "flag_false", body: `{"IsComplete":false}`, complete: false},
		{name: "nested", body: `{"Party":{"OnboardingComplete":false}}`, complete: false},
		{name: "status_pending", body: `{"Status":"pending_kyc"}`, complete: false},
		{name: "status_active", body: `{"Status":"active"}`, complete: true},
		{name: "completeness", body: `{"Completeness":80}`, complete: false},
		{name: "party_dto", body: `{"PartyID":1133539,"Name":"LL-IT","VATNumber":"BE1007132489"}`, complete: true},
		{name: "party_incomplete", body: `{"PartyID":1,"Name":"NoVAT"}`, complete: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(srv.Close)
			st, err := billit.NewClient(srv.URL).CheckParty(context.Background(), "p", "k")
			if err != nil {
				t.Fatal(err)
			}
			if st.Complete != tc.complete {
				t.Fatalf("got %+v want complete=%v", st, tc.complete)
			}
		})
	}
}

func TestParseWebhookUnknownStatusStaysSending(t *testing.T) {
	orderID, _, status, peppol, err := billit.ParseWebhook([]byte(`{"OrderID":7,"Status":"weird-code"}`))
	if err != nil {
		t.Fatal(err)
	}
	if orderID != "7" || status != invoicing.StatusSending || peppol != "unknown" {
		t.Fatalf("%s %s %s", orderID, status, peppol)
	}
}

func TestParseWebhookMessageDelivered(t *testing.T) {
	body := []byte(`{
		"UpdatedEntityID": 690998,
		"UpdatedEntityType": "Message",
		"WebhookUpdateTypeTC": "U",
		"EntityDetail": {
			"OrderMessage": {
				"OrderID": 2615999,
				"Success": true,
				"TransportType": "Peppol",
				"MessageDirection": "Outgoing"
			},
			"AdditionalMessageInformation": {
				"EInvoiceFlowState": "Delivered"
			}
		}
	}`)
	orderID, eventType, status, peppol, err := billit.ParseWebhook(body)
	if err != nil {
		t.Fatal(err)
	}
	if orderID != "2615999" {
		t.Fatalf("orderID=%s", orderID)
	}
	if eventType != "Message" && eventType != "Delivered" {
		t.Fatalf("eventType=%s", eventType)
	}
	if status != invoicing.StatusDelivered || peppol != "delivered" {
		t.Fatalf("status=%s peppol=%s", status, peppol)
	}
}

func TestParseWebhookMessageSentStaysSending(t *testing.T) {
	body := []byte(`{
		"UpdatedEntityID": 1,
		"UpdatedEntityType": "Message",
		"WebhookUpdateTypeTC": "U",
		"EntityDetail": {
			"OrderMessage": {"OrderID": 42, "Success": true},
			"AdditionalMessageInformation": {"EInvoiceFlowState": "Sent"}
		}
	}`)
	orderID, _, status, peppol, err := billit.ParseWebhook(body)
	if err != nil {
		t.Fatal(err)
	}
	if orderID != "42" || status != invoicing.StatusSending || peppol != "sending" {
		t.Fatalf("%s %s %s", orderID, status, peppol)
	}
}

func TestParseWebhookMessageAcceptedNotDelivered(t *testing.T) {
	body := []byte(`{
		"UpdatedEntityType": "Message",
		"EntityDetail": {
			"OrderMessage": {"OrderID": 99},
			"AdditionalMessageInformation": {"EInvoiceFlowState": "Accepted"}
		}
	}`)
	_, _, status, peppol, err := billit.ParseWebhook(body)
	if err != nil {
		t.Fatal(err)
	}
	if status == invoicing.StatusDelivered || peppol == "delivered" {
		t.Fatalf("Accepted must not be delivered: %s %s", status, peppol)
	}
}

func TestParseWebhookMessageRefused(t *testing.T) {
	body := []byte(`{
		"UpdatedEntityID": 690988,
		"UpdatedEntityType": "Message",
		"WebhookUpdateTypeTC": "U",
		"EntityDetail": {
			"OrderID": 455654,
			"MessageAdditionalInformation": {
				"EInvoiceFlowState": "Refused",
				"AdditionalFlowStateInformation": "No Valid VAT"
			}
		}
	}`)
	orderID, _, status, peppol, err := billit.ParseWebhook(body)
	if err != nil {
		t.Fatal(err)
	}
	if orderID != "455654" || status != invoicing.StatusRejected || peppol != "rejected" {
		t.Fatalf("%s %s %s", orderID, status, peppol)
	}
}

func TestParseWebhookMessageMissingOrderID(t *testing.T) {
	body := []byte(`{
		"UpdatedEntityType": "Message",
		"EntityDetail": {
			"AdditionalMessageInformation": {"EInvoiceFlowState": "Delivered"}
		}
	}`)
	_, _, _, _, err := billit.ParseWebhook(body)
	if err == nil {
		t.Fatal("expected missing order id")
	}
}

func TestWebhookDedupeKeyMessageProgression(t *testing.T) {
	sent := []byte(`{
		"UpdatedEntityID": 690998,
		"UpdatedEntityType": "Message",
		"WebhookUpdateTypeTC": "U",
		"EntityDetail": {
			"OrderMessage": {"OrderID": 2615999},
			"AdditionalMessageInformation": {"EInvoiceFlowState": "Sent"}
		}
	}`)
	delivered := []byte(`{
		"UpdatedEntityID": 690998,
		"UpdatedEntityType": "Message",
		"WebhookUpdateTypeTC": "U",
		"EntityDetail": {
			"OrderMessage": {"OrderID": 2615999},
			"AdditionalMessageInformation": {"EInvoiceFlowState": "Delivered"}
		}
	}`)
	if billit.WebhookDedupeKey(sent) == billit.WebhookDedupeKey(delivered) {
		t.Fatal("Sent→Delivered must not share dedupe key")
	}
	if !strings.HasPrefix(billit.WebhookDedupeKey(sent), "msg:") {
		t.Fatalf("want msg: prefix got %s", billit.WebhookDedupeKey(sent))
	}
}

func TestParseWebhookAcceptSuccessNotDelivered(t *testing.T) {
	// Broad Contains("accept"|"success") must not count as Peppol delivered.
	for _, body := range []string{
		`{"OrderID":8,"Status":"accepted"}`,
		`{"OrderID":9,"Status":"success"}`,
		`{"OrderID":10,"Status":"OrderAccepted"}`,
	} {
		_, _, status, peppol, err := billit.ParseWebhook([]byte(body))
		if err != nil {
			t.Fatal(err)
		}
		if status == invoicing.StatusDelivered || peppol == "delivered" {
			t.Fatalf("body=%s status=%s peppol=%s", body, status, peppol)
		}
	}
	_, _, status, peppol, err := billit.ParseWebhook([]byte(`{"OrderID":11,"Status":"OrderDelivered"}`))
	if err != nil {
		t.Fatal(err)
	}
	if status != invoicing.StatusDelivered || peppol != "delivered" {
		t.Fatalf("OrderDelivered → %s %s", status, peppol)
	}
}

// Relecture d'ordre (réconciliation) : le libellé de statut n'est pas au même
// endroit selon le canal, et un état non terminal ne doit rien conclure.
func TestClientFetchStatus(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		status   invoicing.DocStatus
		peppol   string
		terminal bool
	}{
		{"flow state delivered", `{"OrderID":42,"EInvoiceFlowState":"Delivered"}`, invoicing.StatusDelivered, "delivered", true},
		{"refused", `{"OrderID":42,"EInvoiceFlowState":"Refused"}`, invoicing.StatusRejected, "rejected", true},
		{"nested message info", `{"OrderID":42,"AdditionalMessageInformation":{"EInvoiceFlowState":"Delivered"}}`, invoicing.StatusDelivered, "delivered", true},
		{"accepted is not delivered", `{"OrderID":42,"EInvoiceFlowState":"Accepted"}`, invoicing.StatusSending, "sending", false},
		{"no status field", `{"OrderID":42}`, invoicing.StatusSending, "unknown", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(srv.Close)
			st, err := billit.NewClient(srv.URL).FetchStatus(context.Background(), "p", "k", "42")
			if err != nil {
				t.Fatal(err)
			}
			if gotPath != "/v1/orders/42" {
				t.Fatalf("path %s", gotPath)
			}
			if st.Status != tc.status || st.PeppolStatus != tc.peppol || st.Terminal != tc.terminal {
				t.Fatalf("got %+v want %s/%s terminal=%v", st, tc.status, tc.peppol, tc.terminal)
			}
		})
	}
}

func TestCreateDocumentNoRetryOn500(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	t.Cleanup(srv.Close)
	c := billit.NewClient(srv.URL)
	_, err := c.CreateDocument(context.Background(), "p", "k", invoicing.Document{
		Type: invoicing.DocInvoice,
		Counterparty: invoicing.Counterparty{
			Name: "Vet", Country: "BE", VATNumber: "BE1000000021",
			Street: "Rue Test 1", City: "Bruxelles", Postal: "1000",
		},
		Lines: []invoicing.Line{{Description: "A", Quantity: 1, UnitPriceExclCents: 1000, VATPercent: 21}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if hits != 1 {
		t.Fatalf("create must not retry, hits=%d", hits)
	}
}
