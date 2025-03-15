package connection

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type APIKeyAuth struct {
	APIKey           string
	HttpHeaderName   string
	GetParameterName string
	Url              string
	HTTPClient       *http.Client
}

func (a *APIKeyAuth) Authenticate() error {
	if a.APIKey == "" {
		return errors.New("API key is missing")
	}
	if a.Url == "" {
		return errors.New("Url is missing")
	}
	if a.HttpHeaderName == "" && a.GetParameterName == "" {
		return errors.New("HttpHeaderName or GetParameterName need set")
	}
	if a.HttpHeaderName != "" && a.GetParameterName != "" {
		return errors.New("Only HttpHeaderName or GetParameterName need set")
	}
	a.HTTPClient = &http.Client{}

	return nil
}

// FetchReport makes an authenticated API request
func (a *APIKeyAuth) FetchReport(params map[string]string) (any, error) {
	if a.HTTPClient == nil {
		return nil, errors.New("HTTP client not initialized")
	}

	url := a.Url

	// Append API key to the request
	req, _ := http.NewRequest("GET", url, nil)
	if a.HttpHeaderName != "" {
		req.Header.Set(a.HttpHeaderName, a.APIKey)
	} else if a.GetParameterName != "" {
		url += "?" + a.GetParameterName + "=" + a.APIKey
		req, _ = http.NewRequest("GET", url, nil)
	}

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data any
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
