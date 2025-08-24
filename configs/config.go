package configs

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

type Config struct {
	AppName               string `env:"APP_NAME,required"`
	AppPort               int    `env:"APP_PORT,required"`
	BaseURL               string `env:"BASE_URL,required"`
	PostgresConfig        PostgresConfig
	RedisConfig           RedisConfig
	AuthenticationSrevice AuthenticationSrevice
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

type AuthenticationSrevice struct {
	BaseURL         string `env:"AUTHENTICATION_SVC_BASE_URL,required"`
	PathLogin       string `env:"AUTHENTICATION_SVC_PATH_LOGIN,required"`
	PathRefresh     string `env:"AUTHENTICATION_SVC_PATH_REFRESH,required"`
	PathGetUserInfo string `env:"AUTHENTICATION_SVC_PATH_GET_USER_INFO,required"`
	PathVerifyToken string `env:"AUTHENTICATION_SVC_PATH_VERIFY_TOKEN,required"`
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
