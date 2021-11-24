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

func NewRouter(domain string, store storage.Storager) *Shortener {
	if store == nil {
		store = &storage.MemoryStorage{}
	}
	h := &Shortener{
		Mux:     chi.NewMux(),
		domain:  domain,
		storage: store,
	}
	urlGenerator := &urlgenerator.SequenceGenerator{Domain: domain}

	h.Use(middleware.Logger)
	h.Use(middleware.Recoverer)

	h.Post("/", handlers.CreateFullURLHandler(urlGenerator, store))
	h.Get("/{id}", handlers.GetFullURLHandler(store))
	h.MethodNotAllowed(handlers.BadRequestHandler())

	return h
}
