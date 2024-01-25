package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/puregrade-group/sso/internal/domain/models"
)

func (s *Storage) SaveRole(ctx context.Context,
	role models.Role,
) (roleId int32, err error) {
	const op = "storage.sqlite.SaveRole"

	stmt, err := s.db.Prepare("INSERT INTO roles (name, description) VALUES (?, ?)")
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	result, err := stmt.ExecContext(ctx, role.Name, role.Description)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// TODO: переписать с транзакцией и strings.Builder
	q := "INSERT INTO roles_permissions (role_id, permission_id) VALUES"
	for _, p := range role.Permissions {
		q += fmt.Sprintf(" (%s, %s),", strconv.FormatInt(lastId, 10), string(p.Id))
	}
	q = q[:len(q)-1]

	q += "ON CONFLICT DO NOTHING;"

	_, err = s.db.ExecContext(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return int32(lastId), nil
}

func (s *Storage) SaveUserRole(ctx context.Context,
	userId [16]byte,
	roleId int32,
) (err error) {
	const op = "storage.sqlite.SaveUserRole"

	stmt, err := s.db.Prepare("INSERT INTO users_roles (user_id, role_id) VALUES (?, ?)")
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, userId, roleId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetUserRoles(ctx context.Context,
	userId [16]byte,
) (roles []models.Role, err error) {
	const op = "storage.sqlite.GetUserRoles"

	stmt, err := s.db.Prepare(
		`SELECT r.id, r.name, r.description 
		FROM users_roles ur 
		JOIN roles r on ur.role_id = r.id 
		WHERE ur.user_id = ?`,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx, userId)
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for rows.Next() {
		var r models.Role
		if err := rows.Scan(&r); err != nil {
			log.Fatal(err)
		}
		roles = append(roles, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return roles, nil
}

func (s *Storage) CheckUserRoles(ctx context.Context,
	userId [16]byte,
	roles []models.Role,
) (ok bool, err error) {
	const op = "storage.sqlite.CheckUserRoles"

	// TODO: имплементировать
	panic("implement me")

	return false, err
}

func (s *Storage) DeleteRole(ctx context.Context,
	roleName string,
) (roleId int32, err error) {
	const op = "storage.sqlite.DeleteRole"

	stmt, err := s.db.Prepare("DELETE FROM roles WHERE name = ?")
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	result, err := stmt.ExecContext(ctx, roleName)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return int32(id), nil
}

func (s *Storage) DeleteUserRole(ctx context.Context,
	userId [16]byte,
	roleId int32,
) (err error) {
	const op = "storage.sqlite.DeleteUserRole"

	stmt, err := s.db.Prepare("DELETE FROM users_roles WHERE role_id = ? AND user_id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, roleId, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
