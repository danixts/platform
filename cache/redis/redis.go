package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const defaultPingTimeout = 3 * time.Second

type Config struct {
	URL         string
	PingTimeout time.Duration
}

type Client struct {
	underlying *goredis.Client
}

func New(cfg Config) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("platform/cache/redis: empty URL")
	}
	if cfg.PingTimeout == 0 {
		cfg.PingTimeout = defaultPingTimeout
	}
	opts, err := goredis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	underlying := goredis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(context.Background(), cfg.PingTimeout)
	defer cancel()
	if err := underlying.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Client{underlying: underlying}, nil
}

func (c *Client) Raw() *goredis.Client { return c.underlying }

func (c *Client) Close() error {
	if c == nil || c.underlying == nil {
		return nil
	}
	return c.underlying.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	return c.underlying.Ping(ctx).Err()
}

func (c *Client) Get(ctx context.Context, key string) (value string, found bool, err error) {
	stored, err := c.underlying.Get(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return stored, true, nil
}

func (c *Client) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.underlying.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.underlying.Del(ctx, keys...).Err()
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.underlying.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.underlying.Expire(ctx, key, ttl).Err()
}

func (c *Client) SAdd(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	return c.underlying.SAdd(ctx, key, stringsAsAny(members)...).Err()
}

func (c *Client) SRem(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	return c.underlying.SRem(ctx, key, stringsAsAny(members)...).Err()
}

func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.underlying.SMembers(ctx, key).Result()
}

func (c *Client) SIsMember(ctx context.Context, key, member string) (bool, error) {
	return c.underlying.SIsMember(ctx, key, member).Result()
}

func stringsAsAny(members []string) []any {
	out := make([]any, len(members))
	for i := range members {
		out[i] = members[i]
	}
	return out
}
