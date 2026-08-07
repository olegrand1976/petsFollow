package billing_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/billing"
)

func TestMockCustomerID(t *testing.T) {
	got := billing.MockCustomerID("user-abc")
	if got != "cus_mock_user-abc" {
		t.Fatalf("got %q", got)
	}
}

func TestMockGatewayCreatePortalSession(t *testing.T) {
	g := billing.NewMockGateway("whsec_test", "http://localhost:8291/", "test-url-secret")
	sess, err := g.CreatePortalSession(context.Background(), "cus_mock_x", "petsfollow://home")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sess.URL, "/api/v1/billing/dev/mock-portal?") {
		t.Fatalf("url=%q", sess.URL)
	}
	if !strings.Contains(sess.URL, "customer=cus_mock_x") {
		t.Fatalf("missing customer: %q", sess.URL)
	}
	if !strings.Contains(sess.URL, "return=petsfollow") {
		t.Fatalf("missing return: %q", sess.URL)
	}
}

func TestMockGatewayCreateCheckoutSessionIncludesSuccessURL(t *testing.T) {
	g := billing.NewMockGateway("whsec_test", "http://localhost:8291", "test-url-secret")
	sess, err := g.CreateCheckoutSession(context.Background(), billing.CheckoutRequest{
		PriceID:    "price_mock",
		Mode:       "subscription",
		SuccessURL: "petsfollow://payment/success",
		Metadata: map[string]string{
			"pet_id":        "pet-1",
			"owner_user_id": "user-1",
			"plan_code":     "triennial",
			"billing_mode":  "subscription",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sess.URL, "/api/v1/billing/dev/mock-complete?") {
		t.Fatalf("url=%q", sess.URL)
	}
	if !strings.Contains(sess.URL, "pet_id=pet-1") {
		t.Fatalf("missing pet_id: %q", sess.URL)
	}
	if !strings.Contains(sess.URL, "success_url=petsfollow") {
		t.Fatalf("missing success_url: %q", sess.URL)
	}
	if !strings.HasPrefix(sess.ID, "cs_mock_") {
		t.Fatalf("id=%q", sess.ID)
	}
}

// La signature est le seul contrôle sur les callbacks mock (atteints par
// redirection navigateur, sans bearer). Elle doit couvrir *tous* les
// paramètres : sinon on rejoue une URL légitime en changeant le plan ou le
// propriétaire, et on s'octroie un entitlement.
func TestVerifyMockURLRejectsTamperedParams(t *testing.T) {
	const secret = "test-url-secret"
	signed := func() url.Values {
		q := url.Values{}
		q.Set("pet_id", "pet-1")
		q.Set("owner_user_id", "owner-1")
		q.Set("plan_code", "monthly")
		q.Set("billing_mode", "subscription")
		q.Set(billing.MockSignatureParam, billing.SignMockURL(secret, q))
		return q
	}

	if !billing.VerifyMockURL(secret, signed()) {
		t.Fatal("a freshly signed URL must verify")
	}

	tampered := map[string]func(url.Values){
		"plan surclassé":      func(q url.Values) { q.Set("plan_code", "triennial") },
		"propriétaire changé": func(q url.Values) { q.Set("owner_user_id", "someone-else") },
		"animal changé":       func(q url.Values) { q.Set("pet_id", "pet-2") },
		"paramètre ajouté":    func(q url.Values) { q.Set("addon_id", "care-plus") },
		"paramètre retiré":    func(q url.Values) { q.Del("billing_mode") },
		"signature effacée":   func(q url.Values) { q.Del(billing.MockSignatureParam) },
		"signature bricolée":  func(q url.Values) { q.Set(billing.MockSignatureParam, "deadbeef") },
		"valeur multiple":     func(q url.Values) { q.Add("plan_code", "triennial") },
	}
	for name, mutate := range tampered {
		q := signed()
		mutate(q)
		if billing.VerifyMockURL(secret, q) {
			t.Fatalf("%s : l'URL modifiée ne doit pas vérifier", name)
		}
	}

	q := signed()
	if billing.VerifyMockURL("another-secret", q) {
		t.Fatal("une autre clé ne doit pas vérifier la signature")
	}
}

// Garde-fou de canonicalisation : si le payload signé était une simple
// concaténation, déplacer du contenu d'un paramètre à l'autre produirait la
// même signature et rendrait la falsification triviale.
func TestSignMockURLIsUnambiguousAcrossParamBoundaries(t *testing.T) {
	const secret = "test-url-secret"

	twoParams := url.Values{}
	twoParams.Set("a", "b")
	twoParams.Set("c", "d")

	oneParam := url.Values{}
	oneParam.Set("a", "b&c=d")

	if billing.SignMockURL(secret, twoParams) == billing.SignMockURL(secret, oneParam) {
		t.Fatal("deux jeux de paramètres distincts ne doivent pas partager une signature")
	}
}
