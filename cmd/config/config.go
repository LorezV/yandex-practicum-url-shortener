package config

import (
	"github.com/caarlos0/env"
)

var AppConfig Config

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseUrl       string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080/"`
}

func LoadAppConfig() {
	err := env.Parse(&AppConfig)
	if err != nil {
		panic(err)
	}
}
