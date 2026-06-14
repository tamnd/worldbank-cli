package worldbank_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/worldbank-cli/worldbank"
)

// testClient builds a Client pointed at the given test server with pacing off.
func testClient(ts *httptest.Server) *worldbank.Client {
	c := worldbank.NewClient()
	c.HTTP = &http.Client{}
	c.Rate = 0
	// point the client at the test server by overriding BaseURL via a trick:
	// we replace the HTTP client's transport so it redirects to ts.URL.
	c.HTTP = ts.Client()
	c.HTTP.Transport = rewriteTransport{base: ts.URL, inner: ts.Client().Transport}
	return c
}

// rewriteTransport rewrites every request's host+scheme to point at the test
// server, so the worldbank package's internal BaseURL is not hard-coded in tests.
type rewriteTransport struct {
	base  string
	inner http.RoundTripper
}

func (rt rewriteTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// parse the test server base URL
	tsReq := r.Clone(r.Context())
	tsReq.URL.Scheme = "http"
	tsReq.URL.Host = rt.base[len("http://"):]
	return rt.inner.RoundTrip(tsReq)
}

func makePayload(t *testing.T, meta, data any) []byte {
	t.Helper()
	b, err := json.Marshal([2]any{meta, data})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return b
}

func jsonHandler(payload []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	})
}

func TestListCountries(t *testing.T) {
	meta := map[string]any{"page": 1, "pages": 1, "per_page": 20, "total": 1}
	data := []map[string]any{
		{
			"id":          "USA",
			"iso2Code":    "US",
			"name":        "United States",
			"region":      map[string]string{"id": "NAC", "value": "North America"},
			"incomeLevel": map[string]string{"id": "HIC", "value": "High income"},
			"capitalCity": "Washington D.C.",
			"longitude":   "-77.032",
			"latitude":    "38.8895",
		},
	}
	payload := makePayload(t, meta, data)
	ts := httptest.NewServer(jsonHandler(payload))
	defer ts.Close()

	c := testClient(ts)
	countries, err := c.ListCountries(context.Background(), "", 20)
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
	if countries[0].ISO2 != "US" {
		t.Errorf("ISO2 = %q, want US", countries[0].ISO2)
	}
	if countries[0].CapitalCity != "Washington D.C." {
		t.Errorf("CapitalCity = %q, want Washington D.C.", countries[0].CapitalCity)
	}
}

func TestListCountriesRegionFilter(t *testing.T) {
	var gotRegion string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRegion = r.URL.Query().Get("region")
		meta := map[string]any{"page": 1, "pages": 1, "per_page": 20, "total": 0}
		payload := makePayload(t, meta, []any{})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	}))
	defer ts.Close()

	c := testClient(ts)
	_, _ = c.ListCountries(context.Background(), "EAS", 20)
	if gotRegion != "EAS" {
		t.Errorf("region query param = %q, want EAS", gotRegion)
	}
}

func TestListIndicators(t *testing.T) {
	meta := map[string]any{"page": 1, "pages": 1, "per_page": 20, "total": 1}
	data := []map[string]any{
		{
			"id":         "NY.GDP.MKTP.CD",
			"name":       "GDP (current US$)",
			"unit":       "",
			"sourceNote": "GDP at purchaser prices is the sum of gross value added...",
			"source":     map[string]string{"id": "2", "value": "World Development Indicators"},
		},
	}
	payload := makePayload(t, meta, data)
	ts := httptest.NewServer(jsonHandler(payload))
	defer ts.Close()

	c := testClient(ts)
	indicators, err := c.ListIndicators(context.Background(), 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(indicators) != 1 {
		t.Fatalf("got %d indicators, want 1", len(indicators))
	}
	if indicators[0].ID != "NY.GDP.MKTP.CD" {
		t.Errorf("ID = %q, want NY.GDP.MKTP.CD", indicators[0].ID)
	}
	if indicators[0].Name != "GDP (current US$)" {
		t.Errorf("Name = %q, want GDP (current US$)", indicators[0].Name)
	}
}

func TestGetData(t *testing.T) {
	val := 28750956130731.2
	meta := map[string]any{"page": 1, "pages": 1, "per_page": 5, "total": 5}
	data := []map[string]any{
		{
			"country":        map[string]string{"id": "US", "value": "United States"},
			"indicator":      map[string]string{"id": "NY.GDP.MKTP.CD", "value": "GDP (current US$)"},
			"countryiso3code": "USA",
			"date":           "2024",
			"value":          val,
		},
	}
	payload := makePayload(t, meta, data)
	ts := httptest.NewServer(jsonHandler(payload))
	defer ts.Close()

	c := testClient(ts)
	points, err := c.GetData(context.Background(), "US", "NY.GDP.MKTP.CD", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 {
		t.Fatalf("got %d points, want 1", len(points))
	}
	if points[0].CountryID != "US" {
		t.Errorf("CountryID = %q, want US", points[0].CountryID)
	}
	if points[0].CountryName != "United States" {
		t.Errorf("CountryName = %q, want United States", points[0].CountryName)
	}
	if points[0].IndicatorID != "NY.GDP.MKTP.CD" {
		t.Errorf("IndicatorID = %q, want NY.GDP.MKTP.CD", points[0].IndicatorID)
	}
	if points[0].Date != "2024" {
		t.Errorf("Date = %q, want 2024", points[0].Date)
	}
	if points[0].Value != val {
		t.Errorf("Value = %v, want %v", points[0].Value, val)
	}
}

func TestGetDataSkipsNullValues(t *testing.T) {
	meta := map[string]any{"page": 1, "pages": 1, "per_page": 3, "total": 3}
	data := []map[string]any{
		{
			"country":   map[string]string{"id": "US", "value": "United States"},
			"indicator": map[string]string{"id": "NY.GDP.MKTP.CD", "value": "GDP"},
			"date":      "2024",
			"value":     28000000000000.0,
		},
		{
			"country":   map[string]string{"id": "US", "value": "United States"},
			"indicator": map[string]string{"id": "NY.GDP.MKTP.CD", "value": "GDP"},
			"date":      "2023",
			"value":     nil, // null value, should be skipped
		},
		{
			"country":   map[string]string{"id": "US", "value": "United States"},
			"indicator": map[string]string{"id": "NY.GDP.MKTP.CD", "value": "GDP"},
			"date":      "2022",
			"value":     25000000000000.0,
		},
	}
	payload := makePayload(t, meta, data)
	ts := httptest.NewServer(jsonHandler(payload))
	defer ts.Close()

	c := testClient(ts)
	points, err := c.GetData(context.Background(), "US", "NY.GDP.MKTP.CD", 3)
	if err != nil {
		t.Fatal(err)
	}
	// only 2 non-null points
	if len(points) != 2 {
		t.Errorf("got %d points, want 2 (null skipped)", len(points))
	}
}

func TestListTopics(t *testing.T) {
	meta := map[string]any{"page": 1, "pages": 1, "per_page": 50, "total": 5}
	data := []map[string]any{
		{"id": "1", "value": "Agriculture & Rural Development", "sourceNote": "For the 70 percent..."},
		{"id": "3", "value": "Economy & Growth", "sourceNote": "Economic growth is central..."},
	}
	payload := makePayload(t, meta, data)
	ts := httptest.NewServer(jsonHandler(payload))
	defer ts.Close()

	c := testClient(ts)
	topics, err := c.ListTopics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(topics) != 2 {
		t.Fatalf("got %d topics, want 2", len(topics))
	}
	if topics[0].ID != "1" {
		t.Errorf("topics[0].ID = %q, want 1", topics[0].ID)
	}
	if topics[0].Value != "Agriculture & Rural Development" {
		t.Errorf("topics[0].Value = %q", topics[0].Value)
	}
	if topics[1].ID != "3" {
		t.Errorf("topics[1].ID = %q, want 3", topics[1].ID)
	}
}

func TestRetryOn503(t *testing.T) {
	var hits int
	meta := map[string]any{"page": 1, "pages": 1, "per_page": 20, "total": 0}
	payload := makePayload(t, meta, []any{})

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

	c := testClient(ts)
	c.Retries = 5
	_, _ = c.ListCountries(context.Background(), "", 20)
	if hits < 3 {
		t.Errorf("expected at least 3 hits (retries), got %d", hits)
	}
}

func TestGetDataNotFound(t *testing.T) {
	meta := map[string]any{"page": 1, "pages": 0, "per_page": 5, "total": 0}
	payload := makePayload(t, meta, nil)
	ts := httptest.NewServer(jsonHandler(payload))
	defer ts.Close()

	c := testClient(ts)
	_, err := c.GetData(context.Background(), "INVALID", "NY.GDP.MKTP.CD", 5)
	if err == nil {
		t.Error("expected error for null data, got nil")
	}
}
