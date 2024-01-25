package acs

import (
	"github.com/puregrade-group/protos/gen/go/acs"
	"google.golang.org/grpc"
)

type permissionsApi struct {
	acs.UnimplementedPermissionsServer
	perms Permissions
}

type rolesApi struct {
	acs.UnimplementedRolesServer
	roles Roles
}

func Register(gRPC *grpc.Server, roles Roles, perms Permissions) {
	acs.RegisterRolesServer(gRPC, &rolesApi{roles: roles})
	acs.RegisterPermissionsServer(gRPC, &permissionsApi{perms: perms})
}
