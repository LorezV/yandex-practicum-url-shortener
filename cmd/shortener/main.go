package main

import (
	"flag"
	"github.com/LorezV/url-shorter.git/cmd/middlewares"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/LorezV/url-shorter.git/cmd/config"
	"github.com/LorezV/url-shorter.git/cmd/handlers"
	"github.com/LorezV/url-shorter.git/cmd/storage"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	err := config.LoadAppConfig()
	if err != nil {
		panic(err)
	}

	flag.StringVar(&config.AppConfig.ServerAddress, "a", config.AppConfig.ServerAddress, "ip:port")
	flag.StringVar(&config.AppConfig.BaseURL, "b", config.AppConfig.BaseURL, "protocol://ip:port")
	flag.StringVar(&config.AppConfig.FileStoragePath, "f", config.AppConfig.FileStoragePath, "Path to file")
}

func main() {
	flag.Parse()
	storage.Repository = storage.MakeRepository()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.GzipHandle)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", handlers.GetURL)
		r.Post("/", handlers.CreateURL)
	})
	r.Post("/api/shorten", handlers.CreateURLJson)

	log.Fatal(http.ListenAndServe(config.AppConfig.ServerAddress, r))
}
