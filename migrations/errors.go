package migrations

import "errors"

var ErrInvalidMigration = errors.New("invalid migration")

var ErrUnsafeMigration = errors.New("unsafe migration")

var ErrIrreversibleOperation = errors.New("irreversible operation")

var ErrDuplicateMigration = errors.New("duplicate migration")

var ErrMissingDependency = errors.New("missing migration dependency")

var ErrMigrationCycle = errors.New("migration dependency cycle")

var ErrInconsistentMigrationHistory = errors.New("inconsistent migration history")

var ErrMigrationLocked = errors.New("migration lock already held")
