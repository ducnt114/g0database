package g0database

import "errors"

var (
	ErrDatabaseExists      = errors.New("database already exists")
	ErrDatabaseNotFound    = errors.New("database not found")
	ErrNoDatabaseSelected  = errors.New("no database selected")
	ErrTableExists         = errors.New("table already exists")
	ErrTableNotFound       = errors.New("table not found")
	ErrPrimaryKeyExists    = errors.New("duplicate primary key")
	ErrColumnCountMismatch = errors.New("column count mismatch")
	ErrColumnNotFound      = errors.New("column not found")
	ErrInvalidType         = errors.New("invalid value type for column")
	ErrNullNotAllowed      = errors.New("null value not allowed for column")
	ErrRowNotFound         = errors.New("row not found")
	ErrNoPrimaryKey        = errors.New("table has no primary key")
)
