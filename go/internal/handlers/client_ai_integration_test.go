package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/handlers"
	"github.com/olegrand1976/petsFollow/go/internal/platform/gemini"
)

func TestClientAIDisabled(t *testing.T) {
	t.Setenv("CLIENT_AI_ENABLED", "false")
	api := newTestAPI(t)
	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/client-ai/triage/sessions", clientTok, map[string]any{})
	if code != http.StatusNotFound {
		t.Fatalf("want 404 got %d %#v", code, env)
	}
}

func TestClientConsultationExplainCacheAndAuth(t *testing.T) {
	t.Setenv("CLIENT_AI_ENABLED", "true")
	api := newTestAPI(t)
	t.Cleanup(handlers.TestClearClientExplain)

	calls := 0
	handlers.TestSetClientExplain(func(ctx context.Context, in gemini.ClientExplainInput) (*gemini.ClientExplainResult, error) {
		calls++
		return &gemini.ClientExplainResult{
			Disclaimer: "Pas un avis médical.",
			Cards: []gemini.ClientExplainCard{
				{Title: "Examen", Body: "Tout va bien selon le vétérinaire.", Kind: "general"},
			},
		}, nil
	})

	clientTok, visitID, _ := seedFinalConsultation(t, api)

	code, env := doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/visits/"+visitID+"/client-consultation/explain", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("explain %d %#v", code, env)
	}
	data := dataMap(t, env)
	if data["cached"] != false {
		t.Fatalf("first call should not be cached %#v", data)
	}
	cards, _ := data["cards"].([]any)
	if len(cards) != 1 {
		t.Fatalf("cards %#v", data["cards"])
	}
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/visits/"+visitID+"/client-consultation/explain", clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("explain cache %d %#v", code, env)
	}
	data = dataMap(t, env)
	if data["cached"] != true {
		t.Fatalf("second call should be cached %#v", data)
	}
	if calls != 1 {
		t.Fatalf("cache should skip gemini calls=%d", calls)
	}

	otherTok := loginToken(t, api.handler, "client.marie@petsfollow.test", "ClientDemo123!")
	code, _ = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/visits/"+visitID+"/client-consultation/explain", otherTok, nil)
	if code != http.StatusForbidden {
		t.Fatalf("other owner want 403 got %d", code)
	}
}

func TestClientAITriageGeminiFailureLeavesNoOrphan(t *testing.T) {
	t.Setenv("CLIENT_AI_ENABLED", "true")
	api := newTestAPI(t)
	t.Cleanup(handlers.TestClearClientTriage)

	handlers.TestSetClientTriage(func(ctx context.Context, in gemini.ClientTriageInput) (*gemini.ClientTriageTurn, error) {
		return nil, errors.New("gemini_http_500: boom")
	})

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)
	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/client-ai/triage/sessions", clientTok, map[string]any{
		"petId": petID,
	})
	if code != http.StatusCreated {
		t.Fatalf("create session %d %#v", code, env)
	}
	sessionID, _ := dataMap(t, env)["session"].(map[string]any)["id"].(string)

	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/client-ai/triage/sessions/"+sessionID+"/messages", clientTok, map[string]any{
			"body": "Mon chat boite",
		})
	if code != http.StatusBadGateway {
		t.Fatalf("want 502 got %d %#v", code, env)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/client-ai/triage/sessions/"+sessionID, clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get session %d %#v", code, env)
	}
	msgs, _ := dataMap(t, env)["messages"].([]any)
	if len(msgs) != 0 {
		t.Fatalf("expected no orphan messages, got %d %#v", len(msgs), msgs)
	}
}

func TestClientAITriageSessionFlow(t *testing.T) {
	t.Setenv("CLIENT_AI_ENABLED", "true")
	api := newTestAPI(t)
	t.Cleanup(handlers.TestClearClientTriage)

	handlers.TestSetClientTriage(func(ctx context.Context, in gemini.ClientTriageInput) (*gemini.ClientTriageTurn, error) {
		level := gemini.TriageGreen
		lower := strings.ToLower(in.UserText)
		if strings.Contains(lower, "chocolat") || strings.Contains(lower, "chocolate") {
			level = gemini.TriageRed
		}
		return &gemini.ClientTriageTurn{
			Reply:             "Réponse triage test.",
			Level:             level,
			WatchSigns:        []string{"léthargie"},
			RecommendedAction: "Surveiller",
		}, nil
	})

	clientTok := loginToken(t, api.handler, "client.demo@petsfollow.test", "ClientDemo123!")
	petID := activeDemoPetID(t, api.handler, clientTok)

	code, env := doAuthJSON(t, api.handler, http.MethodPost, "/api/v1/client-ai/triage/sessions", clientTok, map[string]any{
		"petId": petID,
	})
	if code != http.StatusCreated {
		t.Fatalf("create session %d %#v", code, env)
	}
	data := dataMap(t, env)
	sess, _ := data["session"].(map[string]any)
	sessionID, _ := sess["id"].(string)
	if sessionID == "" {
		t.Fatalf("no session id %#v", data)
	}
	esc, _ := data["escalation"].(map[string]any)
	if esc["canBookVisit"] != true || esc["canMessage"] != true {
		t.Fatalf("escalation %#v", esc)
	}

	code, env = doAuthJSON(t, api.handler, http.MethodPost,
		"/api/v1/client-ai/triage/sessions/"+sessionID+"/messages", clientTok, map[string]any{
			"body": "Mon chien a mangé du chocolat",
		})
	if code != http.StatusOK {
		t.Fatalf("message %d %#v", code, env)
	}
	data = dataMap(t, env)
	if data["level"] != "red" {
		t.Fatalf("want red got %#v", data["level"])
	}
	if esc2, _ := data["escalation"].(map[string]any); esc2["petId"] != petID {
		t.Fatalf("escalation pet %#v", data["escalation"])
	}

	code, env = doAuthJSON(t, api.handler, http.MethodGet,
		"/api/v1/client-ai/triage/sessions/"+sessionID, clientTok, nil)
	if code != http.StatusOK {
		t.Fatalf("get session %d %#v", code, env)
	}
	msgs, _ := dataMap(t, env)["messages"].([]any)
	if len(msgs) != 2 {
		t.Fatalf("want 2 messages got %d %#v", len(msgs), msgs)
	}
}
