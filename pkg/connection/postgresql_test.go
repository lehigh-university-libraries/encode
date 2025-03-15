package connection_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/lehigh-university-libraries/encode/pkg/connection"
	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresAuth_FetchReport(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	pgAuth := &connection.PostgresAuth{DB: mock}

	// Test cases
	tests := []struct {
		name            string
		setupMock       func()
		params          map[string]string
		expectError     bool
		expectedResults any
	}{
		{
			name: "Valid Query",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, name FROM users").
					WillReturnRows(pgxmock.NewRows([]string{"id", "name"}).
						AddRow(1, "Alice").
						AddRow(2, "Bob"))
			},
			params: map[string]string{"query": "SELECT id, name FROM users"},
			expectedResults: []map[string]any{
				{
					"id":   "1",
					"name": "Alice",
				},
				{
					"id":   "2",
					"name": "Bob",
				},
			},
		},
		{
			name:            "Missing Query Parameter",
			setupMock:       func() {},
			params:          map[string]string{},
			expectError:     true,
			expectedResults: nil,
		},
		{
			name: "Query Execution Error",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, name FROM users").
					WillReturnError(errors.New("database error"))
			},
			params:          map[string]string{"query": "SELECT id, name FROM users"},
			expectError:     true,
			expectedResults: nil,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Call FetchReport
			results, err := pgAuth.FetchReport(tt.params)

			// Check expected behavior
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("FetchReport() failed: %v", err)
			}

			// Verify expectations
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unmet mock expectations: %v", err)
			}

			// if we're not comparing results, bail
			if tt.expectedResults == nil {
				return
			}

			expectedResults, ok := tt.expectedResults.([]map[string]any)
			if !ok {
				t.Fatalf("Expected result type mismatch: got %T, want []map[string]any", tt.expectedResults)
			}

			actualResults, ok := results.([]map[string]any)
			if !ok {
				t.Fatalf("Actual result type mismatch: got %T, want []map[string]any", results)
			}

			if len(expectedResults) != len(actualResults) {
				t.Fatalf("Mismatched lengths: expected %d, got %d", len(expectedResults), len(actualResults))
			}

			// Compare each map in the slices
			for i, expectedMap := range expectedResults {
				actualMap := actualResults[i]
				expectedNormalized := normalizeMap(expectedMap)
				actualNormalized := normalizeMap(actualMap)
				if !reflect.DeepEqual(expectedNormalized, actualNormalized) {
					t.Errorf("Mismatch at index %d:\nExpected: %+v\nGot: %+v", i, expectedNormalized, actualNormalized)
				}

			}

		})
	}
}

func normalizeMap(m map[string]any) map[string]string {
	normalized := make(map[string]string)
	for k, v := range m {
		normalized[k] = fmt.Sprintf("%v", v)
	}
	return normalized
}
