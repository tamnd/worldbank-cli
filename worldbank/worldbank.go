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
	"net/url"
	"strings"
	"time"
)

const (
	defaultBase      = "https://api.worldbank.org/v2"
	DefaultUserAgent = "worldbank/dev (+https://github.com/tamnd/worldbank-cli)"
)

// ErrNotFound is returned when the API returns an empty data array.
var ErrNotFound = errors.New("not found")

// Config holds constructor parameters.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   defaultBase,
		UserAgent: DefaultUserAgent,
		Rate:      200 * time.Millisecond,
		Retries:   5,
		Timeout:   30 * time.Second,
	}
}

// Client talks to the World Bank API.
type Client struct {
	http      *http.Client
	userAgent string
	baseURL   string
	rate      time.Duration
	retries   int
	last      time.Time
}

// NewClient returns a Client configured from cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		http:      &http.Client{Timeout: cfg.Timeout},
		userAgent: cfg.UserAgent,
		baseURL:   strings.TrimRight(cfg.BaseURL, "/"),
		rate:      cfg.Rate,
		retries:   cfg.Retries,
	}
}

// Countries returns a page of countries, optionally filtered by region or income level.
func (c *Client) Countries(ctx context.Context, region, income string, page, perPage int) ([]Country, error) {
	u := c.url("/country", map[string]string{
		"format":   "json",
		"per_page": fmt.Sprintf("%d", perPage),
		"page":     fmt.Sprintf("%d", page),
		"region":   region,
		"incomelevel": income,
	})
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw [2]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse countries: %w", err)
	}
	if raw[1] == nil || string(raw[1]) == "null" {
		return nil, ErrNotFound
	}
	var items []wireCountry
	if err := json.Unmarshal(raw[1], &items); err != nil {
		return nil, fmt.Errorf("parse countries data: %w", err)
	}
	out := make([]Country, 0, len(items))
	for _, it := range items {
		out = append(out, it.toCountry())
	}
	return out, nil
}

// Country returns a single country by ISO2 or ISO3 code.
func (c *Client) Country(ctx context.Context, code string) (Country, error) {
	u := c.url("/country/"+url.PathEscape(strings.ToUpper(code)), map[string]string{
		"format": "json",
	})
	body, err := c.get(ctx, u)
	if err != nil {
		return Country{}, err
	}
	var raw [2]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return Country{}, fmt.Errorf("parse country: %w", err)
	}
	if raw[1] == nil || string(raw[1]) == "null" {
		return Country{}, ErrNotFound
	}
	var items []wireCountry
	if err := json.Unmarshal(raw[1], &items); err != nil {
		return Country{}, fmt.Errorf("parse country data: %w", err)
	}
	if len(items) == 0 {
		return Country{}, ErrNotFound
	}
	return items[0].toCountry(), nil
}

// Indicators returns a page of indicators, optionally filtered by search term.
func (c *Client) Indicators(ctx context.Context, search string, page, perPage int) ([]Indicator, error) {
	u := c.url("/indicator", map[string]string{
		"format":   "json",
		"per_page": fmt.Sprintf("%d", perPage),
		"page":     fmt.Sprintf("%d", page),
	})
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw [2]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse indicators: %w", err)
	}
	if raw[1] == nil || string(raw[1]) == "null" {
		return nil, ErrNotFound
	}
	var items []wireIndicator
	if err := json.Unmarshal(raw[1], &items); err != nil {
		return nil, fmt.Errorf("parse indicators data: %w", err)
	}
	out := make([]Indicator, 0, len(items))
	lower := strings.ToLower(search)
	for _, it := range items {
		ind := it.toIndicator()
		if lower == "" || strings.Contains(strings.ToLower(ind.Name), lower) || strings.Contains(strings.ToLower(ind.ID), lower) {
			out = append(out, ind)
		}
	}
	if len(out) == 0 && search != "" {
		return nil, ErrNotFound
	}
	return out, nil
}

// Data returns indicator data for a country, most recent first.
func (c *Client) Data(ctx context.Context, countryCode, indicatorID string, perPage int) ([]DataPoint, error) {
	path := fmt.Sprintf("/country/%s/indicator/%s",
		url.PathEscape(strings.ToUpper(countryCode)),
		url.PathEscape(indicatorID),
	)
	u := c.url(path, map[string]string{
		"format":   "json",
		"per_page": fmt.Sprintf("%d", perPage),
		"mrv":      fmt.Sprintf("%d", perPage),
	})
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw [2]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse data: %w", err)
	}
	if raw[1] == nil || string(raw[1]) == "null" {
		return nil, ErrNotFound
	}
	var items []wireDataPoint
	if err := json.Unmarshal(raw[1], &items); err != nil {
		return nil, fmt.Errorf("parse data points: %w", err)
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

// url builds a full API URL with query parameters. Empty values are omitted.
func (c *Client) url(path string, params map[string]string) string {
	q := url.Values{}
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	return c.baseURL + path + "?" + q.Encode()
}

// get fetches a URL with pacing and retries.
func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.http.Do(req)
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

func (c *Client) pace() {
	if c.rate <= 0 {
		return
	}
	if wait := c.rate - time.Since(c.last); wait > 0 {
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
