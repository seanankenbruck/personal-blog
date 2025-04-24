package domain

// Role represents a user role in the system
type Role string

const (
	// Editor has full CRUD permissions on posts
	Editor Role = "editor"
	// Reader can only view posts
	Reader Role = "reader"
)

// User represents an application user with a role
type User struct {
	Username string
	Password string // In a real app, this would be hashed
	Role     Role
}