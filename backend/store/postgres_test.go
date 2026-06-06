package store

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsUniqueViolation(t *testing.T) {
	t.Run("detects postgres unique violation", func(t *testing.T) {
		assert.True(t, isUniqueViolation(&pgconn.PgError{Code: "23505"}))
	})

	t.Run("ignores other postgres errors", func(t *testing.T) {
		assert.False(t, isUniqueViolation(&pgconn.PgError{Code: "23503"}))
	})

	t.Run("ignores non-postgres errors", func(t *testing.T) {
		assert.False(t, isUniqueViolation(assert.AnError))
	})
}
