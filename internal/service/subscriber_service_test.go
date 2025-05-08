package service

import (
	"context"
	"testing"

	"github.com/seanankenbruck/blog/internal/domain"
	"github.com/seanankenbruck/blog/internal/repository"
	"github.com/stretchr/testify/assert"
)

func setupTestService() (*SubscriberServiceImpl, *repository.MemorySubscriberRepository) {
	repo := repository.NewMemorySubscriberRepository()
	service := NewSubscriberService(repo)
	return service, repo
}

func TestSubscribe_Success(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()
	subscriber, err := service.Subscribe(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, subscriber)
	assert.Equal(t, "test@example.com", subscriber.Email)
	assert.False(t, subscriber.Confirmed)
	assert.NotEmpty(t, subscriber.ConfirmationToken)
}

func TestSubscribe_AlreadyExists(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()
	_, _ = service.Subscribe(ctx, "test@example.com")
	_, err := service.Subscribe(ctx, "test@example.com")
	assert.ErrorIs(t, err, domain.ErrSubscriberExists)
}

func TestSubscribe_EmptyEmail(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()
	_, err := service.Subscribe(ctx, "")
	assert.Error(t, err)
}

func TestConfirmSubscription_Success(t *testing.T) {
	service, repo := setupTestService()
	ctx := context.Background()
	subscriber, _ := service.Subscribe(ctx, "test@example.com")
	token := subscriber.ConfirmationToken
	err := service.ConfirmSubscription(ctx, token)
	assert.NoError(t, err)
	updated, _ := repo.GetByID(ctx, subscriber.ID)
	assert.True(t, updated.Confirmed)
	assert.Empty(t, updated.ConfirmationToken)
}

func TestConfirmSubscription_InvalidToken(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()
	err := service.ConfirmSubscription(ctx, "invalidtoken")
	assert.ErrorIs(t, err, domain.ErrSubscriberNotFound)
}

func TestUnsubscribe_Success(t *testing.T) {
	service, repo := setupTestService()
	ctx := context.Background()
	subscriber, _ := service.Subscribe(ctx, "test@example.com")
	err := service.Unsubscribe(ctx, "test@example.com")
	assert.NoError(t, err)
	_, err = repo.GetByID(ctx, subscriber.ID)
	assert.ErrorIs(t, err, domain.ErrSubscriberNotFound)
}

func TestUnsubscribe_NotFound(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()
	err := service.Unsubscribe(ctx, "notfound@example.com")
	assert.ErrorIs(t, err, domain.ErrSubscriberNotFound)
}

func TestListSubscribers(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()
	_, _ = service.Subscribe(ctx, "a@example.com")
	_, _ = service.Subscribe(ctx, "b@example.com")
	subs, err := service.ListSubscribers(ctx)
	assert.NoError(t, err)
	assert.Len(t, subs, 2)
}