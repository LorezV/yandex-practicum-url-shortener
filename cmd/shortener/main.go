package main

import (
	"github.com/LorezV/url-shorter.git/cmd/handlers"
	"github.com/LorezV/url-shorter.git/cmd/storage"
	"net/http"
)

func main() {
	storage.Repository = *storage.MakeRepository()
	http.HandleFunc("/", handlers.URLHandler)
	http.ListenAndServe(":8080", nil)
}
