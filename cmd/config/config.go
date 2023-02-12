package config

import "github.com/caarlos0/env"

var AppConfig Config

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	SecretKey       string `env:"SECRET_KEY" envDefault:"ca5ee5227ead"`
}

func LoadAppConfig() error {
	return env.Parse(&AppConfig)
}
