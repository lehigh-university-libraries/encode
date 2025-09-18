package cmd

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/lehigh-university-libraries/encode/pkg/config"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run all reports on their schedule",
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		c, err := config.LoadConfig(f)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		slog.Debug("Got config", "config", c)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	config := os.Getenv("ENCODE_CONFIG_YAML")
	if config == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			slog.Error("Unable to detect home directory", "err", err)
			h = "/tmp"
		}
		config = filepath.Join(h, "encode.yaml")
	}
	runCmd.Flags().String("config", config, "Path to encode.yaml")
}
