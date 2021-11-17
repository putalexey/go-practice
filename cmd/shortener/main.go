package main

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
)

var baseURL = "http://localhost:8080/"

type Sortener struct {
	counter int64
	shorts  map[string]string
}

func (s *Sortener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodGet:
		s.handleGet(w, r)
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}

func (s *Sortener) handleGet(w http.ResponseWriter, r *http.Request) {
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

func (s *Sortener) handlePost(w http.ResponseWriter, r *http.Request) {
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
	s.shorts[id] = fullURL

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(baseURL + id))
}

func (s *Sortener) nextShortURL() string {
	str := strconv.FormatInt(s.counter, 36)
	s.counter += 1
	return str
}

func main() {
	handler := &Sortener{0, make(map[string]string)}

	http.Handle("/", handler)
	http.ListenAndServe(":8080", nil)
}
