package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortener(t *testing.T) {
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
		shorts  ShortsList
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
			name: "resolves",
			request: request{
				method: http.MethodGet,
				target: "/some",
			},
			shorts: ShortsList{"some": "http://test.example.com"},
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

			s := NewShortener(tt.shorts)
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
