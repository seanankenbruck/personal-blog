package repository

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"github.com/seanankenbruck/blog/internal/db"
	"github.com/seanankenbruck/blog/internal/domain"
)

// UserRepository uses GORM to persist Users and handle authentication
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository returns a new UserRepository
func NewUserRepository() *UserRepository {
	return &UserRepository{db: db.DB}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Authenticate finds the user by username and compares the password hash
func (r *UserRepository) Authenticate(ctx context.Context, username, password string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Compare hash with provided password
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}
	return &user, nil
}