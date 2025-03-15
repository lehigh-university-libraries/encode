package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/lehigh-university-libraries/encode/pkg/connection"
	cron "github.com/robfig/cron/v3"
	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Connections []map[string]any `yaml:"connections"`
	Reports     []ReportConfig   `yaml:"reports"`
}

type ReportConfig struct {
	Name         string            `yaml:"name"`
	Connection   string            `yaml:"connection"`
	QueryParams  map[string]string `yaml:"query_params"`
	TemplatePath string            `yaml:"template"`
	Schedule     string            `yaml:"schedule"`
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

	// Validate cron expressions
	for _, report := range config.Reports {
		slog.Debug("Ensuring cron entry is valid", "schedule", report.Schedule)
		if report.Schedule == "" {
			return nil, fmt.Errorf("cron schedule not provided in report '%s'", report.Name)
		}
		_, err := cron.ParseStandard(report.Schedule)
		if err != nil {
			return nil, fmt.Errorf("invalid cron schedule '%s' in report '%s': %v", report.Schedule, report.Name, err)
		}
	}

	return &config, err
}

func InitializeConnections(config *Config) map[string]connection.ConnectionProvider {
	connections := make(map[string]connection.ConnectionProvider)

	for _, connData := range config.Connections {
		name := connData["name"].(string)
		connType := connData["type"].(string)

		var conn connection.ConnectionProvider
		switch connType {
		case "GoogleSheets":
			conn = &connection.GoogleSheetsAuth{CredentialsFile: connData["credentials_file"].(string)}
		case "PostgreSQL":
			conn = &connection.PostgresAuth{DSN: connData["dsn"].(string)}
		default:
			log.Fatalf("Unknown connection type: %s", connType)
		}

		if err := conn.Authenticate(); err != nil {
			log.Fatalf("Failed to authenticate %s: %v", name, err)
		}
		connections[name] = conn
	}
	return connections
}
