package repository

import (
	"context"
	"sync"

	"github.com/seanankenbruck/blog/internal/domain"
)

type InMemoryPostRepository struct {
	posts map[string]*domain.Post
	mu    sync.RWMutex
	nextID uint
}

func NewInMemoryPostRepository() *InMemoryPostRepository {
	return &InMemoryPostRepository{
		posts: make(map[string]*domain.Post),
		nextID: 1,
	}
}

func (r *InMemoryPostRepository) Create(ctx context.Context, post *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post.ID = r.nextID
	r.nextID++
	r.posts[post.Slug] = post
	return nil
}

func (r *InMemoryPostRepository) GetByID(ctx context.Context, id uint) (*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, post := range r.posts {
		if post.ID == id {
			return post, nil
		}
	}
	return nil, domain.ErrPostNotFound
}

func (r *InMemoryPostRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, exists := r.posts[slug]
	if !exists {
		return nil, domain.ErrPostNotFound
	}
	return post, nil
}

func (r *InMemoryPostRepository) GetAll(ctx context.Context) ([]*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	posts := make([]*domain.Post, 0, len(r.posts))
	for _, post := range r.posts {
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *InMemoryPostRepository) Update(ctx context.Context, post *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.posts[post.Slug]; !exists {
		return domain.ErrPostNotFound
	}

	r.posts[post.Slug] = post
	return nil
}

func (r *InMemoryPostRepository) Delete(ctx context.Context, id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for slug, post := range r.posts {
		if post.ID == id {
			delete(r.posts, slug)
			return nil
		}
	}
	return domain.ErrPostNotFound
}