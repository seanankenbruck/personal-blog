package handler

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/seanankenbruck/blog/internal/content"
	"github.com/seanankenbruck/blog/internal/repository"
	"github.com/seanankenbruck/blog/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestPostHandlerCreation(t *testing.T) {
	log.Println("Testing PostHandler creation...")

	// Create a test repository and service
	repo := repository.NewFilePostRepository()
	svc := service.NewPostService(repo)
	postHandler := NewPostHandler(svc)

	// Verify the handler was created successfully
	assert.NotNil(t, postHandler)
	assert.NotNil(t, postHandler.postService)

	log.Println("PostHandler creation test completed")
}

func TestServiceCreation(t *testing.T) {
	log.Println("Testing service creation...")

	// Create a test repository and service
	repo := repository.NewFilePostRepository()
	svc := service.NewPostService(repo)

	// Verify the service was created successfully
	assert.NotNil(t, svc)

	log.Println("Service creation test completed")
}

func TestRepositoryCreation(t *testing.T) {
	log.Println("Testing repository creation...")

	// Create a test repository
	repo := repository.NewFilePostRepository()

	// Verify the repository was created successfully
	assert.NotNil(t, repo)

	log.Println("Repository creation test completed")
}

func TestSetupTemplates(t *testing.T) {
	t.Skip("Skipping template setup test - requires actual template files")
}

func TestGetPosts(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup: create temporary content directory with test posts
	tempDir := t.TempDir()
	testPost := `---
title: "Test Post"
slug: "test-post"
date: 2024-01-15T10:00:00Z
description: "Test description"
published: true
---

Test content.`

	if err := os.WriteFile(filepath.Join(tempDir, "test.md"), []byte(testPost), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Initialize content loader with test directory
	content.Init(tempDir, false)
	if err := content.LoadPosts(); err != nil {
		t.Fatalf("LoadPosts() failed: %v", err)
	}

	// Create a test repository, service, and handler
	repo := repository.NewFilePostRepository()
	svc := service.NewPostService(repo)

	// Setup Gin router and handler
	router := gin.New()
	router.GET("/posts", GetPosts(svc))

	// Create a test request
	req, err := http.NewRequest(http.MethodGet, "/posts", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Post")
}

func TestGetPost(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup: create temporary content directory with test posts
	tempDir := t.TempDir()
	testPost := `---
title: "Test Post"
slug: "test-post"
date: 2024-01-15T10:00:00Z
description: "Test description"
published: true
---

Test content.`

	if err := os.WriteFile(filepath.Join(tempDir, "test.md"), []byte(testPost), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Initialize content loader with test directory
	content.Init(tempDir, false)
	if err := content.LoadPosts(); err != nil {
		t.Fatalf("LoadPosts() failed: %v", err)
	}

	// Create a test repository, service, and handler
	repo := repository.NewFilePostRepository()
	svc := service.NewPostService(repo)

	t.Run("Get existing post returns 200 with JSON", func(t *testing.T) {
		// Setup Gin router and handler
		router := gin.New()
		router.GET("/posts/:slug", GetPost(svc))

		// Create a test request
		req, err := http.NewRequest(http.MethodGet, "/posts/test-post", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Accept", "application/json")

		// Create a response recorder
		w := httptest.NewRecorder()

		// Perform the request
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Test Post")
	})

	// Note: Testing 404 case requires HTML templates to be loaded
	// Skipping this test as it's more appropriate for integration tests
}

func TestHomePage(t *testing.T) {
	t.Skip("Skipping HomePage test - requires HTML templates")
}

func TestAboutPage(t *testing.T) {
	t.Skip("Skipping AboutPage test - requires HTML templates")
}

func TestPortfolioPage(t *testing.T) {
	t.Skip("Skipping PortfolioPage test - requires HTML templates")
}
