package connection_test

import (
	"errors"
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
		expectedResults []map[string]string
	}{
		{
			name: "Valid Query",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, name FROM users").
					WillReturnRows(pgxmock.NewRows([]string{"id", "name"}).
						AddRow("1", "Alice").
						AddRow("2", "Bob"))
			},
			params: map[string]string{"query": "SELECT id, name FROM users"},
			expectedResults: []map[string]string{
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

			if len(tt.expectedResults) != len(results) {
				t.Fatalf("Mismatched lengths: expected %d, got %d", len(tt.expectedResults), len(results))
			}
			if !reflect.DeepEqual(tt.expectedResults, results) {
				t.Errorf("expected %v got %v", tt.expectedResults, results)
			}

		})
	}
}
