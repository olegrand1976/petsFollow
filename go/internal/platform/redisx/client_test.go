package redisx_test

import (
	"context"
	"testing"
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/platform/redisx"
)

func TestPublishSubscribe(t *testing.T) {
	c, err := redisx.New("127.0.0.1:6379", "petsfollow-test:")
	if err != nil || c == nil {
		t.Skip("redis unavailable")
	}
	t.Cleanup(func() { _ = c.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ch := c.Subscribe(ctx, "improve-hub:unit")
	if ch == nil {
		t.Fatal("subscribe nil")
	}
	time.Sleep(50 * time.Millisecond)
	if err := c.Publish(ctx, "improve-hub:unit", `{"type":"step"}`); err != nil {
		t.Fatal(err)
	}
	select {
	case msg := <-ch:
		if msg == "" {
			t.Fatal("empty payload")
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting pubsub")
	}
}
