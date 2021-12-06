package shortener

import (
	"bytes"
	"encoding/json"
	"github.com/putalexey/go-practicum/internal/app/shortener/requests"
	"github.com/putalexey/go-practicum/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortener_Base(t *testing.T) {
	type request struct {
		method string
		target string
		body   string
	}
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name    string
		request request
		shorts  map[string]string
		want    want
	}{
		{
			name: "shortens",
			request: request{
				method: http.MethodPost,
				target: "/",
				body:   "http://test.example.com",
			},
			want: want{
				code: http.StatusCreated,
			},
		},
		{
			name: "post with empty body fails",
			request: request{
				method: http.MethodPost,
				target: "/",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "post with error in url fails",
			request: request{
				method: http.MethodPost,
				target: "/",
				body:   "http//test.example.com",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "resolves",
			request: request{
				method: http.MethodGet,
				target: "/some",
			},
			shorts: map[string]string{"some": "http://test.example.com"},
			want: want{
				code:     http.StatusTemporaryRedirect,
				response: "http://test.example.com",
			},
		},
		{
			name: "returns 404 status",
			request: request{
				method: http.MethodGet,
				target: "/some",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "returns 400 bad request on empty url",
			request: request{
				method: http.MethodGet,
				target: "/",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "returns 400 bad request on wrong method",
			request: request{
				method: http.MethodPatch,
				target: "/some",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestBody io.Reader = nil
			if tt.request.body != "" {
				requestBody = strings.NewReader(tt.request.body)
			}
			request := httptest.NewRequest(tt.request.method, tt.request.target, requestBody)
			w := httptest.NewRecorder()

			s := NewRouter("localhost:8080", storage.NewMemoryStorage(tt.shorts))
			s.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.want.code, result.StatusCode)

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			err = result.Body.Close()
			require.NoError(t, err)

			if tt.want.response != "" {
				assert.Contains(t, string(body), tt.want.response)
			}
		})
	}
}

func TestShortener_JSONCreateFails(t *testing.T) {
	type request struct {
		method string
		target string
		body   string
	}
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name    string
		request request
		shorts  map[string]string
		want    want
	}{
		{
			name: "post with empty body fails",
			request: request{
				method: http.MethodPost,
				target: "/api/shorten",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "post with wrong json fails",
			request: request{
				method: http.MethodPost,
				target: "/api/shorten",
				body:   "http://test.example.com",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "post with error in url fails",
			request: request{
				method: http.MethodPost,
				target: "/api/shorten",
				body:   "{\"url\":\"http//test.example.com\"}",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "returns 400 bad request on wrong method",
			request: request{
				method: http.MethodPatch,
				target: "/api/shorten",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestBody io.Reader = nil
			if tt.request.body != "" {
				requestBody = strings.NewReader(tt.request.body)
			}
			request := httptest.NewRequest(tt.request.method, tt.request.target, requestBody)
			w := httptest.NewRecorder()

			s := NewRouter("localhost:8080", storage.NewMemoryStorage(tt.shorts))
			s.ServeHTTP(w, request)

			result := w.Result()
			assert.Equal(t, tt.want.code, result.StatusCode)

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			err = result.Body.Close()
			require.NoError(t, err)

			if tt.want.response != "" {
				assert.Contains(t, string(body), tt.want.response)
			}
		})
	}
}

func TestShortener_JSONCreates(t *testing.T) {
	t.Run("Creates short url", func(t *testing.T) {
		var requestBody io.Reader = nil
		body, err := json.Marshal(requests.CreateShortRequest{URL: "http://test.example.com"})
		require.NoError(t, err)

		requestBody = bytes.NewReader(body)
		request := httptest.NewRequest(http.MethodPost, "/api/shorten", requestBody)
		w := httptest.NewRecorder()

		s := NewRouter("localhost:8080", nil)
		s.ServeHTTP(w, request)

		result := w.Result()
		assert.Equal(t, http.StatusCreated, result.StatusCode)

		responseBody, err := io.ReadAll(result.Body)
		require.NoError(t, err)

		err = result.Body.Close()
		require.NoError(t, err)

		response := struct {
			Result string `json:"result"`
		}{}
		err = json.Unmarshal(responseBody, &response)
		assert.NoError(t, err)

		assert.NotEmpty(t, response.Result)
	})
}

func TestShortener_NewRouter(t *testing.T) {
	t.Run("default router storage is MemoryStorage ", func(t *testing.T) {
		s := NewRouter("localhost:8080", nil)
		assert.IsType(t, &storage.MemoryStorage{}, s.storage)
	})
}
