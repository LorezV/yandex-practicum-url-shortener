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

	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", handlers.GetURL)
		r.Post("/", handlers.CreateURL)
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", handlers.CreateURLJson)
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
