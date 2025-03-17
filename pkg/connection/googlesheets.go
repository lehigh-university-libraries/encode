package connection

import (
	"context"
	"errors"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsAuth struct {
	CredentialsFile string
	Service         *sheets.Service
}

func (g *GoogleSheetsAuth) Authenticate() error {
	if g.CredentialsFile == "" {
		return errors.New("missing Google Sheets credentials file")
	}

	ctx := context.Background()
	service, err := sheets.NewService(ctx, option.WithCredentialsFile(g.CredentialsFile))
	if err != nil {
		return err
	}

	g.Service = service
	return nil
}

func (g *GoogleSheetsAuth) FetchReport(params map[string]string) ([]map[string]string, error) {
	if g.Service == nil {
		return nil, errors.New("Google Sheets API not initialized")
	}

	spreadsheetID, ok := params["spreadsheet_id"]
	if !ok {
		return nil, errors.New("missing spreadsheet_id parameter")
	}
	readRange, ok := params["range"]
	if !ok {
		return nil, errors.New("missing range parameter")
	}

	// todo: map from result to []map[string]string
	_, err := g.Service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, err
	}
	return nil, nil
}
