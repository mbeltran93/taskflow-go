// Package postgres implements the repository interfaces against a real
// PostgreSQL database using sqlx (a thin, idiomatic layer over database/sql
// that adds struct scanning without hiding SQL behind a query builder).
package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL error codes used to translate driver errors into apperror sentinels.
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

func isUniqueViolation(err error) bool {
	return pgErrorCode(err) == pgUniqueViolation
}

func isForeignKeyViolation(err error) bool {
	return pgErrorCode(err) == pgForeignKeyViolation
}

func pgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
