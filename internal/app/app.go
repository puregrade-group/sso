package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/puregrade-group/sso/internal/app/grpc"
	"github.com/puregrade-group/sso/internal/service/auth"
	"github.com/puregrade-group/sso/internal/storage/postgres"
	"github.com/puregrade-group/sso/pkg/snowflake"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	// Postgres configuration
	postgresHost string, postgresPort uint16, postgresDatabase,
	postgresUser, postgresPassword, postgresSSLMode string,
	// GRPC server configuration
	grpcHost string, grpcPort uint16,
	// AuthService configuration
	nodeID uint16,
	accessTokenTTL time.Duration, accessTokenSecret []byte,
	refreshTokenTTL time.Duration, refreshTokenLength uint,
) *App {
	storage, err := postgres.New(
		postgres.Config{
			Host:     postgresHost,
			Port:     postgresPort,
			Database: postgresDatabase,
			User:     postgresUser,
			Password: postgresPassword,
			SSLMode:  postgresSSLMode,
		},
	)
	if err != nil {
		panic(err)
	}

	sf, err := snowflake.NewSnowflake(nodeID)
	if err != nil {
		panic(err)
	}

	authService := auth.New(
		log, sf,
		storage, storage, storage,
		accessTokenTTL, accessTokenSecret,
		refreshTokenTTL, refreshTokenLength,
	)

	grpcApp := grpcapp.New(log, authService, grpcPort, grpcHost)

	return &App{
		GRPCServer: grpcApp,
	}
}
