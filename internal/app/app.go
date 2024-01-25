package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/puregrade-group/sso/internal/app/grpc"
	"github.com/puregrade-group/sso/internal/service/auth"
	"github.com/puregrade-group/sso/internal/storage/sqlite"
)

type App struct {
	GRPCServer *grpcapp.App
	// TODO: HTTPServer *httpapp.App
}

func New(
	log *slog.Logger,
	grpcPort string,
	grpcHost string,
	storagePath string,
	jwtTokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, jwtTokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort, grpcHost)

	return &App{
		GRPCServer: grpcApp,
	}
}
