package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/sandroJayas/storage-service/utils"
	"go.uber.org/zap"
	"log"
)

type EnvConfig struct {
	DatabaseURL          string `env:"DATABASE_URL,required"`
	JWTSecret            string `env:"JWT_SECRET,required"`
	AppEnv               string `env:"APP_ENV" envDefault:"testing"`
	HoneycombServiceName string `env:"HONEYCOMB_SERVICE_NAME,required"`
	HoneycombEndpoint    string `env:"OTEL_EXPORTER_OTLP_ENDPOINT,required"`
	HoneycombHeaders     string `env:"OTEL_EXPORTER_OTLP_HEADERS,required"`
}

var AppConfig *EnvConfig

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		utils.Logger.Warn("No .env file loaded", zap.Error(err))
	}

	var cfg EnvConfig
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("❌ Failed to parse environment: %v", err)
	}
	AppConfig = &cfg
}
