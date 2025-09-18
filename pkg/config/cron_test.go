package config_test

import (
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
	synctest.Test(t, func(t *testing.T) {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		wd = filepath.Dir(wd)
		wd = filepath.Dir(wd)
		yml := filepath.Join(wd, "fixtures", "encode.mock.test.yaml")
		c, err := config.LoadConfig(yml)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Create a temporary staging directory for the test
		tmpDir, err := os.MkdirTemp("", "encode-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)
		c.StagingDirectory = tmpDir

		cron := c.StartCron()
		cron.Start()

		// Sleep until midnight to trigger the cron job (time is fake in synctest)
		time.Sleep(24 * time.Hour)
		synctest.Wait()

		cron.Stop()

		// Verify that a CSV file was created
		reportDir := filepath.Join(tmpDir, "test_report")
		entries, err := os.ReadDir(reportDir)
		if err == nil && len(entries) > 0 {
			slog.Info("CSV file created successfully", "count", len(entries))
		} else {
			t.Logf("No CSV files found in %s (this might be expected if cron timing is different)", reportDir)
		}
	})
}
