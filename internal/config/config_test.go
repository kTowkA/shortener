package config

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	cfg, err := ParseConfig(slog.Default())
	require.NoError(t, err)
	assert.EqualValues(t, defaultAddress, cfg.Address())
	assert.EqualValues(t, defaultBaseAddress, cfg.BaseAddress())
	assert.EqualValues(t, defaultStorageFilePath, cfg.FileStoragePath())
}
