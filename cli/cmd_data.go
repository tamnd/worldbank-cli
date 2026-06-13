package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) dataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data <country> <indicator>",
		Short: "Fetch indicator data for a country",
		Long: `Fetch time-series data for a country and indicator combination.

country is an ISO2 or ISO3 code (e.g. US, CN, BRA).
indicator is a World Bank indicator ID (e.g. NY.GDP.MKTP.CD).

Use the 'indicators' command to browse available indicators.`,
		Example: `  worldbank data US NY.GDP.MKTP.CD
  worldbank data CN SP.POP.TOTL
  worldbank data IN SH.DYN.MORT`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			limit := a.effectiveLimit(10)
			points, err := a.client.Data(cmd.Context(), args[0], args[1], limit)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(points, len(points))
		},
	}
	return cmd
}
