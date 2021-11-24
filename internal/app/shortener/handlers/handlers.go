package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/putalexey/go-practicum/internal/app/storage"
	"github.com/putalexey/go-practicum/internal/app/urlgenerator"
	"io"
	"net/http"
	"net/url"
)

func GetFullURLHandler(storage storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if fullURL, err := storage.Load(id); err == nil {
			http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func CreateFullURLHandler(generator urlgenerator.URLGenerator, storage storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		short := generator.GenerateShort(fullURL)
		if err := storage.Store(short, fullURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusCreated)
		//_, _ = fmt.Fprintf(w, "http://%s/%s", domain, short)
		_, _ = fmt.Fprint(w, generator.GetURL(short))
	}
}

func BadRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}
