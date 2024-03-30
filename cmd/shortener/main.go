package main

import (
	"log"

	"github.com/kTowkA/shortener/internal/app"
)

func main() {
	srv, err := app.NewServer()
	if err != nil {
		log.Fatal(err)
	}
	if err = srv.ListenAndServe(":8080"); err != nil {
		log.Fatal(err)
	}
}
