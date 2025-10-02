## Architecture

### Core Components

1. **Connection Providers** (`pkg/connection/`)
   - Interface: `ConnectionProvider` with two methods:
     - `Authenticate() error` - establishes connection to remote service
     - `FetchReport(params map[string]string) ([]map[string]string, error)` - retrieves data
   - Implementations:
     - `PostgresAuth`: Executes SQL queries via pgx connection pool
     - `MariaDBAuth`: Executes SQL queries via database/sql with MySQL driver
     - `GoogleSheetsAuth`: Fetches data from Google Sheets (implementation incomplete)
     - `GoogleAnalyticsAuth`: Google Analytics integration (implementation incomplete)
     - `MockConnection`: For testing

2. **Configuration System** (`pkg/config/`)
   - `LoadConfig()` reads YAML, expands environment variables (using `os.ExpandEnv`), validates cron schedules
   - `Config` struct contains:
     - `Connections`: Array of connection definitions (name, type, credentials)
     - `Reports`: Array of report configurations
     - `StagingDirectory`: Where CSV files are written
   - Each `ReportConfig` is initialized with its own connection provider reference

3. **Cron Scheduling** (`pkg/config/cron.go`)
   - `Config.StartCron()` sets up scheduled jobs using robfig/cron
   - Each `ReportConfig` implements `cron.Job` interface via `Run()` method
   - `Run()` executes: fetch report → create directory → write CSV with timestamp filename

4. **CLI** (`cmd/`)
   - Built with spf13/cobra
   - Root command handles logging configuration (DEBUG/INFO/WARN/ERROR)
   - `run` command: loads config and starts cron scheduler

### Data Flow

1. User creates `encode.yaml` with connection definitions and report schedules
2. `encode run` loads config, validates cron expressions, initializes connection providers
3. Cron scheduler calls `ReportConfig.Run()` on schedule
4. `Run()` fetches data via connection provider, writes to CSV at `{stagingDirectory}/{report_name}/{timestamp}.csv`

### Configuration Format

Connection types in YAML:
- `PostgreSQL`: requires `dsn` field
- `MariaDB`: requires `dsn` field
- `GoogleSheets`: requires `credentials_file` field
- `Mock`: for testing

Report parameters vary by connection type:
- PostgreSQL/MariaDB: `query_params.query`
- GoogleSheets: `query_params.spreadsheet_id` and `query_params.range`

### Testing Notes

- PostgreSQL tests use `pashagolub/pgxmock` for mocking database connections
- MariaDB tests use `DATA-DOG/go-sqlmock` for mocking database connections
- Test fixtures in `fixtures/` directory include example YAML configs
- `PostgresAuth.DB` and `MariaDBAuth.DB` fields are exposed to allow injecting mock connections in tests

### Known TODOs/Limitations

- Google Sheets `FetchReport` returns nil (not implemented) - see pkg/connection/googlesheets.go:45
- PostgreSQL type conversion assumes all columns are strings - see pkg/connection/postgresql.go:84-85
- Template processing mentioned in config but not implemented
- Validation definitions mentioned in README but not implemented
