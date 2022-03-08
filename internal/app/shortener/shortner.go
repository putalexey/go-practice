package shortener

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	appMiddleware "github.com/putalexey/go-practicum/internal/app/middleware"
	"github.com/putalexey/go-practicum/internal/app/shortener/handlers"
	"github.com/putalexey/go-practicum/internal/app/storage"
	"github.com/putalexey/go-practicum/internal/app/urlgenerator"
)

type Shortener struct {
	*chi.Mux
	domain  string
	counter int64
	storage storage.Storager
}

func NewRouter(ctx context.Context, baseURL string, store storage.Storager) *Shortener {
	if store == nil {
		store = &storage.MemoryStorage{}
	}
	h := &Shortener{
		Mux:     chi.NewMux(),
		storage: store,
	}
	urlGenerator := &urlgenerator.SequenceGenerator{BaseURL: baseURL}
	batchDeleter := storage.NewBatchDeleterWithContext(ctx, store, 5)

	h.Use(middleware.Logger)
	h.Use(middleware.Recoverer)
	h.Use(appMiddleware.GZipDecoder)
	h.Use(appMiddleware.GZipEncoder)
	h.Use(appMiddleware.AuthCookie(
		"auth",
		"NYiB6/ekacuT53BtdFB2ael09T8vyrnUGbi3NTeedL3tMQy4NpixN9mUzXNod9PH9EVEshAcnSFjgi+QiykVHT0j",
	))

	h.Post("/", handlers.CreateFullURLHandler(urlGenerator, store))
	h.Get("/ping", handlers.PingHandler(store))
	h.Get("/{id}", handlers.GetFullURLHandler(store))
	h.Get("/user/urls", handlers.JSONGetShortsForCurrentUser(urlGenerator, store))
	h.Post("/api/shorten/batch", handlers.JSONCreateShortBatch(urlGenerator, store))
	h.Post("/api/shorten", handlers.JSONCreateShort(urlGenerator, store))
	h.Delete("/api/user/urls", handlers.JSONDeleteUserShorts(store, batchDeleter))
	h.MethodNotAllowed(handlers.BadRequestHandler())

	return h
}
