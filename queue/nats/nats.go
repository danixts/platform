// Package nats wraps github.com/nats-io/nats.go with the publish/subscribe
// surface used across XMart Cloud services. It supports plain byte
// publishes and sonic-marshalled JSON publishes, and queue subscriptions
// for horizontally-scaled consumers.
package nats

import (
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	gonats "github.com/nats-io/nats.go"
)

// Config describes how to connect to a NATS server.
type Config struct {
	// URL is the server URL (nats://host:4222). Required.
	URL string
	// Name is the client identifier shown in server logs. Default: "xmart-service".
	Name string
	// QueueGroup is the default queue group for Subscribe. Can be overridden
	// per subscription via QueueSubscribe.
	QueueGroup string
	// MaxReconnects sets the reconnect budget. -1 means infinite (default).
	MaxReconnects int
	// ReconnectWait between reconnect attempts. Default: 2s.
	ReconnectWait time.Duration
}

// Handler is the function signature invoked for each received message.
type Handler func(subject string, data []byte)

// Client is the wrapper around nats.Conn.
type Client struct {
	conn       *gonats.Conn
	queueGroup string
}

// New opens a connection to the NATS server.
func New(cfg Config) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("xmart-platform/queue/nats: empty URL")
	}
	if cfg.Name == "" {
		cfg.Name = "xmart-service"
	}
	if cfg.MaxReconnects == 0 {
		cfg.MaxReconnects = -1
	}
	if cfg.ReconnectWait == 0 {
		cfg.ReconnectWait = 2 * time.Second
	}

	conn, err := gonats.Connect(cfg.URL,
		gonats.Name(cfg.Name),
		gonats.RetryOnFailedConnect(true),
		gonats.MaxReconnects(cfg.MaxReconnects),
		gonats.ReconnectWait(cfg.ReconnectWait),
	)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}
	return &Client{conn: conn, queueGroup: cfg.QueueGroup}, nil
}

// Raw returns the underlying *nats.Conn for advanced usage (JetStream,
// request/reply, flush, etc).
func (c *Client) Raw() *gonats.Conn { return c.conn }

// IsConnected reports whether the connection is currently healthy. Used
// by readiness probes.
func (c *Client) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

// Close drains outstanding messages and closes the connection.
func (c *Client) Close() {
	if c.conn == nil {
		return
	}
	_ = c.conn.Drain()
}

// Publish sends a raw byte payload to the given subject.
func (c *Client) Publish(subject string, data []byte) error {
	return c.conn.Publish(subject, data)
}

// PublishJSON marshals v with sonic and publishes it. Use this for
// service-to-service event payloads.
func (c *Client) PublishJSON(subject string, v any) error {
	payload, err := sonic.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal nats payload: %w", err)
	}
	return c.conn.Publish(subject, payload)
}

// Subscribe registers a queue-group subscription using the default queue
// group configured at construction time. For multiple consumers to share
// load, they must all use the same queue group.
func (c *Client) Subscribe(subject string, h Handler) error {
	if c.queueGroup == "" {
		_, err := c.conn.Subscribe(subject, func(m *gonats.Msg) {
			h(m.Subject, m.Data)
		})
		return err
	}
	return c.QueueSubscribe(subject, c.queueGroup, h)
}

// QueueSubscribe registers a queue-group subscription with an explicit
// group name, overriding the client default.
func (c *Client) QueueSubscribe(subject, queueGroup string, h Handler) error {
	_, err := c.conn.QueueSubscribe(subject, queueGroup, func(m *gonats.Msg) {
		h(m.Subject, m.Data)
	})
	return err
}
