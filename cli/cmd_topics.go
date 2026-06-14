package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) topicsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topics",
		Short: "List World Bank thematic topics",
		Long: `List all World Bank thematic topics.

Topics organize the World Bank's data into themes such as Agriculture,
Economy & Growth, Education, Energy & Mining, and more.`,
		Example: `  worldbank topics`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			topics, err := a.client.ListTopics(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(topics, len(topics))
		},
	}
	return cmd
}
