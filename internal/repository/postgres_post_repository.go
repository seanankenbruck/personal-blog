package repository

import (
	"context"

	"github.com/seanankenbruck/blog/internal/domain"
	"gorm.io/gorm"
)

type PostgresPostRepository struct {
	db *gorm.DB
}

func NewPostgresPostRepository(db *gorm.DB) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

func (r *PostgresPostRepository) Create(ctx context.Context, post *domain.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *PostgresPostRepository) GetByID(ctx context.Context, id uint) (*domain.Post, error) {
	var post domain.Post
	if err := r.db.WithContext(ctx).First(&post, id).Error; err != nil {
		return nil, domain.ErrPostNotFound
	}
	return &post, nil
}

func (r *PostgresPostRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	var post domain.Post
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&post).Error; err != nil {
		return nil, domain.ErrPostNotFound
	}
	return &post, nil
}

func (r *PostgresPostRepository) GetAll(ctx context.Context) ([]*domain.Post, error) {
	var posts []*domain.Post
	if err := r.db.WithContext(ctx).Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostgresPostRepository) Update(ctx context.Context, post *domain.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *PostgresPostRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Post{}, id).Error
}