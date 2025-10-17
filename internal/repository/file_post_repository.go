package repository

import (
	"context"

	"github.com/seanankenbruck/blog/internal/content"
	"github.com/seanankenbruck/blog/internal/domain"
)

// FilePostRepository implements domain.PostRepository using file-based storage
type FilePostRepository struct{}

// NewFilePostRepository creates a new FilePostRepository
func NewFilePostRepository() *FilePostRepository {
	return &FilePostRepository{}
}


// GetByID is not supported in file-based repository (use GetBySlug instead)
func (r *FilePostRepository) GetByID(ctx context.Context, id uint) (*domain.Post, error) {
	return nil, domain.ErrNotSupported
}

// GetBySlug retrieves a post by its slug from the file system
func (r *FilePostRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	contentPost, err := content.GetPostBySlug(slug)
	if err != nil {
		return nil, domain.ErrPostNotFound
	}

	// Convert content.Post to domain.Post
	return contentPostToDomainPost(contentPost), nil
}

// GetAll retrieves all posts from the file system
func (r *FilePostRepository) GetAll(ctx context.Context) ([]*domain.Post, error) {
	contentPosts, err := content.GetAllPosts()
	if err != nil {
		return nil, err
	}

	// Convert content.Post slice to domain.Post slice
	domainPosts := make([]*domain.Post, len(contentPosts))
	for i, cp := range contentPosts {
		domainPosts[i] = contentPostToDomainPost(cp)
	}

	return domainPosts, nil
}


// contentPostToDomainPost converts a content.Post to a domain.Post
func contentPostToDomainPost(cp *content.Post) *domain.Post {
	return &domain.Post{
		Title:       cp.Title,
		Content:     cp.HTMLContent, // Use pre-rendered HTML
		Description: cp.Description,
		Slug:        cp.Slug,
		Published:   cp.Published,
		CreatedAt:   cp.Date,
		UpdatedAt:   cp.Date,
	}
}
