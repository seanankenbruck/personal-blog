package store

import (
	"errors"
	"github.com/seanankenbruck/blog/internal/domain"
)

type UserStore struct {
	users map[string]*domain.User
}

func NewUserStore() *UserStore {
	store := &UserStore{
		users: make(map[string]*domain.User),
	}

	// Add sample users
	store.users["editor@blog.com"] = &domain.User{
		Username: "editor@blog.com",
		Password: "editor123", // In a real app, this would be hashed
		Role:     domain.Editor,
	}

	store.users["reader@blog.com"] = &domain.User{
		Username: "reader@blog.com",
		Password: "reader123", // In a real app, this would be hashed
		Role:     domain.Reader,
	}

	return store
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists        = errors.New("user already exists")
)

func (s *UserStore) Authenticate(username, password string) (*domain.User, error) {
	user, exists := s.users[username]
	if !exists {
		return nil, ErrInvalidCredentials
	}

	// In a real app, you would hash the password and compare hashes
	if user.Password != password {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserStore) Create(username, password string, role domain.Role) error {
	if _, exists := s.users[username]; exists {
		return ErrUserExists
	}

	s.users[username] = &domain.User{
		Username: username,
		Password: password, // In a real app, this would be hashed
		Role:     role,
	}

	return nil
}