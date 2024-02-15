package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/puregrade-group/sso/internal/domain/models"
	"github.com/puregrade-group/sso/internal/storage"
	"github.com/puregrade-group/sso/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user is already exists")
	ErrUnknownApp         = errors.New("app is unknown") // application is not registered in the system
)

// UserSaver interface must be implemented by the repository layer
type UserSaver interface {
	SaveUser(
		ctx context.Context,
		userId [16]byte,
		email string,
		passHash []byte,
	) (err error)
}

// UserProvider interface must be implemented by the repository layer
type UserProvider interface {
	GetUser(ctx context.Context, email string) (models.User, error)
}

// AppProvider interface must be implemented by the repository layer
type AppProvider interface {
	GetApp(ctx context.Context, appId int32) (models.App, error)
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		log:         log,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
// If app is unknown, returns error.
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appId int32,
) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to login user")

	// get app
	app, err := a.appProvider.GetApp(ctx, appId)
	if err != nil {
		a.log.Warn(
			"unknown app", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return "", fmt.Errorf("%s: %w", op, ErrUnknownApp)
	}

	// get user
	user, err := a.usrProvider.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn(
				"user not found", slog.Attr{
					Key:   "error",
					Value: slog.StringValue(err.Error()),
				},
			)

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error(
			"failed to get user", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	// compare passwords
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info(
			"invalid credentials", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("user logged in successfully")

	// create new jwt token
	token, err := jwt.NewToken(user.Id, app, a.tokenTTL)
	if err != nil {
		a.log.Error(
			"failed to generate token", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	a.log.Debug(
		"user token been successfully created", slog.Attr{
			Key:   "token",
			Value: slog.StringValue(token),
		},
	)

	return token, nil
}

// RegisterNewUser checks if user with given credentials not exists in the system and returns new user userId.
//
// If user already exists, returns error.
// If data don't pass validation process, returns error.
func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
	appId int32,
) (userId string, err error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to register user")

	// get app
	_, err = a.appProvider.GetApp(ctx, appId)
	if err != nil {
		a.log.Warn(
			"unknown app", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return "", fmt.Errorf("%s: %w", op, ErrUnknownApp)
	}

	// generate password hash
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), 5)
	if err != nil {
		a.log.Error(
			"failed to generate password hash", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	// generate new uuid
	id := uuid.New()

	// save user if it not already exists
	err = a.usrSaver.SaveUser(ctx, id, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			a.log.Error(
				"user is already exists", slog.Attr{
					Key:   "error",
					Value: slog.StringValue(err.Error()),
				},
			)

			return "", fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
		}

		a.log.Error(
			"failed to save user", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			},
		)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id.String(), nil
}
