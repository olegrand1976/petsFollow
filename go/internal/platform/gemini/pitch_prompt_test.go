package gemini

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSpinPedagogyBlock(t *testing.T) {
	s := SpinPedagogyBlock()
	for _, want := range []string{"SPIN", "Situation", "Problem", "Implication", "Need-payoff"} {
		if !strings.Contains(s, want) {
			t.Fatalf("SpinPedagogyBlock missing %q", want)
		}
	}
}

func TestBuildVetStreamPromptIncludesSpin(t *testing.T) {
	raw := json.RawMessage(`{
		"basePersona":"persona",
		"productFacts":"facts",
		"difficulty":{"neutre":"ok"},
		"tools":"tools"
	}`)
	p := BuildVetStreamPrompt(raw, "neutre")
	if !strings.Contains(p, "SPIN") {
		t.Fatal("expected SPIN in stream prompt")
	}
	if !strings.Contains(p, "action") {
		t.Fatal("expected action JSON instructions")
	}
}

func TestBuildVetLivePromptIncludesSpin(t *testing.T) {
	raw := json.RawMessage(`{
		"basePersona":"persona",
		"productFacts":"facts",
		"difficulty":{"neutre":"ok"},
		"tools":"tools"
	}`)
	p := BuildVetLivePrompt(raw, "neutre")
	if !strings.Contains(p, "SPIN") {
		t.Fatal("expected SPIN in live prompt")
	}
}

func TestDisplayableStreamTextHidesTrailingJSON(t *testing.T) {
	in := "Oui, je vous écoute.\n{\"action\":\"continue\"}"
	got := DisplayableStreamText(in)
	if got != "Oui, je vous écoute." {
		t.Fatalf("got %q", got)
	}
	partial := "Bonjour\n{"
	if DisplayableStreamText(partial) != "Bonjour" {
		t.Fatalf("partial got %q", DisplayableStreamText(partial))
	}
	if DisplayableStreamText("{\"action\":\"continue\"}") != "" {
		t.Fatal("expected empty while only JSON")
	}
}

func TestParseStreamVetReply(t *testing.T) {
	res := ParseStreamVetReply("Mardi 10h ça me va.\n{\"action\":\"book_appointment\",\"appointmentSlot\":\"mardi 10h\"}")
	if res.Reply != "Mardi 10h ça me va." {
		t.Fatalf("reply=%q", res.Reply)
	}
	if res.Action != "book_appointment" {
		t.Fatalf("action=%q", res.Action)
	}
	if res.AppointmentSlot != "mardi 10h" {
		t.Fatalf("slot=%q", res.AppointmentSlot)
	}
}

func TestBuildVetStreamPromptHostile(t *testing.T) {
	raw := json.RawMessage(`{
		"basePersona":"persona",
		"productFacts":"facts",
		"difficulty":{"hostile":"méchant"},
		"tools":"tools"
	}`)
	p := BuildVetStreamPrompt(raw, "hostile")
	if !strings.Contains(p, "hostile") && !strings.Contains(p, "impatient") {
		t.Fatal("expected hostile turn rules")
	}
}

func TestUtf8SafeSuffix(t *testing.T) {
	delta, ok := utf8SafeSuffix("Bonjour ", "Bonjour café")
	if !ok || delta != "café" {
		t.Fatalf("got %q ok=%v", delta, ok)
	}
	if _, ok := utf8SafeSuffix("Bonjour café", "Bonjour"); ok {
		t.Fatal("shrink should be rejected")
	}
}

func TestDefaultModels(t *testing.T) {
	c := New("key", "", "")
	if c.Model != "gemini-3.6-flash" {
		t.Fatalf("model=%q", c.Model)
	}
	if c.LiteModel != "gemini-3.5-flash-lite" {
		t.Fatalf("lite=%q", c.LiteModel)
	}
	if c.effectiveLite() != "gemini-3.5-flash-lite" {
		t.Fatalf("effectiveLite=%q", c.effectiveLite())
	}
}
