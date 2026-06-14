// Package worldbank is the library behind the worldbank command line:
// the HTTP client, request shaping, and the typed data models for the
// World Bank Open Data API.
//
// The API is public and requires no key. Responses are JSON arrays where
// element [0] is metadata and element [1] is the data slice.
package worldbank

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Host is the World Bank API host.
const Host = "api.worldbank.org"

// BaseURL is the API root every request is built from.
const BaseURL = "https://" + Host + "/v2"

// DefaultUserAgent identifies the client to the World Bank API.
const DefaultUserAgent = "worldbank/dev (+https://github.com/tamnd/worldbank-cli)"

// ErrNotFound is returned when the API returns an empty data array.
var ErrNotFound = errors.New("not found")

// wbResponse is the top-level 2-element wrapper the World Bank API always returns:
// element 0 is pagination metadata, element 1 is the data array.
type wbResponse [2]json.RawMessage

// Client talks to the World Bank Open Data API.
type Client struct {
	HTTP      *http.Client
	UserAgent string
	// Rate is the minimum gap between requests. Zero means no pacing.
	Rate    time.Duration
	Retries int

	last time.Time
}

// NewClient returns a Client with sensible defaults: a 30s timeout, a 500ms
// minimum gap between requests (polite for a public API), and five retries on
// transient errors.
func NewClient() *Client {
	return &Client{
		HTTP:      &http.Client{Timeout: 30 * time.Second},
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Retries:   5,
	}
}

// Get fetches url and returns the response body. It paces and retries according
// to the client's settings.
func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

// pace blocks until at least Rate has passed since the previous request.
func (c *Client) pace() {
	if c.Rate <= 0 {
		return
	}
	if wait := c.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// buildURL constructs an API URL with the given path and query parameters.
// Empty values are omitted.
func buildURL(path string, params map[string]string) string {
	var sb strings.Builder
	sb.WriteString(BaseURL)
	sb.WriteString(path)
	sep := "?"
	for k, v := range params {
		if v != "" {
			sb.WriteString(sep)
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(v)
			sep = "&"
		}
	}
	return sb.String()
}

// fetchWB fetches a URL and unmarshals the outer 2-element wrapper.
func (c *Client) fetchWB(ctx context.Context, rawURL string) (wbResponse, error) {
	body, err := c.Get(ctx, rawURL)
	if err != nil {
		return wbResponse{}, err
	}
	var resp wbResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return wbResponse{}, fmt.Errorf("parse response: %w", err)
	}
	return resp, nil
}

// ListCountries returns countries, optionally filtered by region code.
func (c *Client) ListCountries(ctx context.Context, region string, limit int) ([]Country, error) {
	params := map[string]string{
		"format":   "json",
		"per_page": fmt.Sprintf("%d", limit),
	}
	if region != "" {
		params["region"] = region
	}
	resp, err := c.fetchWB(ctx, buildURL("/country", params))
	if err != nil {
		return nil, err
	}
	if resp[1] == nil || string(resp[1]) == "null" {
		return nil, ErrNotFound
	}
	var items []wireCountry
	if err := json.Unmarshal(resp[1], &items); err != nil {
		return nil, fmt.Errorf("parse countries: %w", err)
	}
	out := make([]Country, 0, len(items))
	for _, it := range items {
		out = append(out, it.toCountry())
	}
	return out, nil
}

// ListIndicators returns indicators from World Development Indicators (source=2).
func (c *Client) ListIndicators(ctx context.Context, limit int) ([]Indicator, error) {
	params := map[string]string{
		"format":   "json",
		"per_page": fmt.Sprintf("%d", limit),
		"source":   "2",
	}
	resp, err := c.fetchWB(ctx, buildURL("/indicator", params))
	if err != nil {
		return nil, err
	}
	if resp[1] == nil || string(resp[1]) == "null" {
		return nil, ErrNotFound
	}
	var items []wireIndicator
	if err := json.Unmarshal(resp[1], &items); err != nil {
		return nil, fmt.Errorf("parse indicators: %w", err)
	}
	out := make([]Indicator, 0, len(items))
	for _, it := range items {
		out = append(out, it.toIndicator())
	}
	return out, nil
}

// GetData returns indicator data for one or more countries (semicolon-separated),
// limited to the most recent mrv values.
func (c *Client) GetData(ctx context.Context, country, indicator string, mrv int) ([]DataPoint, error) {
	path := fmt.Sprintf("/country/%s/indicator/%s", country, indicator)
	params := map[string]string{
		"format": "json",
		"mrv":    fmt.Sprintf("%d", mrv),
	}
	resp, err := c.fetchWB(ctx, buildURL(path, params))
	if err != nil {
		return nil, err
	}
	if resp[1] == nil || string(resp[1]) == "null" {
		return nil, ErrNotFound
	}
	var items []wireDataPoint
	if err := json.Unmarshal(resp[1], &items); err != nil {
		return nil, fmt.Errorf("parse data: %w", err)
	}
	out := make([]DataPoint, 0, len(items))
	for _, it := range items {
		if it.Value != nil {
			out = append(out, it.toDataPoint())
		}
	}
	if len(out) == 0 {
		return nil, ErrNotFound
	}
	return out, nil
}

// ListTopics returns all World Bank topics.
func (c *Client) ListTopics(ctx context.Context) ([]Topic, error) {
	resp, err := c.fetchWB(ctx, buildURL("/topic", map[string]string{"format": "json"}))
	if err != nil {
		return nil, err
	}
	if resp[1] == nil || string(resp[1]) == "null" {
		return nil, ErrNotFound
	}
	var items []wireTopic
	if err := json.Unmarshal(resp[1], &items); err != nil {
		return nil, fmt.Errorf("parse topics: %w", err)
	}
	out := make([]Topic, 0, len(items))
	for _, it := range items {
		out = append(out, it.toTopic())
	}
	return out, nil
}
