package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/seanankenbruck/blog/internal/store"
	"github.com/seanankenbruck/blog/internal/middleware"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Set up session store
	store := cookie.NewStore([]byte("test-secret"))
	r.Use(sessions.Sessions("session", store))

	// apply authentication middleware in tests
	r.Use(middleware.AuthMiddleware())

	// Set up template engine with custom functions
	r.SetFuncMap(map[string]interface{}{
		"safeHTML": func(text string) string {
			extensions := parser.CommonExtensions | parser.AutoHeadingIDs
			p := parser.NewWithExtensions(extensions)
			htmlFlags := html.CommonFlags | html.HrefTargetBlank
			opts := html.RendererOptions{Flags: htmlFlags}
			renderer := html.NewRenderer(opts)
			doc := p.Parse([]byte(text))
			return string(markdown.Render(doc, renderer))
		},
	})

	// Get the absolute path to the templates directory
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	templatesDir := filepath.Join(rootDir, "templates")

	// Load templates
	r.LoadHTMLGlob(filepath.Join(templatesDir, "*.html"))

	return r
}

func TestGetPosts(t *testing.T) {
	s := store.NewStore()
	router := setupRouter()
	router.GET("/posts", GetPosts(s))

	// Create a test post
	post := store.Post{
		Title:   "Test Post",
		Content: "Test Content",
		Author:  "Test Author",
		Slug:    "test-post",
	}
	s.Create(post)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts", nil)
	req.Header.Set("Accept", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var posts []store.Post
	if err := json.NewDecoder(w.Body).Decode(&posts); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}

	if posts[0].Title != "Test Post" {
		t.Errorf("Expected title 'Test Post', got '%s'", posts[0].Title)
	}
}

func TestCreatePost(t *testing.T) {
	s := store.NewStore()
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(s))

	post := store.Post{
		Title:   "New Post",
		Content: "New Content",
		Author:  "New Author",
	}
	jsonData, _ := json.Marshal(post)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	// simulate editor role
	req.Header.Set("X-User-Role", "editor")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var createdPost store.Post
	if err := json.NewDecoder(w.Body).Decode(&createdPost); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if createdPost.Title != "New Post" {
		t.Errorf("Expected title 'New Post', got '%s'", createdPost.Title)
	}
}

func TestGetPost(t *testing.T) {
	s := store.NewStore()
	router := setupRouter()
	router.GET("/posts/:slug", GetPost(s))

	// Create a test post
	post := store.Post{
		Title:   "Test Post",
		Content: "Test Content",
		Author:  "Test Author",
		Slug:    "test-post",
	}
	s.Create(post)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts/test-post", nil)
	req.Header.Set("Accept", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var retrievedPost store.Post
	if err := json.NewDecoder(w.Body).Decode(&retrievedPost); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if retrievedPost.Slug != "test-post" {
		t.Errorf("Expected slug 'test-post', got '%s'", retrievedPost.Slug)
	}
}

func TestUpdatePost(t *testing.T) {
	s := store.NewStore()
	router := setupRouter()
	router.PUT("/posts/:slug", middleware.RequireEditor(), UpdatePost(s))

	// Create a test post
	post := store.Post{
		Title:   "Original Post",
		Content: "Original Content",
		Author:  "Original Author",
		Slug:    "original-post",
	}
	s.Create(post)

	updatedPost := store.Post{
		Title:   "Updated Post",
		Content: "Updated Content",
		Author:  "Updated Author",
	}
	jsonData, _ := json.Marshal(updatedPost)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/posts/original-post", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	// simulate editor role
	req.Header.Set("X-User-Role", "editor")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var result store.Post
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if result.Title != "Updated Post" {
		t.Errorf("Expected title 'Updated Post', got '%s'", result.Title)
	}
}

func TestDeletePost(t *testing.T) {
	s := store.NewStore()
	router := setupRouter()
	router.DELETE("/posts/:slug", middleware.RequireEditor(), DeletePost(s))

	// Create a test post
	post := store.Post{
		Title:   "Test Post",
		Content: "Test Content",
		Author:  "Test Author",
		Slug:    "test-post",
	}
	s.Create(post)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/posts/test-post", nil)
	req.Header.Set("Accept", "application/json")
	// simulate editor role
	req.Header.Set("X-User-Role", "editor")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify the post was deleted
	_, exists := s.GetBySlug("test-post")
	if exists {
		t.Error("Post was not deleted")
	}
}

// TestCreatePostForbidden ensures a non-editor cannot create posts
func TestCreatePostForbidden(t *testing.T) {
	s := store.NewStore()
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(s))

	// Attempt create with reader role
	post := store.Post{Title: "No", Content: "Access", Author: "User"}
	jsonData, _ := json.Marshal(post)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-User-Role", "reader")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
	}
}

// TestUpdatePostForbidden ensures a non-editor cannot update posts
func TestUpdatePostForbidden(t *testing.T) {
	s := store.NewStore()
	// Seed a post
	s.Create(store.Post{Title: "Old", Content: "Content", Author: "User", Slug: "old"})

	router := setupRouter()
	router.PUT("/posts/:slug", middleware.RequireEditor(), UpdatePost(s))

	updated := store.Post{Title: "New", Content: "New", Author: "User"}
	jsonData, _ := json.Marshal(updated)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/posts/old", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-User-Role", "reader")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
	}
}

// TestDeletePostForbidden ensures a non-editor cannot delete posts
func TestDeletePostForbidden(t *testing.T) {
	s := store.NewStore()
	// Seed a post
	s.Create(store.Post{Title: "Old", Content: "Content", Author: "User", Slug: "old"})

	router := setupRouter()
	router.DELETE("/posts/:slug", middleware.RequireEditor(), DeletePost(s))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/posts/old", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-User-Role", "reader")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
	}
}

// TestNewPageForbidden ensures a non-editor cannot view the new-post page
func TestNewPageForbidden(t *testing.T) {
	router := setupRouter()
	router.GET("/posts/new", middleware.RequireEditor(), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts/new", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
	}
}

// TestEditPageForbidden ensures a non-editor cannot view the edit-post page
func TestEditPageForbidden(t *testing.T) {
	router := setupRouter()
	router.GET("/posts/:slug/edit", middleware.RequireEditor(), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts/test/edit", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
	}
}