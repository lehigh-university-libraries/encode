package connection

import (
	"context"
	"errors"

	"google.golang.org/api/analytics/v3"
	"google.golang.org/api/option"
)

type GoogleAuth struct {
	CredentialsFile string
	Service         *analytics.Service
}

func (g *GoogleAuth) Authenticate() error {
	if g.CredentialsFile == "" {
		return errors.New("missing Google Analytics credentials file")
	}

	ctx := context.Background()
	service, err := analytics.NewService(ctx, option.WithCredentialsFile(g.CredentialsFile))
	if err != nil {
		return err
	}
	g.Service = service
	return nil
}

// FetchReport retrieves a report from Google Analytics
func (g *GoogleAuth) FetchReport(params map[string]string) ([]map[string]string, error) {
	if g.Service == nil {
		return nil, errors.New("Google Analytics API not initialized")
	}

	viewID, ok := params["view_id"]
	if !ok {
		return nil, errors.New("missing view_id parameter")
	}
	startDate, ok := params["start_date"]
	if !ok {
		return nil, errors.New("missing start_date parameter")
	}
	endDate, ok := params["end_date"]
	if !ok {
		return nil, errors.New("missing end_date parameter")
	}
	metrics, ok := params["metrics"]
	if !ok {
		return nil, errors.New("missing metrics parameter")
	}
	dimensions := params["dimensions"]

	// Build the Google Analytics request
	req := g.Service.Data.Ga.Get("ga:"+viewID, startDate, endDate, metrics)
	if dimensions != "" {
		req = req.Dimensions(dimensions)
	}

	// todo: map resp to []map[string]string
	_, err := req.Do()
	if err != nil {
		return nil, err
	}

	return nil, nil
}
