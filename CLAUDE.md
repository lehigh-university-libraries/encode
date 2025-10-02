# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`encode` is a Go CLI tool that authenticates to various data sources (PostgreSQL, Google Sheets, Google Analytics), fetches reports on a cron schedule, and saves the output as CSV files.


## 📚 Critical Documentation References
- **Go Conventions**: `./docs/GO_CONVENTIONS.md`
- **Project Architecture**: `./docs/ARCHITECTURE.md`

## Common Commands

### Build and Test
```bash
# Build the binary
make build

# Run tests
make test

# Run a single test file
go test -v ./pkg/config/config_test.go

# Run a specific test
go test -v -run TestFunctionName ./pkg/...

# Lint code
make lint
```

### Running the Application
```bash
# Run with default config (~/encode.yaml or $ENCODE_CONFIG_YAML)
./encode run

# Run with specific config file
./encode run --config path/to/encode.yaml

# Set log level
./encode run --log-level DEBUG
```
