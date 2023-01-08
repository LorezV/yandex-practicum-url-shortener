package main

import (
	"github.com/LorezV/url-shorter.git/cmd/handlers"
	"github.com/LorezV/url-shorter.git/cmd/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func main() {
	storage.Repository = storage.MakeRepository()
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/{id}", handlers.URLHandler)
	r.Post("/", handlers.URLHandler)
	log.Fatal(http.ListenAndServe(":8080", r))
}
