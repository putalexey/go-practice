package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/putalexey/go-practicum/internal/app/shortener"
	"github.com/putalexey/go-practicum/internal/app/storage"
	"log"
	"net/http"
)

type EnvConfig struct {
	Address         string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func main() {
	cfg := EnvConfig{
		Address:         ":8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	var store storage.Storager
	if cfg.FileStoragePath != "" {
		if store, err = storage.NewFileStorage(cfg.FileStoragePath); err != nil {
			log.Fatal(err)
		}
	}

	router := shortener.NewRouter(cfg.BaseURL, store)
	log.Fatal(http.ListenAndServe(cfg.Address, router))
}
