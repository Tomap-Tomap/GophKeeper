package storage

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueViolation проверяет, что переданная ошибка имеет тип pgconn.PgError и это ошибка уникальности
func IsUniqueViolation(err error) bool {
	var tError *pgconn.PgError
	if errors.As(err, &tError) && tError.Code == pgerrcode.UniqueViolation {
		return true
	}

	return false
}

// IsNowRowError проверяет, что переданная ошибка имеет тип pgx.ErrNoRows
func IsNowRowError(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// IsForeignKeyViolation проверяет, что переданная ошибка имеет тип pgconn.PgError и это ошибка нарушения
// внешнего ключа
func IsForeignKeyViolation(err error) bool {
	var tError *pgconn.PgError
	if errors.As(err, &tError) && tError.Code == pgerrcode.ForeignKeyViolation {
		return true
	}

	return false
}
