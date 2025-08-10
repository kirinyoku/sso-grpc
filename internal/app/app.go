// Package app provides the main application setup and initialization.
// It wires together all the components of the SSO service including
// gRPC server, authentication service, and storage layer.
package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/kirinyoku/sso-grpc/internal/app/grpc"
	"github.com/kirinyoku/sso-grpc/internal/services/auth"
	"github.com/kirinyoku/sso-grpc/internal/storage/sqlite"
)

// App is the root application container that holds all the application components.
// It serves as the composition root for the application's dependency graph.
type App struct {
	// GRPCSrv is the gRPC server instance that handles all incoming API requests.
	GRPCSrv *grpcapp.App
}

// New creates and initializes a new instance of the application.
// It sets up all necessary dependencies including storage, services, and the gRPC server.
//
// Parameters:
//   - log: logger instance for application-wide logging
//   - grpcPort: TCP port on which the gRPC server will listen
//   - storagePath: filesystem path for the SQLite database
//   - tokenTTL: duration for which JWT tokens should remain valid
//
// Returns:
//   - *App: fully initialized application instance
//
// Note: The function will panic if it fails to initialize the storage layer,
// as the application cannot function without a working database connection.
func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, tokenTTL)

	grpcApp := grpcapp.New(log, grpcPort, authService)

	return &App{
		GRPCSrv: grpcApp,
	}
}
