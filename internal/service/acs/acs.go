package acs

import (
	"log/slog"
)

type ACS struct {
	log          *slog.Logger
	permSaver    PermissionsSaver
	permProvider PermissionsProvider
	permRemover  PermissionRemover
	roleSaver    RoleSaver
	roleProvider RoleProvider
	roleRemover  RoleRemover
}

func New(
	log *slog.Logger,
	permSaver PermissionsSaver,
	permProvider PermissionsProvider,
	permRemover PermissionRemover,
) *ACS {
	return &ACS{
		log:          log,
		permSaver:    permSaver,
		permProvider: permProvider,
		permRemover:  permRemover,
	}
}
