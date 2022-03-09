package app

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/putalexey/go-practicum/cmd/shortener/config"
	"github.com/putalexey/go-practicum/internal/app/shortener"
	"github.com/putalexey/go-practicum/internal/app/storage"
)

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

func initStorage(cfg config.EnvConfig) (storage.Storager, error) {
	if cfg.FileStoragePath != "" {
		return storage.NewFileStorage(cfg.FileStoragePath)
	}
	if cfg.DatabaseDSN != "" {
		return storage.NewDBStorage(cfg.DatabaseDSN, "migrations")
	}

	return storage.NewMemoryStorage(nil), nil
}
