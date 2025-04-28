package repository

import (
	"context"

	"github.com/seanankenbruck/blog/internal/domain"
	"github.com/seanankenbruck/blog/internal/store"
)

// InMemoryUserRepository wraps store.UserStore to implement authentication
// via a common interface
// It satisfies Authenticate(ctx, username, password)
// so it can be used interchangeably with UserRepository

type InMemoryUserRepository struct {
	store *store.UserStore
}

// NewInMemoryUserRepository returns a new InMemoryUserRepository
func NewInMemoryUserRepository(store *store.UserStore) *InMemoryUserRepository {
	return &InMemoryUserRepository{
		store: store,
	}
}

// Authenticate delegates to the underlying store.UserStore
func (r *InMemoryUserRepository) Authenticate(ctx context.Context, username, password string) (*domain.User, error) {
	return r.store.Authenticate(username, password)
}

func (r *InMemoryUserRepository) Create(ctx context.Context, username, password string, role domain.Role) error {
	return r.store.Create(username, password, role)
}