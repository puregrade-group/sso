package acs

import (
	"context"

	"github.com/puregrade-group/protos/gen/go/acs"
	"github.com/puregrade-group/sso/internal/domain/models"
	"github.com/puregrade-group/sso/pkg/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Permissions interface {
	Create(ctx context.Context,
		requesterToken string,
		resource string,
		action string,
		description string,
	) (err error)
	ReadPermissionOwners(ctx context.Context,
		requesterToken string,
		permissionName string,
	) (apps []models.App, err error)
	CheckPermissions(ctx context.Context,
		requesterToken string,
		appId int32,
		permissionNames []string,
	) (ok bool, err error)
	Delete(ctx context.Context,
		requesterToken string,
		permissionName string,
	) (permissionId int32, err error)
	Add(ctx context.Context,
		requesterToken string,
		appId int32,
		permissionName string,
	) (err error)
	Remove(ctx context.Context,
		requesterToken string,
		appId int32,
		permissionName string,
	) (err error)
}

func (s *permissionsApi) Create(
	ctx context.Context,
	req *acs.CreatePermissionRequest,
) (*acs.CreatePermissionResponse, error) {
	err := validatePermsCreate(req)

	if err != nil {
		return nil, err
	}

	p := req.GetPermission()

	err = s.perms.Create(ctx, req.GetRequesterToken(), p.Resource, p.Action, p.Description)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return nil, nil
}

func (s *permissionsApi) ReadOwners(
	ctx context.Context,
	req *acs.ReadPermissionOwnersRequest,
) (*acs.ReadPermissionOwnersResponse, error) {
	if err := validatePermsReadOwners(req); err != nil {
		return nil, err
	}

	owners, err := s.perms.ReadPermissionOwners(ctx, req.GetRequesterToken(), req.PermissionName)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal")
	}

	appIds := make([]int32, len(owners))
	for _, v := range owners {
		appIds = append(appIds, v.Id)
	}

	return &acs.ReadPermissionOwnersResponse{
		AppIds: appIds,
	}, nil
}

func (s *permissionsApi) CheckPermissions(
	ctx context.Context,
	req *acs.CheckPermissionsRequest,
) (*acs.CheckPermissionsResponse, error) {
	if err := validateCheckPermissions(req); err != nil {
		return nil, err
	}

	ok, err := s.perms.CheckPermissions(ctx, req.GetRequesterToken(), req.GetAppId(), req.GetPermissionNames())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &acs.CheckPermissionsResponse{
		Ok: ok,
	}, err
}

func (s *permissionsApi) Delete(
	ctx context.Context,
	req *acs.DeletePermissionRequest,
) (*acs.DeletePermissionResponse, error) {
	if err := validatePermsDelete(req); err != nil {
		return nil, err
	}

	id, err := s.perms.Delete(ctx, req.GetRequesterToken(), req.GetPermissionName())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &acs.DeletePermissionResponse{
		PermissionId: id,
	}, err
}

func (s *permissionsApi) Add(
	ctx context.Context,
	req *acs.AddPermissionRequest,
) (*acs.AddPermissionResponse, error) {
	if err := validatePermsAdd(req); err != nil {
		return nil, err
	}

	err := s.perms.Add(ctx, req.GetRequesterToken(), req.GetAppId(), req.GetPermissionName())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return nil, nil
}

func (s *permissionsApi) Remove(
	ctx context.Context,
	req *acs.RemovePermissionRequest,
) (*acs.RemovePermissionResponse, error) {
	if err := validatePermsRemove(req); err != nil {
		return nil, err
	}

	err := s.perms.Remove(ctx, req.GetRequesterToken(), req.GetAppId(), req.GetPermissionName())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return nil, nil
}

func validatePermsCreate(req *acs.CreatePermissionRequest) error {
	_, err := jwt.ParseToken(req.RequesterToken)
	if err != nil {
		return status.Error(codes.Unauthenticated, "wrong token")
	}
	perm := req.GetPermission()
	if perm.Resource == "" {
		return status.Error(codes.InvalidArgument, "resource is required")
	}
	if perm.Action == "" {
		return status.Error(codes.InvalidArgument, "action is required")
	}

	return nil
}

func validatePermsReadOwners(req *acs.ReadPermissionOwnersRequest) error {
	_, err := jwt.ParseToken(req.RequesterToken)
	if err != nil {
		return status.Error(codes.Unauthenticated, "wrong token")
	}
	if req.PermissionName == "" {
		return status.Error(codes.InvalidArgument, "permissionName is required")
	}

	return nil
}

func validateCheckPermissions(req *acs.CheckPermissionsRequest) error {
	_, err := jwt.ParseToken(req.RequesterToken)
	if err != nil {
		return status.Error(codes.Unauthenticated, "wrong token")
	}
	if req.GetAppId() == 0 {
		return status.Error(codes.InvalidArgument, "appId is required")
	}
	if len(req.PermissionNames) == 0 {
		return status.Error(codes.InvalidArgument, "at least 1 permissionName must be provided")
	}

	return nil
}

func validatePermsDelete(req *acs.DeletePermissionRequest) error {
	_, err := jwt.ParseToken(req.RequesterToken)
	if err != nil {
		return status.Error(codes.Unauthenticated, "wrong token")
	}
	if req.PermissionName == "" {
		return status.Error(codes.InvalidArgument, "permissionName is required")
	}

	return nil
}

func validatePermsAdd(req *acs.AddPermissionRequest) error {
	_, err := jwt.ParseToken(req.RequesterToken)
	if err != nil {
		return status.Error(codes.Unauthenticated, "wrong token")
	}

	if req.GetAppId() == 0 {
		return status.Error(codes.InvalidArgument, "appId is required")
	}

	if req.PermissionName == "" {
		return status.Error(codes.InvalidArgument, "permissionName is required")
	}

	return nil
}

func validatePermsRemove(req *acs.RemovePermissionRequest) error {
	_, err := jwt.ParseToken(req.RequesterToken)
	if err != nil {
		return status.Error(codes.Unauthenticated, "wrong token")
	}

	if req.GetAppId() == 0 {
		return status.Error(codes.InvalidArgument, "appId is required")
	}

	if req.PermissionName == "" {
		return status.Error(codes.InvalidArgument, "permissionName is required")
	}

	return nil
}
