package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

var globalLogger *slog.Logger

type Config struct {
	Level      string `json:"level"`       // "debug", "info", "warn", "error"
	Format     string `json:"format"`      // "json", "text"
	Output     string `json:"output"`      // "stdout", "stderr"
	ShowSource bool   `json:"show_source"` // показывать файл:линию
}

func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Format:     "text",
		Output:     "stdout",
		ShowSource: false,
	}
}

func Init(cfg Config) error {
	level := parseLevel(cfg.Level)

	var output io.Writer
	switch strings.ToLower(cfg.Output) {
	case "stderr":
		output = os.Stderr
	default:
		output = os.Stdout
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.ShowSource,
	}

	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	globalLogger = slog.New(handler)
	slog.SetDefault(globalLogger)

	return nil
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func GetLogger() *slog.Logger {
	if globalLogger == nil {
		Init(DefaultConfig())
	}
	return globalLogger
}

func Debug(msg string, args ...any) {
	GetLogger().Debug(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	GetLogger().DebugContext(ctx, msg, args...)
}

func Info(msg string, args ...any) {
	GetLogger().Info(msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	GetLogger().InfoContext(ctx, msg, args...)
}

func Warn(msg string, args ...any) {
	GetLogger().Warn(msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	GetLogger().WarnContext(ctx, msg, args...)
}

func Error(msg string, args ...any) {
	GetLogger().Error(msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	GetLogger().ErrorContext(ctx, msg, args...)
}

func With(args ...any) *slog.Logger {
	return GetLogger().With(args...)
}

func WithGroup(name string) *slog.Logger {
	return GetLogger().WithGroup(name)
}

func HTTPRequest(method, path string, statusCode int, duration string, args ...any) {
	allArgs := append([]any{
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status_code", statusCode),
		slog.String("duration", duration),
	}, args...)

	if statusCode >= 500 {
		Error("HTTP request failed", allArgs...)
	} else if statusCode >= 400 {
		Warn("HTTP request warning", allArgs...)
	} else {
		Info("HTTP request", allArgs...)
	}
}

func DBQuery(query string, duration string, err error, args ...any) {
	allArgs := append([]any{
		slog.String("query_type", extractQueryType(query)),
		slog.String("duration", duration),
	}, args...)

	if err != nil {
		allArgs = append(allArgs, slog.String("error", err.Error()))
		Error("Database query failed", allArgs...)
	} else {
		Debug("Database query executed", allArgs...)
	}
}

func ServiceOperation(service, operation string, args ...any) {
	allArgs := append([]any{
		slog.String("service", service),
		slog.String("operation", operation),
	}, args...)

	Info("Service operation", allArgs...)
}

func extractQueryType(query string) string {
	words := strings.Fields(strings.TrimSpace(query))
	if len(words) > 0 {
		return strings.ToUpper(words[0])
	}
	return "UNKNOWN"
}
