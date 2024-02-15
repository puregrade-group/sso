package acs

import (
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
	GetSecret(appId int32) (secret string)
	GetApp(appId int32) (app models.App)
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
