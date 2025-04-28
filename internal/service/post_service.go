package service

import (
	"context"
	"log"
	"time"

	"github.com/seanankenbruck/blog/internal/domain"
)

// postService implements domain.PostService
type postService struct {
	repo domain.PostRepository
}

// NewPostService creates a new PostService with the given repository
func NewPostService(repo domain.PostRepository) domain.PostService {
	return &postService{repo: repo}
}

func (s *postService) CreatePost(ctx context.Context, post *domain.Post) error {
	log.Printf("Creating post with title: %s", post.Title)
	select {
	case <-ctx.Done():
		log.Printf("Context cancelled while creating post: %s", post.Title)
		return ctx.Err()
	default:
		// Get all existing slugs
		posts, err := s.repo.GetAll(ctx)
		if err != nil {
			return err
		}

		// Create a map of existing slugs
		existingSlugs := make(map[string]bool)
		for _, p := range posts {
			existingSlugs[p.Slug] = true
		}

		// Generate a unique slug
		post.Slug = post.GenerateUniqueSlug(existingSlugs)
		return s.repo.Create(ctx, post)
	}
}

func (s *postService) GetPost(ctx context.Context, id uint) (*domain.Post, error) {
	log.Printf("Getting post with ID: %d", id)
	select {
	case <-ctx.Done():
		log.Printf("Context cancelled while getting post: %d", id)
		return nil, ctx.Err()
	default:
		return s.repo.GetByID(ctx, id)
	}
}

func (s *postService) GetPostBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	log.Printf("Getting post with slug: %s", slug)
	select {
	case <-ctx.Done():
		log.Printf("Context cancelled while getting post by slug: %s", slug)
		return nil, ctx.Err()
	default:
		p, err := s.repo.GetBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, domain.ErrPostNotFound
		}
		return p, nil
	}
}

func (s *postService) GetAllPosts(ctx context.Context) ([]*domain.Post, error) {
	log.Println("Getting all posts")
	select {
	case <-ctx.Done():
		log.Println("Context cancelled while getting all posts")
		return nil, ctx.Err()
	default:
		return s.repo.GetAll(ctx)
	}
}

func (s *postService) UpdatePost(ctx context.Context, post *domain.Post) error {
	log.Printf("Updating post with slug: %s", post.Slug)

	// Check if context is done first
	if ctx.Err() != nil {
		log.Printf("Context cancelled before updating post: %s", post.Slug)
		return ctx.Err()
	}

	// Get existing post by slug
	existing, err := s.repo.GetBySlug(ctx, post.Slug)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrPostNotFound
	}

	// Update the existing post with new data
	existing.Title = post.Title
	existing.Content = post.Content
	existing.UpdatedAt = time.Now()

	// Check context again before final update
	if ctx.Err() != nil {
		log.Printf("Context cancelled before final update: %s", post.Slug)
		return ctx.Err()
	}

	return s.repo.Update(ctx, existing)
}

func (s *postService) DeletePost(ctx context.Context, id uint) error {
	log.Printf("Deleting post with ID: %d", id)
	select {
	case <-ctx.Done():
		log.Printf("Context cancelled while deleting post: %d", id)
		return ctx.Err()
	default:
		return s.repo.Delete(ctx, id)
	}
}