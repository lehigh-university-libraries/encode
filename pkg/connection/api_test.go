package connection_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lehigh-university-libraries/encode/pkg/connection"
)

// Mock server that checks for API key in GET parameters
func mockServerGetParam(apiKey string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("api_key")
		if key != apiKey {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		resp := map[string]string{
			"status":  "success",
			"message": "API key valid",
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			slog.Error("Unknown error encoding json", "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
	}))
}

// Mock server that checks for API key in HTTP header
func mockServerHeader(apiKey string, headerName string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get(headerName)
		if key != apiKey {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		resp := map[string]string{
			"status":  "success",
			"message": "API key valid",
		}
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			slog.Error("Unknown error encoding json", "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
	}))
}

func TestAPIKeyAuth_FetchReport(t *testing.T) {
	apiKey := "my-secret-key"

	// Start mock servers
	serverGetParam := mockServerGetParam(apiKey)
	defer serverGetParam.Close()

	serverHeader := mockServerHeader(apiKey, "Authorization")
	defer serverHeader.Close()

	tests := []struct {
		name        string
		auth        connection.APIKeyAuth
		params      map[string]string
		expectError bool
	}{
		{
			name: "Valid API Key via GET parameter",
			auth: connection.APIKeyAuth{
				APIKey:           apiKey,
				GetParameterName: "api_key",
				Url:              serverGetParam.URL,
			},
			params: map[string]string{},
		},
		{
			name: "Invalid API Key via GET parameter",
			auth: connection.APIKeyAuth{
				APIKey:           "wrong-key",
				GetParameterName: "api_key",
				Url:              serverGetParam.URL,
			},
			params:      map[string]string{},
			expectError: true,
		},
		{
			name: "Valid API Key via HTTP header",
			auth: connection.APIKeyAuth{
				APIKey:         apiKey,
				HttpHeaderName: "Authorization",
				Url:            serverHeader.URL,
			},
			params: map[string]string{},
		},
		{
			name: "Invalid API Key via HTTP header",
			auth: connection.APIKeyAuth{
				APIKey:         "wrong-key",
				HttpHeaderName: "Authorization",
				Url:            serverHeader.URL,
			},
			params:      map[string]string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.auth.Authenticate()
			if err != nil {
				t.Fatalf("Authenticate() failed: %v", err)
			}

			// Fetch report
			_, err = tt.auth.FetchReport(tt.params)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("FetchReport() failed: %v", err)
			}
		})
	}
}
