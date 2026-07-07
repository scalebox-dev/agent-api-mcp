package agentapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	AuthHeader string
	HTTPClient *http.Client
}

func New(baseURL string, authHeader string, timeout time.Duration) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		AuthHeader: strings.TrimSpace(authHeader),
		HTTPClient: &http.Client{Timeout: timeout},
	}
}

func BearerFromKey(key string) string {
	if strings.TrimSpace(key) == "" {
		return ""
	}
	return "Bearer " + strings.TrimSpace(key)
}

func PreferRequestAuth(requestAuth string, fallbackKey string) string {
	if strings.TrimSpace(requestAuth) != "" {
		return strings.TrimSpace(requestAuth)
	}
	return BearerFromKey(fallbackKey)
}

func (c *Client) Get(ctx context.Context, endpoint string, query map[string]string) (map[string]any, error) {
	return c.Do(ctx, http.MethodGet, endpoint, query, nil)
}

func (c *Client) Post(ctx context.Context, endpoint string, body any) (map[string]any, error) {
	return c.Do(ctx, http.MethodPost, endpoint, nil, body)
}

func (c *Client) Patch(ctx context.Context, endpoint string, body any) (map[string]any, error) {
	return c.Do(ctx, http.MethodPatch, endpoint, nil, body)
}

func (c *Client) PutRaw(ctx context.Context, endpoint string, content string) (map[string]any, error) {
	return c.doRaw(ctx, http.MethodPut, c.endpoint(endpoint), "text/plain; charset=utf-8", strings.NewReader(content))
}

func (c *Client) Do(ctx context.Context, method string, endpoint string, query map[string]string, body any) (map[string]any, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	return c.doRaw(ctx, method, withQuery(c.endpoint(endpoint), query), "application/json", reader)
}

func (c *Client) doRaw(ctx context.Context, method string, endpoint string, contentType string, body io.Reader) (map[string]any, error) {
	if c.AuthHeader == "" {
		return nil, fmt.Errorf("missing Agent API credential: provide Authorization bearer token or AGENT_API_KEY")
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.AuthHeader)
	if body != nil && contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("agent api %s %s failed: status=%d body=%s", method, endpoint, resp.StatusCode, string(raw))
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return map[string]any{"ok": true}, nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode agent api response: %w", err)
	}
	return out, nil
}

func (c *Client) endpoint(endpoint string) string {
	base := strings.TrimRight(c.BaseURL, "/")
	return base + "/" + strings.TrimLeft(endpoint, "/")
}

func Segment(value string) string {
	return url.PathEscape(strings.TrimSpace(value))
}

func Join(parts ...string) string {
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			clean = append(clean, part)
		}
	}
	return "/" + path.Join(clean...)
}

func Query(values map[string]any) map[string]string {
	out := map[string]string{}
	for key, value := range values {
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				out[key] = v
			}
		case bool:
			out[key] = fmt.Sprintf("%t", v)
		case int:
			if v != 0 {
				out[key] = fmt.Sprintf("%d", v)
			}
		case int64:
			if v != 0 {
				out[key] = fmt.Sprintf("%d", v)
			}
		}
	}
	return out
}

func withQuery(raw string, query map[string]string) string {
	if len(query) == 0 {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	for key, value := range query {
		if value != "" {
			q.Set(key, value)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}
