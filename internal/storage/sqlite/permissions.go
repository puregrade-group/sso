package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/puregrade-group/sso/internal/domain/models"
	"github.com/puregrade-group/sso/internal/storage"
)

func (s *Storage) SavePermission(ctx context.Context,
	permission models.Permission,
) (permissionId int32, err error) {
	const op = "storage.sqlite.SavePermission"

	stmt, err := s.db.Prepare("INSERT INTO permissions (resource, action, description) VALUES (?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	result, err := stmt.ExecContext(ctx, permission.Resource, permission.Action, permission.Description)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return int32(lastId), nil
}

func (s *Storage) SaveRolePermission(ctx context.Context,
	roleId int32,
	permissionId int32,
) (err error) {
	const op = "storage.sqlite.SaveRolePermission"

	stmt, err := s.db.Prepare("INSERT INTO roles_permissions (role_id, permission_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, roleId, permissionId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) CheckUserPermission(ctx context.Context,
	userId [16]byte,
	resource,
	action string,
) (hasPerm bool, err error) {
	const op = "storage.sqlite.CheckUserPermission"

	stmt, err := s.db.Prepare(
		`SELECT EXISTS (
	  	SELECT 1
	  	FROM 
			users u
	  	JOIN 
			users_roles ur ON u.id = ur.user_id
	  	JOIN 
			roles r ON ur.role_id = r.id
	  	JOIN 
			roles_permissions rp ON r.id = rp.role_id
		JOIN 
			permissions p ON rp.permission_id = p.id
	  	WHERE 
			u.id = ?
			AND p.resource = ?
	  		AND p.action = ?
	) AS has_permission;`,
	)

	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	err = stmt.QueryRowContext(ctx, userId, resource, action).Scan(&hasPerm)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return hasPerm, nil
}

func (s *Storage) GetPermissionByName(ctx context.Context,
	resource,
	action string,
) (models.Permission, error) {
	const op = "storage.sqlite.GetPermissionByName"

	stmt, err := s.db.Prepare("SELECT id, description FROM permissions WHERE resource = ? AND action = ? LIMIT 1")
	if err != nil {
		return models.Permission{}, fmt.Errorf("%s: %w", op, err)
	}

	perm := models.Permission{
		Resource: resource,
		Action:   action,
	}
	err = stmt.QueryRowContext(ctx, resource, action).Scan(&perm.Id, &perm.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Permission{}, fmt.Errorf("%s: %w", op, storage.ErrPermissionNotFound)
		}

		return models.Permission{}, fmt.Errorf("%s: %w", op, err)
	}

	return perm, nil
}

func (s *Storage) DeletePermissionById(ctx context.Context,
	permissionId int32,
) (err error) {
	const op = "storage.sqlite.DeletePermissionById"

	stmt, err := s.db.Prepare("DELETE FROM permissions WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, permissionId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeletePermissionByName(ctx context.Context,
	resource,
	action string,
) (err error) {
	const op = "storage.sqlite.DeletePermissionByName"

	stmt, err := s.db.Prepare("DELETE FROM permissions WHERE resource = ? AND action = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, resource, action)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteRolePermission(ctx context.Context,
	roleId int32,
	permissionId int32,
) (err error) {
	const op = "storage.sqlite.DeleteAppPermission"

	stmt, err := s.db.Prepare("DELETE FROM roles_permissions WHERE role_id = ? AND permission_id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, roleId, permissionId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
