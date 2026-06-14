package cli

import (
	"testing"

	"github.com/tamnd/worldbank-cli/worldbank"
)

func TestToIndicatorPoint_Normal(t *testing.T) {
	val := 28750956130731.2
	p := worldbank.DataPoint{
		CountryName: "United States",
		IndicatorID: "NY.GDP.MKTP.CD",
		Date:        "2024",
		Value:       val,
		NullValue:   false,
	}
	got := toIndicatorPoint(p)
	if got.Country != "United States" {
		t.Errorf("Country = %q, want United States", got.Country)
	}
	if got.Indicator != "NY.GDP.MKTP.CD" {
		t.Errorf("Indicator = %q, want NY.GDP.MKTP.CD", got.Indicator)
	}
	if got.Year != "2024" {
		t.Errorf("Year = %q, want 2024", got.Year)
	}
	// formatted with 2 decimal places
	want := "28750956130731.20"
	if got.Value != want {
		t.Errorf("Value = %q, want %q", got.Value, want)
	}
}

func TestToIndicatorPoint_Null(t *testing.T) {
	p := worldbank.DataPoint{
		CountryName: "Eritrea",
		IndicatorID: "NY.GDP.MKTP.CD",
		Date:        "2020",
		Value:       0,
		NullValue:   true,
	}
	got := toIndicatorPoint(p)
	if got.Value != "N/A" {
		t.Errorf("Value = %q, want N/A", got.Value)
	}
}

func TestToIndicatorPoint_Zero(t *testing.T) {
	// A real 0 value (NullValue=false) should be formatted, not "N/A".
	p := worldbank.DataPoint{
		CountryName: "Test",
		IndicatorID: "SH.XPD.CHEX.GD.ZS",
		Date:        "2021",
		Value:       0,
		NullValue:   false,
	}
	got := toIndicatorPoint(p)
	if got.Value != "0.00" {
		t.Errorf("Value = %q, want 0.00", got.Value)
	}
}
