package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) countriesCmd() *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "countries",
		Short: "List World Bank countries",
		Long: `List countries from the World Bank Open Data API.

Optionally filter by region code (e.g. LCN for Latin America, EAS for East Asia,
NAC for North America, SSF for Sub-Saharan Africa).`,
		Example: `  worldbank countries
  worldbank countries --region LCN
  worldbank countries --region EAS --limit 10`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			limit := a.effectiveLimit(20)
			countries, err := a.client.ListCountries(cmd.Context(), region, limit)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(countries, len(countries))
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "filter by region code (e.g. LCN, EAS, NAC, SSF)")
	return cmd
}
