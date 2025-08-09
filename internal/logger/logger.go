// Package logger provides functionality for creating a logger instance.
package logger

import (
	"log/slog"
	"os"

	"github.com/kirinyoku/sso-grpc/internal/config"
)

// New creates a new logger instance based on the application environment.
//
// Parameters:
//   - cfg: application configuration
//
// Returns:
//   - *slog.Logger: new logger instance
func New(cfg *config.Config) *slog.Logger {
	var log *slog.Logger

	switch cfg.Env {
	case "local":
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case "dev":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	default:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	return log
}
