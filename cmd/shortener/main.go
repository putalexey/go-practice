package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

var baseURL = "http://localhost:8080/"

type Shortener struct {
	counter int64
	shorts  map[string]string
}

func (s *Shortener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodGet:
		s.handleGet(w, r)
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}

func (s *Shortener) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
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
	w.Write([]byte(baseURL + id))
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

func main() {
	handler := &Shortener{}

	http.Handle("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
