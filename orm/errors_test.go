package orm

import (
	"errors"
	"fmt"
	"testing"
)

func TestORMErrorSentinelsSupportErrorsIs(t *testing.T) {
	sentinels := []error{
		ErrDoesNotExist,
		ErrMultipleObjectsReturned,
		ErrIntegrity,
		ErrInvalidQuery,
		ErrUnsupportedDialectFeature,
		ErrTransactionClosed,
		ErrNoRowsAffected,
	}
	for _, sentinel := range sentinels {
		if !errors.Is(fmt.Errorf("wrapped: %w", sentinel), sentinel) {
			t.Fatalf("errors.Is failed for %v", sentinel)
		}
	}
}

func TestTypedORMErrorsSupportErrorsAs(t *testing.T) {
	err := &IntegrityError{Constraint: "uniq_slug", Err: errors.New("duplicate key")}
	if !errors.Is(err, ErrIntegrity) {
		t.Fatalf("IntegrityError does not match ErrIntegrity")
	}
	var integrity *IntegrityError
	if !errors.As(err, &integrity) || integrity.Constraint != "uniq_slug" {
		t.Fatalf("errors.As integrity = %#v", integrity)
	}

	queryErr := &QueryError{Operation: "select", Err: ErrInvalidQuery}
	var typedQuery *QueryError
	if !errors.As(queryErr, &typedQuery) || typedQuery.Operation != "select" {
		t.Fatalf("errors.As query = %#v", typedQuery)
	}
	if !errors.Is(queryErr, ErrInvalidQuery) {
		t.Fatalf("QueryError does not unwrap ErrInvalidQuery")
	}
}
