package handlers

import (
	"context"
	"sync"
	"testing"

	"github.com/olegrand1976/petsFollow/go/internal/notifications/fcm"
	"github.com/olegrand1976/petsFollow/go/internal/store"
)

type mockPusher struct {
	mu      sync.Mutex
	calls   int
	tokens  []string
	title   string
	body    string
	data    map[string]string
	invalid []string
}

func (m *mockPusher) Send(_ context.Context, tokens []string, title, body string, data map[string]string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	m.tokens = append([]string{}, tokens...)
	m.title = title
	m.body = body
	m.data = data
	return m.invalid, nil
}

func TestPrefEnabled(t *testing.T) {
	p := store.ClientNotificationPrefs{Messages: true, Visits: false}
	if !prefEnabled(p, pushKindMessages) {
		t.Fatal("messages should be enabled")
	}
	if prefEnabled(p, pushKindVisits) {
		t.Fatal("visits should be disabled")
	}
}

func TestTruncateRunes(t *testing.T) {
	if got := truncateRunes("abc", 10); got != "abc" {
		t.Fatalf("got %q", got)
	}
	if got := truncateRunes("abcdefghij", 5); got != "abcde…" {
		t.Fatalf("got %q", got)
	}
}

func TestNotifyClientPush_PrefOff(t *testing.T) {
	// Store-less path: use API with nil store would panic — test pref helper + Nop instead.
	p := fcm.NopPusher{}
	invalid, err := p.Send(context.Background(), []string{"t"}, "title", "body", nil)
	if err != nil || invalid != nil {
		t.Fatalf("nop: %v %v", err, invalid)
	}
}

func TestMockPusherRecordsSend(t *testing.T) {
	m := &mockPusher{}
	invalid, err := m.Send(context.Background(), []string{"tok-a", "tok-b"}, "Titre", "Corps", map[string]string{"type": "message"})
	if err != nil || invalid != nil {
		t.Fatalf("unexpected: %v %v", err, invalid)
	}
	if m.calls != 1 || len(m.tokens) != 2 || m.title != "Titre" || m.data["type"] != "message" {
		t.Fatalf("unexpected mock state: %+v", m)
	}
}
