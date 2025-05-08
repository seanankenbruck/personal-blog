package repository

import (
	"context"
	"gorm.io/gorm"
	"github.com/seanankenbruck/blog/internal/domain"
)

// PostgresSubscriberRepository implements domain.SubscriberRepository
type PostgresSubscriberRepository struct {
	db *gorm.DB
}

// NewPostgresSubscriberRepository creates a new PostgresSubscriberRepository
func NewPostgresSubscriberRepository(db *gorm.DB) *PostgresSubscriberRepository {
	return &PostgresSubscriberRepository{
		db: db,
	}
}

// Create adds a new subscriber to the database
func (r *PostgresSubscriberRepository) Create(ctx context.Context, subscriber *domain.Subscriber) error {
	return r.db.WithContext(ctx).Create(subscriber).Error
}

// GetByID retrieves a subscriber by their ID
func (r *PostgresSubscriberRepository) GetByID(ctx context.Context, id uint) (*domain.Subscriber, error) {
	var subscriber domain.Subscriber
	if err := r.db.WithContext(ctx).First(&subscriber, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrSubscriberNotFound
		}
		return nil, err
	}
	return &subscriber, nil
}

// GetByEmail retrieves a subscriber by their email address
func (r *PostgresSubscriberRepository) GetByEmail(ctx context.Context, email string) (*domain.Subscriber, error) {
	var subscriber domain.Subscriber
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&subscriber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrSubscriberNotFound
		}
		return nil, err
	}
	return &subscriber, nil
}

// GetAll retrieves all subscribers from the database
func (r *PostgresSubscriberRepository) GetAll(ctx context.Context) ([]*domain.Subscriber, error) {
	var subscribers []*domain.Subscriber
	if err := r.db.WithContext(ctx).Find(&subscribers).Error; err != nil {
		return nil, err
	}
	return subscribers, nil
}

// Update modifies an existing subscriber in the database
func (r *PostgresSubscriberRepository) Update(ctx context.Context, subscriber *domain.Subscriber) error {
	result := r.db.WithContext(ctx).Save(subscriber)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrSubscriberNotFound
	}
	return nil
}

// Delete removes a subscriber from the database
func (r *PostgresSubscriberRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&domain.Subscriber{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrSubscriberNotFound
	}
	return nil
}