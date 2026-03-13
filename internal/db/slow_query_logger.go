package db

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

const slowQueryThreshold = 50 * time.Millisecond

// SlowQueryDB wraps a sql.DB to log queries that exceed the slow query threshold
type SlowQueryDB struct {
	*sql.DB
}

// slowQueryLogger logs a query if it took longer than the threshold
func slowQueryLogger(query string, startTime time.Time, args ...any) {
	duration := time.Since(startTime)
	if duration > slowQueryThreshold {
		slog.Debug("slow query detected",
			"duration_ms", duration.Milliseconds(),
			"query", query,
			"args", args,
		)
	}
}

// QueryContext wraps sql.DB.QueryContext with slow query logging
func (d *SlowQueryDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := d.DB.QueryContext(ctx, query, args...)
	slowQueryLogger(query, start, args...)
	return rows, err
}

// ExecContext wraps sql.DB.ExecContext with slow query logging
func (d *SlowQueryDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := d.DB.ExecContext(ctx, query, args...)
	slowQueryLogger(query, start, args...)
	return result, err
}

// PrepareContext wraps sql.DB.PrepareContext with slow query logging
func (d *SlowQueryDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	start := time.Now()
	stmt, err := d.DB.PrepareContext(ctx, query)
	slowQueryLogger(query, start)
	return stmt, err
}

// QueryRowContext wraps sql.DB.QueryRowContext with slow query logging
func (d *SlowQueryDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	row := d.DB.QueryRowContext(ctx, query, args...)
	slowQueryLogger(query, start, args...)
	return row
}

// WrapSlowQuery wraps a sql.DB with slow query logging if debug mode is enabled
func WrapSlowQuery(sqlDB *sql.DB, debugMode bool) *sql.DB {
	if !debugMode {
		return sqlDB
	}
	return &SlowQueryDB{DB: sqlDB}.DB
}
