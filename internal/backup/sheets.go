package backup

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsClient wraps the Google Sheets API.
type SheetsClient struct {
	srv *sheets.Service
}

// NewSheetsClient creates a Sheets client from service account credentials.
func NewSheetsClient(ctx context.Context, credentialsPath string) (*SheetsClient, error) {
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(config.Client(ctx)))
	if err != nil {
		return nil, fmt.Errorf("create sheets service: %w", err)
	}

	return &SheetsClient{srv: srv}, nil
}

// NewSheetsClientOAuth2 creates a Sheets client using OAuth2 user credentials
// (client_id, client_secret, refresh_token). This is the auth path for
// personal Google Workspace tokens.
func NewSheetsClientOAuth2(ctx context.Context, clientID, clientSecret, refreshToken string) (*SheetsClient, error) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{sheets.SpreadsheetsScope},
	}

	token := &oauth2.Token{RefreshToken: refreshToken}
	tokenSource := config.TokenSource(ctx, token)

	srv, err := sheets.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("create sheets service: %w", err)
	}

	return &SheetsClient{srv: srv}, nil
}

// CreateSpreadsheet creates a new spreadsheet with the given data.
// Returns the spreadsheet ID and URL.
func (c *SheetsClient) CreateSpreadsheet(ctx context.Context, title string, data []SheetData) (id string, url string, err error) {
	// Build sheet properties
	var sheetProps []*sheets.Sheet
	for _, sd := range data {
		sheetProps = append(sheetProps, &sheets.Sheet{
			Properties: &sheets.SheetProperties{Title: sd.Name},
		})
	}

	ss, err := c.srv.Spreadsheets.Create(&sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{Title: title},
		Sheets:     sheetProps,
	}).Context(ctx).Do()
	if err != nil {
		return "", "", fmt.Errorf("create spreadsheet: %w", err)
	}

	// Write data to each sheet
	for _, sd := range data {
		if err := c.writeSheet(ctx, ss.SpreadsheetId, sd); err != nil {
			return "", "", err
		}
	}

	return ss.SpreadsheetId, ss.SpreadsheetUrl, nil
}

// UpdateSpreadsheet overwrites an existing spreadsheet with local data.
func (c *SheetsClient) UpdateSpreadsheet(ctx context.Context, spreadsheetID string, data []SheetData) error {
	// Get existing sheet names
	ss, err := c.srv.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("get spreadsheet: %w", err)
	}

	existingSheets := make(map[string]int64)
	for _, s := range ss.Sheets {
		existingSheets[s.Properties.Title] = s.Properties.SheetId
	}

	// Create missing sheets, clear existing ones
	var requests []*sheets.Request
	for _, sd := range data {
		if _, exists := existingSheets[sd.Name]; !exists {
			requests = append(requests, &sheets.Request{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{Title: sd.Name},
				},
			})
		}
	}

	if len(requests) > 0 {
		_, err := c.srv.Spreadsheets.BatchUpdate(spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
			Requests: requests,
		}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("add sheets: %w", err)
		}
	}

	// Clear and write each sheet
	for _, sd := range data {
		clearRange := fmt.Sprintf("'%s'", sd.Name)
		_, _ = c.srv.Spreadsheets.Values.Clear(spreadsheetID, clearRange, &sheets.ClearValuesRequest{}).Context(ctx).Do()

		if err := c.writeSheet(ctx, spreadsheetID, sd); err != nil {
			return err
		}
	}

	return nil
}

// ReadSpreadsheet reads all data from an existing spreadsheet.
func (c *SheetsClient) ReadSpreadsheet(ctx context.Context, spreadsheetID string) ([]SheetData, error) {
	ss, err := c.srv.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet: %w", err)
	}

	var result []SheetData
	for _, sheet := range ss.Sheets {
		name := sheet.Properties.Title
		resp, err := c.srv.Spreadsheets.Values.Get(spreadsheetID, fmt.Sprintf("'%s'", name)).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("read sheet %s: %w", name, err)
		}

		sd := SheetData{Name: name}
		for i, row := range resp.Values {
			strRow := make([]string, len(row))
			for j, cell := range row {
				strRow[j] = fmt.Sprintf("%v", cell)
			}
			if i == 0 {
				sd.Headers = strRow
			} else {
				sd.Rows = append(sd.Rows, strRow)
			}
		}
		result = append(result, sd)
	}

	return result, nil
}

func (c *SheetsClient) writeSheet(ctx context.Context, spreadsheetID string, sd SheetData) error {
	var values [][]interface{}

	// Header row
	headerRow := make([]interface{}, len(sd.Headers))
	for i, h := range sd.Headers {
		headerRow[i] = h
	}
	values = append(values, headerRow)

	// Data rows
	for _, row := range sd.Rows {
		valRow := make([]interface{}, len(row))
		for i, v := range row {
			valRow[i] = v
		}
		values = append(values, valRow)
	}

	rangeStr := fmt.Sprintf("'%s'!A1", sd.Name)
	_, err := c.srv.Spreadsheets.Values.Update(spreadsheetID, rangeStr, &sheets.ValueRange{
		Values: values,
	}).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("write sheet %s: %w", sd.Name, err)
	}

	return nil
}

// ReadCSVDir reads CSV files from a directory and returns SheetData slices.
// This is useful for testing the diff logic without Google Sheets.
func ReadCSVDir(dir string) ([]SheetData, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	var result []SheetData
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".csv") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".csv")
		f, err := os.Open(fmt.Sprintf("%s/%s", dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", entry.Name(), err)
		}

		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}

		sd := SheetData{Name: name}
		if len(records) > 0 {
			sd.Headers = records[0]
			sd.Rows = records[1:]
		}
		result = append(result, sd)
	}

	return result, nil
}
