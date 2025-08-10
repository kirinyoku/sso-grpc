package models

// User represents a user registered with the SSO service.
type User struct {
	ID       int64
	Email    string
	PassHash []byte
}
