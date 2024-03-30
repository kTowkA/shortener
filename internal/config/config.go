package config

const (
	ENV_SERVER_ADDRESS = "SERVER_ADDRESS"
	ENV_BASE_URL       = "BASE_URL"
)

type Config struct {
	Address     string
	BaseAddress string
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
