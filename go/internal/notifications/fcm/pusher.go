package fcm

import (
	"context"
	"log"
)

// Pusher sends FCM notifications to device tokens.
type Pusher interface {
	// Send delivers a notification to the given tokens.
	// Returns tokens that should be removed (unregistered / invalid).
	Send(ctx context.Context, tokens []string, title, body string, data map[string]string) (invalid []string, err error)
}

// NopPusher is a no-op implementation used when FCM is disabled or unavailable.
type NopPusher struct{}

func (NopPusher) Send(context.Context, []string, string, string, map[string]string) ([]string, error) {
	return nil, nil
}

// NewFromADC creates a Firebase Messaging client using Application Default Credentials.
// When enabled is false, or init fails, returns NopPusher (API keeps working).
func NewFromADC(ctx context.Context, enabled bool) Pusher {
	if !enabled {
		log.Println("fcm: disabled (FCM_ENABLED=false)")
		return NopPusher{}
	}
	client, err := newClient(ctx)
	if err != nil {
		log.Printf("fcm: init failed, push disabled: %v", err)
		return NopPusher{}
	}
	log.Println("fcm: push enabled")
	return client
}
