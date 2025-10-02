package config

import (
	"fmt"
	"log/slog"

	"github.com/lehigh-university-libraries/encode/pkg/connection"
)

func InitializeConnections(config *Config) map[string]connection.ConnectionProvider {
	connections := make(map[string]connection.ConnectionProvider)

	for _, connData := range config.Connections {
		conn, err := InitializeConnection(connData)
		if err != nil {
			slog.Error("Unable to establish connection", "err", err)
			continue
		}

		name := connData["name"].(string)
		connections[name] = conn
	}

	return connections
}

func InitializeConnection(connData map[string]any) (connection.ConnectionProvider, error) {
	connType := connData["type"].(string)
	switch connType {
	case "GoogleSheets":
		return &connection.GoogleSheetsAuth{
			CredentialsFile: connData["credentials_file"].(string),
		}, nil
	case "PostgreSQL":
		return &connection.PostgresAuth{
			DSN: connData["dsn"].(string),
		}, nil
	case "MariaDB":
		return &connection.MariaDBAuth{
			DSN: connData["dsn"].(string),
		}, nil
	case "Mock":
		name := ""
		if n, ok := connData["name"].(string); ok {
			name = n
		}
		return &connection.MockConnection{
			Name: name,
		}, nil
	default:
		return nil, fmt.Errorf("unknown connection type: %s", connType)
	}
}
