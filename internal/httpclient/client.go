package httpclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"htui/internal/types"
)

// DefaultTimeout для исходящих запросов.
const DefaultTimeout = 30 * time.Second

// Client оборачивает net/http.
type Client struct {
	http *http.Client
}

// New создаёт клиент с дефолтным транспортом.
func New() *Client {
	return &Client{
		http: &http.Client{Timeout: DefaultTimeout},
	}
}

// Execute выполняет SavedRequest и возвращает ResponseData.
func (c *Client) Execute(ctx context.Context, r types.SavedRequest) types.ResponseData {
	start := time.Now()
	out := types.ResponseData{}

	finalURL, err := buildURL(r)
	if err != nil {
		out.Error = "Invalid URL"
		out.DurationMs = time.Since(start).Milliseconds()
		return out
	}

	req, err := http.NewRequestWithContext(ctx, r.Method, finalURL, nil)
	if err != nil {
		out.Error = "Could not build request"
		out.DurationMs = time.Since(start).Milliseconds()
		return out
	}

	for _, h := range r.Headers {
		if h.Enabled && h.Key != "" {
			req.Header.Set(h.Key, h.Value)
		}
	}

	if r.Auth.Type == types.AuthBearer && r.Auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Auth.Token)
	}

	bodyReader, err := buildBody(r, req)
	if err != nil {
		out.Error = err.Error()
		out.DurationMs = time.Since(start).Milliseconds()
		return out
	}
	if bodyReader != nil {
		req.Body = bodyReader
	}

	resp, err := c.http.Do(req)
	out.DurationMs = time.Since(start).Milliseconds()
	if err != nil {
		if ctx.Err() != nil {
			out.Error = "Request cancelled"
		} else if strings.Contains(strings.ToLower(err.Error()), "timeout") {
			out.Error = "Request timed out"
		} else {
			out.Error = "Could not connect to server"
		}
		return out
	}
	defer resp.Body.Close()

	out.StatusCode = resp.StatusCode
	out.StatusText = resp.Status

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		out.Error = "Could not read response body"
		return out
	}
	out.SizeBytes = len(bodyBytes)
	out.Body = string(bodyBytes)

	for k, vals := range resp.Header {
		for _, v := range vals {
			out.Headers = append(out.Headers, types.Header{Key: k, Value: v, Enabled: true})
		}
	}

	return out
}

func buildURL(r types.SavedRequest) (string, error) {
	raw := strings.TrimSpace(r.URL)
	if raw == "" {
		return "", errors.New("empty url")
	}
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if r.BodyMode != types.BodyModeForm {
		q := u.Query()
		for _, p := range r.Params {
			if p.Enabled && p.Key != "" {
				q.Set(p.Key, p.Value)
			}
		}
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

func buildBody(r types.SavedRequest, req *http.Request) (io.ReadCloser, error) {
	switch r.BodyMode {
	case types.BodyModeNone:
		return nil, nil
	case types.BodyModeRawText:
		if r.Body == "" {
			return nil, nil
		}
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "text/plain; charset=utf-8")
		}
		return io.NopCloser(strings.NewReader(r.Body)), nil
	case types.BodyModeJSON:
		if r.Body == "" {
			return nil, nil
		}
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
		}
		return io.NopCloser(strings.NewReader(r.Body)), nil
	case types.BodyModeForm:
		form := url.Values{}
		for _, p := range r.Params {
			if p.Enabled && p.Key != "" {
				form.Set(p.Key, p.Value)
			}
		}
		encoded := form.Encode()
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return io.NopCloser(strings.NewReader(encoded)), nil
	default:
		return nil, nil
	}
}
