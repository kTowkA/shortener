package config

const (
	EnvServerAddress = "SERVER_ADDRESS"
	EnvBaseURL       = "BASE_URL"
)

type Config struct {
	Address     string `env:"SERVER_ADDRESS"`
	BaseAddress string `env:"BASE_URL"`
	LogLevel    string
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
func NewConfig(configs ...ConfigParam) (Config, error) {
	cfg := new(Config)
	for _, c := range configs {
		c(cfg)
	}
	return *cfg, nil
}
