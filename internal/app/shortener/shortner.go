package shortener

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/putalexey/go-practicum/internal/app/storage"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Shortener struct {
	*chi.Mux
	domain  string
	counter int64
	storage storage.Storager
}

func NewShortener(domain string) *Shortener {
	h := &Shortener{
		Mux:     chi.NewMux(),
		domain:  domain,
		storage: &storage.MemoryStorage{},
	}
	h.Use(middleware.Logger)
	h.Use(middleware.Recoverer)

	h.Post("/", h.handlePost)
	h.Get("/{id}", h.handleGet)
	h.MethodNotAllowed(h.handleMethodNotAllowed)
	return h
}

func (s *Shortener) handleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if fullURL, err := s.storage.Load(id); err == nil {
		http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
		return
	}
	http.Error(w, "Not found", http.StatusNotFound)
}

func (s *Shortener) handleMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func (s *Shortener) handlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(body) == 0 {
		http.Error(w, "Empty request", http.StatusBadRequest)
		return
	}

	fullURL := string(body)
	if _, err := url.ParseRequestURI(fullURL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	short := s.nextShortURL()
	if err := s.storage.Store(short, fullURL); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = fmt.Fprintf(w, "http://%s/%s", s.domain, short)
}

func (s *Shortener) nextShortURL() string {
	str := strconv.FormatInt(s.counter, 36)
	s.counter += 1
	return str
}
