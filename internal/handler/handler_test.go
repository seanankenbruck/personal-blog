package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/seanankenbruck/blog/internal/store"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

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
	router.POST("/posts", CreatePost(s))

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
	router.PUT("/posts/:slug", UpdatePost(s))

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
	router.DELETE("/posts/:slug", DeletePost(s))

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