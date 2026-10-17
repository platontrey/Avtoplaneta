package postgres

import (
	"context"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("postgres://user:pass@localhost:5432/testdb")
	require.Equal(t, "postgres://user:pass@localhost:5432/testdb", cfg.DatabaseURL)
	require.Equal(t, int32(50), cfg.MaxConns)
	require.Equal(t, int32(5), cfg.MinConns)
	require.Equal(t, 30*time.Minute, cfg.MaxConnLifetime)
	require.Equal(t, 5*time.Minute, cfg.MaxConnIdleTime)
}

func TestRunMigrations_NilPool(t *testing.T) {
	fsys := fstest.MapFS{}
	err := RunMigrations(context.Background(), nil, fsys, "non_existent_dir")
	require.Error(t, err)
	require.Contains(t, err.Error(), "postgres pool cannot be nil")
}
