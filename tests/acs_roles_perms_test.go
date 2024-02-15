package tests

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/puregrade-group/protos/gen/go/acs"
	ssov1 "github.com/puregrade-group/protos/gen/go/sso"
	"github.com/puregrade-group/sso/tests/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRolesPerms_Promotion_Demotion_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, false, 10)

	register, err := st.AuthClient.Register(
		ctx, &ssov1.RegisterRequest{
			Email:    email,
			Password: password,
			AppId:    appId,
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, register.UserId)

	login, err := st.AuthClient.Login(
		ctx, &ssov1.LoginRequest{
			Email:    email,
			Password: password,
			AppId:    appId,
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, login.Token)

	permCreate, err := st.PermsClient.Create(
		ctx, &acs.CreatePermissionRequest{
			RequesterToken: login.Token,
			Permission: &acs.Permission{
				PermissionId: nil,
				Resource:     gofakeit.Username(),
				Action:       gofakeit.Username(),
				Type:         1,
				Description:  gofakeit.Username(),
			},
		},
	)
	require.NoError(t, err)
	assert.NotEmpty(t, permCreate.PermissionId)

	role := &acs.Role{
		RoleId:      nil,
		Name:        gofakeit.Username(),
		Description: gofakeit.Username(),
		Permissions: []int32{permCreate.PermissionId},
	}

	roleCreate, err := st.RolesClient.Create(
		ctx, &acs.CreateRoleRequest{
			RequesterToken: login.Token,
			Role:           role,
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, roleCreate.RoleId)

	_, err = st.RolesClient.Add(
		ctx, &acs.AddRoleRequest{
			RequesterToken: login.Token,
			UserId:         register.UserId,
			RoleId:         roleCreate.RoleId,
		},
	)
	require.NoError(t, err)

	roleGet, err := st.RolesClient.GetUserRoles(
		ctx, &acs.GetUserRolesRequest{
			RequesterToken: login.Token,
			UserId:         register.UserId,
		},
	)
	require.NoError(t, err)
	require.Contains(t, roleGet.Roles, role)
}
