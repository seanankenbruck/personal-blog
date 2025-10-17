package service

import (
	"context"
	"log"

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

