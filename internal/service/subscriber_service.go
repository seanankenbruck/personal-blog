package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"regexp"
	"strings"

	"github.com/seanankenbruck/blog/internal/domain"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

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
	if !emailRegex.MatchString(email) {
		return nil, errors.New("invalid email format")
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

	// Send confirmation email
	if err := sendConfirmationEmail(subscriber.Email, subscriber.ConfirmationToken); err != nil {
		// Log error but do not fail subscription
		fmt.Fprintf(os.Stderr, "Failed to send confirmation email: %v\n", err)
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

// sendConfirmationEmail sends a confirmation email with the token link
var sendConfirmationEmail = func(to, token string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	sender := os.Getenv("EMAIL_SENDER")
	appHost := os.Getenv("APP_HOST")
	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || sender == "" || appHost == "" {
		return nil // Email sending not configured
	}

	confirmURL := fmt.Sprintf("%s/confirm?token=%s", appHost, token)
	subject := "Confirm your subscription"
	body := fmt.Sprintf("Thank you for subscribing! Please confirm your subscription by clicking the link below:\n\n%s\n\nIf you did not request this, please ignore this email.", confirmURL)
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, sender, []string{to}, []byte(msg))
}