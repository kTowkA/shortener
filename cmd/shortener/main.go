package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/kTowkA/shortener/internal/app"
	"github.com/kTowkA/shortener/internal/config"
)

var (
	flagA               string
	flagB               string
	flagStorageFilePath string
	logLevel            string
)

func init() {
	flag.StringVar(&flagA, "a", "localhost:8080", "address:host")
	flag.StringVar(&flagB, "b", "http://localhost:8080", "result address")
	flag.StringVar(&flagStorageFilePath, "f", "/tmp/short-url-db.json", "file on disk with db")
	flag.StringVar(&logLevel, "l", "info", "level (panic,fatal,error,warn,info,debug,trace)")
}

func main() {
	cfg, err := configurate()
	if err != nil {
		log.Fatal(err)
	}
	srv, err := app.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err = srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func configurate() (config.Config, error) {
	flag.Parse()

	cfg := config.Config{
		LogLevel: logLevel,
	}
	err := env.Parse(&cfg)
	if err != nil {
		return config.Config{}, err
	}

	if cfg.Address == "" {
		cfg.Address = flagA
	}

	if cfg.BaseAddress == "" {
		cfg.BaseAddress = flagB
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = flagStorageFilePath
	}
	return cfg, nil
}
