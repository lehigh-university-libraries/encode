package config_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	"github.com/lehigh-university-libraries/encode/pkg/config"
)

func TestRunCommand_WithAssertions(t *testing.T) {
	// see https://go.dev/blog/synctest#testing-time
	synctest.Run(func() {
		ctx := context.Background()

		const timeout = 25 * time.Hour

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// TODO: start postgres

		// Calculate the target time in seconds from midnight
		now := time.Now()
		targetTime := now.Add(24 * time.Hour)
		duration := targetTime.Sub(now)

		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		wd = filepath.Dir(wd)
		wd = filepath.Dir(wd)
		yml := filepath.Join(wd, "fixtures", "encode.test.yaml")
		c, err := config.LoadConfig(yml)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		cron := c.StartCron()
		cron.Start()
		slog.Info("Sleeping for" + duration.String())
		time.Sleep(duration)
		synctest.Wait()

		cron.Stop()

		if err := ctx.Err(); err != nil {
			t.Fatalf("before timeout, ctx.Err() = %v; want nil", err)
		}
		// Assert that no errors occurred
		//	assert.NoError(t, err, "Command should not have errors")

		// Add more assertions here to verify the config is loaded correctly,
		// or that other parts of the command are behaving as expected.
		// Example:
		// assert.Equal(t, "some value", config.C.SomeField, "SomeField should have a value")
	})
}
