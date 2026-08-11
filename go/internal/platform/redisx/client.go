package redisx

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps go-redis with a key prefix (petsfollow:…).
type Client struct {
	rdb    *redis.Client
	prefix string
}

// New connects and pings Redis. Returns nil client (not an error) when addr is empty
// so callers can degrade gracefully.
func New(addr, prefix string) (*Client, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil, nil
	}
	if prefix == "" {
		prefix = "petsfollow:"
	}
	if !strings.HasSuffix(prefix, ":") {
		prefix += ":"
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}
	return &Client{rdb: rdb, prefix: prefix}, nil
}

func (c *Client) key(k string) string {
	return c.prefix + k
}

func (c *Client) Close() error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

func (c *Client) Get(ctx context.Context, k string) (string, error) {
	if c == nil || c.rdb == nil {
		return "", redis.Nil
	}
	return c.rdb.Get(ctx, c.key(k)).Result()
}

func (c *Client) Set(ctx context.Context, k, v string, ttl time.Duration) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Set(ctx, c.key(k), v, ttl).Err()
}

// SetNX sets key only if it does not exist. Returns true when the key was set.
// When Redis is unavailable, returns true (caller should fall back to local guard).
func (c *Client) SetNX(ctx context.Context, k, v string, ttl time.Duration) (bool, error) {
	if c == nil || c.rdb == nil {
		return true, redis.Nil
	}
	return c.rdb.SetNX(ctx, c.key(k), v, ttl).Result()
}

func (c *Client) LPush(ctx context.Context, k string, values ...any) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.LPush(ctx, c.key(k), values...).Err()
}

func (c *Client) LTrim(ctx context.Context, k string, start, stop int64) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.LTrim(ctx, c.key(k), start, stop).Err()
}

func (c *Client) LRange(ctx context.Context, k string, start, stop int64) ([]string, error) {
	if c == nil || c.rdb == nil {
		return nil, nil
	}
	return c.rdb.LRange(ctx, c.key(k), start, stop).Result()
}

// Publish sends a message on a Redis Pub/Sub channel (prefixed).
func (c *Client) Publish(ctx context.Context, channel, payload string) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Publish(ctx, c.key(channel), payload).Err()
}

// Subscribe listens on a Redis channel until ctx is done. Caller must drain ch.
// Returns nil channel when Redis is unavailable.
func (c *Client) Subscribe(ctx context.Context, channel string) <-chan string {
	if c == nil || c.rdb == nil {
		return nil
	}
	pubsub := c.rdb.Subscribe(ctx, c.key(channel))
	out := make(chan string, 32)
	go func() {
		defer close(out)
		defer func() { _ = pubsub.Close() }()
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok || msg == nil {
					return
				}
				select {
				case out <- msg.Payload:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
