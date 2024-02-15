package acs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/puregrade-group/sso/internal/domain/models"
	"github.com/puregrade-group/sso/internal/storage"
	"github.com/puregrade-group/sso/pkg/jwt"
)

var (
	ErrNotEnoughPermissions = errors.New("user do not have enough permissions")
	ErrTokenNotValid        = errors.New("requester token in not valid")
)

type PermissionsSaver interface {
	SavePermission(ctx context.Context,
		permission models.Permission,
	) (permissionId int32, err error)
	SaveRolePermission(ctx context.Context,
		roleId int32,
		permissionId int32,
	) (err error)
}

type PermissionsProvider interface {
	CheckUserPermission(ctx context.Context,
		userId [16]byte,
		resource,
		action string,
	) (hasPerm bool, err error)
	GetPermissionByName(ctx context.Context,
		resource,
		action string,
	) (permission models.Permission, err error)
}

type PermissionRemover interface {
	DeletePermission(ctx context.Context,
		permissionId int32,
	) (err error)
	DeleteRolePermission(ctx context.Context,
		roleId int32,
		permissionId int32,
	) (err error)
}

func (a *ACS) CreatePermission(ctx context.Context,
	requesterToken,
	resource,
	action,
	description string,
) (id int32, err error) {
	const op = "ACS.CreatePermission"

	log := a.log.With(
		slog.String("op", op),
		slog.String("resource:action", resource+":"+action),
	)

	log.Info("attempting to create new permission")

	claims, err := parseToken(log, op, requesterToken, a.appProvider.GetSecret)
	if err != nil {
		return 0, err
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

		return 0, fmt.Errorf("userId is wrong")
	}

	hasPerm, err := a.permProvider.CheckUserPermission(ctx, parsedId, "permission", "create")
	if err != nil {
		log.Error(
			"internal", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if !hasPerm {
		log.Warn(
			"user does not have permission to execute this request", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return 0, fmt.Errorf("%s: %w", op, ErrNotEnoughPermissions)
	}

	p := models.Permission{
		Id:          0,
		Resource:    resource,
		Action:      action,
		Description: description,
	}

	id, err = a.permSaver.SavePermission(ctx, p)

	if err != nil {
		log.Error(
			"internal", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (a *ACS) CheckUserPermission(ctx context.Context,
	requesterToken string,
	userId [16]byte,
	resource,
	action string,
) (ok bool, err error) {
	const op = "ACS.CheckUserPermission"

	log := a.log.With(
		slog.String("op", op),
	)

	_, err = parseToken(log, op, requesterToken, a.appProvider.GetSecret)
	if err != nil {
		return false, err
	}

	hasPerm, err := a.permProvider.CheckUserPermission(ctx, userId, resource, action)
	if err != nil {
		log.Error(
			"db error", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return false, err
	}

	return hasPerm, nil
}

func (a *ACS) DeletePermission(ctx context.Context,
	requesterToken string,
	permissionId int32,
) (err error) {
	const op = "ACS.DeletePermission"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("permId", int(permissionId)),
	)

	claims, err := parseToken(log, op, requesterToken, a.appProvider.GetSecret)
	if err != nil {
		return err
	}

	parsedId, err := uuid.Parse(claims.UID)
	if err != nil {
		log.Error(
			"parse token userId failed", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	hasPerm, err := a.permProvider.CheckUserPermission(ctx, parsedId, "permission", "delete")
	if err != nil {
		log.Error(
			"db error", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return err
	}

	if !hasPerm {
		log.Warn(
			"not enough permissions", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return ErrNotEnoughPermissions
	}

	err = a.permRemover.DeletePermission(ctx, permissionId)
	if err != nil {
		log.Error(
			"parse token userId failed", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *ACS) AddPermission(ctx context.Context,
	requesterToken string,
	roleId int32,
	resource,
	action string,
) (err error) {
	const op = "ACS.AddPermission"

	log := a.log.With(
		slog.String("op", op),
		slog.String("permission", resource+":"+action),
	)

	claims, err := parseToken(log, op, requesterToken, a.appProvider.GetSecret)
	if err != nil {
		return err
	}

	parsedId, err := uuid.Parse(claims.UID)
	if err != nil {
		log.Error(
			"parse token userId failed", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	hasPerm, err := a.permProvider.CheckUserPermission(ctx, parsedId, "permission", "grant")
	if err != nil {
		log.Error(
			"db error", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return err
	}

	if !hasPerm {
		log.Warn(
			"not enough permissions", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return ErrNotEnoughPermissions
	}

	perm, err := a.permProvider.GetPermissionByName(ctx, resource, action)
	if err != nil {
		if errors.Is(err, storage.ErrPermissionNotFound) {
			log.Error(
				"permission was not found", slog.Attr{
					Key:   "error",
					Value: slog.StringValue(err.Error()),
				},
			)

			return storage.ErrPermissionNotFound
		}

		log.Error(
			"db error", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return err
	}

	err = a.permSaver.SaveRolePermission(ctx, roleId, perm.Id)
	if err != nil {
		log.Error(
			"db error", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return err
	}

	return nil
}

func (a *ACS) RemovePermission(ctx context.Context,
	requesterToken string,
	roleId int32,
	resource,
	action string,
) (err error) {
	const op = "ACS.RemovePermission"

	log := a.log.With(
		slog.String("op", op),
		slog.String("permission", resource+":"+action),
	)

	claims, err := parseToken(log, op, requesterToken, a.appProvider.GetSecret)
	if err != nil {
		return err
	}

	parsedId, err := uuid.Parse(claims.UID)
	if err != nil {
		log.Error(
			"parse token userId failed", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	hasPerm, err := a.permProvider.CheckUserPermission(ctx, parsedId, "permission", "revoke")
	if err != nil {
		log.Error(
			"db error", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return err
	}

	if !hasPerm {
		log.Warn(
			"not enough permissions", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return ErrNotEnoughPermissions
	}

	perm, err := a.permProvider.GetPermissionByName(ctx, resource, action)
	if err != nil {
		if errors.Is(err, storage.ErrPermissionNotFound) {
			log.Error(
				"permission was not found", slog.Attr{
					Key:   "error",
					Value: slog.StringValue(err.Error()),
				},
			)

			return storage.ErrPermissionNotFound
		}

		log.Error(
			"db error", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return err
	}

	err = a.permRemover.DeleteRolePermission(ctx, roleId, perm.Id)
	if err != nil {
		log.Error(
			"db error", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return err
	}

	return nil
}

func parseToken(
	log *slog.Logger,
	op, token string,
	secret func(appId int32) string,
) (*jwt.DefaultClaims, error) {
	t, err := jwt.ParseToken(token, secret)
	if err != nil {
		log.Error(
			"token is not valid", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return &jwt.DefaultClaims{}, fmt.Errorf("%s: %w", op, ErrTokenNotValid)
	}

	claims, ok := t.Claims.(*jwt.DefaultClaims)
	if !ok {
		log.Error(
			"token is not valid", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return &jwt.DefaultClaims{}, fmt.Errorf("%s: %w", op, ErrTokenNotValid)
	}

	return claims, nil
}
