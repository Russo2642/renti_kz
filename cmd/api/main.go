package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/russo2642/renti_kz/internal/app"
	"github.com/russo2642/renti_kz/internal/config"
	"github.com/russo2642/renti_kz/pkg/logger"

	_ "github.com/russo2642/renti_kz/docs"
)

// @title Renti.kz API
// @version 1.0
// @description API для платформы аренды квартир Renti.kz
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer token для авторизации

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := logger.Init(logger.Config{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		Output:     cfg.Log.Output,
		ShowSource: cfg.Log.ShowSource,
	}); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info("starting renti.kz application",
		slog.String("version", "1.0"),
		slog.String("environment", cfg.App.Environment),
		slog.String("log_level", cfg.Log.Level))

	app, err := app.InitApp(cfg)
	if err != nil {
		logger.Error("failed to initialize application", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		logger.Info("cleaning up application resources")
		app.Cleanup()
	}()

	srv := app.GetHTTPServer()

	go func() {
		logger.Info("starting HTTP server",
			slog.String("port", cfg.Server.HttpPort),
			slog.Duration("read_timeout", cfg.Server.ReadTimeout),
			slog.Duration("write_timeout", cfg.Server.WriteTimeout))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start HTTP server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("received shutdown signal", slog.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Info("shutting down HTTP server gracefully")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced server shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("application stopped successfully")
}
