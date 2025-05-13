package service

import (
	"context"
	"testing"

	"github.com/seanankenbruck/blog/internal/domain"
	"github.com/stretchr/testify/assert"
)

type mockUserRepo struct {
	users map[string]*domain.User
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if m.users == nil {
		m.users = make(map[string]*domain.User)
	}
	m.users[user.Username] = user
	return nil
}

func (m *mockUserRepo) Authenticate(ctx context.Context, username, password string) (*domain.User, error) {
	user, ok := m.users[username]
	if !ok {
		return nil, domain.ErrInvalidCredentials
	}
	return user, nil
}

func TestCreateUser_ValidPassword(t *testing.T) {
	repo := &mockUserRepo{}
	service := NewUserService(repo)
	user := &domain.User{Username: "testuser", Password: "P@ssw0rd!", Role: domain.Reader}
	err := service.CreateUser(context.Background(), user)
	assert.NoError(t, err)
	assert.NotEqual(t, "P@ssw0rd!", user.Password) // Should be hashed
}

func TestCreateUser_ShortPassword(t *testing.T) {
	repo := &mockUserRepo{}
	service := NewUserService(repo)
	user := &domain.User{Username: "short", Password: "Ab!2", Role: domain.Reader}
	err := service.CreateUser(context.Background(), user)
	assert.Error(t, err)
	assert.Equal(t, "password must be at least 8 characters long", err.Error())
}

func TestCreateUser_NoSpecialChar(t *testing.T) {
	repo := &mockUserRepo{}
	service := NewUserService(repo)
	user := &domain.User{Username: "nospecial", Password: "Password1", Role: domain.Reader}
	err := service.CreateUser(context.Background(), user)
	assert.Error(t, err)
	assert.Equal(t, "password must include at least one special character", err.Error())
}