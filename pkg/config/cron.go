package config

import (
	"encoding/csv"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	cron "github.com/robfig/cron/v3"
)

func (c *Config) StartCron() *cron.Cron {
	cron := cron.New()
	for _, report := range c.Reports {
		slog.Debug("Adding cron for " + report.Name)
		e, err := cron.AddJob(report.Schedule, report)
		if err != nil {
			slog.Error("Unable to start cron", "report", report.Name, "err", err)
			continue
		}
		slog.Info("Scheduled job", "report", report.Name, "cron.entryId", e)
	}

	return cron
}

// https://pkg.go.dev/github.com/robfig/cron#FuncJob.Run
func (r ReportConfig) Run() {
	slog.Debug("Running", "report", r.Name)

	results, err := r.connection.FetchReport(r.QueryParams)
	if err != nil {
		slog.Error("Unable to fetch report", "report", r.Name, "err", err)
		return
	}

	if len(results) == 0 {
		slog.Error("NO results returned", "report", r.Name, "err", err)
		return
	}

	filename := filepath.Join(r.StagingDirectory, r.Name, time.Now().Format("2006-01-02.15.04.05.csv"))
	file, err := os.Create(filename)
	if err != nil {
		slog.Error("Error creating file", "filename", filename, "err", err)
		return
	}
	defer file.Close()
	err = writeToCSV(results, file)
	if err != nil {
		slog.Error("Error writing CSV", "filename", filename, "err", err)
		return
	}

	slog.Info("Saved report", "filename", filename)
}

// writeToCSV writes a slice of maps to a CSV file.
// The first map's keys are used as the header row.
func writeToCSV(data []map[string]string, f *os.File) error {
	// Create a new CSV writer
	file := csv.NewWriter(f)
	defer file.Flush()

	// Get the keys from the first map to use as the header row
	header := make([]string, len(data[0]))
	for _, column := range data[0] {
		header = append(header, column)
	}

	// Write the header row
	err := file.Write(header)
	if err != nil {
		return err
	}

	// Write the data rows
	for _, row := range data {
		record := make([]string, len(row))
		for _, value := range row {
			record = append(record, value)
		}
		err = file.Write(record)
		if err != nil {
			return err
		}
	}

	return nil
}
