package service

import (
	"context"
	"log"
	"testing"

	"github.com/seanankenbruck/blog/internal/domain"
)

// mockPostRepository is a mock implementation of domain.PostRepository
type mockPostRepository struct {
	posts map[string]*domain.Post
}

func newMockPostRepository() *mockPostRepository {
	return &mockPostRepository{
		posts: make(map[string]*domain.Post),
	}
}

func (m *mockPostRepository) GetByID(ctx context.Context, id uint) (*domain.Post, error) {
	return nil, domain.ErrNotSupported
}

func (m *mockPostRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	post, ok := m.posts[slug]
	if !ok {
		return nil, domain.ErrPostNotFound
	}
	return post, nil
}

func (m *mockPostRepository) GetAll(ctx context.Context) ([]*domain.Post, error) {
	posts := make([]*domain.Post, 0, len(m.posts))
	for _, post := range m.posts {
		posts = append(posts, post)
	}
	return posts, nil
}

func TestNewPostService(t *testing.T) {
	log.Println("Testing NewPostService...")

	// Create a mock repository
	mockRepo := &mockPostRepository{}

	// Create the post service
	service := NewPostService(mockRepo)

	// Verify the service is not nil
	if service == nil {
		t.Error("Expected NewPostService to return a non-nil service")
	}

	log.Println("NewPostService test completed")
}

func TestGetAllPosts(t *testing.T) {
	log.Println("Testing GetAllPosts...")

	// Create a mock repository with sample posts
	mockRepo := newMockPostRepository()
	mockRepo.posts["test-post-1"] = &domain.Post{Slug: "test-post-1", Title: "Test Post 1"}
	mockRepo.posts["test-post-2"] = &domain.Post{Slug: "test-post-2", Title: "Test Post 2"}

	// Create the post service
	service := NewPostService(mockRepo)

	// Call GetAllPosts
	posts, err := service.GetAllPosts(context.Background())
	if err != nil {
		t.Fatalf("GetAllPosts() returned error: %v", err)
	}

	// Verify the results
	if len(posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(posts))
	}

	expectedSlugs := map[string]bool{
		"test-post-1": true,
		"test-post-2": true,
	}

	for _, post := range posts {
		if !expectedSlugs[post.Slug] {
			t.Errorf("Unexpected post slug: %s", post.Slug)
		}
	}

	log.Println("GetAllPosts test completed")
}

func TestGetPostBySlug(t *testing.T) {
	log.Println("Testing GetPostBySlug...")

	// Create a mock repository with a sample post
	mockRepo := newMockPostRepository()
	mockRepo.posts["test-post"] = &domain.Post{Slug: "test-post", Title: "Test Post"}

	// Create the post service
	service := NewPostService(mockRepo)

	// Call GetPostBySlug with an existing slug
	post, err := service.GetPostBySlug(context.Background(), "test-post")
	if err != nil {
		t.Fatalf("GetPostBySlug() returned error: %v", err)
	}
	if post.Slug != "test-post" {
		t.Errorf("Expected slug 'test-post', got '%s'", post.Slug)
	}

	// Call GetPostBySlug with a non-existing slug
	_, err = service.GetPostBySlug(context.Background(), "non-existent-post")
	if err != domain.ErrPostNotFound {
		t.Errorf("Expected ErrPostNotFound for non-existent slug, got: %v", err)
	}

	log.Println("GetPostBySlug test completed")
}

func TestGetPost(t *testing.T) {
	log.Println("Testing GetPost...")

	mockRepo := newMockPostRepository()
	service := NewPostService(mockRepo)

	// GetByID returns ErrNotSupported in the mock
	_, err := service.GetPost(context.Background(), 1)
	if err != domain.ErrNotSupported {
		t.Errorf("Expected ErrNotSupported, got: %v", err)
	}

	log.Println("GetPost test completed")
}

func TestGetPostBySlugContextCancelled(t *testing.T) {
	log.Println("Testing GetPostBySlug with cancelled context...")

	mockRepo := newMockPostRepository()
	mockRepo.posts["test-post"] = &domain.Post{Slug: "test-post", Title: "Test Post"}
	service := NewPostService(mockRepo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.GetPostBySlug(ctx, "test-post")
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}

	log.Println("GetPostBySlug context cancelled test completed")
}

func TestGetAllPostsContextCancelled(t *testing.T) {
	log.Println("Testing GetAllPosts with cancelled context...")

	mockRepo := newMockPostRepository()
	service := NewPostService(mockRepo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.GetAllPosts(ctx)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}

	log.Println("GetAllPosts context cancelled test completed")
}

func TestGetPostContextCancelled(t *testing.T) {
	log.Println("Testing GetPost with cancelled context...")

	mockRepo := newMockPostRepository()
	service := NewPostService(mockRepo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.GetPost(ctx, 1)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}

	log.Println("GetPost context cancelled test completed")
}

// mockReturnsNilPostRepository returns nil for GetBySlug to test nil handling
type mockReturnsNilPostRepository struct{}

func (m *mockReturnsNilPostRepository) GetByID(ctx context.Context, id uint) (*domain.Post, error) {
	return nil, nil
}

func (m *mockReturnsNilPostRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	return nil, nil
}

func (m *mockReturnsNilPostRepository) GetAll(ctx context.Context) ([]*domain.Post, error) {
	return nil, nil
}

func TestGetPostBySlugReturnsNil(t *testing.T) {
	log.Println("Testing GetPostBySlug when repo returns nil...")

	mockRepo := &mockReturnsNilPostRepository{}
	service := NewPostService(mockRepo)

	_, err := service.GetPostBySlug(context.Background(), "any-slug")
	if err != domain.ErrPostNotFound {
		t.Errorf("Expected ErrPostNotFound when repo returns nil, got: %v", err)
	}

	log.Println("GetPostBySlug nil test completed")
}
