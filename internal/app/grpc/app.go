package grpc

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"github.com/puregrade-group/sso/internal/transport/grpc/auth"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       uint16
	host       string
}

// New creates new gRPC server app.
func New(
	log *slog.Logger,
	authService auth.Auth,
	port uint16,
	host string,
) *App {
	gRPCServer := grpc.NewServer()

	auth.Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
		host:       host,
	}
}

// MustRun runs gRPC server and panics if any error occurs.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Uint64("port", uint64(a.port)),
	)

	l, err := net.Listen("tcp", net.JoinHostPort(a.host, strconv.Itoa(int(a.port))))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server is running", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Stop stops gRPC server
func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server")

	a.gRPCServer.GracefulStop()
}
