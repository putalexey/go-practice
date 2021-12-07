package shortener

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

func NewRouter(baseURL string, store storage.Storager) *Shortener {
	if store == nil {
		store = &storage.MemoryStorage{}
	}
	h := &Shortener{
		Mux:     chi.NewMux(),
		storage: store,
	}
	urlGenerator := &urlgenerator.SequenceGenerator{BaseURL: baseURL}

	h.Use(middleware.Logger)
	h.Use(middleware.Recoverer)

	h.Post("/", handlers.CreateFullURLHandler(urlGenerator, store))
	h.Get("/{id}", handlers.GetFullURLHandler(store))
	h.Post("/api/shorten", handlers.JSONCreateShort(urlGenerator, store))
	h.MethodNotAllowed(handlers.BadRequestHandler())

	return h
}
