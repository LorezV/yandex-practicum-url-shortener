package main

import (
	"context"
	"fmt"
	"github.com/LorezV/url-shorter.git/internal/config"
	"github.com/LorezV/url-shorter.git/internal/grpc/shortener"
	"github.com/LorezV/url-shorter.git/internal/handlers"
	"github.com/LorezV/url-shorter.git/internal/middlewares"
	repository2 "github.com/LorezV/url-shorter.git/internal/repository"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/LorezV/url-shorter.git/proto"
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
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

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

	srv := &http.Server{Handler: r}
	ln, err := net.Listen("tcp", config.AppConfig.ServerAddress)
	if err != nil {
		panic(err)
	}

	go func() {
		<-sigint

		err := repository2.GlobalRepository.Close()
		if err != nil {
			log.Fatal(err)
		}

		err = srv.Shutdown(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		close(shutdown)
	}()

	go func() {
		gln, grpcError := net.Listen("tcp", config.AppConfig.GRPCAddress)
		if grpcError != nil {
			log.Fatal(grpcError)
			return
		}

		g := grpc.NewServer()
		pb.RegisterShortenerServer(g, shortener.NewGRPCServer(config.AppConfig.EnableHTTPS, config.AppConfig.ServerAddress))

		if grpcErr := g.Serve(gln); grpcErr != nil {
			log.Fatal(grpcErr.Error())
			return
		}
	}()

	if config.AppConfig.EnableHTTPS {
		err = srv.ServeTLS(ln, "cmd/shortener/server.ctr", "cmd/shortener/server.key")
	} else {
		err = srv.Serve(ln)
	}

	if err != http.ErrServerClosed {
		panic(err)
	}

	<-shutdown
}
