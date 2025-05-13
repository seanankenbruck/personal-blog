package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"regexp"

	"github.com/seanankenbruck/blog/internal/domain"
)

var passwordRegex = regexp.MustCompile(`[!@#\$%\^&\*\(\)_\+\-=\[\]{};':"\\|,.<>\/?]+`)

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
	// Validate password
	if len(user.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if !passwordRegex.MatchString(user.Password) {
		return errors.New("password must include at least one special character")
	}
	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword) // Store as bcrypt hash
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