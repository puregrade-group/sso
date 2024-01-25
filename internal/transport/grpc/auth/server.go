package auth

import (
	"context"
	"errors"
	"regexp"

	"github.com/puregrade-group/protos/gen/go/sso"
	"github.com/puregrade-group/sso/internal/service/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	passMinLen = 8
	passMaxlen = 36
)

type serverApi struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

// Auth interface must be implemented by the service layer
type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
		appId int32,
	) (token string, err error)
	RegisterNewUser(ctx context.Context,
		email string,
		password string,
		appId int32,
	) (userId string, err error)
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverApi{auth: auth})
}

func (s *serverApi) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		if errors.Is(err, auth.ErrUnknownApp) {
			return nil, status.Error(codes.PermissionDenied, "app is unknown")
		}
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverApi) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userId, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		if errors.Is(err, auth.ErrUnknownApp) {
			return nil, status.Error(codes.PermissionDenied, "app is unknown")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{
		UserId: userId,
	}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if !isEmail(req.GetEmail()) {
		return status.Error(codes.InvalidArgument, "email has the wrong structure")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if len(req.GetPassword()) < passMinLen {
		return status.Error(codes.InvalidArgument, "invalid email or password")
	}

	if len(req.GetPassword()) > passMaxlen {
		return status.Error(codes.InvalidArgument, "invalid email or password")
	}

	if req.GetAppId() == 0 {
		return status.Error(codes.InvalidArgument, "appId is required")
	}

	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if !isEmail(req.GetEmail()) {
		return status.Error(codes.InvalidArgument, "email has the wrong structure")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if len(req.GetPassword()) < passMinLen {
		return status.Error(codes.InvalidArgument, "password is shorter than minimum password length")
	}

	if len(req.GetPassword()) > passMaxlen {
		return status.Error(codes.InvalidArgument, "password if longer than maximum password length")
	}

	if req.GetAppId() == 0 {
		return status.Error(codes.InvalidArgument, "appId is required")
	}

	return nil
}

func isEmail(email string) bool {
	pattern, err := regexp.Compile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if err != nil {
		return false
	}
	return pattern.MatchString(email)
}
