package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"github.com/puregrade-group/sso/internal/domain/models"
	"github.com/puregrade-group/sso/internal/storage"
)

func (s *Storage) SaveUser(
	ctx context.Context,
	userId [16]byte,
	email string,
	passHash []byte,
) (err error) {
	const op = "storage.sqlite.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users (id, email, pass_hash) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, userId[:], email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetUser(ctx context.Context, email string) (models.User, error) {
	const op = "storage.sqlite.GetUser"

	stmt, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email = ?")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	// TODO: Получить роли пользователя джоином таблицы с ролями
	row := stmt.QueryRowContext(ctx, email)

	var user models.User
	var userIdSlice []byte
	err = row.Scan(&userIdSlice, &user.Email, &user.PassHash)
	if len(userIdSlice) != 16 {
		return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	}
	user.Id = ([16]byte)(userIdSlice) // convert Sqlite blob([]byte) to uuid([16]byte)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) GetApp(ctx context.Context, appId int32) (models.App, error) {
	const op = "storage.sqlite.GetApp"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	// TODO: Получить пермишены приложения джоином таблицы с ролями
	row := stmt.QueryRowContext(ctx, appId)

	var app models.App
	err = row.Scan(&app.Id, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
