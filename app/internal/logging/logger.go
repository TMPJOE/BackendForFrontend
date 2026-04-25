// Package logging provides structured logging for the application.
// It configures a JSON-formatted logger using slog and httplog,
// producing ECS-compatible log output suitable for log aggregation systems.
package logging

import (
	"log/slog"
	"os"
	"strings"

	"github.com/go-chi/httplog/v3"
)

// New creates a new structured logger with the specified level and format
func New(level, format string) *slog.Logger {
	// Parse log level
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Use ECS schema for structured logging
	logFormat := httplog.SchemaECS.Concise(true)

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:       logLevel,
		ReplaceAttr: logFormat.ReplaceAttr,
	}

	// Create JSON handler
	handler := slog.NewJSONHandler(os.Stdout, opts)

	// Create logger with service context
	logger := slog.New(handler).With(
		slog.String("service", "bff-service"),
		slog.String("version", "v1.0.0"),
		slog.String("env", GetEnvOrDefault("ENV", "development")),
	)

	return logger
}

// GetEnvOrDefault returns the value of the environment variable or the default if not set
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
