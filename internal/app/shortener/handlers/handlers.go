package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/putalexey/go-practicum/internal/app/middleware"
	"github.com/putalexey/go-practicum/internal/app/shortener/requests"
	"github.com/putalexey/go-practicum/internal/app/shortener/responses"
	"github.com/putalexey/go-practicum/internal/app/storage"
	"github.com/putalexey/go-practicum/internal/app/urlgenerator"
)

// PingHandler godoc
// @Summary returns "OK" if service is working and storage is available
// @Produce plain
// @Success 200 {string} string "OK"
// @Failure 500	{string} string "DB unavailable"
// @Router /ping [get]
func PingHandler(storage storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := storage.Ping(r.Context())
		if err != nil {
			log.Println("ERROR: DB unavailable:", err)
			http.Error(w, "DB unavailable", http.StatusInternalServerError)
			return
		}
		_, err = w.Write([]byte("OK"))
		if err != nil {
			log.Println("ERROR:", err)
			panic(err)
		}
	}
}

// GetFullURLHandler godoc
// @Summary	Redirects to the full url, if found in storage by {id}
// @Produce	plain
// @Param	id	path	string	true	"url id"
// @Success	307	"redirects to full url"
// @Failure	400	{string}	string	"Bad request"
// @Failure	404	{string}	string	"Not found"
// @Failure	410	{string}	string	"Record has been deleted"
// @Header	307	{string}	Location	"http://example.com/"
// @Router	/{id}	[get]
func GetFullURLHandler(storage storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		record, err := storage.Load(r.Context(), id)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		if record.Deleted {
			http.Error(w, "Record has been deleted", http.StatusGone)
			return
		}
		http.Redirect(w, r, record.Full, http.StatusTemporaryRedirect)
	}
}

// CreateFullURLHandler godoc
// @Summary	Create new short url
// @Accept	plain
// @Produce	plain
// @Param	url	body	string	true	"Full url for shortening"
// @Success	201	{string}	string	"http://shortener.org/123"
// @Failure	409	{string}	string	"http://shortener.org/123"
// @Failure	400	{string}	string	"Empty request"
// @Failure	400	{string}	string	"invalid url: http//example"
// @Failure	500	{string}	string	"Server error"
// @Router	/	[post]
func CreateFullURLHandler(generator urlgenerator.URLGenerator, store storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseStatus := http.StatusCreated
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

		err = store.Store(r.Context(), short)
		if err != nil {
			var conflictError *storage.RecordConflictError
			if errors.As(err, &conflictError) {
				responseStatus = http.StatusConflict
				short = conflictError.OldRecord
			} else {
				jsonError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(responseStatus)
		_, _ = fmt.Fprint(w, generator.GetURL(short.Short))
	}
}

// JSONCreateShort godoc
// @Summary	Create new short url
// @Accept	json
// @Produce	json
// @Param	fullURL	body	requests.CreateShortRequest	true	"Full url for shortening"
// @Success	201	{object}	responses.CreateShortResponse	"URL saved, short url returned in result field"
// @Failure	409	{object}	responses.CreateShortResponse	"Full URL already added earlier, old short url is returned in result field"
// @Failure	400	{object}	responses.ErrorResponse
// @Failure	500	{object}	responses.ErrorResponse
// @Router	/api/shorten	[post]
func JSONCreateShort(generator urlgenerator.URLGenerator, store storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseStatus := http.StatusCreated
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Println("ERROR:", err)
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(body) == 0 {
			jsonError(w, "Empty request", http.StatusBadRequest)
			return
		}

		userID, err := getUserIDFromRequest(r)
		if err != nil {
			log.Println("ERROR:", err)
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
			log.Println("ERROR:", err)
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = store.Store(r.Context(), short)
		if err != nil {
			var conflictError *storage.RecordConflictError
			if errors.As(err, &conflictError) {
				responseStatus = http.StatusConflict
				short = conflictError.OldRecord
			} else {
				log.Println("ERROR:", err)
				jsonError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		createResponse := responses.CreateShortResponse{Result: generator.GetURL(short.Short)}
		data, err := json.Marshal(createResponse)
		if err != nil {
			log.Println("ERROR:", err)
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseStatus)
		_, err = w.Write(data)
		if err != nil {
			log.Println("ERROR:", err)
			panic(err)
		}
	}
}

// JSONCreateShortBatch godoc
// @Summary	Create many new short urls
// @Accept	json
// @Produce	json
// @Param	fullURList	body	requests.CreateShortBatchRequest	true	"List of full urls for shortening"
// @Success	201	{object}	responses.CreateShortBatchResponse
// @Failure	400	{object}	responses.ErrorResponse
// @Failure	500	{object}	responses.ErrorResponse
// @Router	/api/shorten/batch	[post]
func JSONCreateShortBatch(generator urlgenerator.URLGenerator, store storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Println("ERROR:", err)
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(body) == 0 {
			jsonError(w, "Empty request", http.StatusBadRequest)
			return
		}

		userID, err := getUserIDFromRequest(r)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		batch := requests.CreateShortBatchRequest{}
		if err = json.Unmarshal(body, &batch); err != nil {
			jsonError(w, "Request can't be parsed", http.StatusBadRequest)
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

			r, err2 := storage.NewRecord(item.OriginalURL, userID)
			if err2 != nil {
				log.Println("ERROR:", err2)
				jsonError(w, err2.Error(), http.StatusInternalServerError)
				return
			}

			if err2 = batchInserter.AddItem(ctx, r); err2 != nil {
				log.Println("ERROR:", err2)
				jsonError(w, err2.Error(), http.StatusInternalServerError)
				return
			}

			responseItem := responses.CreateShortBatchResponseItem{
				CorrelationID: item.CorrelationID,
				ShortURL:      generator.GetURL(r.Short),
			}
			response = append(response, responseItem)
		}

		if err = batchInserter.Flush(ctx); err != nil {
			log.Println("ERROR:", err)
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(response)
		if err != nil {
			log.Println("ERROR:", err)
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(data)
		if err != nil {
			log.Println("ERROR:", err)
			panic(err)
		}
	}
}
func isValidURL(uri string) bool {
	_, err := url.ParseRequestURI(uri)
	return err == nil
}

// JSONGetShortsForCurrentUser godoc
// @Summary	Get all urls user shortened
// @Produce	json
// @Success	200	{object}	responses.ListShortsResponse	"List of urls, user added"
// @Success	204	"No Content. User not added any urls yet"
// @Failure	400	{object}	responses.ErrorResponse
// @Failure	500	{object}	responses.ErrorResponse
// @Router	/api/user/urls	[get]
func JSONGetShortsForCurrentUser(generator urlgenerator.URLGenerator, storage storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromRequest(r)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		recordsList, err := storage.LoadForUser(r.Context(), userID)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(recordsList) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

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
			log.Println("ERROR:", err)
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(data)
		if err != nil {
			log.Println("ERROR:", err)
			panic(err)
		}
	}
}

// JSONDeleteUserShorts godoc
// @Summary	Delete urls user shortened earlier
// @Accept	json
// @Produce	json
// @Param	deleteURLs	body	requests.DeleteShortBatchRequest	true	"List of urls to delete"
// @Success	202	"Delete request accepted and put on queue, urls will be deleted eventually"
// @Failure	400	{object}	responses.ErrorResponse
// @Failure	500	{object}	responses.ErrorResponse
// @Router	/api/user/urls	[delete]
func JSONDeleteUserShorts(_ storage.Storager, batchDeleter *storage.BatchDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Println("ERROR:", err)
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(body) == 0 {
			jsonError(w, "Empty request", http.StatusBadRequest)
			return
		}

		userID, err := getUserIDFromRequest(r)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		shorts := requests.DeleteShortBatchRequest{}
		if err = json.Unmarshal(body, &shorts); err != nil {
			jsonError(w, "Request can't be parsed", http.StatusBadRequest)
			return
		}

		//ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		//defer cancel()

		batchDeleter.QueueItems(shorts, userID)

		//// check does all shorts exists and belongs to current user
		//records, err := store.LoadBatch(ctx, shorts)
		//if err != nil {
		//	jsonError(w, err.Error(), http.StatusInternalServerError)
		//	return
		//}
		//for _, r := range records {
		//	if r.UserID != userID {
		//		jsonError(w, "you can't delete item: "+r.Short, http.StatusForbidden)
		//		return
		//	}
		//}
		//
		//if err = store.DeleteBatch(ctx, shorts); err != nil {
		//	log.Println("ERROR:", err)
		//	jsonError(w, err.Error(), http.StatusInternalServerError)
		//	return
		//}

		w.WriteHeader(http.StatusAccepted)
	}
}

// JSONInternalStats godoc
// @Summary	Shows service stats, accessible only from trusted IP networks
// @Produce	json
// @Success	200	{object}	responses.InternalStatsResponse	"Stats of the service"
// @Failure	403	{object}	responses.ErrorResponse
// @Failure	500	{object}	responses.ErrorResponse
// @Router	/api/internal/stats	[get]
func JSONInternalStats(storage storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlsCount, err := storage.CountURLs(r.Context())
		if err != nil {
			log.Println("ERROR:", err)
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		usersCount, err := storage.CountUsers(r.Context())
		if err != nil {
			log.Println("ERROR:", err)
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		stats := responses.InternalStatsResponse{URLs: urlsCount, Users: usersCount}

		data, err := json.Marshal(stats)
		if err != nil {
			log.Println("ERROR:", err)
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

// BadRequestHandler handles requests to route with method not supported by route
func BadRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Method not found", http.StatusBadRequest)
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

// getUserIDFromRequest returns user id from request's context
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
