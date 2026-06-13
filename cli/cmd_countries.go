package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) countriesCmd() *cobra.Command {
	var region, income string
	var page int

	cmd := &cobra.Command{
		Use:   "countries",
		Short: "List World Bank countries",
		Long: `List countries from the World Bank Open Data API.

Optionally filter by region code (e.g. LCN for Latin America) or income level
code (e.g. HIC for high income). Use --page to paginate through results.`,
		Example: `  worldbank countries
  worldbank countries --region LCN
  worldbank countries --income HIC
  worldbank countries --page 2`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			limit := a.effectiveLimit(50)
			countries, err := a.client.Countries(cmd.Context(), region, income, page, limit)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(countries, len(countries))
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "filter by region code (e.g. LCN, EAS, NAC)")
	cmd.Flags().StringVar(&income, "income", "", "filter by income level code (e.g. HIC, LIC, MIC)")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}
