## Architecture

### Core Components

1. **Connection Providers** (`pkg/connection/`)
   - Interface: `ConnectionProvider` with two methods:
     - `Authenticate() error` - establishes connection to remote service
     - `FetchReport(params map[string]string) ([]map[string]string, error)` - retrieves data
   - Implementations:
     - `PostgresAuth`: Executes SQL queries via pgx connection pool
     - `MariaDBAuth`: Executes SQL queries via database/sql with MySQL driver
     - `FolioAuth`: Authenticates to FOLIO API and executes SQL queries from GitHub URLs via MetaDB
     - `GoogleSheetsAuth`: Fetches data from Google Sheets (implementation incomplete)
     - `GoogleAnalyticsAuth`: Google Analytics integration (implementation incomplete)
     - `MockConnection`: For testing

2. **Configuration System** (`pkg/config/`)
   - `LoadConfig()` reads YAML, expands environment variables (using `os.ExpandEnv`), validates cron schedules
   - `Config` struct contains:
     - `Connections`: Array of connection definitions (name, type, credentials)
     - `Reports`: Array of report configurations
     - `StagingDirectory`: Where CSV files are written locally
     - `S3`: S3 configuration for AWS upload (optional)
   - Each `ReportConfig` is initialized with its own connection provider reference and shared S3 uploader

3. **Cron Scheduling** (`pkg/config/cron.go`)
   - `Config.StartCron()` sets up scheduled jobs using robfig/cron
   - Each `ReportConfig` implements `cron.Job` interface via `Run()` method
   - `Run()` executes: fetch report → create directory → write CSV with timestamp filename → upload to S3 (if enabled) → generate QuickSight manifest

4. **Storage Layer** (`pkg/storage/`)
   - `S3Uploader`: Handles AWS S3 uploads using AWS SDK v2
   - `UploadFile()`: Uploads CSV files to S3 with path structure: `{prefix}/{report_name}/{filename}.csv`
   - `GenerateManifest()`: Creates AWS QuickSight-compatible JSON manifest files locally
   - `UploadManifest()`: Uploads manifest files to S3 for QuickSight import
   - S3 functionality is optional and controlled by `s3.enabled` config flag

5. **CLI** (`cmd/`)
   - Built with spf13/cobra
   - Root command handles logging configuration (DEBUG/INFO/WARN/ERROR)
   - `run` command: loads config and starts cron scheduler

### Data Flow

1. User creates `encode.yaml` with connection definitions, report schedules, and optional S3 configuration
2. `encode run` loads config, validates cron expressions, initializes connection providers and S3 uploader (if enabled)
3. Cron scheduler calls `ReportConfig.Run()` on schedule
4. `Run()` executes the following pipeline:
   - Fetches data via connection provider
   - Writes CSV locally to `{stagingDirectory}/{report_name}/{timestamp}.csv`
   - If S3 enabled: uploads CSV to `s3://{bucket}/{prefix}/{report_name}/{timestamp}.csv`
   - If S3 enabled: updates cumulative manifest file locally at `{manifest_path}/{report_name}/manifest.json` (appends new S3 URI)
   - If S3 enabled: uploads updated manifest to S3 at `{prefix}/manifests/{report_name}/manifest.json`

### Configuration Format

Connection types in YAML:
- `PostgreSQL`: requires `dsn` field
- `MariaDB`: requires `dsn` field
- `FOLIO`: requires `base_url`, `tenant`, `username`, and `password` fields
- `GoogleSheets`: requires `credentials_file` field
- `Mock`: for testing

Report parameters vary by connection type:
- PostgreSQL/MariaDB: `query_params.query`
- FOLIO: `query_params.query_url` (GitHub raw URL to SQL file)
- GoogleSheets: `query_params.spreadsheet_id` and `query_params.range`

S3 configuration (optional):
- `enabled`: boolean to enable/disable S3 uploads
- `bucket`: S3 bucket name
- `region`: AWS region (e.g., "us-east-1")
- `prefix`: path prefix for organizing files in S3
- `manifest_path`: local directory for storing QuickSight manifest files

AWS credentials are loaded via standard AWS SDK credential chain (environment variables, AWS config files, IAM roles, etc.)

### Testing Notes

- PostgreSQL tests use `pashagolub/pgxmock` for mocking database connections
- MariaDB tests use `DATA-DOG/go-sqlmock` for mocking database connections
- Test fixtures in `fixtures/` directory include example YAML configs
- `PostgresAuth.DB` and `MariaDBAuth.DB` fields are exposed to allow injecting mock connections in tests

### AWS QuickSight Integration

The application generates QuickSight-compatible manifest files following AWS specifications:
- Format: JSON with QuickSight structure
- CSV settings: comma delimiter, double-quote text qualifier, headers included
- Manifest URIs use standard S3 format: `s3://bucket/prefix/report_name/file.csv`
- Manifest files are named: `manifest.json` (one per report)
- **Cumulative approach**: Each cron run appends new S3 URIs to the manifest, preserving historical data

#### How Historical Data is Maintained

1. First cron run: CSV uploaded to S3 → manifest created with 1 URI
2. Second cron run: New CSV uploaded → manifest updated to include both URIs (old + new)
3. Third cron run: Another CSV uploaded → manifest now contains all 3 URIs
4. QuickSight reads the manifest and imports all referenced files, combining the data

This allows QuickSight to maintain historical data across multiple report runs without manual intervention.

#### Setup Instructions

To use in QuickSight:
1. Enable S3 in `encode.yaml` with appropriate bucket and region
2. Ensure QuickSight has read access to the S3 bucket
3. In QuickSight, create dataset from S3 using the manifest URL: `https://{bucket}.s3-{region}.amazonaws.com/{prefix}/manifests/{report_name}/manifest.json`
4. Configure QuickSight dataset to refresh on a schedule that matches or exceeds your cron schedule
5. Each refresh will automatically include new data files as they're added to the manifest

### Known TODOs/Limitations

- Google Sheets `FetchReport` returns nil (not implemented) - see pkg/connection/googlesheets.go:45
- PostgreSQL type conversion assumes all columns are strings - see pkg/connection/postgresql.go:84-85
- MariaDB type conversion handles basic types but may need enhancement for complex types
- Template processing mentioned in config but not implemented
- Validation definitions mentioned in README but not implemented
- Manifest files grow indefinitely; no built-in mechanism to prune old URIs (consider implementing time-based or count-based pruning)
- CSV files must have consistent schemas across all runs for QuickSight to properly combine them
