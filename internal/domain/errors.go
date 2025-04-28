package domain

import "errors"

var (
	// ErrPostNotFound is returned when a post cannot be found
	ErrPostNotFound = errors.New("post not found")
	// ErrUserExists is returned when trying to create a user that already exists
	ErrUserExists = errors.New("user already exists")
	// ErrInvalidCredentials is returned when authentication fails
	ErrInvalidCredentials = errors.New("invalid credentials")
)