package main

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/11SF/tinyurl/app"
	validate "github.com/11SF/tinyurl/app/validator"
	"github.com/11SF/tinyurl/configs"
	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {
	cfg := configs.InitConfig()
	validate.Validator = *validator.New()

	sql, err := sql.Open("pgx", cfg.PostgresConfig.DSN)
	if err != nil {
		panic(err)
	}
	if err := sql.Ping(); err != nil {
		panic(err)
	}

	application := app.NewApp(sql, *cfg)
	application.SetUp()
	application.Start()
}