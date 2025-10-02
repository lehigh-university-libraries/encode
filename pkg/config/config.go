package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/lehigh-university-libraries/encode/pkg/connection"
	"github.com/lehigh-university-libraries/encode/pkg/storage"
	cron "github.com/robfig/cron/v3"
	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Connections      []map[string]any `yaml:"connections"`
	Reports          []ReportConfig   `yaml:"reports"`
	StagingDirectory string           `yaml:"stagingDirectory"`
	S3               storage.S3Config `yaml:"s3"`
	s3Uploader       *storage.S3Uploader
}

type ReportConfig struct {
	Name             string            `yaml:"name"`
	Connection       string            `yaml:"connection"`
	QueryParams      map[string]string `yaml:"query_params"`
	TemplatePath     string            `yaml:"template"`
	Schedule         string            `yaml:"schedule"`
	StagingDirectory string
	connection       connection.ConnectionProvider
	s3Uploader       *storage.S3Uploader
}

func LoadConfig(filename string) (*Config, error) {
	slog.Debug("Loading config", "filename", filename)
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// allow referencing envionrment variables in YML
	expandedYaml := os.ExpandEnv(string(data))

	var config Config
	err = yaml.Unmarshal([]byte(expandedYaml), &config)

	// Initialize S3 uploader if enabled
	if config.S3.Enabled {
		config.s3Uploader, err = storage.NewS3Uploader(config.S3)
		if err != nil {
			slog.Error("Failed to initialize S3 uploader", "err", err)
			return nil, fmt.Errorf("failed to initialize S3 uploader: %w", err)
		}
		slog.Info("S3 uploader initialized", "bucket", config.S3.Bucket, "region", config.S3.Region)
	}

	// Validate cron expressions
	for k, report := range config.Reports {
		slog.Debug("Ensuring cron entry is valid", "schedule", report.Schedule)
		if report.Schedule == "" {
			return nil, fmt.Errorf("cron schedule not provided in report '%s'", report.Name)
		}
		_, err := cron.ParseStandard(report.Schedule)
		if err != nil {
			return nil, fmt.Errorf("invalid cron schedule '%s' in report '%s': %v", report.Schedule, report.Name, err)
		}
		var c connection.ConnectionProvider
		for _, conn := range config.Connections {
			if conn["name"].(string) == report.Connection {
				c, err = InitializeConnection(conn)
				if err != nil {
					slog.Error("Unable to fetch connection details", "conn", conn, "report", report.Name, "err", err)
					c = nil
				}
			}
		}
		if c == nil {
			return nil, fmt.Errorf("invalid connection reference '%s' in report '%s'", report.Connection, report.Name)
		}
		config.Reports[k].StagingDirectory = config.StagingDirectory
		config.Reports[k].connection = c
		config.Reports[k].s3Uploader = config.s3Uploader
	}

	return &config, err
}

// RunReportOnce executes a single report by name and returns immediately
func (c *Config) RunReportOnce(reportName string) error {
	for _, report := range c.Reports {
		if report.Name == reportName {
			slog.Info("Running report once", "report", reportName)
			report.Run()
			return nil
		}
	}
	return fmt.Errorf("report '%s' not found in configuration", reportName)
}
