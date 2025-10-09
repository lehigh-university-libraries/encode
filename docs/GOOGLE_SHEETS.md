# Google Sheets Integration

This document describes how to configure and authenticate with Google Sheets in `encode`.

## Overview

The Google Sheets connection type allows `encode` to fetch data from Google Sheets spreadsheets and export them as CSV files. This is useful for incorporating data from collaborative spreadsheets into your reporting pipeline.

## Prerequisites

1. A Google Cloud Platform (GCP) project
2. Google Sheets API enabled
3. A Service Account with appropriate permissions
4. The target Google Sheet(s) shared with the Service Account

## Setup Instructions

### 1. Create a Service Account

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Select or create a project
3. Navigate to **IAM & Admin** > **Service Accounts**
4. Click **Create Service Account**
5. Provide a name and description (e.g., "encode-sheets-reader")
6. Click **Create and Continue**
7. Skip granting roles (not required for reading sheets)
8. Click **Done**

### 2. Create and Download Credentials

1. Find your newly created Service Account in the list
2. Click on the Service Account email
3. Go to the **Keys** tab
4. Click **Add Key** > **Create new key**
5. Select **JSON** as the key type
6. Click **Create**
7. Save the downloaded JSON file securely (e.g., `~/credentials/encode-service-account.json`)

**Important**: Keep this file secure. It contains private keys that provide access to your Service Account.

### 3. Enable Google Sheets API

1. In Google Cloud Console, go to **APIs & Services** > **Library**
2. Search for "Google Sheets API"
3. Click on it and click **Enable**

### 4. Share Your Google Sheets with the Service Account

For each Google Sheet you want to access:

1. Open the Google Sheet in your browser
2. Click the **Share** button
3. Add the Service Account email address (found in the JSON credentials file, looks like: `your-service-account@your-project.iam.gserviceaccount.com`)
4. Grant **Viewer** access (read-only is sufficient)
5. Uncheck "Notify people" if you don't want to send an email
6. Click **Share**

### 5. Set Environment Variable

Set the `GOOGLE_CREDENTIALS_FILE` environment variable to point to your credentials file:

```bash
export GOOGLE_CREDENTIALS_FILE="/path/to/your-service-account.json"
```

You can add this to your `~/.bashrc`, `~/.zshrc`, or other shell configuration file to make it persistent.

## Configuration

### Connection Configuration

Add a Google Sheets connection to your `encode.yaml`:

```yaml
connections:
  - name: google_sheets
    type: GoogleSheets
    credentials_file: "${GOOGLE_CREDENTIALS_FILE}"
```

### Report Configuration

Configure reports that use the Google Sheets connection:

```yaml
reports:
  - name: my_sheet_report
    connection: google_sheets
    query_params:
      spreadsheet_id: "1FNlPFrGItPDk_kdMw2XtR_AMvX4GQ84L11uCe2DbjS8"
      gid: "0"           # Sheet GID (0 is usually the first sheet)
      header_row: "1"    # Row number containing headers (1-indexed)
    schedule: "0 5 * * *"  # Daily at 5 AM
```

### Query Parameters

- **`spreadsheet_id`** (required): The Google Sheets spreadsheet ID
  - Found in the URL: `https://docs.google.com/spreadsheets/d/{SPREADSHEET_ID}/edit`

- **`gid`** (required): The sheet GID (Grid ID)
  - GID 0 is typically the first sheet
  - For other sheets, find the GID in the URL: `https://docs.google.com/spreadsheets/d/{SPREADSHEET_ID}/edit#gid={GID}`

- **`header_row`** (optional): The row number where column headers are located
  - Defaults to `"1"` if not specified
  - Use `"2"` if your headers are in the second row (e.g., if row 1 contains metadata)
  - Must be a string value, not an integer

## Example: Multiple Sheets from One Spreadsheet

Here's an example of fetching multiple sheets (tabs) from a single spreadsheet:

```yaml
connections:
  - name: google_sheets
    type: GoogleSheets
    credentials_file: "${GOOGLE_CREDENTIALS_FILE}"

reports:
  - name: io_classes
    connection: google_sheets
    query_params:
      spreadsheet_id: "1FNlPFrGItPDk_kdMw2XtR_AMvX4GQ84L11uCe2DbjS8"
      gid: "0"
      header_row: "2"
    schedule: "0 5 * * *"

  - name: io_tours
    connection: google_sheets
    query_params:
      spreadsheet_id: "1FNlPFrGItPDk_kdMw2XtR_AMvX4GQ84L11uCe2DbjS8"
      gid: "1109646791"
      header_row: "2"
    schedule: "0 5 * * *"

  - name: io_external
    connection: google_sheets
    query_params:
      spreadsheet_id: "1FNlPFrGItPDk_kdMw2XtR_AMvX4GQ84L11uCe2DbjS8"
      gid: "1916927317"
      header_row: "2"
    schedule: "0 5 * * *"

  - name: io_misc
    connection: google_sheets
    query_params:
      spreadsheet_id: "1FNlPFrGItPDk_kdMw2XtR_AMvX4GQ84L11uCe2DbjS8"
      gid: "959800694"
      header_row: "2"
    schedule: "0 5 * * *"
```

## Security Best Practices

1. **Never commit credentials to version control**: Use environment variables and keep the JSON file out of your repository
2. **Use minimal permissions**: The Service Account only needs read access to the specific sheets
3. **Rotate credentials regularly**: Periodically create new keys and delete old ones
4. **Restrict file permissions**: Set the credentials file to be readable only by the user running `encode`:
   ```bash
   chmod 600 /path/to/your-service-account.json
   ```

## Troubleshooting

### "Permission denied" or "404 Not Found"

- Verify the spreadsheet is shared with your Service Account email
- Check that the spreadsheet ID is correct
- Ensure the Google Sheets API is enabled in your GCP project

### "Invalid credentials" or authentication errors

- Verify the `GOOGLE_CREDENTIALS_FILE` environment variable is set correctly
- Check that the credentials file is valid JSON
- Ensure the Service Account hasn't been deleted or disabled

### "Invalid GID" or empty data

- Verify the GID matches the sheet you want to access
- Check the URL of the sheet tab: `#gid={GID}`
- Ensure the sheet contains data

### Wrong headers or data

- Adjust the `header_row` parameter to match where your headers actually are
- Row numbers are 1-indexed (first row = "1", second row = "2", etc.)

## Implementation Notes

See `pkg/connection/googlesheets.go` for the implementation details of the Google Sheets connection provider.
