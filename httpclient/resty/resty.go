// Package resty returns a preconfigured go-resty/v2 client wired with
// sonic (de)serialization, sane retry defaults and a browser-like
// User-Agent string that avoids being blocked by most WAFs.
//
// Callers receive a *resty.Client and use it directly — there is no
// custom method surface because resty's own API is already ergonomic.
package resty

import (
	"time"

	"github.com/bytedance/sonic"
	goresty "github.com/go-resty/resty/v2"
)

// Config describes the defaults applied to the returned client. All
// fields are optional.
type Config struct {
	// Timeout per request including retries. Default: 15s.
	Timeout time.Duration
	// RetryCount is the number of retries on transient failures. Default: 2.
	RetryCount int
	// RetryWaitTime is the base backoff between retries. Default: 500ms.
	RetryWaitTime time.Duration
	// UserAgent overrides the default browser-like UA. Leave empty for default.
	UserAgent string
	// Debug toggles resty's debug mode (body logging). Default: false.
	Debug bool
	// OnError is an optional callback invoked on transport errors. Use it
	// to plug in your own logger without coupling this package to one.
	OnError func(req *goresty.Request, err error)
}

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// New returns a *resty.Client with the given defaults applied.
func New(cfg Config) *goresty.Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 15 * time.Second
	}
	if cfg.RetryCount == 0 {
		cfg.RetryCount = 2
	}
	if cfg.RetryWaitTime == 0 {
		cfg.RetryWaitTime = 500 * time.Millisecond
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = defaultUserAgent
	}

	c := goresty.New().
		SetTimeout(cfg.Timeout).
		SetRetryCount(cfg.RetryCount).
		SetRetryWaitTime(cfg.RetryWaitTime).
		SetHeader("User-Agent", cfg.UserAgent).
		SetHeader("Accept", "application/json").
		SetJSONMarshaler(sonic.Marshal).
		SetJSONUnmarshaler(sonic.Unmarshal)

	if cfg.Debug {
		c.SetDebug(true)
	}
	if cfg.OnError != nil {
		c.OnError(cfg.OnError)
	}

	return c
}
