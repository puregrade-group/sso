package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/puregrade-group/sso/internal/domain/models"
)

var (
	ErrWrongClaims  = errors.New("wrong claims")
	ErrWrongMethod  = errors.New("wrong sign method")
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token has expired")
	ErrUnknownApp   = errors.New("app with this appId is unknown")
)

type DefaultClaims struct {
	AppId int32  `json:"appId"`
	UID   uint64 `json:"UID"`
	jwt.StandardClaims
}

// NewToken creates new JWT token for given user.
func NewToken(userId uint64, duration time.Duration, secret []byte) (string, error) {
	claims := DefaultClaims{
		UID: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(duration).Unix(),
		},
	}

	// Creating new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signing token with a secret
	tokenString, err := token.SignedString(secret)

	return tokenString, err
}

// NewToken1 creates new JWT token for given user and app.
func NewToken1(userId uint64, app models.App, duration time.Duration) (string, error) {
	claims := DefaultClaims{
		UID:   userId,
		AppId: app.Id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(duration).Unix(),
		},
	}

	// create new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign with a secret
	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken function checks the validity of the token and parses data from its payload.
// The "Secret" parameter is a function that allows you to obtain the JWT secret key by application ID.
func ParseToken(tokenString string,
	secret func(ctx context.Context, appId int32) (string, error),
) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&DefaultClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrWrongMethod
			}
			if claims, ok := token.Claims.(*DefaultClaims); ok {
				if claims.ExpiresAt < time.Now().Unix() {
					return nil, ErrTokenExpired
				}

				s, err := secret(context.Background(), claims.AppId)
				if err != nil {
					return nil, ErrUnknownApp
				}
				return s, nil
			}
			return nil, ErrWrongClaims
		},
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return token, nil
}

func ParseAppToken(tokenString string, secret string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&DefaultClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrWrongMethod
			}
			if claims, ok := token.Claims.(*DefaultClaims); ok {
				if claims.ExpiresAt < time.Now().Unix() {
					return nil, ErrTokenExpired
				}
			}
			return secret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return token, nil
}
