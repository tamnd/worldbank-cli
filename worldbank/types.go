package worldbank

// Country is a World Bank country record.
type Country struct {
	ID          string `kit:"id" json:"id"`
	ISO2        string `json:"iso2"`
	Name        string `json:"name"`
	Region      string `json:"region"`
	IncomeLevel string `json:"income_level"`
	CapitalCity string `json:"capital_city"`
	Longitude   string `json:"longitude"`
	Latitude    string `json:"latitude"`
}

// Indicator is a World Bank indicator record.
type Indicator struct {
	ID   string `kit:"id" json:"id"`
	Name string `json:"name"`
	Unit string `json:"unit"`
	Note string `json:"source_note"`
}

// DataPoint is a single observation for a country/indicator combination.
// NullValue is true when the API returned JSON null for the value field.
type DataPoint struct {
	CountryID   string  `kit:"id" json:"country_id"`
	CountryName string  `json:"country_name"`
	IndicatorID string  `json:"indicator_id"`
	Date        string  `json:"date"`
	Value       float64 `json:"value"`
	NullValue   bool    `json:"-"`
}

// Topic is a World Bank thematic topic.
type Topic struct {
	ID    string `kit:"id" json:"id"`
	Value string `json:"value"`
	Note  string `json:"source_note"`
}

// --- wire types (JSON shapes from the API) ---

type wireCountry struct {
	ID      string `json:"id"`
	Iso2Code string `json:"iso2Code"`
	Name    string `json:"name"`
	Region  struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	} `json:"region"`
	IncomeLevel struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	} `json:"incomeLevel"`
	CapitalCity string `json:"capitalCity"`
	Longitude   string `json:"longitude"`
	Latitude    string `json:"latitude"`
}

func (w wireCountry) toCountry() Country {
	return Country{
		ID:          w.ID,
		ISO2:        w.Iso2Code,
		Name:        w.Name,
		Region:      w.Region.Value,
		IncomeLevel: w.IncomeLevel.Value,
		CapitalCity: w.CapitalCity,
		Longitude:   w.Longitude,
		Latitude:    w.Latitude,
	}
}

type wireIndicator struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Unit       string `json:"unit"`
	SourceNote string `json:"sourceNote"`
	Source     struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	} `json:"source"`
}

func (w wireIndicator) toIndicator() Indicator {
	return Indicator{
		ID:   w.ID,
		Name: w.Name,
		Unit: w.Unit,
		Note: w.SourceNote,
	}
}

type wireDataPoint struct {
	Country struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	} `json:"country"`
	Indicator struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	} `json:"indicator"`
	CountryISO3 string   `json:"countryiso3code"`
	Date        string   `json:"date"`
	Value       *float64 `json:"value"`
}

func (w wireDataPoint) toDataPoint() DataPoint {
	val := 0.0
	nullVal := w.Value == nil
	if w.Value != nil {
		val = *w.Value
	}
	return DataPoint{
		CountryID:   w.Country.ID,
		CountryName: w.Country.Value,
		IndicatorID: w.Indicator.ID,
		Date:        w.Date,
		Value:       val,
		NullValue:   nullVal,
	}
}

type wireTopic struct {
	ID         string `json:"id"`
	Value      string `json:"value"`
	SourceNote string `json:"sourceNote"`
}

func (w wireTopic) toTopic() Topic {
	return Topic{
		ID:    w.ID,
		Value: w.Value,
		Note:  w.SourceNote,
	}
}
