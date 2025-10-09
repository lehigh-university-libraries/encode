package connection

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

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

	gidParam, ok := params["gid"]
	if !ok {
		return nil, errors.New("missing gid parameter")
	}

	// Parse comma-separated GIDs
	gidStrings := strings.Split(gidParam, ",")
	if len(gidStrings) == 0 {
		return nil, errors.New("gid parameter is empty")
	}

	// Header row position (default: 1)
	headerRow := 1
	if hr, ok := params["header_row"]; ok {
		parsed, err := strconv.Atoi(hr)
		if err != nil {
			return nil, fmt.Errorf("invalid header_row value: %w", err)
		}
		if parsed < 1 {
			return nil, errors.New("header_row must be >= 1")
		}
		headerRow = parsed
	}

	// Data starts at the row after the header
	dataStartRow := headerRow + 1

	// Get spreadsheet metadata to map GIDs to sheet names
	spreadsheet, err := g.Service.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	// Build a map of GID -> Sheet Name
	gidToName := make(map[int64]string)
	for _, sheet := range spreadsheet.Sheets {
		gidToName[sheet.Properties.SheetId] = sheet.Properties.Title
	}

	// Fetch data from each sheet
	var sheetDataList []SheetData
	for i, gidStr := range gidStrings {
		gidStr = strings.TrimSpace(gidStr)
		gidInt, err := strconv.ParseInt(gidStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid gid format '%s': %w", gidStr, err)
		}

		sheetName, exists := gidToName[gidInt]
		if !exists {
			return nil, fmt.Errorf("sheet with gid %s not found", gidStr)
		}

		slog.Debug("Fetching data from sheet", "gid", gidStr, "name", sheetName)

		sheetData := SheetData{Name: sheetName}

		// For the first GID (index 0), read the header row
		if i == 0 {
			headerRange := fmt.Sprintf("%s!A%d:ZZ%d", sheetName, headerRow, headerRow)
			headerResp, err := g.Service.Spreadsheets.Values.Get(spreadsheetID, headerRange).Do()
			if err != nil {
				return nil, fmt.Errorf("failed to read header row from sheet '%s': %w", sheetName, err)
			}

			if len(headerResp.Values) == 0 || len(headerResp.Values[0]) == 0 {
				return nil, fmt.Errorf("header row is empty in sheet '%s'", sheetName)
			}

			sheetData.Header = headerResp.Values[0]
		}

		// Read data from this sheet
		dataRange := fmt.Sprintf("%s!A%d:ZZ10000", sheetName, dataStartRow)
		dataResp, err := g.Service.Spreadsheets.Values.Get(spreadsheetID, dataRange).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to read data from sheet '%s': %w", sheetName, err)
		}

		sheetData.Rows = dataResp.Values
		sheetDataList = append(sheetDataList, sheetData)
	}

	// Parse the fetched data
	return ParseSheetData(sheetDataList, headerRow)
}

// SheetData represents data fetched from a single sheet
type SheetData struct {
	Name   string
	Header []interface{}
	Rows   [][]interface{}
}

// ParseSheetData converts raw Google Sheets API responses into structured data
// This function is separated for easier testing
func ParseSheetData(sheets []SheetData, headerRow int) ([]map[string]string, error) {
	if len(sheets) == 0 {
		return nil, errors.New("no sheets provided")
	}

	var headers []string
	var allResults []map[string]string

	// Process each sheet
	for i, sheet := range sheets {
		slog.Debug("Processing sheet", "name", sheet.Name, "index", i)

		// For the first sheet, extract headers
		if i == 0 {
			if len(sheet.Header) == 0 {
				return nil, fmt.Errorf("header row is empty in sheet '%s'", sheet.Name)
			}

			headers = make([]string, len(sheet.Header))
			for j, v := range sheet.Header {
				if str, ok := v.(string); ok {
					headers[j] = str
				} else {
					headers[j] = fmt.Sprintf("%v", v)
				}
			}

			// Add "sheet" column to headers
			headers = append(headers, "sheet")

			slog.Debug("Extracted headers from first sheet", "headers", headers)
		}

		// Process rows
		for _, row := range sheet.Rows {
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

			// Map row to headers (excluding the "sheet" column)
			rowMap := make(map[string]string)
			for j, header := range headers[:len(headers)-1] { // Skip last "sheet" header
				if j < len(row) && row[j] != nil {
					if str, ok := row[j].(string); ok {
						rowMap[header] = str
					} else {
						rowMap[header] = fmt.Sprintf("%v", row[j])
					}
				} else {
					rowMap[header] = ""
				}
			}

			// Add sheet name column
			rowMap["sheet"] = sheet.Name

			allResults = append(allResults, rowMap)
		}

		slog.Debug("Processed sheet", "name", sheet.Name, "rows", len(sheet.Rows))
	}

	slog.Info("Merged data from multiple sheets", "total_sheets", len(sheets), "total_rows", len(allResults))

	return allResults, nil
}
