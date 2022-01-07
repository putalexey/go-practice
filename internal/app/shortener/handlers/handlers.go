package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/putalexey/go-practicum/internal/app/middleware"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/putalexey/go-practicum/internal/app/shortener/requests"
	"github.com/putalexey/go-practicum/internal/app/shortener/responses"
	"github.com/putalexey/go-practicum/internal/app/storage"
	"github.com/putalexey/go-practicum/internal/app/urlgenerator"
)

func PingHandler(storage storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := storage.Ping(r.Context())
		if err != nil {
			http.Error(w, "DB unavalable", http.StatusInternalServerError)
			return
		}
		_, err = w.Write([]byte("OK"))
		if err != nil {
			panic(err)
		}
	}
}

func GetFullURLHandler(storage storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if record, err := storage.Load(r.Context(), id); err == nil {
			http.Redirect(w, r, record.Full, http.StatusTemporaryRedirect)
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func CreateFullURLHandler(generator urlgenerator.URLGenerator, store storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(body) == 0 {
			http.Error(w, "Empty request", http.StatusBadRequest)
			return
		}

		userID, err := getUserIDFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fullURL := string(body)
		if !isValidURL(fullURL) {
			http.Error(w, invalidURLError(fullURL), http.StatusBadRequest)
			return
		}

		//short := generator.GenerateShort(fullURL)
		short, err := storage.NewRecord(fullURL, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := store.Store(r.Context(), short); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprint(w, generator.GetURL(short.Short))
	}
}

func JSONCreateShort(generator urlgenerator.URLGenerator, store storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(body) == 0 {
			jsonError(w, "Empty request", http.StatusBadRequest)
			return
		}

		userID, err := getUserIDFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		createRequest := requests.CreateShortRequest{}
		if err = json.Unmarshal(body, &createRequest); err != nil {
			jsonError(w, "Request can't be parsed", http.StatusBadRequest)
			return
		}

		if !isValidURL(createRequest.URL) {
			jsonError(w, invalidURLError(createRequest.URL), http.StatusBadRequest)
			return
		}

		short, err := storage.NewRecord(createRequest.URL, userID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := store.Store(r.Context(), short); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		createResponse := responses.CreateShortResponse{Result: generator.GetURL(short.Short)}
		data, err := json.Marshal(createResponse)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(data)
		if err != nil {
			panic(err)
		}
	}
}

func JSONCreateShortBatch(generator urlgenerator.URLGenerator, store storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(body) == 0 {
			jsonError(w, "Empty request", http.StatusBadRequest)
			return
		}

		userID, err := getUserIDFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		batch := requests.CreateShortBatchRequest{}
		if err = json.Unmarshal(body, &batch); err != nil {
			jsonError(w, "Request can't be parsed", http.StatusBadRequest)
			return
		}

		if err := checkBachURLs(batch); err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
		defer cancel()

		response := responses.CreateShortBatchResponse{}
		batchInserter := storage.NewBatchInserter(store, 10)
		for _, item := range batch {
			if !isValidURL(item.OriginalURL) {
				jsonError(w, invalidURLError(item.OriginalURL), http.StatusBadRequest)
				return
			}

			r, err := storage.NewRecord(item.OriginalURL, userID)
			if err != nil {
				jsonError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := batchInserter.AddItem(ctx, r); err != nil {
				jsonError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			responseItem := responses.CreateShortBatchResponseItem{
				CorrelationID: item.CorrelationID,
				ShortURL:      generator.GetURL(r.Short),
			}
			response = append(response, responseItem)
		}

		if err := batchInserter.Flush(ctx); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(response)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(data)
		if err != nil {
			panic(err)
		}
	}
}
func isValidURL(uri string) bool {
	_, err := url.ParseRequestURI(uri)
	return err == nil
}
func checkBachURLs(batch requests.CreateShortBatchRequest) error {
	for _, item := range batch {
		if _, err := url.ParseRequestURI(item.OriginalURL); err != nil {
			return err
		}
	}
	return nil
}

func JSONGetShortsForCurrentUser(generator urlgenerator.URLGenerator, storage storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		recordsList, err := storage.LoadForUser(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(recordsList) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		log.Println(userID, recordsList)

		listResponse := make(responses.ListShortsResponse, len(recordsList))
		i := 0
		for _, record := range recordsList {
			listResponse[i] = responses.ListShortItem{
				ShortURL:    generator.GetURL(record.Short),
				OriginalURL: record.Full,
			}
			i++
		}
		data, err := json.Marshal(listResponse)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(data)
		if err != nil {
			panic(err)
		}
	}
}

func BadRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}

func jsonError(w http.ResponseWriter, errMessage string, code int) {
	response := responses.ErrorResponse{Error: errMessage}
	data, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
}

func getUserIDFromRequest(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(middleware.UIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("user id is not initialized")
	}
	return userID, nil
}

func invalidURLError(uri string) string {
	return fmt.Sprintf("invalid url: %s", uri)
}
