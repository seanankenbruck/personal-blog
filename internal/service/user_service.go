package service

import (
	"context"
	"golang.org/x/crypto/bcrypt"

	"github.com/seanankenbruck/blog/internal/domain"
)

// userService implements the UserService interface
type userService struct {
	repo domain.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo domain.UserRepository) domain.UserService {
	return &userService{repo: repo}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, user *domain.User) error {
	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	return s.repo.Create(ctx, user)
}

// AuthenticateUser authenticates a user
func (s *userService) AuthenticateUser(ctx context.Context, username, password string) (*domain.User, error) {
	user, err := s.repo.Authenticate(ctx, username, password)
	if err != nil {
		return nil, err
	}

	// Verify the password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return user, nil
}