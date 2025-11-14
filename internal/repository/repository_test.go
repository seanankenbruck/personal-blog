package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/seanankenbruck/blog/internal/content"
)

func TestFilePostRepositoryCreation(t *testing.T) {
	repo := NewFilePostRepository()
	if repo == nil {
		t.Error("Expected FilePostRepository to be created, got nil")
	}
}

func TestGetAll(t *testing.T) {
	// Create a temporary directory for test posts
	tempDir := t.TempDir()

	// Create test markdown files with valid front matter
	testPost1 := `---
title: "Test Post 1"
slug: "test-post-1"
date: 2024-01-15T10:00:00Z
tags: ["test", "golang"]
description: "This is test post 1"
published: true
---

This is the content of test post 1.`

	testPost2 := `---
title: "Test Post 2"
slug: "test-post-2"
date: 2024-01-20T10:00:00Z
tags: ["test"]
description: "This is test post 2"
published: true
---

This is the content of test post 2.`

	// Write test files
	if err := os.WriteFile(filepath.Join(tempDir, "2024-01-15-test-post-1.md"), []byte(testPost1), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "2024-01-20-test-post-2.md"), []byte(testPost2), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Initialize the content loader
	content.Init(tempDir, false)
	if err := content.LoadPosts(); err != nil {
		t.Fatalf("LoadPosts() failed: %v", err)
	}

	// Create the repository
	repo := NewFilePostRepository()

	// Call GetAll
	posts, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll() returned error: %v", err)
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
}

func TestGetBySlug(t *testing.T) {
	// Create a temporary directory for test posts
	tempDir := t.TempDir()

	// Create test markdown file
	testPost := `---
title: "Test Post"
slug: "test-post"
date: 2024-01-15T10:00:00Z
description: "This is a test post"
published: true
---

This is the content of the test post.`

	// Write test file
	if err := os.WriteFile(filepath.Join(tempDir, "test-post.md"), []byte(testPost), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Initialize the content loader
	content.Init(tempDir, false)
	if err := content.LoadPosts(); err != nil {
		t.Fatalf("LoadPosts() failed: %v", err)
	}

	// Create the repository
	repo := NewFilePostRepository()

	t.Run("Get existing post by slug", func(t *testing.T) {
		post, err := repo.GetBySlug(context.Background(), "test-post")
		if err != nil {
			t.Fatalf("GetBySlug() returned error: %v", err)
		}

		if post == nil {
			t.Fatal("GetBySlug() returned nil post")
		}

		if post.Title != "Test Post" {
			t.Errorf("Expected title 'Test Post', got '%s'", post.Title)
		}

		if post.Slug != "test-post" {
			t.Errorf("Expected slug 'test-post', got '%s'", post.Slug)
		}
	})

	t.Run("Get non-existent post by slug", func(t *testing.T) {
		_, err := repo.GetBySlug(context.Background(), "non-existent")
		if err == nil {
			t.Error("GetBySlug() expected error for non-existent post, got nil")
		}
	})
}

func TestGetByID(t *testing.T) {
	// Create the repository
	repo := NewFilePostRepository()

	// GetByID is not supported in file-based repository
	_, err := repo.GetByID(context.Background(), 1)
	if err == nil {
		t.Error("GetByID() should return error for file-based repository")
	}
}
