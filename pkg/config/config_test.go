package config_test

import (
	"os"
	"testing"

	"github.com/lehigh-university-libraries/encode/pkg/config"
)

// Helper function to create a temporary YAML file
func createTempYAML(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp YAML file: %v", err)
	}
	_, err = tmpFile.Write([]byte(content))
	if err != nil {
		t.Fatalf("Failed to write to temp YAML file: %v", err)
	}
	tmpFile.Close()
	return tmpFile.Name()
}

// Runs multiple test cases in a loop
func TestLoadConfig(t *testing.T) {
	// Set environment variable for testing
	os.Setenv("SPREADSHEET_ID", "my-spreadsheet-id")
	defer os.Unsetenv("SPREADSHEET_ID")

	// Define test cases
	tests := []struct {
		name         string
		yamlContent  string
		expectError  bool
		validateFunc func(*testing.T, *config.Config) // Custom validation per test case
	}{
		{
			name: "Valid YAML",
			yamlContent: `
connections:
  - name: google_sheets
    type: GoogleSheets
    credentials_file: "path/to/google-service-account.json"

reports:
  - name: Monthly Sales Report
    connection: google_sheets
    query_params:
      spreadsheet_id: "spreadsheet-id"
      range: "Sheet1!A1:C3"
    schedule: "0 0 1 * *"
    template: "templates/sales.tmpl"
`,
			expectError: false,
			validateFunc: func(t *testing.T, cfg *config.Config) {
				if len(cfg.Reports) != 1 {
					t.Errorf("Expected 1 report, got %d", len(cfg.Reports))
				}
				expectedCron := "0 0 1 * *"
				if cfg.Reports[0].Schedule != expectedCron {
					t.Errorf("Expected schedule %q, got %q", expectedCron, cfg.Reports[0].Schedule)
				}
			},
		},
		{
			name: "Invalid Cron Expression",
			yamlContent: `
reports:
  - name: Invalid Report
    connection: google_sheets
    query_params:
      spreadsheet_id: "spreadsheet-id"
      range: "Sheet1!A1:C3"
    schedule: "invalid cron"
    template: "templates/sales.tmpl"
`,
			expectError: true,
		},
		{
			name:        "Missing File",
			yamlContent: "", // File won't exist
			expectError: true,
		},
		{
			name: "Environment Variable Expansion",
			yamlContent: `
reports:
  - name: Env Report
    connection: google_sheets
    query_params:
      spreadsheet_id: "${SPREADSHEET_ID}"
      range: "Sheet1!A1:C3"
    schedule: "0 12 * * *"
    template: "templates/env_report.tmpl"
`,
			expectError: false,
			validateFunc: func(t *testing.T, cfg *config.Config) {
				expectedSpreadsheetID := "my-spreadsheet-id"
				if cfg.Reports[0].QueryParams["spreadsheet_id"] != expectedSpreadsheetID {
					t.Errorf("Expected spreadsheet_id %q, got %q", expectedSpreadsheetID, cfg.Reports[0].QueryParams["spreadsheet_id"])
				}
			},
		},
		{
			name: "Malformed YAML",
			yamlContent: `
reports:
  - name: Bad Report
    connection: google_sheets
    query_params:
      spreadsheet_id: "spreadsheet-id"
      range: "Sheet1!A1:C3"
    schedule: "0 12 * * *
    template: "templates/bad.tmpl"
`, // Syntax error: missing closing quote
			expectError: true,
		},
		{
			name:        "Empty YAML File",
			yamlContent: "",
			expectError: false,
			validateFunc: func(t *testing.T, cfg *config.Config) {
				if len(cfg.Reports) != 0 {
					t.Errorf("Expected 0 reports, got %d", len(cfg.Reports))
				}
			},
		},
	}

	// Iterate over test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filename string
			var err error

			// Create a temp file for YAML content (except for missing file case)
			if tt.name != "Missing File" {
				filename = createTempYAML(t, tt.yamlContent)
				defer os.Remove(filename)
			} else {
				filename = "nonexistent.yaml" // Simulate a missing file
			}

			cfg, err := config.LoadConfig(filename)

			// Check for expected errors
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			// Ensure no errors when not expected
			if err != nil {
				t.Fatalf("LoadConfig() failed unexpectedly: %v", err)
			}

			// Run custom validation function if provided
			if tt.validateFunc != nil {
				tt.validateFunc(t, cfg)
			}
		})
	}
}
