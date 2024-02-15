package acs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/puregrade-group/sso/internal/domain/models"
)

type RoleSaver interface {
	SaveRole(ctx context.Context,
		role models.Role,
	) (roleId int32, err error)
	SaveUserRole(ctx context.Context,
		userId [16]byte,
		roleId int32,
	) (err error)
}

type RoleProvider interface {
	GetUserRoles(ctx context.Context,
		userId [16]byte,
	) (roles []models.Role, err error)
	CheckUserRoles(ctx context.Context,
		userId [16]byte,
		roles []models.Role,
	) (ok bool, err error)
}

type RoleRemover interface {
	DeleteRole(ctx context.Context,
		roleName string,
	) (roleId int32, err error)
	DeleteUserRole(ctx context.Context,
		userId [16]byte,
		roleId int32,
	) (err error)
}

func (a *ACS) CreateRole(ctx context.Context,
	requesterToken string,
	role models.Role,
) (err error) {
	const op = "ACS.CreateRole"

	log := a.log.With(
		slog.String("op", op),
		slog.String("roleName", role.Name),
	)

	log.Info("attempting to create new role")

	claims, err := parseToken(log, op, requesterToken, a.appProvider.GetSecret)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// cast string userId to [16]byte uuid
	parsedId, err := uuid.Parse(claims.UID)
	if err != nil {
		log.Error(
			"UID field in token claims is wrong", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	hasPerm, err := a.permProvider.CheckUserPermission(ctx, parsedId, "role", "create")
	if err != nil {
		log.Error(
			"internal", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	if !hasPerm {
		log.Warn(
			"user does not have permission to execute this request", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, ErrNotEnoughPermissions)
	}

	_, err = a.roleSaver.SaveRole(ctx, role)

	if err != nil {
		log.Error(
			"internal", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *ACS) GetUserRoles(ctx context.Context,
	requesterToken string,
	userId [16]byte,
) (roles []models.Role, err error) {
	const op = "ACS.GetRoles"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to get user roles")

	_, err = parseToken(log, op, requesterToken, a.appProvider.GetSecret)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	roles, err = a.roleProvider.GetUserRoles(ctx, userId)

	if err != nil {
		log.Error(
			"internal", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return roles, nil
}

func (a *ACS) DeleteRole(ctx context.Context,
	requesterToken string,
	roleName string,
) (err error) {
	const op = "ACS.DeleteRole"

	log := a.log.With(
		slog.String("op", op),
		slog.String("roleName", roleName),
	)

	log.Info("attempting to delete role")

	claims, err := parseToken(log, op, requesterToken, a.appProvider.GetSecret)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// cast string userId to [16]byte uuid
	parsedId, err := uuid.Parse(claims.UID)
	if err != nil {
		log.Error(
			"UID field in token claims is wrong", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	hasPerm, err := a.permProvider.CheckUserPermission(ctx, parsedId, "role", "delete")
	if err != nil {
		log.Error(
			"internal", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	if !hasPerm {
		log.Warn(
			"user does not have permission to execute this request", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, ErrNotEnoughPermissions)
	}

	roleId, err := a.roleRemover.DeleteRole(ctx, roleName)

	log.Info(
		"role been deleted", slog.Attr{
			Key:   "roleId",
			Value: slog.IntValue(int(roleId)),
		},
	)

	if err != nil {
		log.Error(
			"internal", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *ACS) AddRole(ctx context.Context,
	requesterToken string,
	userId [16]byte,
	roleId int32,
) (err error) {
	const op = "ACS.AddRole"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to add role to user")

	_, err = parseToken(log, op, requesterToken, a.appProvider.GetSecret)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.roleSaver.SaveUserRole(ctx, userId, roleId)

	if err != nil {
		log.Error(
			"internal", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *ACS) RemoveRole(ctx context.Context,
	requesterToken string,
	userId [16]byte,
	roleId int32,
) (err error) {
	const op = "ACS.RemoveRole"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to add role to user")

	_, err = parseToken(log, op, requesterToken, a.appProvider.GetSecret)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.roleRemover.DeleteUserRole(ctx, userId, roleId)

	if err != nil {
		log.Error(
			"internal", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
