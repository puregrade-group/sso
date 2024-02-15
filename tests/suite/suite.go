package suite

import (
	"context"
	"net"
	"testing"

	"github.com/puregrade-group/protos/gen/go/acs"
	ssov1 "github.com/puregrade-group/protos/gen/go/sso"
	"github.com/puregrade-group/sso/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	Cfg         *config.Config
	AuthClient  ssov1.AuthClient
	RolesClient acs.RolesClient
	PermsClient acs.PermissionsClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadPath("../config/local_tests.yaml")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(
		func() {
			t.Helper()
			cancel()
		},
	)

	cc, err := grpc.DialContext(
		ctx,
		grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:           t,
		Cfg:         cfg,
		AuthClient:  ssov1.NewAuthClient(cc),
		RolesClient: acs.NewRolesClient(cc),
		PermsClient: acs.NewPermissionsClient(cc),
	}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(cfg.GRPC.Host, cfg.GRPC.Port)
}
