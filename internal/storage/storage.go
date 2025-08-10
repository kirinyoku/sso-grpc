// Package storage provides storage functionality for the SSO service.
package storage

import "errors"

var (
	// ErrUserExists is returned when a user with the given email already exists
	ErrUserExists = errors.New("user already exists")
	// ErrUserNotFound is returned when a user with the given email does not exist
	ErrUserNotFound = errors.New("user not found")
	// ErrAppNotFound is returned when an application with the given ID does not exist
	ErrAppNotFound = errors.New("app not found")
)
