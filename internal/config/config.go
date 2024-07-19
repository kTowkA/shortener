// модуль config отвечает за создание экземпляра конфигурации из переменных окружения, флагов запуска и значений по умолчанию
package config

import (
	"flag"
	"fmt"
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v6"
)

const (
	defaultSecretKey = "super strong secret key"

	defaultAddress         = "localhost:8080"
	defaultBaseAddress     = "http://localhost:8080/"
	defaultStorageFilePath = "/tmp/short-url-db.json"
)

var (
	flagA               string
	flagB               string
	flagStorageFilePath string
	flagDatabaseDSN     string
)

// Config конфигурация приложения
type Config struct {
	address         string
	baseAddress     string
	fileStoragePath string
	databaseDSN     string
	secretKey       string
}

// Address возвращает строку с адресом запускаемого сервера
func (c *Config) Address() string {
	return c.address
}

// BaseAddress возвращает строку с базовым адресом для сокращения ссылок
func (c *Config) BaseAddress() string {
	return c.baseAddress
}

// FileStoragePath возвращает строку с файлом-хранилищем
func (c *Config) FileStoragePath() string {
	return c.fileStoragePath
}

// DatabaseDSN возвращает строку для подключения к БД
func (c *Config) DatabaseDSN() string {
	return c.databaseDSN
}

// SecretKey возвращает строку содержащую секретный ключ
func (c *Config) SecretKey() string {
	return c.secretKey
}

// DefaultConfig конфигурация по умолчанию для быстрой настройки
var DefaultConfig = Config{
	address:         defaultAddress,
	baseAddress:     defaultBaseAddress,
	fileStoragePath: defaultStorageFilePath,
	secretKey:       defaultSecretKey,
}

// ParseConfig запускает создание конфигурации читая значения переменных окружения и флагов командной строки
func ParseConfig(logger *slog.Logger) (Config, error) {
	flag.StringVar(&flagA, "a", defaultAddress, "address:host")
	flag.StringVar(&flagB, "b", defaultBaseAddress, "result address")
	flag.StringVar(&flagDatabaseDSN, "d", "", "connect string. example postgres://username:password@localhost:5432/database_name")
	flag.StringVar(&flagStorageFilePath, "f", defaultStorageFilePath, "file on disk with db")

	flag.Parse()

	type PublicConfig struct {
		Address         string `env:"SERVER_ADDRESS"`
		BaseAddress     string `env:"BASE_URL"`
		FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"/tmp/short-url-db.json"`
		DatabaseDSN     string `env:"DATABASE_DSN"`
		SecretKey       string `env:"SECRET_KEY" envDefault:"my_super_secret_key"`
	}

	cfg := PublicConfig{}
	err := env.Parse(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("сопостовление переменных окружения с объектом конфигурации. %w", err)
	}

	if cfg.Address == "" {
		logger.Debug("адрес приложения. используется значение флага")
		cfg.Address = flagA
	}

	if cfg.BaseAddress == "" {
		logger.Debug("базовый адрес шортера. используется значение флага")
		cfg.BaseAddress = flagB
	}
	cfg.BaseAddress = strings.TrimSuffix(cfg.BaseAddress, "/") + "/"

	if cfg.DatabaseDSN == "" {
		logger.Debug("строка соединения с БД. используется значение флага")
		cfg.DatabaseDSN = flagDatabaseDSN
	}
	if cfg.FileStoragePath == "" {
		logger.Debug("путь к файлу-хранилищу. используется значение флага")
		cfg.FileStoragePath = flagStorageFilePath
	}
	if cfg.SecretKey == "" {
		cfg.SecretKey = defaultSecretKey
	}
	logger.Debug("конфигурация",
		slog.String("адрес", cfg.Address),
		slog.String("базовый адрес", cfg.BaseAddress),
		slog.String("путь к файлу-хранилищу", cfg.FileStoragePath),
		slog.String("строка соединения с БД", cfg.DatabaseDSN),
	)
	return Config{
		address:         cfg.Address,
		baseAddress:     cfg.BaseAddress,
		fileStoragePath: cfg.FileStoragePath,
		databaseDSN:     cfg.DatabaseDSN,
		secretKey:       cfg.SecretKey,
	}, nil
}
