package main

import (
	"flag"
	"github.com/LorezV/url-shorter.git/internal/config"
	"github.com/LorezV/url-shorter.git/internal/handlers"
	"github.com/LorezV/url-shorter.git/internal/middlewares"
	repository2 "github.com/LorezV/url-shorter.git/internal/repository"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	flag.StringVar(&config.AppConfig.DatabaseDsn, "d", config.AppConfig.DatabaseDsn, "Database connection URL")
}

func main() {
	flag.Parse()
	if len(config.AppConfig.DatabaseDsn) > 0 {
		repository2.GlobalRepository = repository2.MakePostgresRepository()
	} else {
		repository2.GlobalRepository = repository2.MakeMemoryRepository()
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.GzipHandle)
	r.Use(middlewares.Authorization)

	r.Mount("/debug", middleware.Profiler())

	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", handlers.GetURL)
		r.Post("/", handlers.CreateURL)
	})
	r.Post("/api/shorten/batch", handlers.BatchURLJson)
	r.Post("/api/shorten", handlers.CreateURLJson)
	r.Route("/api/user/urls", func(r chi.Router) {
		r.Get("/", handlers.GetUserUrls)
		r.Delete("/", handlers.DeleteUserUrls)
	})
	r.Get("/ping", handlers.CheckPing)

	log.Fatal(http.ListenAndServe(config.AppConfig.ServerAddress, r))
}
