// Package app implements api service of shortening urls
package app

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/putalexey/go-practicum/cmd/shortener/config"
	_ "github.com/putalexey/go-practicum/internal/app/docs"
	"github.com/putalexey/go-practicum/internal/app/shortener"
	"github.com/putalexey/go-practicum/internal/app/storage"
)

// @title Shortener API
// @version 1.0
// @description API server for shorting log urls to short ones
// @BasePath /

// Run starts http server with shortener module as router. If ctx context is canceled,
// then http server will gracefully shutdown
func Run(ctx context.Context, cfg config.EnvConfig) {
	var err error

	store, err := initStorage(cfg)
	if err != nil {
		log.Fatal(err)
	}

	router := shortener.NewRouter(ctx, cfg.BaseURL, store)
	srv := http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
	}
	wg.Wait()
}

// initStorage initializes one of supported storagers
func initStorage(cfg config.EnvConfig) (storage.Storager, error) {
	if cfg.FileStoragePath != "" {
		return storage.NewFileStorage(cfg.FileStoragePath)
	}
	if cfg.DatabaseDSN != "" {
		return storage.NewDBStorage(cfg.DatabaseDSN, "migrations")
	}

	return storage.NewMemoryStorage(nil), nil
}
