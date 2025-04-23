package domain

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
)

// Post represents a blog post
type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GenerateSlug creates a URL-friendly slug from the post title
func (p *Post) GenerateSlug() string {
	// Convert to lowercase
	slug := strings.ToLower(p.Title)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove all non-alphanumeric characters except hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}

// Validate checks if the post has all required fields
func (p *Post) Validate() error {
	if p.Title == "" {
		return errors.New("title is required")
	}
	if p.Content == "" {
		return errors.New("content is required")
	}
	if p.Author == "" {
		return errors.New("author is required")
	}
	return nil
}

// Update updates the post with new values
func (p *Post) Update(newPost *Post) {
	if newPost.Title != "" {
		p.Title = newPost.Title
		p.Slug = p.GenerateSlug() // Regenerate slug when title changes
	}
	if newPost.Content != "" {
		p.Content = newPost.Content
	}
	if newPost.Author != "" {
		p.Author = newPost.Author
	}
	p.UpdatedAt = time.Now()
}

// PostRepository defines the interface for post data access
type PostRepository interface {
	Create(ctx context.Context, post *Post) error
	GetByID(ctx context.Context, id int64) (*Post, error)
	GetBySlug(ctx context.Context, slug string) (*Post, error)
	GetAll(ctx context.Context) ([]*Post, error)
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id int64) error
}

// PostService defines the interface for post business logic
type PostService interface {
	CreatePost(ctx context.Context, post *Post) error
	GetPost(ctx context.Context, id int64) (*Post, error)
	GetPostBySlug(ctx context.Context, slug string) (*Post, error)
	GetAllPosts(ctx context.Context) ([]*Post, error)
	UpdatePost(ctx context.Context, post *Post) error
	DeletePost(ctx context.Context, id int64) error
}