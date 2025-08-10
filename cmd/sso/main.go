package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kirinyoku/sso-grpc/internal/app"
	"github.com/kirinyoku/sso-grpc/internal/config"
	"github.com/kirinyoku/sso-grpc/internal/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.New(cfg)

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	go application.GRPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sig := <-stop

	log.Info("stopping application", slog.String("signal", sig.String()))

	application.GRPCSrv.Stop()
}
