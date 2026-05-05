package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	baseURL     string
	httpClient  *http.Client
	syncClient  *http.Client
}

func closeBody(c io.Closer) {
	_ = c.Close()
}

func New(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	syncTimeout := cfg.SyncTimeout
	if syncTimeout == 0 {
		// Default: server tope ~30s; damos 60s para cubrir red + retry interno.
		syncTimeout = 60 * time.Second
		if timeout > syncTimeout {
			syncTimeout = timeout
		}
	}
	return &Client{
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
		syncClient: &http.Client{Timeout: syncTimeout},
	}
}

func (c *Client) Upload(ctx context.Context, filename string, content io.Reader, contentType string, opts UploadOptions) (*MediaResp, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, fmt.Errorf("media upload: create form file: %w", err)
	}
	if _, err := io.Copy(part, content); err != nil {
		return nil, fmt.Errorf("media upload: copy content: %w", err)
	}

	fields := map[string]string{
		"bucket":      opts.Bucket,
		"service":     opts.Service,
		"entity_type": opts.EntityType,
		"entity_id":   opts.EntityID,
		"preset":      opts.Preset,
		"subpath":     opts.Subpath,
		"category":    opts.Category,
	}
	if len(opts.Tags) > 0 {
		fields["tags"] = strings.Join(opts.Tags, ",")
	}
	if opts.WaitForProcessed {
		fields["wait"] = "true"
	}
	for k, v := range fields {
		if v != "" {
			_ = w.WriteField(k, v)
		}
	}
	_ = w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/media", &buf)
	if err != nil {
		return nil, fmt.Errorf("media upload: build request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("x-account-id", opts.AccountUID)

	httpClient := c.httpClient
	if opts.WaitForProcessed {
		httpClient = c.syncClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("media upload: http: %w", err)
	}
	defer closeBody(resp.Body)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("media upload: status %d: %s", resp.StatusCode, body)
	}

	var env envelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, fmt.Errorf("media upload: decode response: %w", err)
	}
	return &env.Data, nil
}

func (c *Client) Get(ctx context.Context, accountUID, uid string) (*MediaResp, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/media/"+uid, nil)
	if err != nil {
		return nil, fmt.Errorf("media get: %w", err)
	}
	req.Header.Set("x-account-id", accountUID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("media get: http: %w", err)
	}
	defer closeBody(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("media get: status %d: %s", resp.StatusCode, body)
	}

	var env envelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, fmt.Errorf("media get: decode response: %w", err)
	}
	return &env.Data, nil
}

func (c *Client) Delete(ctx context.Context, accountUID, uid string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/api/v1/media/"+uid, nil)
	if err != nil {
		return fmt.Errorf("media delete: %w", err)
	}
	req.Header.Set("x-account-id", accountUID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("media delete: http: %w", err)
	}
	defer closeBody(resp.Body)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("media delete: status %d: %s", resp.StatusCode, body)
	}
	return nil
}
