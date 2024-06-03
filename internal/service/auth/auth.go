package auth

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	"github.com/puregrade-group/sso/internal/domain/models"
	"github.com/puregrade-group/sso/internal/transport/grpc/auth"
	"github.com/puregrade-group/sso/pkg/jwt"
	"github.com/puregrade-group/sso/pkg/snowflake"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	snowflake *snowflake.Snowflake
	log       *slog.Logger
	// Database layer interfaces
	usrSaver             UserSaver
	usrProvider          UserProvider
	refreshTokenProvider RefreshTokenProvider
	// Service configs
	accessTokenTTL     time.Duration
	accessTokenSecret  []byte
	refreshTokenTTL    time.Duration
	refreshTokenLength uint
}

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrRefreshTokenNotFound = errors.New("token with such content has not been provided to anyone")
	ErrUserAlreadyExists    = errors.New("user is already exists")
	ErrInternal             = errors.New("internal error")
	ErrUnknown              = errors.New("unknown error")
)

// UserSaver interface must be implemented by the repository layer
type UserSaver interface {
	SaveUser(ctx context.Context,
		creds models.UserCredentials,
		profile models.Profile,
	) (err error)
}

// UserProvider interface must be implemented by the repository layer
type UserProvider interface {
	GetUserCreds(ctx context.Context,
		email string,
	) (models.UserCredentials, error)
}

type RefreshTokenProvider interface {
	Upsert(ctx context.Context,
		userId uint64,
		token string,
		expiresIn time.Time,
	) (err error)
	GetUserId(ctx context.Context,
		token string,
	) (userId uint64, err error)
}

func New(
	log *slog.Logger,
	snowflake *snowflake.Snowflake,
	userSaver UserSaver,
	userProvider UserProvider,
	refreshTokenProvider RefreshTokenProvider,
	accessTokenTTL time.Duration,
	accessTokenSecret []byte,
	refreshTokenTTL time.Duration,
	refreshTokenLength uint,
) *Auth {
	return &Auth{
		usrSaver:             userSaver,
		usrProvider:          userProvider,
		refreshTokenProvider: refreshTokenProvider,
		snowflake:            snowflake,
		log:                  log,
		accessTokenTTL:       accessTokenTTL,
		accessTokenSecret:    accessTokenSecret,
		refreshTokenTTL:      refreshTokenTTL,
		refreshTokenLength:   refreshTokenLength,
	}
}

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
func (a *Auth) Login(ctx context.Context,
	creds models.Credentials,
) (accessToken, refreshToken string, err error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to login user")

	// Getting user credentials
	user, err := a.usrProvider.GetUserCreds(ctx, creds.Email)

	// Error handling
	if err != nil {
		log.Error(err.Error())

		switch {
		case err == nil: // Do nothing
		case errors.Is(err, ErrUserNotFound):
			return "", "", auth.ErrWrongCredentials
		case errors.Is(err, ErrInternal):
			return "", "", auth.ErrInternal
		default:
			return "", "", auth.ErrUnknown
		}
	}

	// Password comparing
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(creds.Password)); err != nil {
		log.Error(err.Error())

		return "", "", auth.ErrWrongCredentials
	}

	// Creating new JWT token
	accessToken, err = jwt.NewToken(user.Id, a.accessTokenTTL, a.accessTokenSecret)
	if err != nil {
		log.Error(err.Error())

		return "", "", auth.ErrInternal
	}

	log.Debug(
		"access token was successfully created",
		slog.String("token", accessToken),
	)

	// Creating refresh token
	refreshToken = genRandomString(a.refreshTokenLength)

	log.Debug("refresh token was generated", slog.String("token", refreshToken))

	err = a.refreshTokenProvider.Upsert(ctx, user.Id, refreshToken, time.Now().Add(a.refreshTokenTTL))

	// Error handling
	if err != nil {
		log.Error(err.Error())

		switch {
		case errors.Is(err, ErrInternal):
			return "", "", auth.ErrInternal
		default:
			return "", "", auth.ErrUnknown
		}
	}

	log.Debug(
		"refresh token was successfully upserted",
		slog.String("token", refreshToken),
	)

	log.Info("user logged in successfully")

	return accessToken, refreshToken, nil
}

// RegisterNewUser checks if user with given credentials not exists in the system and returns new user userId.
//
// If user already exists, returns error.
// If data don't pass validation process, returns error.
func (a *Auth) RegisterNewUser(ctx context.Context,
	creds models.Credentials,
	profile models.BriefProfile,
) (uint64, error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to register user")

	// Generating password hash
	passHash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 5)
	if err != nil {
		log.Error(err.Error())

		return 0, auth.ErrInternal
	}

	// Creating internal version of user credentials with a new Snowflake ID
	userCreds := models.UserCredentials{
		Id:           a.snowflake.Generate(),
		Email:        creds.Email,
		PasswordHash: passHash,
	}

	fullProfile := models.Profile{BriefProfile: profile}

	// Save user if he is not already exists
	err = a.usrSaver.SaveUser(ctx, userCreds, fullProfile)
	if err != nil {
		log.Error(err.Error())

		switch {
		case errors.Is(err, ErrUserAlreadyExists):
			return 0, auth.ErrUserAlreadyExists
		case errors.Is(err, ErrInternal):
			return 0, auth.ErrInternal
		default:
			return 0, auth.ErrUnknown
		}
	}

	log.Debug(
		"new user was created",
		slog.Uint64("id", userCreds.Id),
	)

	return userCreds.Id, nil
}

func (a *Auth) RefreshTokens(
	ctx context.Context,
	token string,
) (accessToken, refreshToken string, err error) {
	const op = "Auth.RefreshTokens"

	log := a.log.With(slog.String("op", op))

	// Getting user id by refresh token
	id, err := a.refreshTokenProvider.GetUserId(ctx, token)

	// Error handling
	if err != nil {
		log.With(slog.String("token", token)).Error(err.Error())

		switch {
		case errors.Is(err, ErrRefreshTokenNotFound):
			return "", "", auth.ErrTokenNotFound
		case errors.Is(err, ErrInternal):
			return "", "", auth.ErrInternal
		default:
			return "", "", auth.ErrUnknown
		}
	}

	// Creating new JWT token
	accessToken, err = jwt.NewToken(id, a.accessTokenTTL, a.accessTokenSecret)
	if err != nil {
		log.Error(err.Error())

		return "", "", auth.ErrInternal
	}

	// Generating new refresh token
	refreshToken = genRandomString(a.refreshTokenLength)
	log.Debug("refresh token was generated", slog.String("token", refreshToken))

	// Upserting refresh token
	err = a.refreshTokenProvider.Upsert(ctx, id, refreshToken, time.Now().Add(a.refreshTokenTTL))

	// Error handling
	if err != nil {
		log.Error(err.Error())

		switch {
		case errors.Is(err, ErrInternal):
			return "", "", auth.ErrInternal
		default:
			return "", "", auth.ErrUnknown
		}
	}

	log.Debug(
		"",
		slog.String("accessToken", accessToken),
		slog.String("refreshToken", refreshToken),
	)

	log.Info("tokens was refreshed")

	return accessToken, refreshToken, nil
}

func genRandomString(length uint) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]uint8, length)

	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}

	return string(b)
}
