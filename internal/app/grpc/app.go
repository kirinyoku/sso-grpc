// Package grpcapp provides the gRPC server implementation for the SSO service.
// It handles the lifecycle of the gRPC server including startup, shutdown,
// and request routing to the appropriate services.
package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	authgrpc "github.com/kirinyoku/sso-grpc/internal/grpc/auth"
	"google.golang.org/grpc"
)

// App represents the gRPC server application.
// It encapsulates the gRPC server and its configuration.
type App struct {
	log        *slog.Logger // Logger for application events
	gRPCServer *grpc.Server // gRPC server instance
	port       int          // TCP port on which the server listens
}

// New creates and initializes a new gRPC application instance.
//
// Parameters:
//   - log: logger for application events
//   - port: TCP port on which the gRPC server will listen
//   - authService: authentication service implementation
//
// Returns:
//   - *App: new gRPC application instance with registered services
func New(log *slog.Logger, port int, authService authgrpc.Auth) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:        log,
		port:       port,
		gRPCServer: gRPCServer,
	}
}

// MustRun starts the gRPC server and panics if it fails to start.
// This is a convenience method for use in main() where a failure to start
// the server should terminate the application.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run starts the gRPC server and begins serving requests.
// It blocks until the server is stopped or encounters a fatal error.
//
// Returns:
//   - error: non-nil if the server fails to start or encounters a fatal error
//
// The server will attempt a graceful shutdown on receiving an interrupt signal.
func (a *App) Run() error {
	const op = "grpcapp.App.Run"

	log := a.log.With(slog.String("op", op), slog.Int("port", a.port))

	log.Info("starting gRPC server")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server started successfully", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Stop gracefully shuts down the gRPC server.
// It stops the server from accepting new connections and waits for
// existing RPCs to complete before shutting down.
//
// The function logs the shutdown process and any errors that occur.
// It's safe to call Stop multiple times.
func (a *App) Stop() {
	const op = "grpcapp.App.Stop"

	log := a.log.With(slog.String("op", op))

	log.Info("stopping gRPC server")

	a.gRPCServer.GracefulStop()

	log.Info("gRPC server stopped successfully")
}
