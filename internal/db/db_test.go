package db_test

import (
	"testing"

	"github.com/anjakDev/hourglass/internal/db"
	"github.com/stretchr/testify/require"
)

func TestOpen_InMemory(t *testing.T) {
	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()
}

func TestOpen_TablesExist(t *testing.T) {
	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	defer conn.Close()

	tables := []string{"projects", "sessions", "breaks"}
	for _, table := range tables {
		var name string
		err := conn.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		require.NoError(t, err, "table %q should exist", table)
		require.Equal(t, table, name)
	}
}

func TestOpen_ForeignKeysEnabled(t *testing.T) {
	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	defer conn.Close()

	var fk int
	err = conn.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	require.NoError(t, err)
	require.Equal(t, 1, fk, "foreign keys should be enabled")
}
