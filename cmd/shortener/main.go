package main

import (
	"github.com/LorezV/url-shorter.git/cmd/config"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/LorezV/url-shorter.git/cmd/handlers"
	"github.com/LorezV/url-shorter.git/cmd/storage"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	err := config.LoadAppConfig()
	if err != nil {
		panic(err)
	}

	storage.Repository = storage.MakeRepository()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", handlers.GetURL)
		r.Post("/", handlers.CreateURL)
	})

	r.Post("/api/shorten", handlers.CreateURLJson)

	log.Fatal(http.ListenAndServe(config.AppConfig.ServerAddress, r))
}
