package connection_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/lehigh-university-libraries/encode/pkg/connection"
)

func TestFolioAuth_Authenticate(t *testing.T) {
	tests := []struct {
		name           string
		auth           *connection.FolioAuth
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    bool
		expectedToken  string
	}{
		{
			name: "Successful Authentication",
			auth: &connection.FolioAuth{
				Tenant:   "lu",
				Username: "testuser",
				Password: "testpass",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/authn/login" {
					t.Errorf("Expected path /authn/login, got %s", r.URL.Path)
				}
				if r.Header.Get("x-okapi-tenant") != "lu" {
					t.Errorf("Expected tenant header 'lu', got %s", r.Header.Get("x-okapi-tenant"))
				}
				w.Header().Set("x-okapi-token", "test-token-123")
				w.WriteHeader(http.StatusCreated)
			},
			expectError:   false,
			expectedToken: "test-token-123",
		},
		{
			name: "Missing BaseURL",
			auth: &connection.FolioAuth{
				Tenant:   "lu",
				Username: "testuser",
				Password: "testpass",
			},
			serverResponse: nil,
			expectError:    true,
		},
		{
			name: "Missing Tenant",
			auth: &connection.FolioAuth{
				BaseURL:  "http://test",
				Username: "testuser",
				Password: "testpass",
			},
			serverResponse: nil,
			expectError:    true,
		},
		{
			name: "Authentication Failed",
			auth: &connection.FolioAuth{
				Tenant:   "lu",
				Username: "wronguser",
				Password: "wrongpass",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("Invalid credentials"))
			},
			expectError: true,
		},
		{
			name: "No Token in Response",
			auth: &connection.FolioAuth{
				Tenant:   "lu",
				Username: "testuser",
				Password: "testpass",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.serverResponse != nil {
				server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
				defer server.Close()
				tt.auth.BaseURL = server.URL
			}

			err := tt.auth.Authenticate()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Authenticate() failed: %v", err)
			}
			if !tt.expectError && tt.auth.Token != tt.expectedToken {
				t.Errorf("Expected token %s, got %s", tt.expectedToken, tt.auth.Token)
			}
		})
	}
}

func TestFolioAuth_FetchReport(t *testing.T) {
	tests := []struct {
		name            string
		auth            *connection.FolioAuth
		params          map[string]string
		serverResponse  func(w http.ResponseWriter, r *http.Request)
		expectError     bool
		expectedResults []map[string]string
	}{
		{
			name: "Successful Report Fetch",
			auth: &connection.FolioAuth{
				Token: "test-token-123",
			},
			params: map[string]string{
				"query_url": "https://example.com/query.sql",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/authn/login" {
					w.Header().Set("x-okapi-token", "test-token-123")
					w.WriteHeader(http.StatusCreated)
					return
				}

				if r.URL.Path != "/ldp/db/reports" {
					t.Errorf("Expected path /ldp/db/reports, got %s", r.URL.Path)
				}
				if r.Header.Get("x-okapi-token") != "test-token-123" {
					t.Errorf("Expected token header, got %s", r.Header.Get("x-okapi-token"))
				}

				body, _ := io.ReadAll(r.Body)
				expectedBody := `{"url":"https://example.com/query.sql"}`
				if string(body) != expectedBody {
					t.Errorf("Expected body %s, got %s", expectedBody, string(body))
				}

				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("id,name,amount\n1,Alice,100\n2,Bob,200\n"))
			},
			expectError: false,
			expectedResults: []map[string]string{
				{"id": "1", "name": "Alice", "amount": "100"},
				{"id": "2", "name": "Bob", "amount": "200"},
			},
		},
		{
			name: "Missing Query URL Parameter",
			auth: &connection.FolioAuth{
				Token: "test-token-123",
			},
			params:      map[string]string{},
			expectError: true,
		},
		{
			name: "Empty CSV Response",
			auth: &connection.FolioAuth{
				Token: "test-token-123",
			},
			params: map[string]string{
				"query_url": "https://example.com/query.sql",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(""))
			},
			expectError:     false,
			expectedResults: []map[string]string{},
		},
		{
			name: "Report Request Failed",
			auth: &connection.FolioAuth{
				Token: "test-token-123",
			},
			params: map[string]string{
				"query_url": "https://example.com/query.sql",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Internal server error"))
			},
			expectError: true,
		},
		{
			name: "Authentication Required Before Fetch",
			auth: &connection.FolioAuth{
				Tenant:   "lu",
				Username: "testuser",
				Password: "testpass",
			},
			params: map[string]string{
				"query_url": "https://example.com/query.sql",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/authn/login" {
					w.Header().Set("x-okapi-token", "new-token-456")
					w.WriteHeader(http.StatusCreated)
					return
				}

				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("id,value\n1,test\n"))
			},
			expectError: false,
			expectedResults: []map[string]string{
				{"id": "1", "value": "test"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.serverResponse != nil {
				server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
				defer server.Close()
				tt.auth.BaseURL = server.URL
				tt.auth.Tenant = "lu"
			}

			results, err := tt.auth.FetchReport(tt.params)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("FetchReport() failed: %v", err)
			}

			if tt.expectedResults == nil {
				return
			}

			if len(results) != len(tt.expectedResults) {
				t.Fatalf("Mismatched lengths: expected %d, got %d", len(tt.expectedResults), len(results))
			}

			if !reflect.DeepEqual(results, tt.expectedResults) {
				t.Errorf("Expected %v, got %v", tt.expectedResults, results)
			}
		})
	}
}
