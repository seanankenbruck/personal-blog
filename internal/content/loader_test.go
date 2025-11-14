package content

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	dir := "test/content"
	devMode := true

	Init(dir, devMode)

	if contentDir != dir {
		t.Errorf("Init failed: expected contentDir %s, got %s", dir, contentDir)
	}
	if isDev != devMode {
		t.Errorf("Init failed: expected isDev %v, got %v", devMode, isDev)
	}
}

func TestLoadPosts(t *testing.T) {
	t.Run("Non-existent content directory", func(t *testing.T) {
		// Initialize with a non-existent directory
		Init("nonexistent/dir", false)

		// Call LoadPosts should return an error
		err := LoadPosts()
		if err == nil {
			t.Error("LoadPosts() expected error for non-existent directory, got nil")
		}
	})

	t.Run("Empty content directory", func(t *testing.T) {
		// Create a temporary directory for test posts
		tempDir := t.TempDir()

		// Initialize the content loader with the empty temp directory
		Init(tempDir, false)

		// Call LoadPosts - should succeed but load no posts
		err := LoadPosts()
		if err != nil {
			t.Errorf("LoadPosts() unexpected error: %v", err)
		}

		// Verify no posts were loaded
		loadedPosts, err := GetAllPosts()
		if err != nil {
			t.Errorf("GetAllPosts() unexpected error: %v", err)
		}
		if len(loadedPosts) != 0 {
			t.Errorf("Expected 0 posts, got %d", len(loadedPosts))
		}
	})

	t.Run("Valid content directory with published posts", func(t *testing.T) {
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
		Init(tempDir, false)

		// Call LoadPosts
		err := LoadPosts()
		if err != nil {
			t.Errorf("LoadPosts() unexpected error: %v", err)
		}

		// Verify posts were loaded
		loadedPosts, err := GetAllPosts()
		if err != nil {
			t.Errorf("GetAllPosts() unexpected error: %v", err)
		}
		if len(loadedPosts) != 2 {
			t.Errorf("Expected 2 posts, got %d", len(loadedPosts))
		}

		// Verify posts are sorted by date (newest first)
		if len(loadedPosts) == 2 {
			if loadedPosts[0].Title != "Test Post 2" {
				t.Errorf("Expected first post to be 'Test Post 2', got '%s'", loadedPosts[0].Title)
			}
			if loadedPosts[1].Title != "Test Post 1" {
				t.Errorf("Expected second post to be 'Test Post 1', got '%s'", loadedPosts[1].Title)
			}
		}

		// Verify we can retrieve a post by slug
		post, err := GetPostBySlug("test-post-1")
		if err != nil {
			t.Errorf("GetPostBySlug() unexpected error: %v", err)
		}
		if post.Title != "Test Post 1" {
			t.Errorf("Expected post title 'Test Post 1', got '%s'", post.Title)
		}
	})

	t.Run("Unpublished posts are filtered out", func(t *testing.T) {
		// Create a temporary directory for test posts
		tempDir := t.TempDir()

		// Create one published and one unpublished post
		publishedPost := `---
title: "Published Post"
slug: "published-post"
date: 2024-01-15T10:00:00Z
description: "This is published"
published: true
---

Published content.`

		unpublishedPost := `---
title: "Unpublished Post"
slug: "unpublished-post"
date: 2024-01-20T10:00:00Z
description: "This is unpublished"
published: false
---

Unpublished content.`

		// Write test files
		if err := os.WriteFile(filepath.Join(tempDir, "published.md"), []byte(publishedPost), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tempDir, "unpublished.md"), []byte(unpublishedPost), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Initialize and load posts
		Init(tempDir, false)
		err := LoadPosts()
		if err != nil {
			t.Errorf("LoadPosts() unexpected error: %v", err)
		}

		// Verify only published post is loaded
		loadedPosts, err := GetAllPosts()
		if err != nil {
			t.Errorf("GetAllPosts() unexpected error: %v", err)
		}
		if len(loadedPosts) != 1 {
			t.Errorf("Expected 1 published post, got %d", len(loadedPosts))
		}
		if len(loadedPosts) == 1 && loadedPosts[0].Title != "Published Post" {
			t.Errorf("Expected 'Published Post', got '%s'", loadedPosts[0].Title)
		}
	})

	t.Run("Invalid front matter returns error", func(t *testing.T) {
		// Create a temporary directory
		tempDir := t.TempDir()

		// Create a file with invalid YAML front matter
		invalidPost := `---
title: "Test Post"
date: this-is-not-a-valid-date
---

Content here.`

		if err := os.WriteFile(filepath.Join(tempDir, "invalid.md"), []byte(invalidPost), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Initialize and try to load posts
		Init(tempDir, false)
		err := LoadPosts()
		if err == nil {
			t.Error("LoadPosts() expected error for invalid front matter, got nil")
		}
	})
}

func TestParseFrontMatter(t *testing.T) {
	t.Run("Invalid front matter format returns error", func(t *testing.T) {
		invalidFM := `title: "Missing dashes"
		date: 2024-01-15T10:00:00Z
		`

		_, _, err := parseFrontMatter(invalidFM)
		if err == nil {
			t.Error("parseFrontMatter() expected error for invalid format, got nil")
		}
	})

	t.Run("Invalid yaml content returns error", func(t *testing.T) {
		invalidYAML := `---
		title: "Test Post"
		date: invalid-date-format
		---`

		_, _, err := parseFrontMatter(invalidYAML)
		if err == nil {
			t.Error("parseFrontMatter() expected error for invalid YAML, got nil")
		}
	})

	t.Run("Valid front matter parsed correctly", func(t *testing.T) {
		validFM := `---
title: "Valid Post"
slug: "valid-post"
date: 2024-01-15T10:00:00Z
tags: ["test"]
description: "This is a valid post"
published: true
---

Unpublished content.`

		fm, content, err := parseFrontMatter(validFM)
		if err != nil {
			t.Errorf("parseFrontMatter() unexpected error: %v", err)
		}

		if fm.Title != "Valid Post" {
			t.Errorf("Expected title 'Valid Post', got '%s'", fm.Title)
		}
		if fm.Slug != "valid-post" {
			t.Errorf("Expected slug 'valid-post', got '%s'", fm.Slug)
		}
		if !fm.Published {
			t.Error("Expected published to be true")
		}
		expectedContent := "Unpublished content."
		if content != expectedContent {
			t.Errorf("Expected content '%s', got '%s'", expectedContent, content)
		}
	})
}

func TestGenerateSlugFromFileName(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"2024-01-15-my-first-post.md", "my-first-post"},
		{"2023-12-31-year-end-review.md", "year-end-review"},
		{"no-date-slug.md", "no-date-slug"},
		{"2024-02-29-leap-year-post.md", "leap-year-post"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			slug := generateSlugFromFilename(tt.filename)
			if slug != tt.expected {
				t.Errorf("generateSlugFromFilename(%s) = %s; want %s", tt.filename, slug, tt.expected)
			}
		})
	}
}

func TestGetRecentPosts(t *testing.T) {
	// Setup: create a temporary content directory with test posts
	tempDir := t.TempDir()

	postsContent := []string{
		`---
title: "Post 1"
slug: "post-1"
date: 2024-01-10T10:00:00Z
description: "First post"
published: true
---

Content of post 1.`,
		`---
title: "Post 2"
slug: "post-2"
date: 2024-01-15T10:00:00Z
description: "Second post"
published: true
---

Content of post 2.`,
		`---
title: "Post 3"
slug: "post-3"
date: 2024-01-20T10:00:00Z
description: "Third post"
published: true
---

Content of post 3.`,
	}

	for i, content := range postsContent {
		filename := filepath.Join(tempDir, fmt.Sprintf("post%d.md", i))
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Initialize and load posts
	Init(tempDir, false)
	if err := LoadPosts(); err != nil {
		t.Fatalf("LoadPosts() unexpected error: %v", err)
	}

	// Test GetRecentPosts
	recentPosts, err := GetRecentPosts(2)
	if err != nil {
		t.Errorf("GetRecentPosts() unexpected error: %v", err)
	}

	if len(recentPosts) != 2 {
		t.Errorf("Expected 2 recent posts, got %d", len(recentPosts))
	}

	if recentPosts[0].Title != "Post 3" || recentPosts[1].Title != "Post 2" {
		t.Errorf("GetRecentPosts() returned incorrect posts")
	}
}

func TestReload(t *testing.T) {
	// Setup: create a temporary content directory with test posts
	tempDir := t.TempDir()

	initialPost := `---
title: "Initial Post"
slug: "initial-post"
date: 2024-01-10T10:00:00Z
description: "Initial post"
published: true
---

Content of initial post.`

	if err := os.WriteFile(filepath.Join(tempDir, "initial.md"), []byte(initialPost), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Initialize and load posts
	Init(tempDir, false)
	if err := LoadPosts(); err != nil {
		t.Fatalf("LoadPosts() unexpected error: %v", err)
	}

	// Verify initial post is loaded
	posts, err := GetAllPosts()
	if err != nil {
		t.Fatalf("GetAllPosts() unexpected error: %v", err)
	}
	if len(posts) != 1 || posts[0].Title != "Initial Post" {
		t.Fatalf("Initial post not loaded correctly")
	}

	// Add a new post
	newPost := `---
title: "New Post"
slug: "new-post"
date: 2024-01-15T10:00:00Z
description: "Newly added post"
published: true
---

Content of new post.`

	if err := os.WriteFile(filepath.Join(tempDir, "new.md"), []byte(newPost), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Call Reload to refresh posts
	if err := Reload(); err != nil {
		t.Fatalf("Reload() unexpected error: %v", err)
	}

	// Verify both posts are now loaded
	posts, err = GetAllPosts()
	if err != nil {
		t.Fatalf("GetAllPosts() unexpected error: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("Expected 2 posts after reload, got %d", len(posts))
	}

	foundNewPost := false
	for _, post := range posts {
		if post.Title == "New Post" {
			foundNewPost = true
			break
		}
	}
	if !foundNewPost {
		t.Fatalf("Newly added post not found after reload")
	}
}
