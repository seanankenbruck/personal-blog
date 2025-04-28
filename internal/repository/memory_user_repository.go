package repository

import (
	"context"
	"sync"

	"github.com/seanankenbruck/blog/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// MemoryUserRepository implements domain.UserRepository
type MemoryUserRepository struct {
	users map[string]*domain.User
	mu    sync.RWMutex
}

// NewMemoryUserRepository creates a new MemoryUserRepository with test users
func NewMemoryUserRepository() *MemoryUserRepository {
	repo := &MemoryUserRepository{
		users: make(map[string]*domain.User),
	}

	// Create test users
	testUsers := []struct {
		username string
		password string
		role     domain.Role
	}{
		{"editor", "editor123", domain.Editor},
		{"reader", "reader123", domain.Reader},
	}

	for _, user := range testUsers {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.password), bcrypt.DefaultCost)
		repo.users[user.username] = &domain.User{
			Username: user.username,
			Password: string(hashedPassword),
			Role:     user.role,
		}
	}

	return repo
}

func (r *MemoryUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Username]; exists {
		return domain.ErrUserExists
	}

	r.users[user.Username] = user
	return nil
}

func (r *MemoryUserRepository) Authenticate(ctx context.Context, username, password string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[username]
	if !exists {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return user, nil
}