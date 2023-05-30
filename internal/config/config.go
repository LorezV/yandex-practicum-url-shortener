package config

import (
	"database/sql"
	"github.com/caarlos0/env"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// AppConfig contains main app configs, such as ServerAddress, DatabaseDsn and more...
var AppConfig config

type config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	SecretKey       string `env:"SECRET_KEY" envDefault:"ca5ee5227ead"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	EnableHttp      bool   `env:"ENABLE_HTTPS"`
	//DatabaseDsn string `env:"DATABASE_DSN" envDefault:"postgres://postgres:admin@localhost:5432/go-learn?sslmode=disable"`
}

// DB it is a global app lifetime connection to database. If AppConfig.DatabaseDsb is null, DB refer to null pointer.
var DB *sql.DB

// LoadAppConfig parse config data from env and after from program flags.
func LoadAppConfig() error {
	return env.Parse(&AppConfig)
}
