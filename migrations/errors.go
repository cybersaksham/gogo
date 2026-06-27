package migrations

import "errors"

var ErrInvalidMigration = errors.New("invalid migration")

var ErrUnsafeMigration = errors.New("unsafe migration")
