package resty

import (
	"time"

	"github.com/bytedance/sonic"
	goresty "github.com/go-resty/resty/v2"
)

type Config struct {
	Timeout       time.Duration
	RetryCount    int
	RetryWaitTime time.Duration
	UserAgent     string
	Debug         bool
	OnError       func(req *goresty.Request, err error)
}

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

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
