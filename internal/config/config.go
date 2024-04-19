package config

const (
	EnvServerAddress = "SERVER_ADDRESS"
	EnvBaseURL       = "BASE_URL"
)

type Config struct {
	Address         string `env:"SERVER_ADDRESS"`
	BaseAddress     string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"/tmp/short-url-db.json"`
	LogLevel        string
}

type ConfigParam func(c *Config)

func ConfigAddress(address string) ConfigParam {
	return func(c *Config) {
		c.Address = address
	}
}
func ConfigBaseAddress(baseAddress string) ConfigParam {
	return func(c *Config) {
		c.BaseAddress = baseAddress
	}
}
func ConfigFileSstoragePath(fileStoragePath string) ConfigParam {
	return func(c *Config) {
		c.FileStoragePath = fileStoragePath
	}
}
func NewConfig(configs ...ConfigParam) (Config, error) {
	cfg := new(Config)
	for _, c := range configs {
		c(cfg)
	}
	return *cfg, nil
}
