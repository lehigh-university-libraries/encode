package connection

import (
	"context"
	"errors"
	"fmt"
	"strconv"

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
	// Only authenticate if Service is not already set (e.g., for testing with mocks)
	if g.Service == nil {
		err := g.Authenticate()
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	spreadsheetID, ok := params["spreadsheet_id"]
	if !ok {
		return nil, errors.New("missing spreadsheet_id parameter")
	}

	gid, ok := params["gid"]
	if !ok {
		return nil, errors.New("missing gid parameter")
	}

	// Header row position (1 or 2)
	headerRow := 1
	if hr, ok := params["header_row"]; ok {
		parsed, err := strconv.Atoi(hr)
		if err != nil {
			return nil, fmt.Errorf("invalid header_row value: %w", err)
		}
		if parsed != 1 && parsed != 2 {
			return nil, errors.New("header_row must be 1 or 2")
		}
		headerRow = parsed
	}

	// Data starts at the row after the header
	dataStartRow := headerRow + 1

	// First, get the sheet name from the gid
	spreadsheet, err := g.Service.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	var sheetName string
	gidInt, err := strconv.ParseInt(gid, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid gid format: %w", err)
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.SheetId == gidInt {
			sheetName = sheet.Properties.Title
			break
		}
	}

	if sheetName == "" {
		return nil, fmt.Errorf("sheet with gid %s not found", gid)
	}

	// Read header row
	headerRange := fmt.Sprintf("%s!A%d:ZZ%d", sheetName, headerRow, headerRow)
	headerResp, err := g.Service.Spreadsheets.Values.Get(spreadsheetID, headerRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read header row: %w", err)
	}

	if len(headerResp.Values) == 0 || len(headerResp.Values[0]) == 0 {
		return nil, errors.New("header row is empty")
	}

	// Extract headers
	headers := make([]string, len(headerResp.Values[0]))
	for i, v := range headerResp.Values[0] {
		if str, ok := v.(string); ok {
			headers[i] = str
		} else {
			headers[i] = fmt.Sprintf("%v", v)
		}
	}

	// Read all data starting from dataStartRow
	// We'll read a large range and iterate until we find a blank cell in column A
	dataRange := fmt.Sprintf("%s!A%d:ZZ10000", sheetName, dataStartRow)
	dataResp, err := g.Service.Spreadsheets.Values.Get(spreadsheetID, dataRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	var results []map[string]string

	// Iterate through rows until column A is blank
	for _, row := range dataResp.Values {
		// Check if column A (index 0) is blank or empty
		if len(row) == 0 {
			break
		}

		firstCell := ""
		if row[0] != nil {
			if str, ok := row[0].(string); ok {
				firstCell = str
			} else {
				firstCell = fmt.Sprintf("%v", row[0])
			}
		}

		// Stop when column A is blank
		if firstCell == "" {
			break
		}

		// Map row to headers
		rowMap := make(map[string]string)
		for i, header := range headers {
			if i < len(row) && row[i] != nil {
				if str, ok := row[i].(string); ok {
					rowMap[header] = str
				} else {
					rowMap[header] = fmt.Sprintf("%v", row[i])
				}
			} else {
				rowMap[header] = ""
			}
		}
		results = append(results, rowMap)
	}

	return results, nil
}
