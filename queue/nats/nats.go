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

	// defaultRequestTimeout bounds Request/RequestJSON/Call when the caller's
	// ctx carries no deadline (e.g. context.Background()), which would
	// otherwise hold the reply subscription open forever.
	defaultRequestTimeout = 10 * time.Second

	// defaultDispatchQueueSize bounds the per-worker backlog for
	// SubscribeOptions.Workers > 1 before Subscribe/QueueSubscribe blocks
	// nats.go's own dispatcher goroutine.
	defaultDispatchQueueSize = 4
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
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultRequestTimeout)
		defer cancel()
	}
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

// SubscribeOptions tunes dispatch for Subscribe/QueueSubscribeWithOptions.
// The zero value keeps nats.go's default: one dispatcher goroutine per
// subscription (handlers for a subject run strictly in order) and its
// default pending buffer.
type SubscribeOptions struct {
	// Workers, set > 1, fans a subscription's messages out to a fixed pool
	// of goroutines instead of nats.go's single dispatcher goroutine, which
	// otherwise serializes every handler call for that subject. Ordering
	// across messages is not preserved when Workers > 1.
	Workers int
	// PendingMsgLimit and PendingBytesLimit raise nats.go's default pending
	// buffer, past which the client silently drops messages for a consumer
	// that falls behind. 0 keeps the nats.go default for that limit.
	PendingMsgLimit   int
	PendingBytesLimit int
}

// SubscribeWithOptions is Subscribe with SubscribeOptions control over
// dispatch concurrency and pending buffer limits.
func (c *Client) SubscribeWithOptions(subject string, h Handler, opts SubscribeOptions) (*gonats.Subscription, error) {
	if c.queueGroup == "" {
		return c.subscribe(subject, "", h, opts)
	}
	return c.QueueSubscribeWithOptions(subject, c.queueGroup, h, opts)
}

// QueueSubscribeWithOptions is QueueSubscribe with SubscribeOptions control
// over dispatch concurrency and pending buffer limits.
func (c *Client) QueueSubscribeWithOptions(subject, queueGroup string, h Handler, opts SubscribeOptions) (*gonats.Subscription, error) {
	return c.subscribe(subject, queueGroup, h, opts)
}

func (c *Client) subscribe(subject, queueGroup string, h Handler, opts SubscribeOptions) (*gonats.Subscription, error) {
	dispatch := h
	if opts.Workers > 1 {
		dispatch = workerPoolDispatch(opts.Workers, h)
	}
	cb := func(m *gonats.Msg) { dispatch(m.Subject, m.Data) }

	var sub *gonats.Subscription
	var err error
	if queueGroup == "" {
		sub, err = c.underlying.Subscribe(subject, cb)
	} else {
		sub, err = c.underlying.QueueSubscribe(subject, queueGroup, cb)
	}
	if err != nil {
		return nil, err
	}

	if opts.PendingMsgLimit > 0 || opts.PendingBytesLimit > 0 {
		msgLimit, bytesLimit := opts.PendingMsgLimit, opts.PendingBytesLimit
		if msgLimit <= 0 {
			msgLimit = gonats.DefaultSubPendingMsgsLimit
		}
		if bytesLimit <= 0 {
			bytesLimit = gonats.DefaultSubPendingBytesLimit
		}
		if err := sub.SetPendingLimits(msgLimit, bytesLimit); err != nil {
			return sub, fmt.Errorf("set pending limits: %w", err)
		}
	}
	return sub, nil
}

// workerPoolDispatch runs h on a fixed pool of goroutines fed by one
// channel, so a subscription's messages no longer serialize on nats.go's
// single dispatcher goroutine. The pool lives for the subscription's
// lifetime; there is no drain-on-unsubscribe, matching nats.go's own
// fire-and-forget async dispatch.
func workerPoolDispatch(workers int, h Handler) Handler {
	type job struct {
		subject string
		data    []byte
	}
	jobs := make(chan job, workers*defaultDispatchQueueSize)
	for range workers {
		go func() {
			for j := range jobs {
				h(j.subject, j.data)
			}
		}()
	}
	return func(subject string, data []byte) {
		jobs <- job{subject: subject, data: data}
	}
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
