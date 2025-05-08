package repository

import (
	"context"
	"sync"

	"github.com/seanankenbruck/blog/internal/domain"
)

// MemorySubscriberRepository implements domain.SubscriberRepository
type MemorySubscriberRepository struct {
	subscribers map[uint]*domain.Subscriber
	mu         sync.RWMutex
	nextID     uint
}

// NewMemorySubscriberRepository creates a new MemorySubscriberRepository
func NewMemorySubscriberRepository() *MemorySubscriberRepository {
	return &MemorySubscriberRepository{
		subscribers: make(map[uint]*domain.Subscriber),
		nextID:     1,
	}
}

// Create adds a new subscriber to memory
func (r *MemorySubscriberRepository) Create(ctx context.Context, subscriber *domain.Subscriber) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for existing email
	for _, s := range r.subscribers {
		if s.Email == subscriber.Email {
			return domain.ErrSubscriberExists
		}
	}

	subscriber.ID = r.nextID
	r.nextID++
	r.subscribers[subscriber.ID] = subscriber
	return nil
}

// GetByID retrieves a subscriber by their ID from memory
func (r *MemorySubscriberRepository) GetByID(ctx context.Context, id uint) (*domain.Subscriber, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	subscriber, exists := r.subscribers[id]
	if !exists {
		return nil, domain.ErrSubscriberNotFound
	}
	return subscriber, nil
}

// GetByEmail retrieves a subscriber by their email from memory
func (r *MemorySubscriberRepository) GetByEmail(ctx context.Context, email string) (*domain.Subscriber, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, subscriber := range r.subscribers {
		if subscriber.Email == email {
			return subscriber, nil
		}
	}
	return nil, domain.ErrSubscriberNotFound
}

// GetAll retrieves all subscribers from memory
func (r *MemorySubscriberRepository) GetAll(ctx context.Context) ([]*domain.Subscriber, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	subscribers := make([]*domain.Subscriber, 0, len(r.subscribers))
	for _, subscriber := range r.subscribers {
		subscribers = append(subscribers, subscriber)
	}
	return subscribers, nil
}

// Update modifies an existing subscriber in memory
func (r *MemorySubscriberRepository) Update(ctx context.Context, subscriber *domain.Subscriber) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.subscribers[subscriber.ID]; !exists {
		return domain.ErrSubscriberNotFound
	}

	r.subscribers[subscriber.ID] = subscriber
	return nil
}

// Delete removes a subscriber from memory
func (r *MemorySubscriberRepository) Delete(ctx context.Context, id uint) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.subscribers[id]; !exists {
		return domain.ErrSubscriberNotFound
	}

	delete(r.subscribers, id)
	return nil
}