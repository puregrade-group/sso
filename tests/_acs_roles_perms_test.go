package tests

//
// import (
// 	"testing"
//
// 	"github.com/brianvoe/gofakeit/v6"
// 	"github.com/puregrade-group/sso/pkg/protos/gen/go/acs"
// 	ssov1 "github.com/puregrade-group/sso/pkg/protos/gen/go/sso"
// 	"github.com/puregrade-group/sso/tests/suite"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )
//
// func TestRolesPerms_Promotion_Demotion_HappyPath(t *testing.T) {
// 	ctx, st := suite.New(t)
//
// 	email := gofakeit.Email()
// 	password := gofakeit.Password(true, true, true, true, false, 10)
//
// 	register, err := st.AuthClient.Register(
// 		ctx, &ssov1.RegisterRequest{
// 			Email:    email,
// 			Password: password,
// 			AppId:    appId,
// 		},
// 	)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, register.UserId)
//
// 	login, err := st.AuthClient.Login(
// 		ctx, &ssov1.LoginRequest{
// 			Email:    email,
// 			Password: password,
// 			AppId:    appId,
// 		},
// 	)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, login.GetAccessToken())
// 	require.NotEmpty(t, login.GetRefreshToken())
//
// 	permCreate, err := st.PermsClient.Create(
// 		ctx, &acs.CreatePermissionRequest{
// 			RequesterToken: login.GetAccessToken(),
// 			Permission: &acs.Permission{
// 				PermissionId: nil,
// 				Resource:     gofakeit.Username(),
// 				Action:       gofakeit.Username(),
// 				Description:  gofakeit.Username(),
// 			},
// 		},
// 	)
// 	require.NoError(t, err)
// 	assert.NotEmpty(t, permCreate.PermissionId)
//
// 	role := &acs.Role{
// 		RoleId:      nil,
// 		Name:        gofakeit.Username(),
// 		Description: gofakeit.Username(),
// 		Permissions: []int32{permCreate.PermissionId},
// 	}
//
// 	roleCreate, err := st.RolesClient.Create(
// 		ctx, &acs.CreateRoleRequest{
// 			RequesterToken: login.GetAccessToken(),
// 			Role:           role,
// 		},
// 	)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, roleCreate.RoleId)
//
// 	_, err = st.RolesClient.Add(
// 		ctx, &acs.AddRoleRequest{
// 			RequesterToken: login.GetAccessToken(),
// 			UserId:         register.UserId,
// 			RoleId:         roleCreate.RoleId,
// 		},
// 	)
// 	require.NoError(t, err)
//
// 	roleGet, err := st.RolesClient.GetUserRoles(
// 		ctx, &acs.GetUserRolesRequest{
// 			RequesterToken: login.GetAccessToken(),
// 			UserId:         register.UserId,
// 		},
// 	)
// 	require.NoError(t, err)
// 	require.Contains(t, roleGet.Roles, role)
// }
//
// func TestPermissionCreate_FailCases(t *testing.T) {
// 	ctx, st := suite.New(t)
//
// 	email := gofakeit.Email()
// 	password := gofakeit.Password(true, true, true, true, false, passDefaultLen)
//
// 	_, err := st.AuthClient.Register(
// 		ctx, &ssov1.RegisterRequest{
// 			Email:    email,
// 			Password: password,
// 			AppId:    1,
// 		},
// 	)
// 	require.NoError(t, err)
//
// 	login, err := st.AuthClient.Login(
// 		ctx, &ssov1.LoginRequest{
// 			Email:    email,
// 			Password: password,
// 			AppId:    1,
// 		},
// 	)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, login.GetAccessToken())
// 	require.NotEmpty(t, login.GetRefreshToken())
//
// 	tests := []struct {
// 		name           string
// 		requesterToken string
// 		permission     *acs.Permission
// 		expectedError  string
// 	}{
// 		{
// 			name:           "Empty requester access token",
// 			requesterToken: "",
// 			permission: &acs.Permission{
// 				Resource:    "",
// 				Action:      "",
// 				Description: "",
// 			},
// 			expectedError: "unregistered",
// 		},
// 		{
// 			name:           "Wrong requester access token",
// 			requesterToken: "wrong123.secret123.token123",
// 			permission: &acs.Permission{
// 				Resource:    "",
// 				Action:      "",
// 				Description: "",
// 			},
// 			expectedError: "unregistered",
// 		},
// 		{
// 			name:           "Try to create via token without required permission",
// 			requesterToken: login.GetAccessToken(),
// 			permission: &acs.Permission{
// 				Resource:    "",
// 				Action:      "",
// 				Description: "",
// 			},
// 			expectedError: "permission denied",
// 		},
// 		{
// 			name:           "Empty resource param",
// 			requesterToken: login.GetAccessToken(),
// 			permission: &acs.Permission{
// 				Resource:    "",
// 				Action:      "action",
// 				Description: "Description",
// 			},
// 			expectedError: "empty resource parameter",
// 		},
// 		{
// 			name:           "Resource param via numbers",
// 			requesterToken: login.GetAccessToken(),
// 			permission: &acs.Permission{
// 				Resource:    gofakeit.Password(true, true, true, false, false, 10),
// 				Action:      "action",
// 				Description: "Description",
// 			},
// 			expectedError: "wrong resource parameter",
// 		},
// 		{
// 			name:           "Resource param via special symbols",
// 			requesterToken: login.GetAccessToken(),
// 			permission: &acs.Permission{
// 				Resource:    gofakeit.Password(true, true, false, true, false, 10),
// 				Action:      "action",
// 				Description: "Description",
// 			},
// 			expectedError: "wrong resource parameter",
// 		},
// 		{
// 			name:           "Resource param via spaces",
// 			requesterToken: login.GetAccessToken(),
// 			permission: &acs.Permission{
// 				Resource:    gofakeit.Password(true, true, false, false, true, 10),
// 				Action:      "action",
// 				Description: "Description",
// 			},
// 			expectedError: "wrong resource parameter",
// 		},
// 		{
// 			name:           "Too long resource parameter",
// 			requesterToken: login.GetAccessToken(),
// 			permission: &acs.Permission{
// 				Resource:    gofakeit.Password(true, true, false, false, false, 64),
// 				Action:      "action",
// 				Description: "Description",
// 			},
// 			expectedError: "wrong resource parameter",
// 		},
// 		{
// 			name:           "Empty action param",
// 			requesterToken: login.GetAccessToken(),
// 			permission: &acs.Permission{
// 				Resource:    "resource",
// 				Action:      "",
// 				Description: "Description",
// 			},
// 			expectedError: "empty action parameter",
// 		},
// 		{
// 			name:           "Action param via numbers",
// 			requesterToken: login.GetAccessToken(),
// 			permission: &acs.Permission{
// 				Resource:    "resource",
// 				Action:      gofakeit.Password(true, true, true, false, false, 10),
// 				Description: "Description",
// 			},
// 			expectedError: "wrong action parameter",
// 		},
// 		{
// 			name:           "Action param via special symbols",
// 			requesterToken: login.GetAccessToken(),
// 			permission: &acs.Permission{
// 				Resource:    "resource",
// 				Action:      gofakeit.Password(true, true, false, true, false, 10),
// 				Description: "Description",
// 			},
// 			expectedError: "wrong action parameter",
// 		},
// 		{
// 			name:           "Action param via spaces",
// 			requesterToken: login.GetAccessToken(),
// 			permission: &acs.Permission{
// 				Resource:    "resource",
// 				Action:      gofakeit.Password(true, true, false, false, true, 10),
// 				Description: "Description",
// 			},
// 			expectedError: "wrong action parameter",
// 		},
// 		{
// 			name:           "Too long action parameter",
// 			requesterToken: login.GetAccessToken(),
// 			permission: &acs.Permission{
// 				Resource:    "resource",
// 				Action:      gofakeit.Password(true, true, false, false, false, 64),
// 				Description: "Description",
// 			},
// 			expectedError: "wrong action parameter",
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				_, err := st.PermsClient.Create(
// 					ctx, &acs.CreatePermissionRequest{
// 						RequesterToken: tt.requesterToken,
// 						Permission:     tt.permission,
// 					},
// 				)
//
// 				require.Error(t, err)
// 				require.Contains(t, err.Error(), tt.expectedError)
// 			},
// 		)
// 	}
// }
