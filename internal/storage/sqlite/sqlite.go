// Package sqlite provides a SQLite implementation of the storage interface
// for the SSO service. It handles all database operations including
// user management, authentication, and application data storage.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kirinyoku/sso-grpc/internal/domain/models"
	"github.com/kirinyoku/sso-grpc/internal/storage"
	"github.com/mattn/go-sqlite3"
)

// Storage implements the Storage interface using SQLite as the backing store.
// It provides methods for user management, authentication, and application data access.
type Storage struct {
	db *sql.DB // Database connection handle
}

// New creates a new SQLite storage instance and establishes a database connection.
//
// Parameters:
//   - storagePath: filesystem path where the SQLite database file is located or should be created
//
// Returns:
//   - *Storage: a new Storage instance on success
//   - error: non-nil if database connection fails
//
// The function ensures the database connection is working by pinging it before returning.
func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// SaveUser creates a new user record in the database with the provided email and password hash.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - email: user's email address (must be unique)
//   - passHash: bcrypt hashed password
//
// Returns:
//   - int64: ID of the newly created user
//   - error: storage.ErrUserExists if a user with the email already exists,
//     or another error if the operation fails
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.sqlite.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users (email, pass_hash) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var sqliteErr *sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// User retrieves a user from the database by email.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - email: email address of the user to retrieve
//
// Returns:
//   - *models.User: user information if found
//   - error: storage.ErrUserNotFound if no user exists with the email,
//     or another error if the operation fails
func (s *Storage) User(ctx context.Context, email string) (*models.User, error) {
	const op = "storage.sqlite.User"

	stmt, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email = ?")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, email)

	var user models.User

	if err := row.Scan(&user.ID, &user.Email, &user.PassHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

// IsAdmin checks if a user has administrative privileges.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userID: ID of the user to check
//
// Returns:
//   - bool: true if the user is an administrator, false otherwise
//   - error: storage.ErrUserNotFound if no user exists with the ID,
//     or another error if the operation fails
func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.sqlite.IsAdmin"

	stmt, err := s.db.Prepare("SELECT is_admin FROM users WHERE id = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool

	if err := row.Scan(&isAdmin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

// App retrieves application information by ID.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - appID: ID of the application to retrieve
//
// Returns:
//   - *models.App: application information if found
//   - error: storage.ErrAppNotFound if no application exists with the ID,
//     or another error if the operation fails
func (s *Storage) App(ctx context.Context, appID int32) (*models.App, error) {
	const op = "storage.sqlite.App"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, appID)

	var app models.App

	if err := row.Scan(&app.ID, &app.Name, &app.Secret); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &app, nil
}
