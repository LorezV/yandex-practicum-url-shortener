package main

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
)

var storage map[string]string

func GenerateURL() string {
	runes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 6)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}

	// Простая проверка на уникальность
	_, ok := storage[string(b)]
	if ok {
		return GenerateURL()
	}

	return string(b)
}

func URLHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		b, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "Unknown error", http.StatusBadRequest)
		}

		url := GenerateURL()
		storage[url] = string(b)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://127.0.0.1:8080/" + url))
	case http.MethodGet:
		id := strings.Split(r.URL.Path, "/")[1]

		if id == "" {
			http.Error(w, "The query parameter ID is missing", http.StatusBadRequest)
			return
		}

		long, ok := storage[id]

		if ok {
			w.WriteHeader(http.StatusTemporaryRedirect)
			w.Write([]byte(long))
			return
		} else {
			http.Error(w, "Url with this id not found!", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method not allowed!", http.StatusMethodNotAllowed)
	}
}

func main() {
	storage = make(map[string]string)
	http.HandleFunc("/", URLHandler)
	http.ListenAndServe(":8080", nil)
}
