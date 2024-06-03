package auth

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/puregrade-group/sso/internal/domain/models"
	"github.com/puregrade-group/sso/pkg/protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	passMinLen = 8
	passMaxlen = 36
)

var (
	ErrWrongCredentials  = errors.New("invalid email or password")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrTokenNotFound     = errors.New("provided refresh token is not exists")
	ErrInternal          = errors.New("internal error")
	ErrUnknown           = errors.New("unknown error")
)

type serverApi struct {
	auth.UnimplementedAuthServer
	auth Auth
}

// Auth interface must be implemented by the service layer
type Auth interface {
	Login(ctx context.Context,
		creds models.Credentials,
	) (accessToken, refreshToken string, err error)
	RegisterNewUser(ctx context.Context,
		creds models.Credentials,
		profile models.BriefProfile,
	) (userId uint64, err error)
	RefreshTokens(ctx context.Context,
		token string,
	) (accessToken, refreshToken string, err error)
}

func Register(gRPC *grpc.Server, authService Auth) {
	auth.RegisterAuthServer(gRPC, &serverApi{auth: authService})
}

func (s *serverApi) Login(
	ctx context.Context,
	req *auth.LoginRequest,
) (*auth.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	creds := models.Credentials{
		Email:    req.GetCreds().GetEmail(),
		Password: req.GetCreds().GetPassword(),
	}

	access, refresh, err := s.auth.Login(ctx, creds)
	switch err {
	case nil: // Do nothing
	case ErrWrongCredentials:
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case ErrInternal:
		return nil, status.Error(codes.Internal, err.Error())
	default:
		return nil, status.Error(codes.Unknown, ErrUnknown.Error())
	}

	return &auth.LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (s *serverApi) Register(
	ctx context.Context,
	req *auth.RegisterRequest,
) (*auth.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	creds := models.Credentials{
		Email:    req.GetCreds().GetEmail(),
		Password: req.GetCreds().GetPassword(),
	}

	profile := models.BriefProfile{
		FirstName:   req.GetProfile().GetFirstName(),
		LastName:    req.GetProfile().GetLastName(),
		DateOfBirth: req.GetProfile().GetDateOfBirth().AsTime(),
	}

	userId, err := s.auth.RegisterNewUser(ctx, creds, profile)
	switch err {
	case nil: // Do nothing
	case ErrUserAlreadyExists:
		return nil, status.Error(codes.AlreadyExists, err.Error())
	default:
		return nil, status.Error(codes.Unknown, ErrUnknown.Error())
	}

	return &auth.RegisterResponse{
		UserId: userId,
	}, nil
}

func (s *serverApi) Refresh(
	ctx context.Context,
	req *auth.RefreshRequest,
) (*auth.RefreshResponse, error) {
	if err := validateRefresh(req); err != nil {
		return nil, err
	}

	access, refresh, err := s.auth.RefreshTokens(ctx, req.GetRefreshToken())
	switch err {
	case nil: // Do nothing
	case ErrTokenNotFound:
		return nil, status.Error(codes.NotFound, err.Error())
	default:
		return nil, status.Error(codes.Unknown, ErrUnknown.Error())
	}

	return &auth.RefreshResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func validateLogin(req *auth.LoginRequest) error {
	if req.GetCreds().GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if !isEmail(req.GetCreds().GetEmail()) {
		return status.Error(codes.InvalidArgument, "email has the wrong structure")
	}

	if req.GetCreds().GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if len(req.GetCreds().GetPassword()) < passMinLen || len(req.GetCreds().GetPassword()) > passMaxlen {
		return status.Error(codes.InvalidArgument, "password length is not within the allowed range")
	}

	return nil
}

func validateRegister(req *auth.RegisterRequest) error {
	if req.GetCreds().GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if !isEmail(req.GetCreds().GetEmail()) {
		return status.Error(codes.InvalidArgument, "email has the wrong structure")
	}

	if req.GetCreds().GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if len(req.GetCreds().GetPassword()) < passMinLen || len(req.GetCreds().GetPassword()) > passMaxlen {
		return status.Error(codes.InvalidArgument, "password length is not within the allowed range")
	}

	if req.GetProfile().GetFirstName() == "" {
		return status.Error(codes.InvalidArgument, "first_name is required")
	}

	if req.GetProfile().GetDateOfBirth().AsTime().Add(time.Hour*24*365*6).Compare(time.Now()) == 1 { // if the DOB was less than 6 years ago
		return status.Error(codes.InvalidArgument, "too young")
	}

	if req.GetProfile().GetDateOfBirth().AsTime().Add(time.Hour*24*365*100).Compare(time.Now()) == -1 { // // if the DOB was less than 100 years ago
		return status.Error(codes.InvalidArgument, "too old")
	}

	return nil
}

func validateRefresh(req *auth.RefreshRequest) error {
	if req.GetRefreshToken() == "" {
		return status.Error(codes.InvalidArgument, "refresh token is required")
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
