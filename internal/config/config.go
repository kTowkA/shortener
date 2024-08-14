// модуль config отвечает за создание экземпляра конфигурации из переменных окружения, флагов запуска и значений по умолчанию
package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
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
	flagDomainName      string
	flagEnableHTTPS     bool
)

// Config конфигурация приложения
type Config struct {
	address         string
	baseAddress     string
	fileStoragePath string
	databaseDSN     string
	secretKey       string
	configHTTPS
}

type configHTTPS struct {
	enable bool
	domain string
}

// Domain возвращает доменное имя, если оно было установлено
func (c *Config) Domain() string {
	return c.configHTTPS.domain
}

// HTTPS возвращает true если в настройках установлено, что требуется поднять https-соединение, в противном случае - false
func (c *Config) HTTPS() bool {
	return c.configHTTPS.enable
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
	configHTTPS: configHTTPS{
		enable: false,
		domain: "",
	},
}

func init() {
	flag.StringVar(&flagA, "a", defaultAddress, "address:host")
	flag.StringVar(&flagB, "b", defaultBaseAddress, "result address")
	flag.StringVar(&flagDatabaseDSN, "d", "", "connect string. example postgres://username:password@localhost:5432/database_name")
	flag.StringVar(&flagStorageFilePath, "f", defaultStorageFilePath, "file on disk with db")
	flag.StringVar(&flagDomainName, "dn", "", "domain name")
	flag.BoolVar(&flagEnableHTTPS, "s", false, "enable https")
}

// ParseConfig запускает создание конфигурации читая значения переменных окружения и флагов командной строки
func ParseConfig(logger *slog.Logger) (Config, error) {
	flag.Parse()

	type PublicConfig struct {
		Address         string `env:"SERVER_ADDRESS"`
		BaseAddress     string `env:"BASE_URL"`
		FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"/tmp/short-url-db.json"`
		DatabaseDSN     string `env:"DATABASE_DSN"`
		SecretKey       string `env:"SECRET_KEY" envDefault:"my_super_secret_key"`
		EnableHTTPS     bool   `env:"ENABLE_HTTPS"`
		DomainName      string `env:"DOMAIN"`
	}

	cfg := PublicConfig{}

	// чтобы работало ENABLE_HTTPS= как валидная установка значения
	if val, ok := os.LookupEnv("ENABLE_HTTPS"); ok && val == "" {
		os.Setenv("ENABLE_HTTPS", "true")
	}

	err := env.Parse(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("сопостовление переменных окружения с объектом конфигурации. %w", err)
	}

	cfg.Address = setConfigValue(cfg.Address, flagA, defaultAddress)

	cfg.BaseAddress = setConfigValue(cfg.BaseAddress, flagB, defaultBaseAddress)
	cfg.BaseAddress = strings.TrimSuffix(cfg.BaseAddress, "/") + "/"

	cfg.DatabaseDSN = setConfigValue(cfg.DatabaseDSN, flagDatabaseDSN, "")
	cfg.FileStoragePath = setConfigValue(cfg.FileStoragePath, flagStorageFilePath, defaultStorageFilePath)
	cfg.SecretKey = setConfigValue(cfg.SecretKey, "", defaultSecretKey)
	cfg.DomainName = setConfigValue(cfg.DomainName, flagDomainName, "")

	if cfg.EnableHTTPS {
		if strings.HasPrefix(cfg.Address, "http://") {
			cfg.Address = "https://" + strings.TrimPrefix(cfg.Address, "http://")
		}
		adrSl := strings.Split(cfg.Address, ":")
		if len(adrSl) == 1 {
			cfg.Address += ":443"
		} else {
			adrSl[len(adrSl)-1] = "443"
			cfg.Address = strings.Join(adrSl, ":")
		}
		logger.Info("установлен HTTPS. Установили порт на 443")
	}

	logger.Debug("конфигурация",
		slog.String("адрес", cfg.Address),
		slog.String("базовый адрес", cfg.BaseAddress),
		slog.String("путь к файлу-хранилищу", cfg.FileStoragePath),
		slog.String("строка соединения с БД", cfg.DatabaseDSN),
		slog.Bool("статус https", cfg.EnableHTTPS),
		slog.String("доменное имя", cfg.DomainName),
	)
	return Config{
		address:         cfg.Address,
		baseAddress:     cfg.BaseAddress,
		fileStoragePath: cfg.FileStoragePath,
		databaseDSN:     cfg.DatabaseDSN,
		secretKey:       cfg.SecretKey,
		configHTTPS: configHTTPS{
			enable: cfg.EnableHTTPS,
			domain: cfg.DomainName,
		},
	}, nil
}

func setConfigValue(envValue, flagValue, defaultValue string) string {
	if envValue == "" {
		if flagValue != "" {
			return flagValue
		}
		return defaultValue
	}
	return envValue
}
