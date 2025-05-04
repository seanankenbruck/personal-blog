package domain

import (
	"testing"
	"time"
	"gorm.io/gorm"
)

func TestPost_Validate(t *testing.T) {
	tests := []struct {
		name    string
		post    *Post
		wantErr bool
	}{
		{
			name: "valid post",
			post: &Post{
				Title:   "Test Post",
				Content: "This is a test post",
				Description: "This is a test description",
			},
			wantErr: false,
		},
		{
			name: "empty title",
			post: &Post{
				Title:   "",
				Content: "This is a test post",
				Description: "This is a test description",
			},
			wantErr: true,
		},
		{
			name: "empty content",
			post: &Post{
				Title:   "Test Post",
				Content: "",
				Description: "This is a test description",
			},
			wantErr: true,
		},
		{
			name: "empty description",
			post: &Post{
				Title:   "Test Post",
				Content: "Test Content",
				Description: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.post.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Post.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPost_Update(t *testing.T) {
	now := time.Now()
	post := &Post{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Title:     "Original Title",
		Content:   "Original Content",
		Description: "Original Description",
	}

	newPost := &Post{
		Title:   "New Title",
		Content: "New Content",
		Description: "New Description",
	}

	post.Update(newPost)

	if post.Title != newPost.Title {
		t.Errorf("Expected title %v, got %v", newPost.Title, post.Title)
	}
	if post.Content != newPost.Content {
		t.Errorf("Expected content %v, got %v", newPost.Content, post.Content)
	}
	if post.Description != newPost.Description {
		t.Errorf("Expected description %v, got %v", newPost.Description, post.Description)
	}
	if post.UpdatedAt == now {
		t.Error("UpdatedAt should have changed")
	}
}