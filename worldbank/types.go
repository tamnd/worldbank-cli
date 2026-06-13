package worldbank

// Country is a World Bank country record.
type Country struct {
	ID          string `json:"id"`
	ISO2Code    string `json:"iso2code"`
	Name        string `json:"name"`
	Region      string `json:"region"`
	IncomeLevel string `json:"income_level"`
	CapitalCity string `json:"capital_city"`
	Longitude   string `json:"longitude"`
	Latitude    string `json:"latitude"`
}

// Indicator is a World Bank indicator record.
type Indicator struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Unit   string `json:"unit"`
	Source string `json:"source"`
	Topics string `json:"topics"`
}

// DataPoint is a single observation for a country/indicator combination.
type DataPoint struct {
	Country     string  `json:"country"`
	IndicatorID string  `json:"indicator_id"`
	Year        string  `json:"year"`
	Value       float64 `json:"value"`
}

// --- wire types (JSON shapes from the API) ---

type wireMeta struct {
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

type wireCountry struct {
	ID       string `json:"id"`
	Iso2Code string `json:"iso2Code"`
	Name     string `json:"name"`
	Region   struct {
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
		ISO2Code:    w.Iso2Code,
		Name:        w.Name,
		Region:      w.Region.Value,
		IncomeLevel: w.IncomeLevel.Value,
		CapitalCity: w.CapitalCity,
		Longitude:   w.Longitude,
		Latitude:    w.Latitude,
	}
}

type wireIndicator struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Unit   string `json:"unit"`
	Source struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	} `json:"source"`
	Topics []struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	} `json:"topics"`
}

func (w wireIndicator) toIndicator() Indicator {
	topics := make([]string, 0, len(w.Topics))
	for _, t := range w.Topics {
		if t.Value != "" {
			topics = append(topics, t.Value)
		}
	}
	topicsStr := ""
	for i, t := range topics {
		if i > 0 {
			topicsStr += "; "
		}
		topicsStr += t
	}
	return Indicator{
		ID:     w.ID,
		Name:   w.Name,
		Unit:   w.Unit,
		Source: w.Source.Value,
		Topics: topicsStr,
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
	Date  string   `json:"date"`
	Value *float64 `json:"value"`
}

func (w wireDataPoint) toDataPoint() DataPoint {
	val := 0.0
	if w.Value != nil {
		val = *w.Value
	}
	return DataPoint{
		Country:     w.Country.Value,
		IndicatorID: w.Indicator.ID,
		Year:        w.Date,
		Value:       val,
	}
}
