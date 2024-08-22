// модуль config отвечает за создание экземпляра конфигурации из переменных окружения, флагов запуска и значений по умолчанию
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net"
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
	flagConfig          string
	flagTrustedSubnet   string
	flagEnableHTTPS     bool
)

// Config конфигурация приложения
type Config struct {
	address         string
	baseAddress     string
	fileStoragePath string
	databaseDSN     string
	secretKey       string
	trustedSubnet   *net.IPNet
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

// TrustedSubnet возвращает строковое представление бесклассовой адресации (CIDR)
func (c *Config) TrustedSubnet() *net.IPNet {
	return c.trustedSubnet
}

// DefaultConfig конфигурация по умолчанию для быстрой настройки
var DefaultConfig = Config{
	address:         defaultAddress,
	baseAddress:     defaultBaseAddress,
	fileStoragePath: defaultStorageFilePath,
	secretKey:       defaultSecretKey,
	trustedSubnet:   &net.IPNet{},
	configHTTPS: configHTTPS{
		enable: false,
		domain: "",
	},
}

func init() {
	flag.StringVar(&flagA, "a", "", "address:host")
	flag.StringVar(&flagB, "b", "", "result address")
	flag.StringVar(&flagDatabaseDSN, "d", "", "connect string. example postgres://username:password@localhost:5432/database_name")
	flag.StringVar(&flagStorageFilePath, "f", "", "file on disk with db")
	flag.StringVar(&flagDomainName, "dn", "", "domain name")
	flag.StringVar(&flagConfig, "c", "", "config file(only JSON)")
	flag.StringVar(&flagTrustedSubnet, "t", "", "trusted subnet")
	flag.BoolVar(&flagEnableHTTPS, "s", false, "enable https")
}

// ParseConfig запускает создание конфигурации читая значения переменных окружения и флагов командной строки
func ParseConfig(logger *slog.Logger) (Config, error) {
	flag.Parse()

	type PublicConfig struct {
		Address         string `env:"SERVER_ADDRESS" json:"server_address"`
		BaseAddress     string `env:"BASE_URL" json:"base_url"`
		FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
		DatabaseDSN     string `env:"DATABASE_DSN" json:"database_dsn"`
		SecretKey       string `env:"SECRET_KEY" envDefault:"my_super_secret_key"`
		Config          string `env:"CONFIG"`
		DomainName      string `env:"DOMAIN" json:"domain_name"`
		TrustedSubnet   string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
		EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	}

	cfg := PublicConfig{}

	// чтобы работало ENABLE_HTTPS= как валидная установка значения. Иначе пакет github.com/caarlos0/env/v6 говорит, что ENABLE_HTTPS false
	if val, ok := os.LookupEnv("ENABLE_HTTPS"); ok && val == "" {
		os.Setenv("ENABLE_HTTPS", "true")
	}

	err := env.Parse(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("сопостовление переменных окружения с объектом конфигурации. %w", err)
	}

	// работаем с файлом конфигурации если он установлен
	cfg.Config = getConfigValue(cfg.Config, flagConfig, "", "", "")
	cfgFromFile := PublicConfig{}
	if cfg.Config != "" {
		err = parseConfigFromFile(&cfgFromFile, cfg.Config)
		if err != nil {
			return Config{}, fmt.Errorf("чтение файла конфигурации. %w", err)
		}
	}

	// устанавливаем окончательные значения конфигурации выбирая из переменных окружения, флагов командной строки, значений из файла или значений по умолчанию
	cfg.Address = getConfigValue(cfg.Address, flagA, cfgFromFile.Address, defaultAddress, "")
	cfg.BaseAddress = getConfigValue(cfg.BaseAddress, flagB, cfgFromFile.BaseAddress, defaultBaseAddress, "")
	cfg.BaseAddress = strings.TrimSuffix(cfg.BaseAddress, "/") + "/"
	cfg.DatabaseDSN = getConfigValue(cfg.DatabaseDSN, flagDatabaseDSN, cfgFromFile.DatabaseDSN, "", "")
	cfg.FileStoragePath = getConfigValue(cfg.FileStoragePath, flagStorageFilePath, cfgFromFile.FileStoragePath, defaultStorageFilePath, "")
	cfg.SecretKey = getConfigValue(cfg.SecretKey, "", "", defaultSecretKey, "")
	cfg.DomainName = getConfigValue(cfg.DomainName, flagDomainName, cfgFromFile.DomainName, "", "")
	cfg.EnableHTTPS = getConfigValue(cfg.EnableHTTPS, flagEnableHTTPS, cfgFromFile.EnableHTTPS, false, false)

	cfg.TrustedSubnet = getConfigValue(cfg.TrustedSubnet, flagTrustedSubnet, cfgFromFile.TrustedSubnet, "", "")
	_, ipnet, err := net.ParseCIDR(cfg.TrustedSubnet)
	// не стал выходить из функции, просто выведем ошибку
	if err != nil {
		logger.Error("конвертация CIDR", slog.String("ошибка", err.Error()))
		ipnet = &net.IPNet{}
	}

	if cfg.EnableHTTPS {
		cfg.Address = addressForHTTPS(cfg.Address)
	}

	logger.Debug("конфигурация",
		slog.String("файл конфигурации", cfg.Config),
		slog.String("адрес", cfg.Address),
		slog.String("базовый адрес", cfg.BaseAddress),
		slog.String("путь к файлу-хранилищу", cfg.FileStoragePath),
		slog.String("строка соединения с БД", cfg.DatabaseDSN),
		slog.Bool("статус https", cfg.EnableHTTPS),
		slog.String("доменное имя", cfg.DomainName),
		slog.String("CIDR", cfg.TrustedSubnet),
	)
	return Config{
		address:         cfg.Address,
		baseAddress:     cfg.BaseAddress,
		fileStoragePath: cfg.FileStoragePath,
		databaseDSN:     cfg.DatabaseDSN,
		secretKey:       cfg.SecretKey,
		trustedSubnet:   ipnet,
		configHTTPS: configHTTPS{
			enable: cfg.EnableHTTPS,
			domain: cfg.DomainName,
		},
	}, nil
}

// getConfigValue возвращает нужное для установки значение конфигурации
// приоритет envValue,далее flagValue, следом fileValue и значение по умолчанию defaultValue
// notSetValue служит для проверки что значение не установлено
func getConfigValue[T comparable](envValue, flagValue, fileValue, defaultValue, notSetValue T) T {
	if envValue == notSetValue {
		if flagValue == notSetValue {
			if fileValue != notSetValue {
				return fileValue
			}
			return defaultValue
		}
		return flagValue
	}
	return envValue
}

// parseConfigFromFile читаем конфигурацию из файла filename, publicConfig должны передать указать на структуру PublicConfig из ParseConfig
// не объявляю PublicConfig вне фукнции ParseConfig, чтобы у пользователя не было соблазна ее использовать
func parseConfigFromFile(publicConfig any, filename string) error {
	configFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("невозможно прочитать файл %s. %w", filename, err)
	}
	defer configFile.Close()
	return json.NewDecoder(configFile).Decode(publicConfig)
}

func addressForHTTPS(address string) string {
	if strings.HasPrefix(address, "http://") {
		address = "https://" + strings.TrimPrefix(address, "http://")
	}
	adrSl := strings.Split(address, ":")
	if len(adrSl) == 1 {
		address += ":443"
	} else {
		adrSl[len(adrSl)-1] = "443"
		address = strings.Join(adrSl, ":")
	}
	return address
}
