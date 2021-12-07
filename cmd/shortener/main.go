package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/putalexey/go-practicum/internal/app/shortener"
	"log"
	"net/http"
)

type EnvConfig struct {
	Address string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

func main() {
	cfg := EnvConfig{
		Address: ":8080",
		BaseURL: "http://localhost:8080",
	}
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	router := shortener.NewRouter(cfg.BaseURL, nil)
	log.Fatal(http.ListenAndServe(cfg.Address, router))
}
