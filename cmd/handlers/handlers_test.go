package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/LorezV/url-shorter.git/cmd/middlewares"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LorezV/url-shorter.git/cmd/handlers"
	"github.com/LorezV/url-shorter.git/cmd/repository"
)

func TestGetURL(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name string
		urls []repository.URL
		path string
		want want
	}{
		{
			name: "Test GET request with exiting url in repository.",
			urls: []repository.URL{
				{
					ID:       "xhxKQF",
					Original: "https://practicum.yandex.ru",
					Short:    "http://127.0.0.1:8080/xhxKQF",
					UserID:   "",
				},
			},
			path: "/xhxKQF",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://practicum.yandex.ru",
			},
		},
		{
			name: "Test GET request with empty repository.",
			urls: []repository.URL{},
			path: "/xhxKQF",
			want: want{
				statusCode: http.StatusNotFound,
				location:   "",
			},
		},
		{
			name: "Test GET request with different urls in the request and repository.",
			urls: []repository.URL{
				{
					ID:       "ASKTTG",
					Original: "https://practicum.yandex.ru",
					Short:    "http://127.0.0.1:8080/ASKTTG",
					UserID:   "",
				},
			},
			path: "/xhxKQF",
			want: want{
				statusCode: http.StatusNotFound,
				location:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository.GlobalRepository = repository.MakeMemoryRepository()
			for _, url := range tt.urls {
				repository.GlobalRepository.Insert(context.Background(), url)
			}

			r := chi.NewRouter()
			r.Use(middlewares.Authorization)
			r.Get("/{id}", handlers.GetURL)
			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, _ := testRequest(t, ts, http.MethodGet, tt.path, nil)

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.location, resp.Header.Get("Location"))
		})
	}
}

func TestCreateURL(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name string
		path string
		body string
		want want
	}{
		{
			name: "Test POST request.",
			path: "/",
			body: "https://practicum.yandex.ru",
			want: want{
				statusCode: http.StatusCreated,
			},
		},
		{
			name: "Test POST path with empty body.",
			path: "/",
			body: "",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(middlewares.Authorization)
			r.Post("/", handlers.CreateURL)
			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, _ := testRequest(t, ts, http.MethodPost, tt.path, strings.NewReader(tt.body))

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}

func TestCreateURLJson(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name string
		path string
		body string
		want want
	}{
		{
			name: "Test POST request with valid body.",
			path: "/api/shorten",
			body: `{"url":"https://practicum.yandex.ru"}`,
			want: want{
				statusCode: http.StatusCreated,
			},
		},
		{
			name: "Test POST request with empty url parameter in body.",
			path: "/api/shorten",
			body: "{\"url\":\"\"}",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		}, {
			name: "Test POST request with invalid json.",
			path: "/api/shorten",
			body: `{"url:"https://practicum.yandex.ru"}`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Test POST request with empty body.",
			path: "/api/shorten",
			body: "",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(middlewares.Authorization)
			r.Post("/api/shorten", handlers.CreateURLJson)
			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, _ := testRequest(t, ts, http.MethodPost, tt.path, strings.NewReader(tt.body))

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}

func TestGetUserUrls(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name   string
		path   string
		method string
		want   want
	}{
		{
			name:   "Try to getting user urls",
			path:   "/api/user/urls",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusNoContent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(middlewares.Authorization)
			r.Get("/api/user/urls", handlers.GetUserUrls)
			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, _ := testRequest(t, ts, tt.method, tt.path, nil)

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}

func TestBatchURLJson(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name string
		path string
		body string
		want want
	}{
		{
			name: "Test with valid data",
			path: "/api/shorten/batch",
			body: `[{"correlation_id":"12123", "original_url": "http://yandex.practicum.ru"}, {"correlation_id":"2321312", "original_url": "google.com"}]`,
			want: want{
				statusCode: http.StatusCreated,
			},
		},
		{
			name: "Test with invalid body",
			path: "/api/shorten/batch",
			body: ``,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(middlewares.Authorization)
			r.Post("/api/shorten/batch", handlers.BatchURLJson)
			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, _ := testRequest(t, ts, http.MethodPost, tt.path, strings.NewReader(tt.body))

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}

func makeRequest(ts *httptest.Server, method, path string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, ts.URL+path, body)
}

func makeClient() *http.Client {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return client
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (http.Response, string) {
	req, err := makeRequest(ts, method, path, body)
	require.NoError(t, err)

	client := makeClient()
	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return *resp, string(respBody)
}

func ExampleGetURL() {
	r, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/1244543", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}

func ExampleCreateURL() {
	r, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/", strings.NewReader("yandex.lyceum"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}

func ExampleCreateURLJson() {
	r, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/shorten", strings.NewReader(`{"url": "yandex.lyceum"}`))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}

func ExampleBatchURLJson() {
	type RequestElement struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	urls := []RequestElement{
		{
			CorrelationID: "SOMEID1",
			OriginalURL:   "yandex.lyceum",
		},
		{
			CorrelationID: "SOMEID1",
			OriginalURL:   "google.com",
		},
	}

	body, err := json.Marshal(urls)
	if err != nil {
		log.Fatal(err)
	}

	r, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/shorten/batch", strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}

func ExampleGetUserUrls() {
	r, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/user/urls", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}

func ExampleDeleteUserUrls() {
	r, err := http.NewRequest(http.MethodDelete, "http://127.0.0.1:8080/api/user/urls", strings.NewReader(`["URLID1", "URLID2", "URLID3"]`))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}

func ExampleCheckPing() {
	r, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/ping", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(r)
}
