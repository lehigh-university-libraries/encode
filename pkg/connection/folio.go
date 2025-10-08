package connection

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// FolioAuth handles authentication and data fetching from FOLIO MetaDB API
type FolioAuth struct {
	BaseURL  string
	Tenant   string
	Username string
	Password string
	Token    string
	Client   *http.Client
}

// loginRequest represents the FOLIO login payload
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// reportRequest represents the FOLIO report request payload
type reportRequest struct {
	URL    string            `json:"url"`
	Params map[string]string `json:"params,omitempty"`
}

// reportResponse represents the FOLIO report API response
type reportResponse struct {
	TotalRecords int                      `json:"totalRecords"`
	Records      []map[string]interface{} `json:"records"`
}

// Authenticate logs into FOLIO and retrieves an authentication token
func (f *FolioAuth) Authenticate() error {
	if f.BaseURL == "" {
		return errors.New("missing FOLIO base_url")
	}
	if f.Tenant == "" {
		return errors.New("missing FOLIO tenant")
	}
	if f.Username == "" {
		return errors.New("missing FOLIO username")
	}
	if f.Password == "" {
		return errors.New("missing FOLIO password")
	}

	if f.Client == nil {
		f.Client = &http.Client{
			Timeout: 60 * time.Second,
		}
	}

	loginPayload := loginRequest{
		Username: f.Username,
		Password: f.Password,
	}

	body, err := json.Marshal(loginPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal login payload: %w", err)
	}

	loginURL := fmt.Sprintf("%s/authn/login", f.BaseURL)
	req, err := http.NewRequestWithContext(context.Background(), "POST", loginURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-okapi-tenant", f.Tenant)

	resp, err := f.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	token := resp.Header.Get("x-okapi-token")
	if token == "" {
		return errors.New("no authentication token received from FOLIO")
	}

	f.Token = token
	slog.Debug("FOLIO authentication successful", "tenant", f.Tenant)
	return nil
}

// FetchReport executes a SQL query from a GitHub URL and returns results as CSV data
func (f *FolioAuth) FetchReport(params map[string]string) ([]map[string]string, error) {
	// Initialize client if not set
	if f.Client == nil {
		f.Client = &http.Client{
			Timeout: 60 * time.Second,
		}
	}

	// Authenticate if token is not set
	if f.Token == "" {
		err := f.Authenticate()
		if err != nil {
			slog.Warn("Unable to authenticate to FOLIO")
			return nil, err
		}
	}

	queryURL, ok := params["query_url"]
	if !ok {
		return nil, errors.New("missing query_url parameter")
	}

	// Extract any additional params (excluding query_url)
	queryParams := make(map[string]string)
	for k, v := range params {
		if k != "query_url" {
			queryParams[k] = v
		}
	}

	reportPayload := reportRequest{
		URL:    queryURL,
		Params: queryParams,
	}

	body, err := json.Marshal(reportPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report payload: %w", err)
	}

	reportURL := fmt.Sprintf("%s/ldp/db/reports", f.BaseURL)
	req, err := http.NewRequestWithContext(context.Background(), "POST", reportURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create report request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-okapi-tenant", f.Tenant)
	req.Header.Set("x-okapi-token", f.Token)

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute report request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("report request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read JSON response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var reportResp reportResponse
	err = json.Unmarshal(bodyBytes, &reportResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Convert []map[string]interface{} to []map[string]string
	results := make([]map[string]string, len(reportResp.Records))
	for i, record := range reportResp.Records {
		rowMap := make(map[string]string)
		for key, value := range record {
			// Convert all values to strings
			if value == nil {
				rowMap[key] = ""
			} else {
				rowMap[key] = fmt.Sprintf("%v", value)
			}
		}
		results[i] = rowMap
	}

	slog.Debug("FOLIO report fetched successfully", "rows", len(results))
	return results, nil
}
