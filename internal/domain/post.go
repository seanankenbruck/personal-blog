package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"gorm.io/gorm"
)

// Post represents a blog post persisted to the database
// GORM tags specify column types and constraints
// gorm.Model brings ID, CreatedAt, UpdatedAt, DeletedAt
// DeletedAt enables soft deletes if needed
// Unique index on slug ensures URL uniqueness
// Slug is generated from Title
// Content is stored as text
// Author can be a username or identifier
// -------------------------------------------
// Migrate this model with AutoMigrate
//   db.AutoMigrate(&Post{})
// -------------------------------------------

type Post struct {
	gorm.Model
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Slug      string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Published bool      `gorm:"default:false" json:"published"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
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
	Create(ctx context.Context, post *Post) error
	GetByID(ctx context.Context, id uint) (*Post, error)
	GetBySlug(ctx context.Context, slug string) (*Post, error)
	GetAll(ctx context.Context) ([]*Post, error)
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id uint) error
}

// PostService defines the interface for post business logic
type PostService interface {
	CreatePost(ctx context.Context, post *Post) error
	GetPost(ctx context.Context, id uint) (*Post, error)
	GetPostBySlug(ctx context.Context, slug string) (*Post, error)
	GetAllPosts(ctx context.Context) ([]*Post, error)
	UpdatePost(ctx context.Context, post *Post) error
	DeletePost(ctx context.Context, id uint) error
}