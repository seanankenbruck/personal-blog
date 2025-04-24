package store

import (
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

func (s *UserStore) Authenticate(username, password string) (*domain.User, bool) {
	user, exists := s.users[username]
	if !exists {
		return nil, false
	}

	// In a real app, you would hash the password and compare hashes
	if user.Password != password {
		return nil, false
	}

	return user, true
}