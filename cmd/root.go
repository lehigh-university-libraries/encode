package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "encode",
	Short: "Fetch reports from various sources",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level := slog.LevelInfo
		ll, err := cmd.Flags().GetString("log-level")
		if err != nil {
			return err
		}

		switch strings.ToUpper(ll) {
		case "DEBUG":
			level = slog.LevelDebug
		case "WARN":
			level = slog.LevelWarn
		case "ERROR":
			level = slog.LevelError
		}

		opts := &slog.HandlerOptions{
			Level: level,
		}
		handler := slog.New(slog.NewTextHandler(os.Stdout, opts))
		slog.SetDefault(handler)

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)
}

func init() {
	ll := os.Getenv("LOG_LEVEL")
	if ll == "" {
		ll = "INFO"
	}
	rootCmd.PersistentFlags().String("log-level", ll, "The logging level for the command")
}
