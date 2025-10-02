package connection

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
)

type SqlQuerier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Close() error
}

type MariaDBAuth struct {
	DSN string
	DB  SqlQuerier
}

func (m *MariaDBAuth) Authenticate() error {
	if m.DSN == "" {
		return errors.New("missing MariaDB DSN")
	}

	db, err := sql.Open("mysql", m.DSN)
	if err != nil {
		return err
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return err
	}

	m.DB = db
	return nil
}

// FetchReport executes a SQL query and returns results
func (m *MariaDBAuth) FetchReport(params map[string]string) ([]map[string]string, error) {
	// Only authenticate if DB is not already set (e.g., for testing with mocks)
	if m.DB == nil {
		err := m.Authenticate()
		if err != nil {
			slog.Warn("Unable to authenticate")
			return nil, err
		}
	}

	if m.DB == nil {
		return nil, errors.New("MariaDB database not initialized")
	}

	query, ok := params["query"]
	if !ok {
		return nil, errors.New("missing query parameter")
	}

	rows, err := m.DB.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]string

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err = rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		rowMap := make(map[string]string)
		for i, col := range cols {
			var v string
			val := values[i]
			if val != nil {
				// Convert bytes to string if necessary
				if b, ok := val.([]byte); ok {
					v = string(b)
				} else {
					// todo: reflect on val type and convert properly to string
					v = val.(string)
				}
			}
			rowMap[col] = v
		}
		results = append(results, rowMap)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
