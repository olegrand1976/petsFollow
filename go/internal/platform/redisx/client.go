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
