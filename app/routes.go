package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpclient "github.com/11SF/go-common/http_client"
	"github.com/11SF/tinyurl/app/url"
	"github.com/11SF/tinyurl/app/user"
	"github.com/11SF/tinyurl/configs"
	"github.com/ansrivas/fiberprometheus/v2"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type App struct {
	db     *sql.DB
	config configs.Config
	logger *slog.Logger
	fiber  *fiber.App
}

func NewApp(db *sql.DB, config configs.Config) *App {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	app := fiber.New(fiber.Config{
		WriteTimeout:  1 * time.Minute,
		ReadTimeout:   30 * time.Second,
		ErrorHandler:  customErrorHandler,
		AppName:       "tinyurl",
		CaseSensitive: true,
		StrictRouting: true,
		Prefork:       false,
	})

	return &App{
		db:     db,
		config: config,
		logger: logger,
		fiber:  app,
	}
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	slog.Error("request error",
		"path", c.Path(),
		"method", c.Method(),
		"error", err.Error(),
		"status", code,
	)

	return c.Status(code).JSON(fiber.Map{
		"status":  "error",
		"message": err.Error(),
		"code":    code,
	})
}

func (a *App) registerRoutes() {
	urlRepository := url.NewRepository(a.db)
	urlService := url.NewService(urlRepository)
	
	client := httpclient.NewHTTPClient(httpclient.ClientConfig{
		BaseURL:     a.config.AuthenticationSrevice.BaseURL,
		Timeout:     5 * time.Minute,
		ContentType: httpclient.ContentTypeJSON,
	})
	authClient := user.NewAuthenticationClient(a.config, *client)

	urlHandler := url.NewHandler(a.config, urlService, authClient)
	userHandler := user.NewHandler(a.config, authClient)

	api := a.fiber.Group("/api/tinyurl/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/login", userHandler.Login)
	auth.Get("/validate", userHandler.ValidateToken)

	// URL routes
	api.Post("/shorten", urlHandler.CreateURL)
	api.Get("/urls", urlHandler.GetUserURLs)
	api.Put("/urls", urlHandler.UpdateURL)
	api.Delete("/urls/:id", urlHandler.DeleteURL)
	api.Get("/analytics/:id", urlHandler.GetAnalytics)

	// Redirect route
	a.fiber.Get("/:shortCode", urlHandler.RedirectURL)
}

func (a *App) setupMiddleware() {
	a.fiber.Use(recover.New())

	a.fiber.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} | ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))

	a.fiber.Use(cors.New())

	prometheus := fiberprometheus.New("tinyurl")
	prometheus.RegisterAt(a.fiber, "/metrics")
	a.fiber.Use(prometheus.Middleware)
}

func (a *App) setupHealthAndMetrics() {
	a.fiber.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "UP",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	a.fiber.Get("/metrics/dashboard", monitor.New(monitor.Config{
		Title: "TinyURL Metrics",
	}))
}

func (a *App) SetUp() *fiber.App {
	a.setupMiddleware()
	a.setupHealthAndMetrics()
	a.registerRoutes()

	return a.fiber
}

func (a *App) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdownCh
		a.logger.Info("shutdown signal received")

		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
		defer shutdownCancel()

		if err := a.fiber.ShutdownWithContext(shutdownCtx); err != nil {
			a.logger.Error("server shutdown failed", "error", err)
		}

		a.logger.Info("server gracefully shutdown")
		cancel()
	}()

	serverAddr := fmt.Sprintf(":%d", a.config.AppPort)
	a.logger.Info("server starting", "address", serverAddr)

	go func() {
		if err := a.fiber.Listen(serverAddr); err != nil && ctx.Err() == nil {
			a.logger.Error("server error", "error", err)
			cancel()
		}
	}()

	<-ctx.Done()
	a.logger.Info("application terminated")
}

func (a *App) Stop() error {
	return a.fiber.Shutdown()
}

