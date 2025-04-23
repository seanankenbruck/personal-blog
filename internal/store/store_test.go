package store

import (
	"testing"
)

func TestStoreOperations(t *testing.T) {
	store := NewStore()

	// Test Create
	post := Post{
		Title:   "Test Post",
		Content: "# Test Content\n\nThis is a test post.",
		Author:  "Test Author",
		Slug:    "test-post",
	}
	createdPost := store.Create(post)

	if createdPost.ID == 0 {
		t.Error("Expected post ID to be set")
	}
	if createdPost.Title != post.Title {
		t.Errorf("Expected title %q, got %q", post.Title, createdPost.Title)
	}
	if createdPost.Content != post.Content {
		t.Errorf("Expected content %q, got %q", post.Content, createdPost.Content)
	}
	if createdPost.Author != post.Author {
		t.Errorf("Expected author %q, got %q", post.Author, createdPost.Author)
	}
	if createdPost.Slug != post.Slug {
		t.Errorf("Expected slug %q, got %q", post.Slug, createdPost.Slug)
	}
	if createdPost.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if createdPost.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	// Test Get
	retrievedPost, exists := store.Get(createdPost.ID)
	if !exists {
		t.Error("Expected post to exist")
	}
	if retrievedPost.ID != createdPost.ID {
		t.Errorf("Expected ID %d, got %d", createdPost.ID, retrievedPost.ID)
	}

	// Test Update
	updatedPost := Post{
		Title:   "Updated Post",
		Content: "# Updated Content\n\nThis is an updated post.",
		Author:  "Updated Author",
		Slug:    "updated-post",
	}
	updated, exists := store.Update(createdPost.ID, updatedPost)
	if !exists {
		t.Error("Expected post to exist for update")
	}
	if updated.Title != updatedPost.Title {
		t.Errorf("Expected updated title %q, got %q", updatedPost.Title, updated.Title)
	}
	if updated.Content != updatedPost.Content {
		t.Errorf("Expected updated content %q, got %q", updatedPost.Content, updated.Content)
	}
	if updated.Author != updatedPost.Author {
		t.Errorf("Expected updated author %q, got %q", updatedPost.Author, updated.Author)
	}
	if updated.UpdatedAt.Before(updated.CreatedAt) {
		t.Error("Expected UpdatedAt to be after CreatedAt")
	}

	// Test Delete
	if !store.DeleteBySlug(updatedPost.Slug) {
		t.Error("Expected delete to succeed")
	}
	if _, exists := store.GetBySlug(updatedPost.Slug); exists {
		t.Error("Expected post to be deleted")
	}

	// Test GetAll
	posts := store.GetAll()
	if len(posts) != 0 {
		t.Errorf("Expected 0 posts, got %d", len(posts))
	}
}

func TestStoreConcurrentOperations(t *testing.T) {
	store := NewStore()
	done := make(chan bool)

	// Create posts concurrently
	for i := 0; i < 10; i++ {
		go func(i int) {
			post := Post{
				Title:   "Test Post",
				Content: "Test Content",
				Author:  "Test Author",
			}
			store.Create(post)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	posts := store.GetAll()
	if len(posts) != 10 {
		t.Errorf("Expected 10 posts, got %d", len(posts))
	}
}