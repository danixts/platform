package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	gonats "github.com/nats-io/nats.go"
)

const (
	defaultClientName    = "platform-service"
	defaultReconnectWait = 2 * time.Second
)

type Config struct {
	URL           string
	Name          string
	QueueGroup    string
	MaxReconnects int
	ReconnectWait time.Duration
}

type Handler func(subject string, data []byte)

type MsgHandler func(msg *gonats.Msg)

type Client struct {
	underlying *gonats.Conn
	queueGroup string
}

func New(cfg Config) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("platform/queue/nats: empty URL")
	}
	if cfg.Name == "" {
		cfg.Name = defaultClientName
	}
	if cfg.MaxReconnects == 0 {
		cfg.MaxReconnects = -1
	}
	if cfg.ReconnectWait == 0 {
		cfg.ReconnectWait = defaultReconnectWait
	}

	underlying, err := gonats.Connect(cfg.URL,
		gonats.Name(cfg.Name),
		gonats.RetryOnFailedConnect(true),
		gonats.MaxReconnects(cfg.MaxReconnects),
		gonats.ReconnectWait(cfg.ReconnectWait),
	)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}
	return &Client{underlying: underlying, queueGroup: cfg.QueueGroup}, nil
}

func (c *Client) Raw() *gonats.Conn { return c.underlying }

func (c *Client) IsConnected() bool {
	return c.underlying != nil && c.underlying.IsConnected()
}

func (c *Client) Close() {
	if c.underlying == nil {
		return
	}
	_ = c.underlying.Drain()
}

func (c *Client) Publish(subject string, data []byte) error {
	return c.underlying.Publish(subject, data)
}

func (c *Client) PublishJSON(subject string, v any) error {
	payload, err := sonic.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal nats payload: %w", err)
	}
	return c.underlying.Publish(subject, payload)
}

func (c *Client) Request(ctx context.Context, subject string, data []byte) ([]byte, error) {
	msg, err := c.underlying.RequestWithContext(ctx, subject, data)
	if err != nil {
		return nil, fmt.Errorf("nats request %s: %w", subject, err)
	}
	return msg.Data, nil
}

func (c *Client) RequestJSON(ctx context.Context, subject string, in, out any) error {
	payload, err := sonic.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshal nats request: %w", err)
	}
	raw, err := c.Request(ctx, subject, payload)
	if err != nil {
		return err
	}
	return sonic.Unmarshal(raw, out)
}

func (c *Client) Subscribe(subject string, h Handler) error {
	if c.queueGroup == "" {
		_, err := c.underlying.Subscribe(subject, func(m *gonats.Msg) {
			h(m.Subject, m.Data)
		})
		return err
	}
	return c.QueueSubscribe(subject, c.queueGroup, h)
}

func (c *Client) SubscribeMsg(subject string, h MsgHandler) error {
	if c.queueGroup == "" {
		_, err := c.underlying.Subscribe(subject, gonats.MsgHandler(h))
		return err
	}
	_, err := c.underlying.QueueSubscribe(subject, c.queueGroup, gonats.MsgHandler(h))
	return err
}

func (c *Client) QueueSubscribe(subject, queueGroup string, h Handler) error {
	_, err := c.underlying.QueueSubscribe(subject, queueGroup, func(m *gonats.Msg) {
		h(m.Subject, m.Data)
	})
	return err
}

func SubscribeJSON[T any](c *Client, subject string, h func(msg *gonats.Msg, data T)) error {
	return c.SubscribeMsg(subject, func(msg *gonats.Msg) {
		var v T
		if err := sonic.Unmarshal(msg.Data, &v); err != nil {
			return
		}
		h(msg, v)
	})
}

func Call[In, Out any](ctx context.Context, c *Client, subject string, in In) (Out, error) {
	var out Out
	if err := c.RequestJSON(ctx, subject, in, &out); err != nil {
		return out, err
	}
	return out, nil
}

func Reply(msg *gonats.Msg, v any) error {
	data, err := sonic.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal nats reply: %w", err)
	}
	return msg.Respond(data)
}
