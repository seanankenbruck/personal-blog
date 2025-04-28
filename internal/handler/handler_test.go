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
	"github.com/seanankenbruck/blog/internal/domain"
	"github.com/seanankenbruck/blog/internal/store"
	"github.com/seanankenbruck/blog/internal/middleware"
	"github.com/seanankenbruck/blog/internal/auth"
)

// setupRouter creates a test router with all middleware and templates configured
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Apply authentication middleware
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

// getAuthToken is a helper function to generate JWT tokens for testing
func getAuthToken(username string, role domain.Role) string {
	token, _ := auth.GenerateToken(username, role)
	return token
}

// getGinContext is a helper function to create a Gin context for testing
func getGinContext(r *gin.Engine, w *httptest.ResponseRecorder, req *http.Request) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c
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

	// Test JSON response
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

	// Test HTML response
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/posts", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCreatePost(t *testing.T) {
	s := store.NewStore()
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(s))

	// Create test data
	post := store.Post{
		Title:   "New Post",
		Content: "New Content",
		Author:  "New Author",
	}
	jsonData, _ := json.Marshal(post)

	// Test without token - should fail
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Test with editor token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+getAuthToken("editor@blog.com", domain.Editor))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestLogin(t *testing.T) {
	userStore := store.NewUserStore()
	router := setupRouter()
	router.POST("/login", Login(userStore))

	// Test successful login
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	req.PostForm = map[string][]string{
		"username": {"editor@blog.com"},
		"password": {"editor123"},
	}
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response struct {
		Token string `json:"token"`
		User  struct {
			Username string      `json:"username"`
			Role     domain.Role `json:"role"`
		} `json:"user"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token in response")
	}
}

func TestLogout(t *testing.T) {
	router := setupRouter()
	router.GET("/logout", Logout())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/logout", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if response.Message != "logged out successfully" {
		t.Errorf("Expected message 'logged out successfully', got '%s'", response.Message)
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

	// Test with editor token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/posts/original-post", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+getAuthToken("editor@blog.com", domain.Editor))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
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

	// Test with editor token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/posts/test-post", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+getAuthToken("editor@blog.com", domain.Editor))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, w.Code)
	}
}

// TestCreatePostForbidden ensures a non-editor cannot create posts
func TestCreatePostForbidden(t *testing.T) {
	s := store.NewStore()
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(s))

	// Attempt create with reader token
	post := store.Post{Title: "No", Content: "Access", Author: "User"}
	jsonData, _ := json.Marshal(post)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+getAuthToken("reader@blog.com", domain.Reader))
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
	req.Header.Set("Authorization", "Bearer "+getAuthToken("reader@blog.com", domain.Reader))
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
	req.Header.Set("Authorization", "Bearer "+getAuthToken("reader@blog.com", domain.Reader))
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
	req.Header.Set("Authorization", "Bearer "+getAuthToken("reader@blog.com", domain.Reader))
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
	req.Header.Set("Authorization", "Bearer "+getAuthToken("reader@blog.com", domain.Reader))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
	}
}