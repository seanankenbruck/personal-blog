package domain

import (
	"testing"
	"time"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "Basic title",
			title:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "Title with multiple spaces",
			title:    "This Is A Test",
			expected: "this-is-a-test",
		},
		{
			name:     "Title with single quotes",
			title:    "Don't Stop",
			expected: "dont-stop",
		},
		{
			name:     "Title with double quotes",
			title:    `The "Best" Post`,
			expected: "the-best-post",
		},
		{
			name:     "Title with mixed quotes",
			title:    `It's "The" Best`,
			expected: "its-the-best",
		},
		{
			name:     "Already lowercase",
			title:    "already lowercase",
			expected: "already-lowercase",
		},
		{
			name:     "Single word",
			title:    "Hello",
			expected: "hello",
		},
		{
			name:     "Empty title",
			title:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post := &Post{Title: tt.title}
			got := post.GenerateSlug()
			if got != tt.expected {
				t.Errorf("GenerateSlug() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateUniqueSlug(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		existingSlugs map[string]bool
		expected      string
	}{
		{
			name:          "No collision",
			title:         "Hello World",
			existingSlugs: map[string]bool{},
			expected:      "hello-world",
		},
		{
			name:  "Single collision",
			title: "Hello World",
			existingSlugs: map[string]bool{
				"hello-world": true,
			},
			expected: "hello-world-1",
		},
		{
			name:  "Multiple collisions",
			title: "Hello World",
			existingSlugs: map[string]bool{
				"hello-world":   true,
				"hello-world-1": true,
				"hello-world-2": true,
			},
			expected: "hello-world-3",
		},
		{
			name:  "Non-sequential collision",
			title: "Test Post",
			existingSlugs: map[string]bool{
				"test-post":   true,
				"test-post-2": true,
			},
			expected: "test-post-1",
		},
		{
			name:          "Empty existing slugs",
			title:         "New Post",
			existingSlugs: nil,
			expected:      "new-post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post := &Post{Title: tt.title}
			got := post.GenerateUniqueSlug(tt.existingSlugs)
			if got != tt.expected {
				t.Errorf("GenerateUniqueSlug() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		post    *Post
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid post",
			post: &Post{
				Title:       "Test Title",
				Content:     "Test content",
				Description: "Test description",
			},
			wantErr: false,
		},
		{
			name: "Missing title",
			post: &Post{
				Title:       "",
				Content:     "Test content",
				Description: "Test description",
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "Missing content",
			post: &Post{
				Title:       "Test Title",
				Content:     "",
				Description: "Test description",
			},
			wantErr: true,
			errMsg:  "content is required",
		},
		{
			name: "Missing description",
			post: &Post{
				Title:       "Test Title",
				Content:     "Test content",
				Description: "",
			},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "All fields missing",
			post: &Post{
				Title:       "",
				Content:     "",
				Description: "",
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "Valid post with optional fields",
			post: &Post{
				Title:       "Test Title",
				Content:     "Test content",
				Description: "Test description",
				Slug:        "test-slug",
				Published:   true,
				ID:          1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.post.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	// Create an initial post with all fields set
	originalTime := time.Now().Add(-24 * time.Hour)
	originalPost := &Post{
		ID:          1,
		Title:       "Original Title",
		Content:     "Original content",
		Description: "Original description",
		Slug:        "original-slug",
		Published:   false,
		CreatedAt:   originalTime,
		UpdatedAt:   originalTime,
	}

	// Create a new post with updated values
	newPost := &Post{
		Title:       "Updated Title",
		Content:     "Updated content",
		Description: "Updated description",
		Published:   true,
	}

	// Record the time before update
	beforeUpdate := time.Now()

	// Perform the update
	originalPost.Update(newPost)

	// Verify all fields were updated correctly
	if originalPost.Title != "Updated Title" {
		t.Errorf("Update() Title = %v, want %v", originalPost.Title, "Updated Title")
	}
	if originalPost.Content != "Updated content" {
		t.Errorf("Update() Content = %v, want %v", originalPost.Content, "Updated content")
	}
	if originalPost.Description != "Updated description" {
		t.Errorf("Update() Description = %v, want %v", originalPost.Description, "Updated description")
	}
	if originalPost.Published != true {
		t.Errorf("Update() Published = %v, want %v", originalPost.Published, true)
	}

	// Verify slug was regenerated from the new title
	expectedSlug := "updated-title"
	if originalPost.Slug != expectedSlug {
		t.Errorf("Update() Slug = %v, want %v", originalPost.Slug, expectedSlug)
	}

	// Verify UpdatedAt was updated to a recent time
	if originalPost.UpdatedAt.Before(beforeUpdate) {
		t.Errorf("Update() UpdatedAt was not updated to a recent time")
	}

	// Verify ID and CreatedAt were not changed
	if originalPost.ID != 1 {
		t.Errorf("Update() ID should not change, got %v, want %v", originalPost.ID, 1)
	}
	if !originalPost.CreatedAt.Equal(originalTime) {
		t.Errorf("Update() CreatedAt should not change, got %v, want %v", originalPost.CreatedAt, originalTime)
	}
}

func TestUpdate_WithQuotesInTitle(t *testing.T) {
	originalPost := &Post{
		ID:          1,
		Title:       "Original Title",
		Content:     "Original content",
		Description: "Original description",
		Slug:        "original-title",
	}

	newPost := &Post{
		Title:       `The "Best" Post Ever`,
		Content:     "New content",
		Description: "New description",
	}

	originalPost.Update(newPost)

	// Verify slug was generated correctly (quotes removed)
	expectedSlug := "the-best-post-ever"
	if originalPost.Slug != expectedSlug {
		t.Errorf("Update() Slug = %v, want %v", originalPost.Slug, expectedSlug)
	}
}

func TestUpdate_PreservesImmutableFields(t *testing.T) {
	createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	originalPost := &Post{
		ID:        42,
		CreatedAt: createdAt,
	}

	newPost := &Post{
		Title:       "New Title",
		Content:     "New content",
		Description: "New description",
	}

	originalPost.Update(newPost)

	// Verify immutable fields are preserved
	if originalPost.ID != 42 {
		t.Errorf("Update() should preserve ID, got %v, want %v", originalPost.ID, 42)
	}
	if !originalPost.CreatedAt.Equal(createdAt) {
		t.Errorf("Update() should preserve CreatedAt, got %v, want %v", originalPost.CreatedAt, createdAt)
	}
}
