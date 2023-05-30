package main

import (
	"context"
	"fmt"
	"github.com/LorezV/url-shorter.git/internal/config"
	"github.com/LorezV/url-shorter.git/internal/handlers"
	"github.com/LorezV/url-shorter.git/internal/middlewares"
	repository2 "github.com/LorezV/url-shorter.git/internal/repository"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func init() {
	err := config.LoadAppConfig()
	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().UnixNano())
}

func main() {
	shutdown := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build version:", buildDate)
	fmt.Println("Build version:", buildCommit)

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

	srv := &http.Server{Addr: config.AppConfig.ServerAddress, Handler: r}

	go func() {
		<-sigint
		srv.Shutdown(context.Background())

		close(shutdown)
	}()

	if config.AppConfig.EnableHTTPS {
		log.Fatal(srv.ListenAndServeTLS("cmd/shortener/server.ctr", "cmd/shortener/server.key"))
	} else {
		log.Fatal(srv.ListenAndServe())
	}

	<-shutdown
}
