package domain

import (
	"context"
	"gorm.io/gorm"
)

// Role represents a user's role in the system
// Used in JWT claims and database
type Role string

const (
	// Editor can create, edit, and delete posts
	Editor Role = "editor"
	// Reader can only view posts
	Reader Role = "reader"
)

// User represents an application user persisted to the database
// Password is stored as a bcrypt hash
// GORM tags specify primary key and column constraints
// -------------------------------------------
type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Password string `gorm:"type:varchar(255);not null"`
	Role     Role   `gorm:"type:varchar(20);not null" json:"role"`
}

// UserRepository defines the interface for user data access
// Implementations can be backed by GORM or in-memory store
// -------------------------------------------
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	Authenticate(ctx context.Context, username, password string) (*User, error)
}

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(ctx context.Context, user *User) error
	AuthenticateUser(ctx context.Context, username, password string) (*User, error)
}