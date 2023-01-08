package handlers_test

import (
	"github.com/LorezV/url-shorter.git/cmd/handlers"
	"github.com/LorezV/url-shorter.git/cmd/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestURLHandler(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name    string
		urls    []storage.URL
		method  string
		body    string
		request string
		want    want
	}{
		{
			name: "Test GET request with exiting url in repository.",
			urls: []storage.URL{
				{
					ID:       "xhxKQF",
					Original: "https://practicum.yandex.ru",
					Short:    "http://127.0.0.1:8080/xhxKQF",
				},
			},
			request: "http://127.0.0.1:8080/xhxKQF",
			method:  http.MethodGet,
			body:    "",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://practicum.yandex.ru",
			},
		},
		{
			name:    "Test GET request with an existing link in the repository.",
			urls:    []storage.URL{},
			request: "http://127.0.0.1:8080/xhxKQF",
			want: want{
				statusCode: http.StatusNotFound,
				location:   "",
			},
		},
		{
			name:    "Test GET request with different urls in the request and repository.",
			urls:    []storage.URL{},
			request: "http://127.0.0.1:8080/xhxKQF",
			method:  http.MethodGet,
			body:    "",
			want: want{
				statusCode: http.StatusNotFound,
				location:   "",
			},
		},
		{
			name:    "Test POST request.",
			urls:    []storage.URL{},
			request: "http://127.0.0.1:8080/",
			method:  http.MethodPost,
			body:    "https://practicum.yandex.ru",
			want: want{
				statusCode: http.StatusCreated,
				location:   "",
			},
		},
		{
			name:    "Test POST request with empty body.",
			urls:    []storage.URL{},
			request: "http://127.0.0.1:8080/",
			method:  http.MethodPost,
			body:    "",
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Repository = storage.MakeRepository()
			for _, url := range tt.urls {
				storage.Repository.Add(url)
			}
			w := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.request, strings.NewReader(tt.body))
			h := http.HandlerFunc(handlers.URLHandler)
			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.location, res.Header.Get("Location"))

			if tt.method == http.MethodPost {
				urls := storage.Repository.GetAll()
				assert.Len(t, urls, 1)
			}
		})
	}
}
