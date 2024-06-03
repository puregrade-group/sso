package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/puregrade-group/sso/internal/app"
	"github.com/puregrade-group/sso/internal/config"
	"github.com/puregrade-group/sso/pkg/slogpretty"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting application", slog.Any("config", cfg))

	application := app.New(
		log,
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Database,
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.SSLMode,
		cfg.GRPC.Host, cfg.GRPC.Port,
		cfg.App.NodeID,
		cfg.AccessTokenTTL, []byte(cfg.AccessTokenSecret),
		cfg.RefreshTokenTTL, cfg.RefreshTokenLength,
	)

	go application.GRPCServer.MustRun()

	// Graceful shutdown

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("stopping application", slog.String("signal", sign.String()))

	application.GRPCServer.Stop()

	log.Info("gracefully stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
