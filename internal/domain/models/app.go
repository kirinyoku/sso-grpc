// Package models provides data models for the SSO service.
package models

// App represents an application registered with the SSO service.
type App struct {
	ID     int
	Name   string
	Secret string
}
