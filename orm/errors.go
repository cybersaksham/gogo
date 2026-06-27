package orm

import (
	"errors"
	"fmt"
)

var (
	ErrDoesNotExist              = errors.New("object does not exist")
	ErrMultipleObjectsReturned   = errors.New("multiple objects returned")
	ErrIntegrity                 = errors.New("integrity error")
	ErrInvalidQuery              = errors.New("invalid query")
	ErrUnsupportedDialectFeature = errors.New("unsupported dialect feature")
	ErrTransactionClosed         = errors.New("transaction closed")
	ErrNoRowsAffected            = errors.New("no rows affected")
)

// IntegrityError carries database integrity violation metadata.
type IntegrityError struct {
	Constraint string
	Err        error
}

func (e *IntegrityError) Error() string {
	if e.Constraint == "" {
		return ErrIntegrity.Error()
	}
	return fmt.Sprintf("%s: %s", ErrIntegrity, e.Constraint)
}

func (e *IntegrityError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}
	return ErrIntegrity
}

func (e *IntegrityError) Is(target error) bool {
	return target == ErrIntegrity
}

// QueryError carries operation metadata for invalid query failures.
type QueryError struct {
	Operation string
	Err       error
}

func (e *QueryError) Error() string {
	if e.Operation == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Operation, e.Err)
}

func (e *QueryError) Unwrap() error {
	return e.Err
}
