package config

import (
	"database/sql"
	"encoding/json"
	"flag"
	"github.com/caarlos0/env"
	_ "github.com/jackc/pgx/v5/stdlib"
	"io"
	"os"
)

// AppConfig contains main app configs, such as ServerAddress, DatabaseDsn and more...
var AppConfig config

type config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080" json:"server_address"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080" json:"base_url"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	SecretKey       string `env:"SECRET_KEY" envDefault:"ca5ee5227ead" json:"secret_key"`
	DatabaseDsn     string `env:"DATABASE_DSN" json:"database_dsn"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	ConfigFile      string `env:"CONFIG"`
	//DatabaseDsn string `env:"DATABASE_DSN" envDefault:"postgres://postgres:admin@localhost:5432/go-learn?sslmode=disable"`
}

// DB it is a global app lifetime connection to database. If AppConfig.DatabaseDsb is null, DB refer to null pointer.
var DB *sql.DB

// LoadAppConfig parse config data from env and after from program flags.
func LoadAppConfig() error {
	err := env.Parse(&AppConfig)
	if err != nil {
		return err
	}

	flag.StringVar(&AppConfig.ServerAddress, "a", AppConfig.ServerAddress, "ip:port")
	flag.StringVar(&AppConfig.BaseURL, "b", AppConfig.BaseURL, "protocol://ip:port")
	flag.StringVar(&AppConfig.FileStoragePath, "f", AppConfig.FileStoragePath, "Path to file")
	flag.StringVar(&AppConfig.DatabaseDsn, "d", AppConfig.DatabaseDsn, "Database connection URL")
	flag.BoolVar(&AppConfig.EnableHTTPS, "s", AppConfig.EnableHTTPS, "Enable tls")
	flag.StringVar(&AppConfig.ConfigFile, "c", AppConfig.ConfigFile, "Path to config.json")
	flag.Parse()

	if len(AppConfig.ConfigFile) > 0 {
		file, err := os.Open(AppConfig.ConfigFile)
		if err != nil {
			return err
		}

		defer file.Close()

		bt, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		var tempConfig config

		err = json.Unmarshal(bt, &tempConfig)
		if err != nil {
			return err
		}

		if len(AppConfig.ServerAddress) == 0 {
			AppConfig.ServerAddress = tempConfig.ServerAddress
		}

		if len(AppConfig.BaseURL) == 0 {
			AppConfig.BaseURL = tempConfig.BaseURL
		}

		if len(AppConfig.FileStoragePath) == 0 {
			AppConfig.FileStoragePath = tempConfig.FileStoragePath
		}

		if len(AppConfig.SecretKey) == 0 {
			AppConfig.SecretKey = tempConfig.SecretKey
		}

		if len(AppConfig.DatabaseDsn) == 0 {
			AppConfig.DatabaseDsn = tempConfig.DatabaseDsn
		}

		if !AppConfig.EnableHTTPS || tempConfig.EnableHTTPS {
			AppConfig.EnableHTTPS = tempConfig.EnableHTTPS
		}
	}

	return nil
}
