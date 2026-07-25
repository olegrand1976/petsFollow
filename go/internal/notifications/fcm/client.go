package fcm

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

type client struct {
	msg *messaging.Client
}

func newClient(ctx context.Context) (*client, error) {
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("firebase app: %w", err)
	}
	msg, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("messaging: %w", err)
	}
	return &client{msg: msg}, nil
}

func (c *client) Send(ctx context.Context, tokens []string, title, body string, data map[string]string) ([]string, error) {
	if len(tokens) == 0 {
		return nil, nil
	}
	// FCM multicast limit is 500 tokens per request.
	const batchSize = 500
	var invalid []string
	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		batch := tokens[i:end]
		resp, err := c.msg.SendEachForMulticast(ctx, &messaging.MulticastMessage{
			Tokens: batch,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{Sound: "default"},
				},
			},
		})
		if err != nil {
			return invalid, err
		}
		for j, r := range resp.Responses {
			if r.Success {
				continue
			}
			if messaging.IsUnregistered(r.Error) {
				invalid = append(invalid, batch[j])
				continue
			}
			log.Printf("fcm: send token error: %v", r.Error)
		}
	}
	return invalid, nil
}
