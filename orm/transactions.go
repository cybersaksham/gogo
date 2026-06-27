package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"
)

type transactionContextKey struct{}

var ErrTransactionClosed = errors.New("transaction closed")

// TxOption mutates sql transaction options.
type TxOption func(*sql.TxOptions)

// WithIsolation sets transaction isolation level.
func WithIsolation(level sql.IsolationLevel) TxOption {
	return func(options *sql.TxOptions) {
		options.Isolation = level
	}
}

// ReadOnly marks a transaction read-only.
func ReadOnly() TxOption {
	return func(options *sql.TxOptions) {
		options.ReadOnly = true
	}
}

// TransactionManager coordinates root and nested transactions.
type TransactionManager struct {
	Database *Database
	counter  atomic.Uint64
}

// NewTransactionManager creates a manager for one database.
func NewTransactionManager(database *Database) *TransactionManager {
	return &TransactionManager{Database: database}
}

// Transaction wraps a sql.Tx and on-commit callbacks.
type Transaction struct {
	manager   *TransactionManager
	tx        *sql.Tx
	parent    *Transaction
	savepoint string
	callbacks []func()
	closed    bool
}

// ExecContext executes SQL inside the transaction.
func (t *Transaction) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if t.closed {
		return nil, ErrTransactionClosed
	}
	return t.tx.ExecContext(ctx, query, args...)
}

// QueryContext queries SQL inside the transaction.
func (t *Transaction) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if t.closed {
		return nil, ErrTransactionClosed
	}
	return t.tx.QueryContext(ctx, query, args...)
}

// QueryRowContext queries one row inside the transaction.
func (t *Transaction) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// OnCommit registers a callback to run after the outer transaction commits.
func (t *Transaction) OnCommit(callback func()) {
	t.callbacks = append(t.callbacks, callback)
}

// Atomic runs a function inside an atomic transaction.
func (m *TransactionManager) Atomic(ctx context.Context, fn func(context.Context, *Transaction) error, options ...TxOption) (err error) {
	if parent, ok := CurrentTransaction(ctx); ok {
		return m.atomicNested(ctx, parent, fn)
	}
	return m.atomicRoot(ctx, fn, options...)
}

// CurrentTransaction returns the transaction stored in context.
func CurrentTransaction(ctx context.Context) (*Transaction, bool) {
	tx, ok := ctx.Value(transactionContextKey{}).(*Transaction)
	return tx, ok
}

func (m *TransactionManager) atomicRoot(ctx context.Context, fn func(context.Context, *Transaction) error, options ...TxOption) (err error) {
	if m.Database == nil || m.Database.SQLDB() == nil {
		return ErrDatabaseNotFound
	}
	txOptions := &sql.TxOptions{}
	for _, option := range options {
		option(txOptions)
	}
	sqlTx, err := m.Database.SQLDB().BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	tx := &Transaction{manager: m, tx: sqlTx}
	txCtx := context.WithValue(ctx, transactionContextKey{}, tx)

	defer func() {
		if recovered := recover(); recovered != nil {
			_ = sqlTx.Rollback()
			tx.closed = true
			panic(recovered)
		}
	}()

	if err := fn(txCtx, tx); err != nil {
		_ = sqlTx.Rollback()
		tx.closed = true
		return err
	}
	if err := sqlTx.Commit(); err != nil {
		tx.closed = true
		return err
	}
	tx.closed = true
	for _, callback := range tx.callbacks {
		callback()
	}
	return nil
}

func (m *TransactionManager) atomicNested(ctx context.Context, parent *Transaction, fn func(context.Context, *Transaction) error) (err error) {
	name := fmt.Sprintf("sp_%d", m.counter.Add(1))
	if _, err := parent.tx.ExecContext(ctx, m.Database.Dialect.SavepointSQL(name)); err != nil {
		return err
	}
	child := &Transaction{manager: m, tx: parent.tx, parent: parent, savepoint: name}
	childCtx := context.WithValue(ctx, transactionContextKey{}, child)

	defer func() {
		if recovered := recover(); recovered != nil {
			_ = child.rollbackToSavepoint(ctx)
			child.closed = true
			panic(recovered)
		}
	}()

	if err := fn(childCtx, child); err != nil {
		_ = child.rollbackToSavepoint(ctx)
		child.closed = true
		return err
	}
	if _, err := parent.tx.ExecContext(ctx, m.Database.Dialect.ReleaseSavepointSQL(name)); err != nil {
		child.closed = true
		return err
	}
	child.closed = true
	parent.callbacks = append(parent.callbacks, child.callbacks...)
	return nil
}

func (t *Transaction) rollbackToSavepoint(ctx context.Context) error {
	if _, err := t.tx.ExecContext(ctx, t.manager.Database.Dialect.RollbackToSavepointSQL(t.savepoint)); err != nil {
		return err
	}
	_, err := t.tx.ExecContext(ctx, t.manager.Database.Dialect.ReleaseSavepointSQL(t.savepoint))
	return err
}
