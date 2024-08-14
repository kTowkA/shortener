package config

import (
	"log/slog"
	"os"
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

func TestConfigEnv(t *testing.T) {
	var (
		domain        = "test_domain"
		serverAddress = "test_server_address"
		baseURL       = "test_base_url"
		fileStorage   = "test_file_storage"
		database      = "test_database"
		secretKey     = "test_secret_key"
		enableHTTPS   = "true"
	)
	os.Setenv("DOMAIN", domain)
	os.Setenv("ENABLE_HTTPS", enableHTTPS)
	os.Setenv("SECRET_KEY", secretKey)
	os.Setenv("DATABASE_DSN", database)
	os.Setenv("FILE_STORAGE_PATH", fileStorage)
	os.Setenv("BASE_URL", baseURL)
	os.Setenv("SERVER_ADDRESS", serverAddress)
	cfg, err := ParseConfig(slog.Default())
	require.NoError(t, err)
	assert.EqualValues(t, domain, cfg.Domain())
	assert.EqualValues(t, serverAddress+":443", cfg.Address())
	assert.EqualValues(t, baseURL+"/", cfg.BaseAddress())
	assert.EqualValues(t, fileStorage, cfg.FileStoragePath())
	assert.EqualValues(t, database, cfg.DatabaseDSN())
	assert.EqualValues(t, secretKey, cfg.SecretKey())
	assert.True(t, cfg.HTTPS())

	os.Setenv("ENABLE_HTTPS", "")
	cfg, err = ParseConfig(slog.Default())
	require.NoError(t, err)
	assert.True(t, cfg.HTTPS())

	os.Setenv("ENABLE_HTTPS", "t")
	cfg, err = ParseConfig(slog.Default())
	require.NoError(t, err)
	assert.True(t, cfg.HTTPS())

	os.Setenv("ENABLE_HTTPS", "1")
	cfg, err = ParseConfig(slog.Default())
	require.NoError(t, err)
	assert.True(t, cfg.HTTPS())

	os.Setenv("ENABLE_HTTPS", "false")
	cfg, err = ParseConfig(slog.Default())
	require.NoError(t, err)
	assert.False(t, cfg.HTTPS())

	os.Setenv("ENABLE_HTTPS", "f")
	cfg, err = ParseConfig(slog.Default())
	require.NoError(t, err)
	assert.False(t, cfg.HTTPS())

	os.Setenv("ENABLE_HTTPS", "0")
	cfg, err = ParseConfig(slog.Default())
	require.NoError(t, err)
	assert.False(t, cfg.HTTPS())
}

func TestHTTPS(t *testing.T) {
	os.Setenv("SERVER_ADDRESS", "http://localhost:80")
	cfg, err := ParseConfig(slog.Default())
	require.NoError(t, err)
	assert.EqualValues(t, "http://localhost:80", cfg.Address())
	os.Setenv("ENABLE_HTTPS", "")
	cfg, err = ParseConfig(slog.Default())
	require.NoError(t, err)
	assert.EqualValues(t, "https://localhost:443", cfg.Address())
	os.Setenv("SERVER_ADDRESS", "localhost:80")
	cfg, err = ParseConfig(slog.Default())
	require.NoError(t, err)
	assert.EqualValues(t, "localhost:443", cfg.Address())
}
