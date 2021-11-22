package shortener

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Shortener struct {
	*chi.Mux
	domain  string
	counter int64
	shorts  ShortsList
}

type ShortsList map[string]string

func NewShortener(domain string) *Shortener {
	h := &Shortener{
		Mux:    chi.NewMux(),
		shorts: map[string]string{},
		domain: domain,
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

	if fullURL, ok := s.shorts[id]; ok {
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

	id := s.nextShortURL()
	s.setShort(id, fullURL)

	w.WriteHeader(http.StatusCreated)

	_, _ = fmt.Fprintf(w, "http://%s/%s", s.domain, id)
}

func (s *Shortener) setShort(short, full string) {
	if s.shorts == nil {
		s.shorts = make(map[string]string)
	}
	s.shorts[short] = full
}

func (s *Shortener) nextShortURL() string {
	str := strconv.FormatInt(s.counter, 36)
	s.counter += 1
	return str
}
