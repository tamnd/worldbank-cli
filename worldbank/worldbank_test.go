package worldbank_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/worldbank-cli/worldbank"
)

func newTestClient(ts *httptest.Server) *worldbank.Client {
	cfg := worldbank.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return worldbank.NewClient(cfg)
}

func TestCountries(t *testing.T) {
	meta := map[string]any{"page": 1, "pages": 1, "per_page": 50, "total": 1}
	data := []map[string]any{
		{
			"id":       "USA",
			"iso2Code": "US",
			"name":     "United States",
			"region":   map[string]string{"id": "NAC", "value": "North America"},
			"incomeLevel": map[string]string{"id": "HIC", "value": "High income"},
			"capitalCity": "Washington D.C.",
			"longitude": "-77.032",
			"latitude":  "38.8895",
		},
	}
	payload, _ := json.Marshal([2]any{meta, data})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	countries, err := c.Countries(context.Background(), "", "", 1, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(countries) != 1 {
		t.Fatalf("got %d countries, want 1", len(countries))
	}
	if countries[0].ID != "USA" {
		t.Errorf("ID = %q, want USA", countries[0].ID)
	}
	if countries[0].Name != "United States" {
		t.Errorf("Name = %q, want United States", countries[0].Name)
	}
	if countries[0].Region != "North America" {
		t.Errorf("Region = %q, want North America", countries[0].Region)
	}
}

func TestCountry(t *testing.T) {
	meta := map[string]any{"page": 1, "pages": 1, "per_page": 50, "total": 1}
	data := []map[string]any{
		{
			"id":       "BRA",
			"iso2Code": "BR",
			"name":     "Brazil",
			"region":   map[string]string{"id": "LCN", "value": "Latin America & Caribbean"},
			"incomeLevel": map[string]string{"id": "UMC", "value": "Upper middle income"},
			"capitalCity": "Brasilia",
			"longitude": "-47.9292",
			"latitude":  "-15.7801",
		},
	}
	payload, _ := json.Marshal([2]any{meta, data})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	country, err := c.Country(context.Background(), "BR")
	if err != nil {
		t.Fatal(err)
	}
	if country.ISO2Code != "BR" {
		t.Errorf("ISO2Code = %q, want BR", country.ISO2Code)
	}
	if country.CapitalCity != "Brasilia" {
		t.Errorf("CapitalCity = %q, want Brasilia", country.CapitalCity)
	}
}

func TestCountryNotFound(t *testing.T) {
	meta := map[string]any{"page": 1, "pages": 0, "per_page": 50, "total": 0}
	payload, _ := json.Marshal([2]any{meta, nil})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Country(context.Background(), "XXXX")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestIndicators(t *testing.T) {
	meta := map[string]any{"page": 1, "pages": 1, "per_page": 50, "total": 1}
	data := []map[string]any{
		{
			"id":   "NY.GDP.MKTP.CD",
			"name": "GDP (current US$)",
			"unit": "",
			"source": map[string]string{"id": "2", "value": "World Development Indicators"},
			"topics": []map[string]string{{"id": "3", "value": "Economy & Growth"}},
		},
	}
	payload, _ := json.Marshal([2]any{meta, data})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	indicators, err := c.Indicators(context.Background(), "", 1, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(indicators) != 1 {
		t.Fatalf("got %d indicators, want 1", len(indicators))
	}
	if indicators[0].ID != "NY.GDP.MKTP.CD" {
		t.Errorf("ID = %q, want NY.GDP.MKTP.CD", indicators[0].ID)
	}
}

func TestData(t *testing.T) {
	val := 25462700000000.0
	meta := map[string]any{"page": 1, "pages": 1, "per_page": 10, "total": 10}
	data := []map[string]any{
		{
			"country":   map[string]string{"id": "US", "value": "United States"},
			"indicator": map[string]string{"id": "NY.GDP.MKTP.CD", "value": "GDP (current US$)"},
			"date":      "2022",
			"value":     val,
		},
	}
	payload, _ := json.Marshal([2]any{meta, data})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	points, err := c.Data(context.Background(), "US", "NY.GDP.MKTP.CD", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 {
		t.Fatalf("got %d points, want 1", len(points))
	}
	if points[0].Year != "2022" {
		t.Errorf("Year = %q, want 2022", points[0].Year)
	}
	if points[0].Country != "United States" {
		t.Errorf("Country = %q, want United States", points[0].Country)
	}
}

func TestRetryOn503(t *testing.T) {
	var hits int
	meta := map[string]any{"page": 1, "pages": 1, "per_page": 50, "total": 0}
	payload, _ := json.Marshal([2]any{meta, []map[string]any{}})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	}))
	defer ts.Close()

	cfg := worldbank.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := worldbank.NewClient(cfg)
	_, _ = c.Countries(context.Background(), "", "", 1, 50)
	if hits < 3 {
		t.Errorf("expected at least 3 hits (retries), got %d", hits)
	}
}
