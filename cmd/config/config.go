package config

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/caarlos0/env"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
	"time"
)

var AppConfig Config

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	SecretKey       string `env:"SECRET_KEY" envDefault:"ca5ee5227ead"`
	//DatabaseDsn     string `env:"DATABASE_DSN" envDefault:"user=postgres password=admin host=localhost port=5432 dbname=go-learn sslmode=verify-ca pool_max_conns=10"`
	DatabaseDsn string `env:"DATABASE_DSN" envDefault:"postgres://postgres:admin@localhost:5432/go-learn?sslmode=disable"`
}

func LoadAppConfig() error {
	return env.Parse(&AppConfig)
}

var DB *sql.DB

func InitDatabase() {
	var err error
	DB, err = sql.Open("pgx", AppConfig.DatabaseDsn)
	if err != nil {
		fmt.Println("Unable to connect to database.")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
}
