package app

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/russo2642/renti_kz/internal/config"
	"github.com/russo2642/renti_kz/pkg/auth"
	"github.com/russo2642/renti_kz/pkg/logger"
	"github.com/russo2642/renti_kz/pkg/storage/s3"
)

func initDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	logger.Info("initializing database connection",
		slog.String("host", cfg.Host),
		slog.String("port", cfg.Port),
		slog.String("database", cfg.Name))

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		logger.Error("failed to open database connection", slog.String("error", err.Error()))
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", slog.String("error", err.Error()))
		return nil, err
	}

	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(100)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	logger.Info("database connection established with optimized pool",
		slog.Int("max_open_conns", 200),
		slog.Int("max_idle_conns", 100),
		slog.Duration("conn_max_lifetime", 30*time.Minute))

	return db, nil
}

func initS3Storage(cfg config.S3Config) (*s3.Storage, error) {
	logger.Info("initializing S3 storage",
		slog.String("region", cfg.Region),
		slog.String("bucket", cfg.Bucket),
		slog.String("access_key", maskCredential(cfg.AccessKey)))

	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		logger.Warn("AWS credentials are not set, S3 storage may not work correctly")
	}

	s3Storage, err := s3.NewStorage(cfg.Region, cfg.Bucket, cfg.AccessKey, cfg.SecretKey)
	if err != nil {
		logger.Error("failed to initialize S3 storage", slog.String("error", err.Error()))
		return nil, err
	}

	logger.Info("S3 storage initialized successfully")
	return s3Storage, nil
}

func initTokenManager(cfg config.JWTConfig) auth.TokenManager {
	logger.Info("initializing JWT token manager",
		slog.Duration("access_ttl", cfg.AccessTTL),
		slog.Duration("refresh_ttl", cfg.RefreshTTL))

	return auth.NewJWTManager(cfg.AccessSecret, cfg.RefreshSecret, cfg.AccessTTL, cfg.RefreshTTL)
}

func maskCredential(credential string) string {
	if len(credential) <= 8 {
		return "***"
	}
	return credential[:4] + "***" + credential[len(credential)-4:]
}
