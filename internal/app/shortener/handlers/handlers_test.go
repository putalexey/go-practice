package handlers_test

import (
	"github.com/go-chi/chi/v5"
	"github.com/putalexey/go-practicum/internal/app/shortener/handlers"
	"github.com/putalexey/go-practicum/internal/app/storage"
	"net/http"
)

func Example() {
	mux := chi.NewMux()
	store := storage.NewMemoryStorage(nil)

	mux.Get("/ping", handlers.PingHandler(store))
	mux.Get("/{id}", handlers.GetFullURLHandler(store))

	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
