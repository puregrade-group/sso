package tests

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt"
	ssov1 "github.com/puregrade-group/protos/gen/go/sso"
	myjwt "github.com/puregrade-group/sso/internal/utils/jwt"
	"github.com/puregrade-group/sso/tests/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppId       = 0
	appId      int32 = 1
	appSecret        = "test-secret"

	passDefaultLen = 10
	passMinLen     = 8
	passMaxlen     = 36
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, false, passDefaultLen)

	regResp, err := st.AuthClient.Register(
		ctx, &ssov1.RegisterRequest{
			Email:    email,
			Password: password,
			AppId:    appId,
		},
	)
	require.NoError(t, err)
	assert.NotEmpty(t, regResp.GetUserId())

	loginResp, err := st.AuthClient.Login(
		ctx, &ssov1.LoginRequest{
			Email:    email,
			Password: password,
			AppId:    appId,
		},
	)
	require.NoError(t, err)

	loginTime := time.Now()

	token := loginResp.GetToken()
	require.NotEmpty(t, token)

	parsedToken, err := jwt.ParseWithClaims(
		token, &myjwt.CustomClaims{}, func(*jwt.Token) (interface{}, error) {
			return []byte(appSecret), nil
		},
	)
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(*myjwt.CustomClaims)
	assert.True(t, ok)

	assert.Equal(t, regResp.GetUserId(), claims.UID)
	assert.Equal(t, appId, claims.AppId)

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(st.Cfg.JWTTokenTTL).Unix(), claims.ExpiresAt, deltaSeconds)
}

func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, false, passDefaultLen)

	regResp, err := st.AuthClient.Register(
		ctx, &ssov1.RegisterRequest{
			Email:    email,
			Password: password,
			AppId:    appId,
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, regResp.GetUserId())

	regResp, err = st.AuthClient.Register(
		ctx, &ssov1.RegisterRequest{
			Email:    email,
			Password: password,
			AppId:    appId,
		},
	)
	require.Error(t, err)
	assert.Empty(t, regResp.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		appId       int32
		expectedErr string
	}{
		{
			name:        "Register with empty password",
			email:       gofakeit.Email(),
			password:    "",
			appId:       appId,
			expectedErr: "password is required",
		},
		{
			name:        "Register with a password longer than passMaxLen",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, false, passMaxlen+1),
			appId:       appId,
			expectedErr: "password if longer than maximum password length",
		},
		{
			name:        "Register with a password shorter than passMinLen",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, false, passMinLen-1),
			appId:       appId,
			expectedErr: "password is shorter than minimum password length",
		},
		{
			name:        "Register with empty email",
			email:       "",
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			appId:       appId,
			expectedErr: "email is required",
		},
		{
			name:        "Register with wrong email",
			email:       gofakeit.MinecraftTool(), // for example
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			appId:       appId,
			expectedErr: "email has the wrong structure",
		},
		{
			name:        "Register with both empty",
			email:       "",
			password:    "",
			appId:       appId,
			expectedErr: "email is required",
		},
		{
			name:        "Register with empty appId",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			appId:       emptyAppId,
			expectedErr: "appId is required",
		},
		{
			name:        "Register with unknown appId",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			appId:       10000,
			expectedErr: "app is unknown",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				_, err := st.AuthClient.Register(
					ctx, &ssov1.RegisterRequest{
						Email:    tt.email,
						Password: tt.password,
						AppId:    tt.appId,
					},
				)

				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			},
		)
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		appId       int32
		expectedErr string
	}{
		{
			name:        "Login with empty password",
			email:       gofakeit.Email(),
			password:    "",
			appId:       appId,
			expectedErr: "password is required",
		},
		{
			name:        "Login with a non-matching password",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			appId:       appId,
			expectedErr: "invalid email or password",
		},
		{
			name:        "Login with empty email",
			email:       "",
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			appId:       appId,
			expectedErr: "email is required",
		},
		{
			name:        "Login with wrong email",
			email:       gofakeit.MinecraftTool(), // for example
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			appId:       appId,
			expectedErr: "email has the wrong structure",
		},
		{
			name:        "Login with both empty",
			email:       "",
			password:    "",
			appId:       appId,
			expectedErr: "email is required",
		},
		{
			name:        "Login with empty appId",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			appId:       emptyAppId,
			expectedErr: "appId is required",
		},
		{
			name:        "Login with unknown appId",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			appId:       10000,
			expectedErr: "app is unknown",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				_, err := st.AuthClient.Login(
					ctx, &ssov1.LoginRequest{
						Email:    tt.email,
						Password: tt.password,
						AppId:    tt.appId,
					},
				)

				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			},
		)
	}
}
