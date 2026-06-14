package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) indicatorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "indicators",
		Short: "List World Development Indicators",
		Long: `List indicators from the World Development Indicators (WDI) dataset.

Indicators have IDs like NY.GDP.MKTP.CD (GDP in current US$) and cover topics
such as economy, health, education, environment, and infrastructure.`,
		Example: `  worldbank indicators
  worldbank indicators --limit 50`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			limit := a.effectiveLimit(20)
			indicators, err := a.client.ListIndicators(cmd.Context(), limit)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(indicators, len(indicators))
		},
	}
	return cmd
}
