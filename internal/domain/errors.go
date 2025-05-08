package domain

import "errors"

var (
	// ErrPostNotFound is returned when a post cannot be found
	ErrPostNotFound = errors.New("post not found")
	// ErrUserExists is returned when trying to create a user that already exists
	ErrUserExists = errors.New("user already exists")
	// ErrInvalidCredentials is returned when authentication fails
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrSubscriberNotFound is returned when a subscriber cannot be found
	ErrSubscriberNotFound = errors.New("subscriber not found")
	// ErrSubscriberExists is returned when trying to create a subscriber with an email that already exists
	ErrSubscriberExists = errors.New("subscriber with this email already exists")
)