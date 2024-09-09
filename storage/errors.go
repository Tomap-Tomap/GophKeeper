package storage

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// ErrUserNotFound is returned when a user is not found in the database.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when attempting to create a user that already exists.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrPasswordNotFound is returned when a password is not found in the database.
	ErrPasswordNotFound = errors.New("password not found")
	// ErrFileNotFound is returned when a file is not found in the storage.
	ErrFileNotFound = errors.New("file not found")
	// ErrBankNotFound is returned when a bank record is not found in the database.
	ErrBankNotFound = errors.New("bank not found")
	// ErrTextNotFound is returned when a text record is not found in the database.
	ErrTextNotFound = errors.New("text not found")
)

// IsUniqueViolation checks if the given error is of type pgconn.PgError and is a unique violation error.
func IsUniqueViolation(err error) bool {
	var tError *pgconn.PgError
	if errors.As(err, &tError) && tError.Code == pgerrcode.UniqueViolation {
		return true
	}

	return false
}

// IsNoRowError checks if the given error is of type pgx.ErrNoRows.
func IsNoRowError(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// IsForeignKeyViolation checks if the given error is of type pgconn.PgError and is a foreign key violation error.
func IsForeignKeyViolation(err error) bool {
	var tError *pgconn.PgError
	if errors.As(err, &tError) && tError.Code == pgerrcode.ForeignKeyViolation {
		return true
	}

	return false
}
