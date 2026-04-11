package nats

import (
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	gonats "github.com/nats-io/nats.go"
)

type Config struct {
	URL           string
	Name          string
	QueueGroup    string
	MaxReconnects int
	ReconnectWait time.Duration
}

type Handler func(subject string, data []byte)

type Client struct {
	conn       *gonats.Conn
	queueGroup string
}

func New(cfg Config) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("platform/queue/nats: empty URL")
	}
	if cfg.Name == "" {
		cfg.Name = "platform-service"
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

func (c *Client) Raw() *gonats.Conn { return c.conn }

func (c *Client) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

func (c *Client) Close() {
	if c.conn == nil {
		return
	}
	_ = c.conn.Drain()
}

func (c *Client) Publish(subject string, data []byte) error {
	return c.conn.Publish(subject, data)
}

func (c *Client) PublishJSON(subject string, v any) error {
	payload, err := sonic.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal nats payload: %w", err)
	}
	return c.conn.Publish(subject, payload)
}

func (c *Client) Subscribe(subject string, h Handler) error {
	if c.queueGroup == "" {
		_, err := c.conn.Subscribe(subject, func(m *gonats.Msg) {
			h(m.Subject, m.Data)
		})
		return err
	}
	return c.QueueSubscribe(subject, c.queueGroup, h)
}

func (c *Client) QueueSubscribe(subject, queueGroup string, h Handler) error {
	_, err := c.conn.QueueSubscribe(subject, queueGroup, func(m *gonats.Msg) {
		h(m.Subject, m.Data)
	})
	return err
}
