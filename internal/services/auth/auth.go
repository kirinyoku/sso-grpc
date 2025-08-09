// Package auth provides authentication and authorization functionality for the SSO service.
// It handles user registration, login, and admin role verification.
package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kirinyoku/sso-grpc/internal/domain/models"
	"github.com/kirinyoku/sso-grpc/internal/lib/jwt"
	"github.com/kirinyoku/sso-grpc/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

// Auth provides authentication and authorization services.
type Auth struct {
	log      *slog.Logger  // logger for structured logging
	storage  Storage       // storage dependency for data persistence
	tokenTTL time.Duration // duration for which JWT tokens are valid
}

// Storage defines the interface that must be implemented by any storage provider
// used by the Auth service.
type Storage interface {
	// SaveUser persists a new user with the given email and password hash.
	// Returns the ID of the created user or an error if the operation fails.
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)

	// User retrieves a user by email.
	// Returns the user if found, or an error if the user doesn't exist or the operation fails.
	User(ctx context.Context, email string) (*models.User, error)

	// IsAdmin checks if a user has administrative privileges.
	// Returns true if the user is an admin, false otherwise.
	IsAdmin(ctx context.Context, userID int64) (bool, error)

	// App retrieves application information by ID.
	// Returns the app if found, or an error if the app doesn't exist or the operation fails.
	App(ctx context.Context, appID int) (*models.App, error)
}

// Common authentication errors
var (
	// ErrInvalidCredentials is returned when authentication fails due to invalid credentials
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidAppID is returned when the provided application ID is invalid or not found
	ErrInvalidAppID = errors.New("invalid app ID")

	// ErrUserExists is returned when attempting to register a user that already exists
	ErrUserExists = errors.New("user already exists")
)

// New creates a new instance of the Auth service with the provided dependencies.
//
// Parameters:
//   - log: logger instance for structured logging
//   - storage: storage implementation for data persistence
//   - tokenTTL: duration for which JWT tokens should be valid
//
// Returns a new *Auth instance ready to use.
func New(log *slog.Logger, storage Storage, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:      log,
		storage:  storage,
		tokenTTL: tokenTTL,
	}
}

// Register creates a new user account with the provided email and password.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - email: user's email address (must be unique)
//   - password: user's password (will be hashed before storage)
//
// Returns:
//   - int64: ID of the newly created user
//   - error: nil on success, or an error if registration fails
//
// Possible errors:
//   - ErrUserExists: if a user with the given email already exists
//   - other errors: for any other failure during user creation
func (a *Auth) Register(ctx context.Context, email string, password string) (int64, error) {
	const op = "auth.Auth.Register"

	log := a.log.With(
		slog.String("op", op),
	)

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	userID, err := a.storage.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", slog.String("error", err.Error()))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to save user", slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered successfully", slog.Int64("user_id", userID))

	return userID, nil
}

// Login authenticates a user and generates a JWT token for the specified application.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - email: user's email address
//   - password: user's password
//   - appID: ID of the application the user is logging into
//
// Returns:
//   - string: JWT token for authenticated sessions
//   - error: nil on success, or an error if authentication fails
//
// Possible errors:
//   - ErrInvalidCredentials: if email/password is incorrect or user doesn't exist
//   - ErrInvalidAppID: if the specified appID is invalid
//   - other errors: for any other failure during authentication
func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	const op = "auth.Auth.Login"

	log := a.log.With(
		slog.String("op", op),
	)

	user, err := a.storage.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", slog.String("error", err.Error()))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		log.Error("failed to get user", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Error("invalid credentials", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.storage.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", slog.String("error", err.Error()))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		log.Error("failed to get app", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully", slog.Int64("user_id", user.ID))

	return token, nil
}

// IsAdmin checks if the specified user has administrative privileges.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userID: ID of the user to check
//
// Returns:
//   - bool: true if the user is an admin, false otherwise
//   - error: nil on success, or an error if the check fails
//
// Possible errors:
//   - ErrInvalidAppID: if there's an issue with application configuration
//   - other errors: for any other failure during the admin check
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.Auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	isAdmin, err := a.storage.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", slog.String("error", err.Error()))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}

		log.Error("failed to check if user is admin", slog.String("error", err.Error()))

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
