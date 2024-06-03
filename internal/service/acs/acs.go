package acs

import (
	"context"
	"log/slog"

	"github.com/puregrade-group/sso/internal/domain/models"
)

type ACS struct {
	log          *slog.Logger
	permSaver    PermissionsSaver
	permProvider PermissionsProvider
	permRemover  PermissionRemover
	roleSaver    RoleSaver
	roleProvider RoleProvider
	roleRemover  RoleRemover
	appProvider  AppProvider
}

type AppProvider interface {
	GetSecret(ctx context.Context, appId int32) (string, error)
	GetApp(ctx context.Context, appId int32) (models.App, error)
}

func New(
	log *slog.Logger,
	permSaver PermissionsSaver,
	permProvider PermissionsProvider,
	permRemover PermissionRemover,
	roleSaver RoleSaver,
	roleProvider RoleProvider,
	roleRemover RoleRemover,
	appProvider AppProvider,
) *ACS {
	return &ACS{
		log:          log,
		permSaver:    permSaver,
		permProvider: permProvider,
		permRemover:  permRemover,
		roleSaver:    roleSaver,
		roleProvider: roleProvider,
		roleRemover:  roleRemover,
		appProvider:  appProvider,
	}
}
