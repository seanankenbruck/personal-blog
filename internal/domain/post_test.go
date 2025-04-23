package domain

import (
	"testing"
	"time"
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
				Author:  "Test Author",
			},
			wantErr: false,
		},
		{
			name: "empty title",
			post: &Post{
				Title:   "",
				Content: "This is a test post",
				Author:  "Test Author",
			},
			wantErr: true,
		},
		{
			name: "empty content",
			post: &Post{
				Title:   "Test Post",
				Content: "",
				Author:  "Test Author",
			},
			wantErr: true,
		},
		{
			name: "empty author",
			post: &Post{
				Title:   "Test Post",
				Content: "This is a test post",
				Author:  "",
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
		ID:        1,
		Title:     "Original Title",
		Content:   "Original Content",
		Author:    "Original Author",
		CreatedAt: now,
		UpdatedAt: now,
	}

	newPost := &Post{
		Title:   "New Title",
		Content: "New Content",
		Author:  "New Author",
	}

	post.Update(newPost)

	if post.Title != newPost.Title {
		t.Errorf("Expected title %v, got %v", newPost.Title, post.Title)
	}
	if post.Content != newPost.Content {
		t.Errorf("Expected content %v, got %v", newPost.Content, post.Content)
	}
	if post.Author != newPost.Author {
		t.Errorf("Expected author %v, got %v", newPost.Author, post.Author)
	}
	if post.UpdatedAt == now {
		t.Error("UpdatedAt should have changed")
	}
}