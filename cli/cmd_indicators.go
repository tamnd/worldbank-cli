package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) indicatorsCmd() *cobra.Command {
	var search string
	var page int

	cmd := &cobra.Command{
		Use:   "indicators",
		Short: "List or search World Bank indicators",
		Long: `List World Bank indicators. Use --search to filter by name or ID substring.

Indicators have IDs like NY.GDP.MKTP.CD (GDP in current US$) and cover topics
such as economy, health, education, environment, and infrastructure.`,
		Example: `  worldbank indicators
  worldbank indicators --search gdp
  worldbank indicators --search population --page 1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			limit := a.effectiveLimit(50)
			indicators, err := a.client.Indicators(cmd.Context(), search, page, limit)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(indicators, len(indicators))
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "filter by name or ID substring")
	cmd.Flags().IntVar(&page, "page", 1, "page number")
	return cmd
}
