package orm

import (
	"context"
	"database/sql"
	"fmt"
)

// RowScanner is implemented by sql row/rows scanners.
type RowScanner interface {
	Scan(dest ...any) error
}

// RawExecutor executes raw SQL against a database.
type RawExecutor struct {
	Database *Database
}

// RawExecResult stores raw exec metadata.
type RawExecResult struct {
	RowsAffected int64
	LastInsertID int64
}

// NewRawExecutor creates a raw SQL executor.
func NewRawExecutor(database *Database) RawExecutor {
	return RawExecutor{Database: database}
}

// ParameterizedRaw creates a raw query with bound parameters.
func ParameterizedRaw(sql string, args ...any) RawQuery {
	return RawQuery{SQL: sql, Args: append([]any(nil), args...)}
}

// UnsafeRawQuery creates an explicitly marked unparameterized raw query.
func UnsafeRawQuery(sql string, args ...any) RawQuery {
	return RawQuery{SQL: sql, Args: append([]any(nil), args...), Unsafe: true}
}

// Exec executes a raw SQL statement.
func (e RawExecutor) Exec(ctx context.Context, query RawQuery) (RawExecResult, error) {
	if err := ValidateRawQuery(query); err != nil {
		return RawExecResult{}, err
	}
	if e.Database == nil || e.Database.SQLDB() == nil {
		return RawExecResult{}, ErrDatabaseNotFound
	}
	result, err := e.Database.SQLDB().ExecContext(ctx, query.SQL, query.Args...)
	if err != nil {
		return RawExecResult{}, err
	}
	rowsAffected, _ := result.RowsAffected()
	lastInsertID, _ := result.LastInsertId()
	return RawExecResult{RowsAffected: rowsAffected, LastInsertID: lastInsertID}, nil
}

// Query executes a raw SELECT statement.
func (e RawExecutor) Query(ctx context.Context, query RawQuery) (*sql.Rows, error) {
	if err := ValidateRawQuery(query); err != nil {
		return nil, err
	}
	if e.Database == nil || e.Database.SQLDB() == nil {
		return nil, ErrDatabaseNotFound
	}
	return e.Database.SQLDB().QueryContext(ctx, query.SQL, query.Args...)
}

// RawSelect maps raw SELECT rows to values.
func RawSelect[T any](ctx context.Context, executor RawExecutor, query RawQuery, mapper func(RowScanner) (T, error)) ([]T, error) {
	rows, err := executor.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []T
	for rows.Next() {
		value, err := mapper(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

// ValidateRawQuery rejects accidental unparameterized raw SQL.
func ValidateRawQuery(query RawQuery) error {
	if query.SQL == "" {
		return fmt.Errorf("%w: raw SQL is required", ErrInvalidQuery)
	}
	if len(query.Args) == 0 && !query.Unsafe {
		return ErrUnsafeRawSQL
	}
	return nil
}
