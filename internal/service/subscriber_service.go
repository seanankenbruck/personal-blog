package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/seanankenbruck/blog/internal/domain"
)

// SubscriberServiceImpl implements domain.SubscriberService
//
type SubscriberServiceImpl struct {
	repo domain.SubscriberRepository
}

// NewSubscriberService creates a new SubscriberServiceImpl
func NewSubscriberService(repo domain.SubscriberRepository) *SubscriberServiceImpl {
	return &SubscriberServiceImpl{repo: repo}
}

// Subscribe adds a new subscriber and generates a confirmation token
func (s *SubscriberServiceImpl) Subscribe(ctx context.Context, email string) (*domain.Subscriber, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, errors.New("email is required")
	}
	// Check if already exists
	_, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		return nil, domain.ErrSubscriberExists
	} else if err != domain.ErrSubscriberNotFound {
		return nil, err
	}

	token, err := generateToken(32)
	if err != nil {
		return nil, err
	}

	subscriber := &domain.Subscriber{
		Email:             email,
		Confirmed:         false,
		ConfirmationToken: token,
	}
	if err := s.repo.Create(ctx, subscriber); err != nil {
		return nil, err
	}
	return subscriber, nil
}

// ConfirmSubscription confirms a subscriber using the confirmation token
func (s *SubscriberServiceImpl) ConfirmSubscription(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("confirmation token is required")
	}
	// Find by token
	subscribers, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}
	for _, sub := range subscribers {
		if sub.ConfirmationToken == token {
			sub.Confirmed = true
			sub.ConfirmationToken = ""
			return s.repo.Update(ctx, sub)
		}
	}
	return domain.ErrSubscriberNotFound
}

// Unsubscribe removes a subscriber by email
func (s *SubscriberServiceImpl) Unsubscribe(ctx context.Context, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	sub, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, sub.ID)
}

// ListSubscribers returns all subscribers
func (s *SubscriberServiceImpl) ListSubscribers(ctx context.Context) ([]*domain.Subscriber, error) {
	return s.repo.GetAll(ctx)
}

// generateToken creates a random hex string of n bytes
func generateToken(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}