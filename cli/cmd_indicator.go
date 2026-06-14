package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tamnd/worldbank-cli/worldbank"
)

// IndicatorPoint is the display record for the indicator and compare commands.
// Value is pre-formatted as "%.2f" or "N/A" when null.
type IndicatorPoint struct {
	Country   string `json:"country"`
	Indicator string `json:"indicator"`
	Year      string `json:"year"`
	Value     string `json:"value"`
}

func toIndicatorPoint(p worldbank.DataPoint) IndicatorPoint {
	var val string
	if p.Value == 0 && p.NullValue {
		val = "N/A"
	} else {
		val = fmt.Sprintf("%.2f", p.Value)
	}
	return IndicatorPoint{
		Country:   p.CountryName,
		Indicator: p.IndicatorID,
		Year:      p.Date,
		Value:     val,
	}
}

func (a *App) indicatorCmd() *cobra.Command {
	var years int

	cmd := &cobra.Command{
		Use:   "indicator <country> <indicator>",
		Short: "Get indicator data for a country",
		Long: `Get time-series indicator data for a country.

country is an ISO2 or ISO3 code (e.g. US, CN, BRA).
indicator is a World Bank indicator ID (e.g. NY.GDP.MKTP.CD for GDP).

Use --years to control how many most-recent years to return (default: 5).

Use the 'indicators' command to browse available indicator IDs.`,
		Example: `  worldbank indicator US NY.GDP.MKTP.CD
  worldbank indicator CN SP.POP.TOTL --years 3
  worldbank indicator BR NY.GDP.MKTP.CD --years 10 -o json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			country := args[0]
			indicatorID := args[1]
			if years <= 0 {
				years = 5
			}
			points, err := a.client.GetData(cmd.Context(), country, indicatorID, years)
			if err != nil {
				return mapFetchErr(err)
			}
			out := make([]IndicatorPoint, 0, len(points))
			for _, p := range points {
				out = append(out, toIndicatorPoint(p))
			}
			return a.renderOrEmpty(out, len(out))
		},
	}
	cmd.Flags().IntVar(&years, "years", 5, "number of most recent years to return")
	return cmd
}

func (a *App) compareCmd() *cobra.Command {
	var countriesFlag string
	var year string

	cmd := &cobra.Command{
		Use:   "compare <indicator>",
		Short: "Compare an indicator across multiple countries",
		Long: `Compare a World Bank indicator across multiple countries for the latest year.

indicator is a World Bank indicator ID (e.g. NY.GDP.MKTP.CD for GDP).

Use --countries to specify a comma-separated list of ISO2 country codes.
Use --year to pin a specific year; omit for the latest available data.

Results are sorted by value descending.`,
		Example: `  worldbank compare NY.GDP.MKTP.CD
  worldbank compare SP.POP.TOTL --countries US,CN,IN,BR
  worldbank compare NY.GDP.MKTP.CD --countries US,CN,JP,DE,IN --year 2022`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			indicatorID := args[0]
			codes := strings.Split(countriesFlag, ",")
			for i, c := range codes {
				codes[i] = strings.TrimSpace(c)
			}
			// Build semicolon-joined list for the multi-country API call.
			joined := strings.Join(codes, ";")
			mrv := 1
			points, err := a.client.GetData(cmd.Context(), joined, indicatorID, mrv)
			if err != nil {
				return mapFetchErr(err)
			}
			// Filter by year if specified.
			if year != "" {
				filtered := points[:0]
				for _, p := range points {
					if p.Date == year {
						filtered = append(filtered, p)
					}
				}
				points = filtered
			}
			// Sort by value descending.
			sort.Slice(points, func(i, j int) bool {
				return points[i].Value > points[j].Value
			})
			out := make([]IndicatorPoint, 0, len(points))
			for _, p := range points {
				out = append(out, toIndicatorPoint(p))
			}
			return a.renderOrEmpty(out, len(out))
		},
	}
	cmd.Flags().StringVar(&countriesFlag, "countries", "US,CN,JP,DE,IN", "comma-separated ISO2 country codes")
	cmd.Flags().StringVar(&year, "year", "", "specific year (default: latest available)")
	return cmd
}
