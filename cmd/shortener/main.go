package main

import (
	"flag"
	"log"

	"github.com/kTowkA/shortener/internal/app"
	"github.com/kTowkA/shortener/internal/config"
)

var (
	flagA string
	flagB string
)

func init() {
	flag.StringVar(&flagA, "a", "localhost:8080", "address:host")
	flag.StringVar(&flagB, "b", "http://localhost:8080", "result address")
}

func main() {
	flag.Parse()
	cfg, err := config.NewConfig(config.ConfigAddress(flagA), config.ConfigBaseAddress(flagB))
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
