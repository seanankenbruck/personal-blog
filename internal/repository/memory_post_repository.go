package repository

import (
	"context"
	"sync"

	"github.com/seanankenbruck/blog/internal/domain"
)

// MemoryPostRepository implements domain.PostRepository
type MemoryPostRepository struct {
	posts  map[uint]*domain.Post
	mu     sync.RWMutex
	nextID uint
}

// NewMemoryPostRepository creates a new MemoryPostRepository
func NewMemoryPostRepository() *MemoryPostRepository {
	return &MemoryPostRepository{
		posts:  make(map[uint]*domain.Post),
		nextID: 1,
	}
}

func (r *MemoryPostRepository) Create(ctx context.Context, post *domain.Post) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Get existing slugs
	existingSlugs := make(map[string]bool)
	for _, p := range r.posts {
		existingSlugs[p.Slug] = true
	}

	// Generate unique slug
	post.Slug = post.GenerateUniqueSlug(existingSlugs)
	post.ID = r.nextID
	r.nextID++
	r.posts[post.ID] = post
	return nil
}

func (r *MemoryPostRepository) GetByID(ctx context.Context, id uint) (*domain.Post, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	post, exists := r.posts[id]
	if !exists {
		return nil, domain.ErrPostNotFound
	}
	return post, nil
}

func (r *MemoryPostRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, post := range r.posts {
		if post.Slug == slug {
			return post, nil
		}
	}
	return nil, domain.ErrPostNotFound
}

func (r *MemoryPostRepository) GetAll(ctx context.Context) ([]*domain.Post, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	posts := make([]*domain.Post, 0, len(r.posts))
	for _, post := range r.posts {
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *MemoryPostRepository) Update(ctx context.Context, post *domain.Post) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if post exists
	existing, exists := r.posts[post.ID]
	if !exists {
		return domain.ErrPostNotFound
	}

	// If slug is changing, check for uniqueness
	if post.Slug != existing.Slug {
		existingSlugs := make(map[string]bool)
		for _, p := range r.posts {
			if p.ID != post.ID { // Don't count the current post's slug
				existingSlugs[p.Slug] = true
			}
		}
		post.Slug = post.GenerateUniqueSlug(existingSlugs)
	}

	// Update the post
	r.posts[post.ID] = post
	return nil
}

func (r *MemoryPostRepository) Delete(ctx context.Context, id uint) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.posts[id]; !exists {
		return domain.ErrPostNotFound
	}
	delete(r.posts, id)
	return nil
}