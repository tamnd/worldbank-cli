package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) countryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "country <code>",
		Short: "Show details for a single country",
		Long: `Show details for a single country by its ISO2 or ISO3 code.

Examples of codes: US, CN, BR, IN, GB, DE, FR, JP, AU, CA`,
		Example: `  worldbank country US
  worldbank country CN
  worldbank country BRA`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			country, err := a.client.Country(cmd.Context(), args[0])
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render(country)
		},
	}
	return cmd
}
