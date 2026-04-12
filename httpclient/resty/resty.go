package resty

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	goresty "github.com/go-resty/resty/v2"
)

const (
	defaultTimeout       = 15 * time.Second
	defaultRetryCount    = 2
	defaultRetryWaitTime = 500 * time.Millisecond
	defaultUserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

	headerContentType   = "Content-Type"
	headerAccept        = "Accept"
	headerAuthorization = "Authorization"
	headerUserAgent     = "User-Agent"

	mimeJSON = "application/json"
	mimeForm = "application/x-www-form-urlencoded"
)

type Config struct {
	BaseURL string
	Headers map[string][]string

	Timeout       time.Duration
	RetryCount    int
	RetryWaitTime time.Duration
	UserAgent     string
	Debug         bool
	OnError       func(req *goresty.Request, err error)
}

type Service struct {
	underlying *goresty.Client
}

type Call struct {
	req    *goresty.Request
	strict bool
}

type HTTPStatusError struct {
	StatusCode int
	Method     string
	URL        string
	Body       string
}

func New(cfg Config) *goresty.Client {
	return buildClient(cfg)
}

func NewService(cfg Config) *Service {
	return &Service{underlying: buildClient(cfg)}
}

func (s *Service) Raw() *goresty.Client {
	if s == nil {
		return nil
	}
	return s.underlying
}

func (s *Service) R(ctx context.Context) *goresty.Request {
	return s.underlying.R().SetContext(ctx)
}

func (s *Service) Call(ctx context.Context) *Call {
	return &Call{req: s.underlying.R().SetContext(ctx)}
}

func (c *Call) Header(key string, values ...string) *Call {
	switch len(values) {
	case 0:
		return c
	case 1:
		c.req.SetHeader(key, values[0])
	default:
		c.req.SetHeaderMultiValues(map[string][]string{key: values})
	}
	return c
}

func (c *Call) AddHeader(key, value string) *Call {
	c.req.Header.Add(key, value)
	return c
}

func (c *Call) Headers(headers map[string][]string) *Call {
	if len(headers) > 0 {
		c.req.SetHeaderMultiValues(headers)
	}
	return c
}

func (c *Call) JSON() *Call {
	c.req.SetHeader(headerContentType, mimeJSON)
	c.req.SetHeader(headerAccept, mimeJSON)
	return c
}

func (c *Call) Form() *Call {
	c.req.SetHeader(headerContentType, mimeForm)
	return c
}

func (c *Call) Authorization(value string) *Call {
	if value != "" {
		c.req.SetHeader(headerAuthorization, value)
	}
	return c
}

func (c *Call) Bearer(token string) *Call {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return c
	}
	if !strings.HasPrefix(trimmed, "Bearer ") {
		trimmed = "Bearer " + trimmed
	}
	c.req.SetHeader(headerAuthorization, trimmed)
	return c
}

func (c *Call) Basic(username, password string) *Call {
	if username == "" || password == "" {
		return c
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	c.req.SetHeader(headerAuthorization, "Basic "+encoded)
	return c
}

func (c *Call) Body(payload any) *Call {
	if payload != nil {
		c.req.SetBody(payload)
	}
	return c
}

func (c *Call) Into(target any) *Call {
	if target != nil {
		c.req.SetResult(target)
	}
	return c
}

func (c *Call) Strict() *Call {
	c.strict = true
	return c
}

func (c *Call) Raw() *goresty.Request {
	return c.req
}

func (c *Call) Do(method, path string) (*goresty.Response, error) {
	resp, err := c.req.Execute(method, path)
	if c.strict {
		if sErr := expectSuccess(resp, err); sErr != nil {
			return resp, sErr
		}
	}
	return resp, err
}

func (c *Call) Get(path string) (*goresty.Response, error)    { return c.Do(http.MethodGet, path) }
func (c *Call) Post(path string) (*goresty.Response, error)   { return c.Do(http.MethodPost, path) }
func (c *Call) Put(path string) (*goresty.Response, error)    { return c.Do(http.MethodPut, path) }
func (c *Call) Patch(path string) (*goresty.Response, error)  { return c.Do(http.MethodPatch, path) }
func (c *Call) Delete(path string) (*goresty.Response, error) { return c.Do(http.MethodDelete, path) }

func Do[Out any](c *Call, method, path string) (Out, error) {
	var out Out
	_, err := c.Into(&out).Do(method, path)
	return out, err
}

func Get[Out any](c *Call, path string) (Out, error) {
	return Do[Out](c, http.MethodGet, path)
}

func Post[Out, In any](c *Call, path string, body In) (Out, error) {
	return Do[Out](c.Body(body), http.MethodPost, path)
}

func Put[Out, In any](c *Call, path string, body In) (Out, error) {
	return Do[Out](c.Body(body), http.MethodPut, path)
}

func Patch[Out, In any](c *Call, path string, body In) (Out, error) {
	return Do[Out](c.Body(body), http.MethodPatch, path)
}

func Delete[Out any](c *Call, path string) (Out, error) {
	return Do[Out](c, http.MethodDelete, path)
}

func (e *HTTPStatusError) Error() string {
	if e == nil {
		return "HTTPStatusError: nil"
	}
	return fmt.Sprintf("http %s %s: status %d", e.Method, e.URL, e.StatusCode)
}

func AsHTTPError(err error) (*HTTPStatusError, bool) {
	if he, ok := errors.AsType[*HTTPStatusError](err); ok {
		return he, true
	}
	return nil, false
}

func HTTPStatusCode(err error) (int, bool) {
	if he, ok := AsHTTPError(err); ok {
		return he.StatusCode, true
	}
	return 0, false
}

func buildClient(cfg Config) *goresty.Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.RetryCount == 0 {
		cfg.RetryCount = defaultRetryCount
	}
	if cfg.RetryWaitTime == 0 {
		cfg.RetryWaitTime = defaultRetryWaitTime
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = defaultUserAgent
	}

	client := goresty.New().
		SetTimeout(cfg.Timeout).
		SetRetryCount(cfg.RetryCount).
		SetRetryWaitTime(cfg.RetryWaitTime).
		SetHeader(headerUserAgent, cfg.UserAgent).
		SetHeader(headerAccept, mimeJSON).
		SetJSONMarshaler(sonic.Marshal).
		SetJSONUnmarshaler(sonic.Unmarshal)

	if cfg.BaseURL != "" {
		client.SetBaseURL(cfg.BaseURL)
	}
	for key, values := range cfg.Headers {
		client.Header.Del(key)
		for _, value := range values {
			client.Header.Add(key, value)
		}
	}
	if cfg.Debug {
		client.SetDebug(true)
	}
	if cfg.OnError != nil {
		client.OnError(cfg.OnError)
	}
	return client
}

func expectSuccess(resp *goresty.Response, err error) error {
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("resty: nil response")
	}
	if !resp.IsError() {
		return nil
	}
	statusErr := &HTTPStatusError{
		StatusCode: resp.StatusCode(),
		Body:       string(resp.Body()),
	}
	if resp.Request != nil {
		statusErr.Method = resp.Request.Method
		statusErr.URL = resp.Request.URL
	}
	return statusErr
}
