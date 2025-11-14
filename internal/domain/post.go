package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Content in this file is reserved for future database integration.
// Post represents a blog post
type Post struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Description string    `json:"description"`
	Slug        string    `json:"slug"`
	Published   bool      `json:"published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GenerateSlug creates a URL-friendly slug from the post title
func (p *Post) GenerateSlug() string {
	slug := strings.ToLower(p.Title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, "\"", "")
	return slug
}

// GenerateUniqueSlug creates a unique slug by appending a number if needed
func (p *Post) GenerateUniqueSlug(existingSlugs map[string]bool) string {
	baseSlug := p.GenerateSlug()
	slug := baseSlug
	counter := 1

	for existingSlugs[slug] {
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

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
	if p.Description == "" {
		return errors.New("description is required")
	}
	return nil
}

// Update updates the post with new values
func (p *Post) Update(newPost *Post) {
	p.Title = newPost.Title
	p.Content = newPost.Content
	p.Description = newPost.Description
	p.Slug = newPost.GenerateSlug()
	p.Published = newPost.Published
	p.UpdatedAt = time.Now()
}

// PostRepository defines the interface for post data access
type PostRepository interface {
	GetByID(ctx context.Context, id uint) (*Post, error)
	GetBySlug(ctx context.Context, slug string) (*Post, error)
	GetAll(ctx context.Context) ([]*Post, error)
}

// PostService defines the interface for post business logic
type PostService interface {
	GetPost(ctx context.Context, id uint) (*Post, error)
	GetPostBySlug(ctx context.Context, slug string) (*Post, error)
	GetAllPosts(ctx context.Context) ([]*Post, error)
}
