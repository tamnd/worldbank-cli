package worldbank

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

// These tests are offline: they exercise the URI driver's pure string functions
// and the host wiring, which need no network.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "worldbank" {
		t.Errorf("Scheme = %q, want worldbank", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s ...]", info.Hosts, Host)
	}
	if info.Identity.Binary != "worldbank" {
		t.Errorf("Identity.Binary = %q, want worldbank", info.Identity.Binary)
	}
	found := false
	for _, a := range info.Aliases {
		if a == "wb" {
			found = true
		}
	}
	if !found {
		t.Errorf("aliases %v missing wb", info.Aliases)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		in      string
		wantTyp string
		wantID  string
		wantErr bool
	}{
		{"US", "country", "US", false},
		{"CN", "country", "CN", false},
		{"NY.GDP.MKTP.CD", "indicator", "NY.GDP.MKTP.CD", false},
		{"SP.POP.TOTL", "indicator", "SP.POP.TOTL", false},
		{"3", "topic", "3", false},
		{"12", "topic", "12", false},
		{"https://data.worldbank.org/country/US", "country", "US", false},
		{"https://data.worldbank.org/indicator/NY.GDP.MKTP.CD", "indicator", "NY.GDP.MKTP.CD", false},
		{"", "", "", true},
	}
	d := Domain{}
	for _, tc := range cases {
		typ, id, err := d.Classify(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("Classify(%q): want error, got nil", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("Classify(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if typ != tc.wantTyp || id != tc.wantID {
			t.Errorf("Classify(%q) = (%q, %q), want (%q, %q)", tc.in, typ, id, tc.wantTyp, tc.wantID)
		}
	}
}

func TestLocate(t *testing.T) {
	d := Domain{}

	got, err := d.Locate("country", "US")
	if err != nil || got != "https://data.worldbank.org/country/US" {
		t.Errorf("Locate(country, US) = (%q, %v)", got, err)
	}

	got, err = d.Locate("indicator", "NY.GDP.MKTP.CD")
	if err != nil || got != "https://data.worldbank.org/indicator/NY.GDP.MKTP.CD" {
		t.Errorf("Locate(indicator, NY.GDP.MKTP.CD) = (%q, %v)", got, err)
	}

	got, err = d.Locate("topic", "3")
	if err != nil || got != "https://data.worldbank.org/topic/3" {
		t.Errorf("Locate(topic, 3) = (%q, %v)", got, err)
	}

	_, err = d.Locate("unknown", "foo")
	if err == nil {
		t.Error("Locate(unknown, foo): want error, got nil")
	}
}

func TestHostWiring(t *testing.T) {
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}
	_, ok := h.Domain("worldbank")
	if !ok {
		t.Fatal("worldbank not mounted on host")
	}
}

func TestIsNumeric(t *testing.T) {
	cases := []struct {
		s    string
		want bool
	}{
		{"3", true},
		{"12", true},
		{"0", true},
		{"", false},
		{"a", false},
		{"3a", false},
	}
	for _, tc := range cases {
		got := isNumeric(tc.s)
		if got != tc.want {
			t.Errorf("isNumeric(%q) = %v, want %v", tc.s, got, tc.want)
		}
	}
}

func TestIsUpperAlpha(t *testing.T) {
	cases := []struct {
		s    string
		want bool
	}{
		{"US", true},
		{"CN", true},
		{"us", false},
		{"U1", false},
		{"", false},
	}
	for _, tc := range cases {
		got := isUpperAlpha(tc.s)
		if got != tc.want {
			t.Errorf("isUpperAlpha(%q) = %v, want %v", tc.s, got, tc.want)
		}
	}
}
