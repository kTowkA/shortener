package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

const (
	EnvServerAddress = "SERVER_ADDRESS"
	EnvBaseURL       = "BASE_URL"
	EnvDSN           = "DATABASE_DSN"
	EnvSecretKey     = "SECRET_KEY"
)

var (
	flagA               string
	flagB               string
	flagStorageFilePath string
	flagDatabaseDSN     string
	logLevel            string
)

type Config struct {
	Address         string `env:"SERVER_ADDRESS"`
	BaseAddress     string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"/tmp/short-url-db.json"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	LogLevel        string
	SecretKey       string `env:"SECRET_KEY" envDefault:"my_super_secret_key"`
}

func ParseConfig() (Config, error) {
	flag.StringVar(&flagA, "a", "localhost:8080", "address:host")
	flag.StringVar(&flagB, "b", "http://localhost:8080", "result address")
	flag.StringVar(&flagDatabaseDSN, "d", "", "connect string. example postgres://username:password@localhost:5432/database_name")
	flag.StringVar(&flagStorageFilePath, "f", "/tmp/short-url-db.json", "file on disk with db")
	flag.StringVar(&logLevel, "l", "info", "level (panic,fatal,error,warn,info,debug,trace)")

	flag.Parse()

	cfg := Config{
		LogLevel: logLevel,
	}
	err := env.Parse(&cfg)
	if err != nil {
		return Config{}, err
	}

	if cfg.Address == "" {
		cfg.Address = flagA
	}

	if cfg.BaseAddress == "" {
		cfg.BaseAddress = flagB
	}
	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = flagDatabaseDSN
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = flagStorageFilePath
	}
	return cfg, nil
}
