package connection_test

import (
	"reflect"
	"testing"

	"github.com/lehigh-university-libraries/encode/pkg/connection"
)

func TestParseSheetData_SingleSheet(t *testing.T) {
	tests := []struct {
		name            string
		sheets          []connection.SheetData
		headerRow       int
		expectedResults []map[string]string
		expectError     bool
	}{
		{
			name: "Single sheet with data",
			sheets: []connection.SheetData{
				{
					Name: "Classes",
					Header: []interface{}{
						"Date", "Class Type", "Count",
					},
					Rows: [][]interface{}{
						{"2024-01-01", "Undergraduate", "25"},
						{"2024-01-02", "Graduate", "15"},
					},
				},
			},
			headerRow: 1,
			expectedResults: []map[string]string{
				{
					"Date":       "2024-01-01",
					"Class Type": "Undergraduate",
					"Count":      "25",
					"sheet":      "Classes",
				},
				{
					"Date":       "2024-01-02",
					"Class Type": "Graduate",
					"Count":      "15",
					"sheet":      "Classes",
				},
			},
			expectError: false,
		},
		{
			name: "Single sheet with empty first cell stops processing",
			sheets: []connection.SheetData{
				{
					Name: "Data",
					Header: []interface{}{
						"ID", "Name",
					},
					Rows: [][]interface{}{
						{"1", "Alice"},
						{"", "Bob"}, // Empty first cell - should stop here
						{"3", "Charlie"},
					},
				},
			},
			headerRow: 1,
			expectedResults: []map[string]string{
				{
					"ID":    "1",
					"Name":  "Alice",
					"sheet": "Data",
				},
			},
			expectError: false,
		},
		{
			name: "Single sheet with numeric values",
			sheets: []connection.SheetData{
				{
					Name: "Stats",
					Header: []interface{}{
						"Year", "Count",
					},
					Rows: [][]interface{}{
						{2024, 100}, // Numeric values
						{2023, 95},
					},
				},
			},
			headerRow: 1,
			expectedResults: []map[string]string{
				{
					"Year":  "2024",
					"Count": "100",
					"sheet": "Stats",
				},
				{
					"Year":  "2023",
					"Count": "95",
					"sheet": "Stats",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := connection.ParseSheetData(tt.sheets, tt.headerRow)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ParseSheetData() failed: %v", err)
			}

			if !tt.expectError && !reflect.DeepEqual(tt.expectedResults, results) {
				t.Errorf("Expected:\n%v\nGot:\n%v", tt.expectedResults, results)
			}
		})
	}
}

func TestParseSheetData_MultipleSheets(t *testing.T) {
	tests := []struct {
		name            string
		sheets          []connection.SheetData
		headerRow       int
		expectedResults []map[string]string
		expectError     bool
	}{
		{
			name: "Multiple sheets with same structure",
			sheets: []connection.SheetData{
				{
					Name: "Classes",
					Header: []interface{}{
						"Date", "Type", "Count",
					},
					Rows: [][]interface{}{
						{"2024-01-01", "Undergraduate", "25"},
						{"2024-01-02", "Graduate", "15"},
					},
				},
				{
					Name: "Tours",
					Rows: [][]interface{}{
						{"2024-01-03", "Campus Tour", "30"},
						{"2024-01-04", "Virtual Tour", "20"},
					},
				},
				{
					Name: "External",
					Rows: [][]interface{}{
						{"2024-01-05", "Conference", "50"},
					},
				},
			},
			headerRow: 2,
			expectedResults: []map[string]string{
				{"Date": "2024-01-01", "Type": "Undergraduate", "Count": "25", "sheet": "Classes"},
				{"Date": "2024-01-02", "Type": "Graduate", "Count": "15", "sheet": "Classes"},
				{"Date": "2024-01-03", "Type": "Campus Tour", "Count": "30", "sheet": "Tours"},
				{"Date": "2024-01-04", "Type": "Virtual Tour", "Count": "20", "sheet": "Tours"},
				{"Date": "2024-01-05", "Type": "Conference", "Count": "50", "sheet": "External"},
			},
			expectError: false,
		},
		{
			name: "Multiple sheets with varying row counts",
			sheets: []connection.SheetData{
				{
					Name: "Sheet1",
					Header: []interface{}{
						"Col1", "Col2",
					},
					Rows: [][]interface{}{
						{"A1", "B1"},
					},
				},
				{
					Name: "Sheet2",
					Rows: [][]interface{}{
						{"A2", "B2"},
						{"A3", "B3"},
						{"A4", "B4"},
					},
				},
			},
			headerRow: 1,
			expectedResults: []map[string]string{
				{"Col1": "A1", "Col2": "B1", "sheet": "Sheet1"},
				{"Col1": "A2", "Col2": "B2", "sheet": "Sheet2"},
				{"Col1": "A3", "Col2": "B3", "sheet": "Sheet2"},
				{"Col1": "A4", "Col2": "B4", "sheet": "Sheet2"},
			},
			expectError: false,
		},
		{
			name: "Multiple sheets with missing values",
			sheets: []connection.SheetData{
				{
					Name: "Sheet1",
					Header: []interface{}{
						"A", "B", "C",
					},
					Rows: [][]interface{}{
						{"1", "2"},      // Missing C
						{"3", "4", "5"}, // Full row
					},
				},
				{
					Name: "Sheet2",
					Rows: [][]interface{}{
						{"6"},           // Missing B and C
						{"7", "8", "9"}, // Full row
					},
				},
			},
			headerRow: 1,
			expectedResults: []map[string]string{
				{"A": "1", "B": "2", "C": "", "sheet": "Sheet1"},
				{"A": "3", "B": "4", "C": "5", "sheet": "Sheet1"},
				{"A": "6", "B": "", "C": "", "sheet": "Sheet2"},
				{"A": "7", "B": "8", "C": "9", "sheet": "Sheet2"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := connection.ParseSheetData(tt.sheets, tt.headerRow)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ParseSheetData() failed: %v", err)
			}

			if !tt.expectError && !reflect.DeepEqual(tt.expectedResults, results) {
				t.Errorf("Expected:\n%v\nGot:\n%v", tt.expectedResults, results)
			}
		})
	}
}

func TestParseSheetData_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		sheets      []connection.SheetData
		headerRow   int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Empty sheets list",
			sheets:      []connection.SheetData{},
			headerRow:   1,
			expectError: true,
			errorMsg:    "no sheets provided",
		},
		{
			name: "Empty header row",
			sheets: []connection.SheetData{
				{
					Name:   "Sheet1",
					Header: []interface{}{},
					Rows:   [][]interface{}{{"data"}},
				},
			},
			headerRow:   1,
			expectError: true,
			errorMsg:    "header row is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := connection.ParseSheetData(tt.sheets, tt.headerRow)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ParseSheetData() failed: %v", err)
			}
		})
	}
}

func TestGoogleSheetsAuth_MissingParameters(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
	}{
		{
			name:   "Missing spreadsheet_id",
			params: map[string]string{"gid": "0"},
		},
		{
			name:   "Missing gid",
			params: map[string]string{"spreadsheet_id": "test-id"},
		},
		{
			name:   "Empty params",
			params: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &connection.GoogleSheetsAuth{
				Service: nil, // Will trigger authentication error first
			}

			_, err := auth.FetchReport(tt.params)
			if err == nil {
				t.Error("Expected error for missing parameters")
			}
		})
	}
}
