package repository

import (
	"context"

	"gorm.io/gorm"
	"github.com/seanankenbruck/blog/internal/db"
	"github.com/seanankenbruck/blog/internal/domain"
)

// PostRepository uses GORM to persist Posts
type PostRepository struct {
	db *gorm.DB
}

// NewPostRepository returns a GORM implementation of PostRepository
func NewPostRepository() domain.PostRepository {
	return &PostRepository{db: db.DB}
}

func (r *PostRepository) Create(ctx context.Context, post *domain.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *PostRepository) GetByID(ctx context.Context, id uint) (*domain.Post, error) {
	var post domain.Post
	if err := r.db.WithContext(ctx).First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrPostNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	var post domain.Post
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrPostNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) GetAll(ctx context.Context) ([]*domain.Post, error) {
	var posts []*domain.Post
	if err := r.db.WithContext(ctx).Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) Update(ctx context.Context, post *domain.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *PostRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Post{}, id).Error
}