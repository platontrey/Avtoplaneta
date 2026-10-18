package grpcclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_Options(t *testing.T) {
	conn, err := NewClient("passthrough://127.0.0.1:9099")
	require.NoError(t, err)
	defer conn.Close()

	assert.NotNil(t, conn)
}
