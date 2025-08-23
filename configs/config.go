package configs

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

type Config struct {
	AppName        string `env:"APP_NAME,required"`
	AppPort        int    `env:"APP_PORT,required"`
	BaseURL        string `env:"BASE_URL,required"`
	PostgresConfig PostgresConfig
	RedisConfig    RedisConfig
	AuthConfig     AuthConfig
}

type PostgresConfig struct {
	DSN string `env:"DB_DSN,required"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST,required"`
	Port     int    `env:"REDIS_PORT,required"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB"`
}

type AuthConfig struct {
	ServiceURL string `env:"AUTH_SERVICE_URL,required"`
}

func InitConfig() *Config {
	slog.Info("Loading environment variables", slog.String("tag", "config"))
	err := godotenv.Load("./configs/.env")
	if err != nil {
		slog.Error("Error loading .env file", slog.String("err", err.Error()), slog.String("tag", "config"))
	}

	config := &Config{}
	err = env.Parse(config)
	if err != nil {
		slog.Error("Error loading environment variables", slog.String("err", err.Error()), slog.String("tag", "config"))
		panic(err)
	}

	slog.Info("Environment variables loaded", slog.String("tag", "config"))
	return config
}
