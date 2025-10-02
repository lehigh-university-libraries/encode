# encode

Authenticate to various sources, generate reports on a schedule, and upload to AWS S3 for QuickSight visualization.

## Overview

`encode` is a Go CLI tool that:
- Connects to data sources (PostgreSQL, MariaDB, Google Sheets)
- Runs SQL queries on a cron schedule
- Saves results as CSV files locally
- Uploads to AWS S3 with cumulative manifest files for QuickSight

## Configure

A YAML file `encode.yaml` is required. See [encode.example.yaml](./encode.example.yaml) for an example.

### Connections

Define a list of `connections` - each is a way to authenticate to a remote service:

- `PostgreSQL`: Requires `dsn` field
- `MariaDB`: Requires `dsn` field
- `GoogleSheets`: Requires `credentials_file` field

### Reports

Each report needs:
- A reference to a connection name defined in `connections`
- A cron schedule for when the report will run
- Query parameters specific to the connection type

### S3 (Optional)

Enable S3 uploads for QuickSight integration:
- `enabled`: Set to `true` to enable S3 uploads
- `bucket`: S3 bucket name
- `region`: AWS region
- `prefix`: Path prefix for organizing files
- `manifest_path`: Local directory for manifest files

AWS credentials are loaded via standard AWS SDK credential chain (environment variables, AWS config files, IAM roles).

## QuickSight Integration

See [docs/AWS_QUICKSIGHT.md](./docs/AWS_QUICKSIGHT.md)
