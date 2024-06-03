package tests

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt"
	myjwt "github.com/puregrade-group/sso/pkg/jwt"
	"github.com/puregrade-group/sso/pkg/protos/gen/go/auth"
	"github.com/puregrade-group/sso/tests/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	accessTokenSecret = "secret"

	passDefaultLen = 10
	passMinLen     = 8
	passMaxlen     = 36
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, false, passDefaultLen)

	regResp, err := st.AuthClient.Register(
		ctx, &auth.RegisterRequest{
			Creds: &auth.Credentials{
				Email:    email,
				Password: password,
			},
			Profile: &auth.BriefProfile{
				FirstName:   "first_name",
				DateOfBirth: timestamppb.New(gofakeit.Date()),
			},
		},
	)
	require.NoError(t, err)
	assert.NotEmpty(t, regResp.GetUserId())

	loginResp, err := st.AuthClient.Login(
		ctx, &auth.LoginRequest{
			Creds: &auth.Credentials{
				Email:    email,
				Password: password,
			},
		},
	)
	require.NoError(t, err)

	loginTime := time.Now()

	token := loginResp.GetAccessToken()
	require.NotEmpty(t, token)

	parsedToken, err := jwt.ParseWithClaims(
		token, &myjwt.DefaultClaims{}, func(*jwt.Token) (interface{}, error) {
			return []byte(accessTokenSecret), nil
		},
	)
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(*myjwt.DefaultClaims)
	assert.True(t, ok)

	assert.Equal(t, regResp.GetUserId(), claims.UID)

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(st.Cfg.AccessTokenTTL).Unix(), claims.ExpiresAt, deltaSeconds)
}

func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, false, passDefaultLen)

	regResp, err := st.AuthClient.Register(
		ctx, &auth.RegisterRequest{
			Creds: &auth.Credentials{
				Email:    email,
				Password: password,
			},
			Profile: &auth.BriefProfile{
				FirstName:   "first_name",
				LastName:    "last_name",
				DateOfBirth: timestamppb.New(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)),
			},
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, regResp.GetUserId())

	regResp, err = st.AuthClient.Register(
		ctx, &auth.RegisterRequest{
			Creds: &auth.Credentials{
				Email:    email,
				Password: password,
			},
			Profile: &auth.BriefProfile{
				FirstName:   "first_name",
				LastName:    "last_name",
				DateOfBirth: timestamppb.New(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)),
			},
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
			expectedErr: "password is required",
		},
		{
			name:        "Register with a password longer than passMaxLen",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, false, passMaxlen+1),
			expectedErr: "password length is not within the allowed range",
		},
		{
			name:        "Register with a password shorter than passMinLen",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, false, passMinLen-1),
			expectedErr: "password length is not within the allowed range",
		},
		{
			name:        "Register with empty email",
			email:       "",
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			expectedErr: "email is required",
		},
		{
			name:        "Register with wrong email",
			email:       gofakeit.MinecraftTool(), // for example
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			expectedErr: "email has the wrong structure",
		},
		{
			name:        "Register with both empty",
			email:       "",
			password:    "",
			expectedErr: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				_, err := st.AuthClient.Register(
					ctx, &auth.RegisterRequest{
						Creds: &auth.Credentials{
							Email:    tt.email,
							Password: tt.password,
						},
						Profile: &auth.BriefProfile{
							FirstName:   "first_name",
							LastName:    "last_name",
							DateOfBirth: timestamppb.New(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)),
						},
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
			expectedErr: "password is required",
		},
		{
			name:        "Login with a non-matching password",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			expectedErr: "invalid email or password",
		},
		{
			name:        "Login with empty email",
			email:       "",
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			expectedErr: "email is required",
		},
		{
			name:        "Login with wrong email",
			email:       gofakeit.MinecraftTool(), // for example
			password:    gofakeit.Password(true, true, true, true, false, passDefaultLen),
			expectedErr: "email has the wrong structure",
		},
		{
			name:        "Login with both empty",
			email:       "",
			password:    "",
			expectedErr: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				_, err := st.AuthClient.Login(
					ctx, &auth.LoginRequest{
						Creds: &auth.Credentials{
							Email:    tt.email,
							Password: tt.password,
						},
					},
				)

				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			},
		)
	}
}
