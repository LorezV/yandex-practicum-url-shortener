package config

import (
	"database/sql"
	"github.com/caarlos0/env"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var AppConfig Config

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	SecretKey       string `env:"SECRET_KEY" envDefault:"ca5ee5227ead"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	//DatabaseDsn string `env:"DATABASE_DSN" envDefault:"postgres://postgres:admin@localhost:5432/go-learn?sslmode=disable"`
}

var DB *sql.DB

func LoadAppConfig() error {
	return env.Parse(&AppConfig)
}
