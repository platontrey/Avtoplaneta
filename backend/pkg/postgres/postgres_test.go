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

func TestRunInTransaction_NilPool(t *testing.T) {
	called := false
	err := RunInTransaction(context.Background(), nil, func(ctx context.Context) error {
		called = true
		return nil
	})
	require.NoError(t, err)
	require.True(t, called)
}

func TestTxFromContext_Empty(t *testing.T) {
	tx, ok := TxFromContext(context.Background())
	require.False(t, ok)
	require.Nil(t, tx)
}

