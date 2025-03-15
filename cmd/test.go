package cmd

import (
	"log"
	"log/slog"

	"github.com/lehigh-university-libraries/encode/pkg/config"
	"github.com/spf13/cobra"
)

var composeCmd = &cobra.Command{
	Use:   "test",
	Short: "test config",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.LoadConfig("encode.yaml")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		slog.Debug("Got config", "config", c)
		//		_ = config.InitializeConnections(c)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(composeCmd)
}
