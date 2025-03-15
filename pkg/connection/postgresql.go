package connection

import (
	"context"
	"errors"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Close()
}

type PostgresAuth struct {
	DSN string
	DB  PgxQuerier
}

func (p *PostgresAuth) Authenticate() error {
	if p.DSN == "" {
		return errors.New("missing PostgreSQL DSN")
	}

	// Initialize connection pool
	config, err := pgxpool.ParseConfig(p.DSN)
	if err != nil {
		return err
	}

	db, err := pgxpool.New(context.Background(), config.ConnString())
	if err != nil {
		return err
	}

	p.DB = db
	return nil
}

// FetchReport executes a SQL query and returns results
func (p *PostgresAuth) FetchReport(params map[string]string) (any, error) {
	if p.DB == nil {
		return nil, errors.New("PostgreSQL database not initialized")
	}

	query, ok := params["query"]
	if !ok {
		return nil, errors.New("missing query parameter")
	}

	rows, err := p.DB.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	cols := rows.FieldDescriptions()

	for rows.Next() {
		rowData := make([]any, len(cols))
		rowPointers := make([]any, len(cols))
		for i := range rowData {
			rowPointers[i] = &rowData[i]
		}
		err = rows.Scan(rowPointers...)
		if err != nil {
			return nil, err
		}

		rowMap := make(map[string]any)
		for i, col := range cols {
			rowMap[string(col.Name)] = rowData[i]
		}
		results = append(results, rowMap)
	}

	return results, nil
}
