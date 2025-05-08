package domain

import (
	"context"
	"errors"
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

// Subscriber represents a user who has subscribed to new content
// GORM tags specify primary key and column constraints
type Subscriber struct {
	gorm.Model
	Email             string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Confirmed         bool      `gorm:"not null" json:"confirmed"`
	ConfirmationToken string    `gorm:"type:varchar(255)" json:"confirmation_token,omitempty"`
}

// SubscriberRepository defines the interface for subscriber data access
type SubscriberRepository interface {
	Create(ctx context.Context, subscriber *Subscriber) error
	GetByID(ctx context.Context, id uint) (*Subscriber, error)
	GetByEmail(ctx context.Context, email string) (*Subscriber, error)
	GetAll(ctx context.Context) ([]*Subscriber, error)
	Update(ctx context.Context, subscriber *Subscriber) error
	Delete(ctx context.Context, id uint) error
}

// SubscriberService defines the interface for subscriber business logic
type SubscriberService interface {
	Subscribe(ctx context.Context, email string) (*Subscriber, error)
	ConfirmSubscription(ctx context.Context, token string) error
	Unsubscribe(ctx context.Context, email string) error
	ListSubscribers(ctx context.Context) ([]*Subscriber, error)
}