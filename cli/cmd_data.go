package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) dataCmd() *cobra.Command {
	var indicator string
	var mrv int

	cmd := &cobra.Command{
		Use:   "data <country>",
		Short: "Fetch indicator data for a country",
		Long: `Fetch time-series data for one or more countries and an indicator.

country is an ISO2 or ISO3 code (e.g. US, CN, BRA). For multiple countries,
separate them with semicolons: US;CN;DE.

Use --indicator to pick the indicator (default: NY.GDP.MKTP.CD for GDP).
Use --mrv to control how many most-recent values to return (default: 5).

Use the 'indicators' command to browse available indicator IDs.`,
		Example: `  worldbank data US
  worldbank data US --indicator SP.POP.TOTL
  worldbank data US;CN;DE --indicator NY.GDP.MKTP.CD --mrv 3`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if mrv <= 0 {
				mrv = 5
			}
			if indicator == "" {
				indicator = "NY.GDP.MKTP.CD"
			}
			points, err := a.client.GetData(cmd.Context(), args[0], indicator, mrv)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(points, len(points))
		},
	}
	cmd.Flags().StringVar(&indicator, "indicator", "NY.GDP.MKTP.CD", "indicator ID")
	cmd.Flags().IntVar(&mrv, "mrv", 5, "most recent N values")
	return cmd
}
