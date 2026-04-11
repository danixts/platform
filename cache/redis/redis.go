// Package redis wraps github.com/redis/go-redis/v9 with the method
// surface used across XMart Cloud services. It covers string kv, sets,
// TTL management, existence checks and ping. For anything else, callers
// can use Raw() to get the underlying *redis.Client.
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Config describes how to connect to Redis. URL is required and must be a
// standard redis:// or rediss:// URL.
type Config struct {
	URL string
	// PingTimeout bounds the initial connectivity check. Default: 3s.
	PingTimeout time.Duration
}

// Client is the wrapper around go-redis. It is safe for concurrent use.
type Client struct {
	rdb *goredis.Client
}

// New parses the URL, opens the client and pings the server. It returns
// an error if the URL is invalid or the server is unreachable.
func New(cfg Config) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("xmart-platform/cache/redis: empty URL")
	}
	if cfg.PingTimeout == 0 {
		cfg.PingTimeout = 3 * time.Second
	}
	opts, err := goredis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	rdb := goredis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingTimeout)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

// Raw returns the underlying go-redis client for operations the wrapper
// does not cover (pipelines, streams, pub/sub, scripting).
func (c *Client) Raw() *goredis.Client { return c.rdb }

// Close releases the connection pool. Safe to call on a nil client.
func (c *Client) Close() error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

// Ping is the readiness-probe helper. It returns nil iff the server
// responds to a PING within the given context deadline.
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// ----- strings -----

// Get returns the value for key. The found flag distinguishes a missing
// key from an empty string — callers should check it before using the value.
func (c *Client) Get(ctx context.Context, key string) (value string, found bool, err error) {
	v, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

// Set writes a string value with an optional TTL. Pass ttl=0 for no expiry.
func (c *Client) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Delete removes one or more keys. Missing keys are ignored.
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}

// Exists reports whether a key is present.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Expire sets a TTL on an existing key.
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}

// ----- sets -----

// SAdd adds members to a set.
func (c *Client) SAdd(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	args := make([]any, len(members))
	for i, m := range members {
		args[i] = m
	}
	return c.rdb.SAdd(ctx, key, args...).Err()
}

// SRem removes members from a set.
func (c *Client) SRem(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	args := make([]any, len(members))
	for i, m := range members {
		args[i] = m
	}
	return c.rdb.SRem(ctx, key, args...).Err()
}

// SMembers returns all members of a set.
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.rdb.SMembers(ctx, key).Result()
}

// SIsMember reports whether member belongs to the set.
func (c *Client) SIsMember(ctx context.Context, key, member string) (bool, error) {
	return c.rdb.SIsMember(ctx, key, member).Result()
}
