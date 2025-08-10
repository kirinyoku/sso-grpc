// Package auth implements the gRPC server for the authentication service.
// It provides user registration, login, and admin verification functionality.
package auth

import (
	"context"
	"errors"

	pb "github.com/kirinyoku/sso-grpc/internal/proto/auth"
	"github.com/kirinyoku/sso-grpc/internal/services/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Auth defines the interface that must be implemented by the authentication service.
type Auth interface {
	// Register creates a new user account with the provided credentials.
	Register(ctx context.Context, email, password string) (userID int64, err error)
	// Login authenticates a user and returns an authentication token.
	Login(ctx context.Context, email, password string, appID int32) (token string, err error)
	// IsAdmin checks if the specified user has administrative privileges.
	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
}

// server implements the gRPC Auth service.
type server struct {
	pb.UnimplementedAuthServer      // Embed the unimplemented server for forward compatibility
	auth                       Auth // Authentication service implementation
}

// Register registers the authentication service implementation with the gRPC server.
//
// Parameters:
//   - s: The gRPC server instance
//   - auth: Implementation of the Auth interface
func Register(s *grpc.Server, auth Auth) {
	pb.RegisterAuthServer(s, &server{auth: auth})
}

const (
	// emptyValue represents the zero value for numeric fields in protobuf messages
	emptyValue = 0
)

// Register handles user registration requests.
//
// It validates the request and delegates to the underlying Auth service.
// Returns a user ID on success or an appropriate gRPC error on failure.
//
// Possible errors:
//   - codes.InvalidArgument: if request validation fails
//   - codes.Internal: if the registration process fails
func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := validateRegisterRequest(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.Register(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.RegisterResponse{
		UserId: userID,
	}, nil
}

// Login handles user authentication requests.
//
// It validates the request, authenticates the user, and returns an authentication token.
// Returns a JWT token on success or an appropriate gRPC error on failure.
//
// Possible errors:
//   - codes.InvalidArgument: if request validation fails
//   - codes.Unauthenticated: if authentication fails
//   - codes.Internal: if the login process fails
func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := validateLoginRequest(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.Email, req.Password, req.AppId)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}

		if errors.Is(err, auth.ErrInvalidAppID) {
			return nil, status.Error(codes.InvalidArgument, "invalid app ID")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.LoginResponse{
		Token: token,
	}, nil
}

// IsAdmin checks if a user has administrative privileges.
//
// It validates the request and checks the user's admin status.
// Returns the admin status or an appropriate gRPC error on failure.
//
// Possible errors:
//   - codes.InvalidArgument: if user_id is invalid or missing
//   - codes.Internal: if the admin check fails
func (s *server) IsAdmin(ctx context.Context, req *pb.IsAdminRequest) (*pb.IsAdminResponse, error) {
	if err := validateIsAdminRequest(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

// validateRegisterRequest validates the registration request parameters.
// Returns nil if the request is valid, otherwise returns a gRPC error.
func validateRegisterRequest(req *pb.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	return nil
}

// validateLoginRequest validates the login request parameters.
// Returns nil if the request is valid, otherwise returns a gRPC error.
func validateLoginRequest(req *pb.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if req.AppId == emptyValue {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

// validateIsAdminRequest validates the admin check request parameters.
// Returns nil if the request is valid, otherwise returns a gRPC error.
func validateIsAdminRequest(req *pb.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}

	if req.GetUserId() < 0 {
		return status.Error(codes.InvalidArgument, "invalid user_id")
	}

	return nil
}
