// Package acs is the transport layer package who provides convenient work with roles & permissions
package acs

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/puregrade-group/protos/gen/go/acs"
	"github.com/puregrade-group/sso/internal/domain/models"
	"github.com/puregrade-group/sso/internal/utils/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Roles interface {
	CreateRole(ctx context.Context,
		requesterToken string,
		role models.Role,
	) (err error)
	GetUserRoles(ctx context.Context,
		requesterToken string,
		userId [16]byte,
	) (roles []models.Role, err error)
	DeleteRole(ctx context.Context,
		requesterToken string,
		roleName string,
	) (err error)
	AddRole(ctx context.Context,
		requesterToken string,
		userId [16]byte,
		roleId int32,
	) (err error)
	RemoveRole(ctx context.Context,
		requesterToken string,
		userId [16]byte,
		roleId int32,
	) (err error)
}

func (s *rolesApi) Create(
	ctx context.Context,
	req *acs.CreateRoleRequest,
) (*acs.CreateRoleResponse, error) {
	err := validateRolesCreate(req)
	if err != nil {
		return nil, err
	}

	// get and parse role
	r := req.GetRole()
	var role models.Role
	role.Name = r.GetName()
	role.Description = r.GetDescription()
	for _, p := range r.GetPermissions() {
		temp := strings.Split(p, ":")
		role.Permissions = append(
			role.Permissions, models.Permission{
				Resource: temp[0],
				Action:   temp[1],
			},
		)
	}

	err = s.roles.CreateRole(ctx, req.GetRequesterToken(), role)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return nil, nil
}

func (s *rolesApi) GetUserRoles(
	ctx context.Context,
	req *acs.GetUserRolesRequest,
) (*acs.GetUserRolesResponse, error) {
	if err := validateGetUserRoles(req); err != nil {
		return nil, err
	}

	parsedId, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "wrong userId")
	}

	roles, err := s.roles.GetUserRoles(ctx, req.GetRequesterToken(), parsedId)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal")
	}

	// struct arrays conversion
	var grpcRoles []*acs.Role
	for _, r := range roles {
		var permsNames []string
		for _, p := range r.Permissions {
			permsNames = append(permsNames, p.Resource+":"+p.Action)
		}

		grpcRoles = append(
			grpcRoles, &acs.Role{
				RoleId:      &r.Id,
				Name:        r.Name,
				Description: r.Description,
				Permissions: permsNames,
			},
		)
	}

	return &acs.GetUserRolesResponse{
		Roles: grpcRoles,
	}, nil
}

func (s *rolesApi) Delete(
	ctx context.Context,
	req *acs.DeleteRoleRequest,
) (*acs.DeleteRoleResponse, error) {
	if err := validateRolesDelete(req); err != nil {
		return nil, err
	}

	err := s.roles.DeleteRole(ctx, req.GetRequesterToken(), req.GetRoleName())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return nil, err
}

func (s *rolesApi) Add(
	ctx context.Context,
	req *acs.AddRoleRequest,
) (*acs.AddRoleResponse, error) {
	if err := validateRolesAdd(req); err != nil {
		return nil, err
	}

	// parse userId
	uuidObj, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "userId is not uuid")
	}

	err = s.roles.AddRole(ctx, req.GetRequesterToken(), uuidObj, req.GetRoleId())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return nil, nil
}

func (s *rolesApi) Remove(
	ctx context.Context,
	req *acs.RemoveRoleRequest,
) (*acs.RemoveRoleResponse, error) {
	if err := validateRolesRemove(req); err != nil {
		return nil, err
	}

	// parse userId
	uuidObj, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "userId is not uuid")
	}

	err = s.roles.RemoveRole(ctx, req.GetRequesterToken(), uuidObj, req.GetRoleId())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return nil, nil
}

func validateRolesCreate(req *acs.CreateRoleRequest) error {
	_, err := jwt.ParseToken(req.RequesterToken)
	if err != nil {
		return status.Error(codes.Unauthenticated, "wrong token")
	}
	role := req.GetRole()
	if role.Name == "" {
		return status.Error(codes.InvalidArgument, "role.name is required")
	}
	if len(role.Permissions) == 0 {
		return status.Error(codes.InvalidArgument, "at least 1 role.permission must be provided")
	}

	return nil
}

func validateGetUserRoles(req *acs.GetUserRolesRequest) error {
	_, err := jwt.ParseToken(req.RequesterToken)
	if err != nil {
		return status.Error(codes.Unauthenticated, "wrong token")
	}
	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "UserId is required")
	}

	return nil
}

func validateRolesDelete(req *acs.DeleteRoleRequest) error {
	_, err := jwt.ParseToken(req.RequesterToken)
	if err != nil {
		return status.Error(codes.Unauthenticated, "wrong token")
	}
	if req.RoleName == "" {
		return status.Error(codes.InvalidArgument, "roleName is required")
	}

	return nil
}

func validateRolesAdd(req *acs.AddRoleRequest) error {
	_, err := jwt.ParseToken(req.RequesterToken)
	if err != nil {
		return status.Error(codes.Unauthenticated, "wrong token")
	}

	if req.GetUserId() == "" {
		return status.Error(codes.InvalidArgument, "userId is required")
	}

	if req.GetRoleId() == 0 {
		return status.Error(codes.InvalidArgument, "roleName is required")
	}

	return nil
}

func validateRolesRemove(req *acs.RemoveRoleRequest) error {
	_, err := jwt.ParseToken(req.RequesterToken)
	if err != nil {
		return status.Error(codes.Unauthenticated, "wrong token")
	}

	if req.GetUserId() == "" {
		return status.Error(codes.InvalidArgument, "userId is required")
	}

	if req.GetRoleId() == 0 {
		return status.Error(codes.InvalidArgument, "roleName is required")
	}

	return nil
}
