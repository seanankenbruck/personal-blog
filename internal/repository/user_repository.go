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

// Create creates a new user, hashing the password before saving
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	// Hash the plaintext password
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return r.db.WithContext(ctx).Create(user).Error
}

// Authenticate finds the user by username and compares the password hash
func (r *UserRepository) Authenticate(ctx context.Context, username, password string) (*domain.User, bool) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false
		}
		return nil, false
	}

	// Compare hash with provided password
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, false
	}
	return &user, true
}