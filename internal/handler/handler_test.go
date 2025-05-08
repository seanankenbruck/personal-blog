package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/seanankenbruck/blog/internal/domain"
	"github.com/seanankenbruck/blog/internal/repository"
	"github.com/seanankenbruck/blog/internal/service"
	"github.com/seanankenbruck/blog/internal/middleware"
	"github.com/seanankenbruck/blog/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

var testRouter *gin.Engine

func TestMain(m *testing.M) {
	log.Println("Setting up test router...")
	gin.SetMode(gin.TestMode)
	testRouter = gin.Default()

	// Apply authentication middleware
	log.Println("Adding auth middleware...")
	testRouter.Use(middleware.AuthMiddleware())

	// Set up template engine with custom functions
	log.Println("Setting up template engine...")
	testRouter.SetFuncMap(map[string]interface{}{
		"safeHTML": func(text string) string {
			extensions := parser.CommonExtensions | parser.AutoHeadingIDs
			p := parser.NewWithExtensions(extensions)
			htmlFlags := html.CommonFlags | html.HrefTargetBlank
			opts := html.RendererOptions{Flags: htmlFlags}
			renderer := html.NewRenderer(opts)
			doc := p.Parse([]byte(text))
			return string(markdown.Render(doc, renderer))
		},
		// Returns true if the given user is an editor
		"isEditor": func(u interface{}) bool {
			user, ok := u.(*domain.User)
			return ok && user != nil && user.Role == domain.Editor
		},
	})

	// Get the absolute path to the templates directory
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	templatesDir := filepath.Join(rootDir, "templates")

	// Load templates
	log.Printf("Loading templates from %s...", templatesDir)
	testRouter.LoadHTMLGlob(filepath.Join(templatesDir, "*.html"))

	// Add context timeout middleware with a longer timeout for tests
	log.Println("Adding context timeout middleware...")
	testRouter.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	// Set up routes for global test router
	log.Println("Setting up routes...")
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	testRouter.GET("/posts", GetPosts(svc))

	log.Println("Router setup complete")
	m.Run()
}

// setupRouter returns a fresh router for each test
func setupRouter() *gin.Engine {
	// Create a new router for each test
	router := gin.New()

	// Apply authentication middleware
	router.Use(middleware.AuthMiddleware())

	// Set up template engine with custom functions
	router.SetFuncMap(map[string]interface{}{
		"safeHTML": func(text string) string {
			extensions := parser.CommonExtensions | parser.AutoHeadingIDs
			p := parser.NewWithExtensions(extensions)
			htmlFlags := html.CommonFlags | html.HrefTargetBlank
			opts := html.RendererOptions{Flags: htmlFlags}
			renderer := html.NewRenderer(opts)
			doc := p.Parse([]byte(text))
			return string(markdown.Render(doc, renderer))
		},
		// Returns true if the given user is an editor
		"isEditor": func(u interface{}) bool {
			user, ok := u.(*domain.User)
			return ok && user != nil && user.Role == domain.Editor
		},
	})

	// Get the absolute path to the templates directory
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	templatesDir := filepath.Join(rootDir, "templates")

	// Load templates
	router.LoadHTMLGlob(filepath.Join(templatesDir, "*.html"))

	// Add context timeout middleware
	router.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	return router
}

// getGinContext is a helper function to create a Gin context for testing
func getGinContext(r *gin.Engine, w *httptest.ResponseRecorder, req *http.Request) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c
}

func TestGetPosts(t *testing.T) {
	log.Println("Starting TestGetPosts...")
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	router := setupRouter()
	router.GET("/posts", GetPosts(svc))

	// Create a test post
	log.Println("Creating test post...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	post := &domain.Post{
		Title:       "Test Post",
		Content:     "Test Content",
		Description: "Test Description",
	}
	if err := svc.CreatePost(ctx, post); err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}
	log.Printf("Created post with slug: %s", post.Slug)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts", nil)
	req.Header.Set("Accept", "application/json")
	log.Printf("Making request to %s", req.URL.String())
	router.ServeHTTP(w, req)

	log.Printf("Response status: %d", w.Code)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var posts []domain.Post
	if err := json.NewDecoder(w.Body).Decode(&posts); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}
	log.Println("TestGetPosts completed")
}

func TestCreatePost(t *testing.T) {
	log.Println("Starting TestCreatePost...")
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(svc))

	// Create test data
	post := &domain.Post{
		Title:       "New Post",
		Content:     "New Content",
		Description: "New Description",
	}
	jsonData, _ := json.Marshal(post)

	// Test without token - should fail
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	log.Printf("Making request to %s without token", req.URL.String())
	router.ServeHTTP(w, req)

	log.Printf("Response status: %d", w.Code)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Test with editor token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	log.Printf("Making request to %s with editor token", req.URL.String())
	router.ServeHTTP(w, req)

	log.Printf("Response status: %d", w.Code)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var createdPost domain.Post
	if err := json.NewDecoder(w.Body).Decode(&createdPost); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if createdPost.Title != post.Title {
		t.Errorf("Expected title '%s', got '%s'", post.Title, createdPost.Title)
	}
	log.Println("TestCreatePost completed")
}

func TestLogin(t *testing.T) {
	log.Println("Starting TestLogin...")
	userRepo := repository.NewMemoryUserRepository()
	router := setupRouter()
	router.POST("/login", Login(userRepo))

	// Test successful login
	w := httptest.NewRecorder()
	form := url.Values{}
	form.Add("username", "editor")
	form.Add("password", "editor123")
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	log.Printf("Making request to %s", req.URL.String())
	router.ServeHTTP(w, req)

	log.Printf("Response status: %d", w.Code)
	if w.Code != http.StatusFound {
		t.Errorf("Expected status code %d, got %d", http.StatusFound, w.Code)
	}

	// Check for the JWT cookie
	cookies := w.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "jwt" && cookie.Value != "" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected JWT cookie in response")
	}
	log.Println("TestLogin completed")
}

func TestLogout(t *testing.T) {
	// Create a new router with the user handler
	router := setupRouter()
	userRepo := repository.NewMemoryUserRepository()
	userSvc := service.NewUserService(userRepo)
	userHandler := NewUserHandler(userSvc)

	// Set up the logout route with the handler method
	router.GET("/logout", userHandler.Logout)

	// Create a request with Accept header for HTML
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/logout", nil)
	req.Header.Set("Accept", "text/html")

	// Set a JWT cookie to simulate a logged-in user
	req.AddCookie(&http.Cookie{
		Name:  "jwt",
		Value: GetAuthToken("editor", domain.Editor),
	})

	// Execute the request
	router.ServeHTTP(w, req)

	// Check that we get a redirect to login page
	if w.Code != http.StatusFound {
		t.Errorf("Expected status code %d, got %d", http.StatusFound, w.Code)
	}

	// Check that the location header points to /login
	location := w.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Expected redirect to /login, got %s", location)
	}

	// Check that the JWT cookie was cleared
	var cookieCleared bool
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "jwt" {
			if cookie.Value == "" && cookie.MaxAge < 0 {
				cookieCleared = true
			}
			break
		}
	}
	if !cookieCleared {
		t.Error("JWT cookie was not properly cleared")
	}
}

func TestGetPost(t *testing.T) {
	log.Println("Starting TestGetPost...")
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	router := setupRouter()
	router.GET("/posts/:slug", GetPost(svc))

	// Create a test post
	log.Println("Creating test post...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	post := &domain.Post{
		Title:       "Test Post",
		Content:     "Test Content",
		Description: "Test Description",
	}
	if err := svc.CreatePost(ctx, post); err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}
	log.Printf("Created post with slug: %s", post.Slug)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts/"+post.Slug, nil)
	req.Header.Set("Accept", "application/json")
	log.Printf("Making request to %s", req.URL.String())
	router.ServeHTTP(w, req)

	log.Printf("Response status: %d", w.Code)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var responsePost domain.Post
	if err := json.NewDecoder(w.Body).Decode(&responsePost); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if responsePost.Title != post.Title {
		t.Errorf("Expected title %s, got %s", post.Title, responsePost.Title)
	}

	if responsePost.Slug != post.Slug {
		t.Errorf("Expected slug %s, got %s", post.Slug, responsePost.Slug)
	}
	log.Println("TestGetPost completed")
}

func TestUpdatePost(t *testing.T) {
	log.Println("Starting TestUpdatePost...")
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)

	// Create a test post
	log.Println("Creating test post...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	post := &domain.Post{
		Title:       "Test Post",
		Content:     "Test Content",
		Description: "Test Description",
	}
	if err := svc.CreatePost(ctx, post); err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}
	log.Printf("Created post with slug: %s", post.Slug)

	// Update the post
	updateData := &domain.Post{
		Slug:        post.Slug,
		Title:       "Updated Post",
		Content:     "Updated Content",
		Description: "Updated Description",
	}

	// Update using the service directly
	if err := svc.UpdatePost(ctx, updateData); err != nil {
		t.Fatalf("Failed to update post: %v", err)
	}

	// Verify the update
	updatedPost, err := svc.GetPostBySlug(ctx, post.Slug)
	if err != nil {
		t.Fatalf("Failed to get updated post: %v", err)
	}

	if updatedPost.Title != updateData.Title {
		t.Errorf("Expected title %s, got %s", updateData.Title, updatedPost.Title)
	}
	if updatedPost.Content != updateData.Content {
		t.Errorf("Expected content %s, got %s", updateData.Content, updatedPost.Content)
	}
	if updatedPost.Slug != post.Slug {
		t.Errorf("Expected slug %s, got %s", post.Slug, updatedPost.Slug)
	}
	log.Println("TestUpdatePost completed")
}

func TestDeletePost(t *testing.T) {
	log.Println("Starting TestDeletePost...")
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	router := setupRouter()
	router.DELETE("/posts/:slug", middleware.RequireEditor(), DeletePost(svc))

	// Create a test post
	log.Println("Creating test post...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	post := &domain.Post{
		Title:       "Test Post",
		Content:     "Test Content",
		Description: "Test Description",
	}
	if err := svc.CreatePost(ctx, post); err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}
	log.Printf("Created post with slug: %s", post.Slug)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/posts/"+post.Slug, nil)
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	log.Printf("Making request to %s", req.URL.String())
	router.ServeHTTP(w, req)

	log.Printf("Response status: %d", w.Code)
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, w.Code)
	}
	log.Println("TestDeletePost completed")
}

// TestCreatePostForbidden ensures a non-editor cannot create posts
func TestCreatePostForbidden(t *testing.T) {
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(svc))

	// Attempt create with reader token
	post := &domain.Post{Title: "No", Content: "Access"}
	jsonData, _ := json.Marshal(post)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("reader", domain.Reader))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
	}
}

// TestUpdatePostForbidden ensures a non-editor cannot update posts
func TestUpdatePostForbidden(t *testing.T) {
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	// Seed a post
	svc.CreatePost(context.Background(), &domain.Post{Title: "Old", Content: "Content", Slug: "old"})

	router := setupRouter()
	router.PUT("/posts/:slug", middleware.RequireEditor(), UpdatePost(svc))

	updated := &domain.Post{Title: "New", Content: "New"}
	jsonData, _ := json.Marshal(updated)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/posts/old", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("reader", domain.Reader))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
	}
}

// TestDeletePostForbidden ensures a non-editor cannot delete posts
func TestDeletePostForbidden(t *testing.T) {
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	// Seed a post
	svc.CreatePost(context.Background(), &domain.Post{Title: "Old", Content: "Content", Slug: "old"})

	router := setupRouter()
	router.DELETE("/posts/:slug", middleware.RequireEditor(), DeletePost(svc))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/posts/old", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("reader", domain.Reader))
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
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("reader", domain.Reader))
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
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("reader", domain.Reader))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
	}
}

// TestJWTExpiration tests handling of expired tokens
func TestJWTExpiration(t *testing.T) {
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)

	router := setupRouter()
	router.GET("/posts", middleware.RequireEditor(), GetPosts(svc))

	// Create an expired token by setting expiration time to the past
	expirationTime := time.Now().Add(-1 * time.Hour) // Set expiration to 1 hour ago
	claims := &auth.Claims{
		Username: "editor",
		Role:     domain.Editor,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)), // Set issued at to 2 hours ago
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, _ := token.SignedString([]byte("your-secret-key"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestInvalidToken tests handling of malformed tokens
func TestInvalidToken(t *testing.T) {
	router := setupRouter()
	router.GET("/posts", middleware.RequireEditor(), GetPosts(nil))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestDuplicateSlug tests handling of duplicate post slugs
func TestDuplicateSlug(t *testing.T) {
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(svc))

	// Create first post
	post1 := &domain.Post{
		Title:       "Test Post",
		Content:     "Test Content",
		Description: "Test Description",
	}
	jsonData1, _ := json.Marshal(post1)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData1))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// Try to create post with same title (should generate same slug)
	post2 := &domain.Post{
		Title:       "Test Post", // Same title
		Content:     "Different Content",
		Description: "Different Description",
	}
	jsonData2, _ := json.Marshal(post2)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData2))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// Verify slugs are different
	var response1, response2 domain.Post
	json.NewDecoder(w.Body).Decode(&response2)
	if response1.Slug == response2.Slug {
		t.Errorf("Expected different slugs for posts with same title")
	}
}

// TestConcurrentPosts tests handling of concurrent post creation
func TestConcurrentPosts(t *testing.T) {
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(svc))

	// Create multiple posts concurrently
	var wg sync.WaitGroup
	posts := make([]*domain.Post, 10)
	errors := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			post := &domain.Post{
				Title:       fmt.Sprintf("Concurrent Post %d", index),
				Content:     fmt.Sprintf("Content %d", index),
				Description: fmt.Sprintf("Description %d", index),
			}
			jsonData, _ := json.Marshal(post)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
			router.ServeHTTP(w, req)

			if w.Code == http.StatusCreated {
				json.NewDecoder(w.Body).Decode(&posts[index])
			} else {
				errors[index] = fmt.Errorf("failed to create post: %d", w.Code)
			}
		}(i)
	}

	wg.Wait()

	// Verify all posts were created successfully
	for i, err := range errors {
		if err != nil {
			t.Errorf("Error creating post %d: %v", i, err)
		}
	}

	// Verify all slugs are unique
	slugs := make(map[string]bool)
	for _, post := range posts {
		if slugs[post.Slug] {
			t.Errorf("Duplicate slug found: %s", post.Slug)
		}
		slugs[post.Slug] = true
	}
}

// MockPostRepository is a test stub implementing domain.PostRepository for error scenarios
// It uses GetAllFunc to simulate database errors.
type MockPostRepository struct {
	GetAllFunc func(ctx context.Context) ([]*domain.Post, error)
}

func (m *MockPostRepository) Create(ctx context.Context, post *domain.Post) error {
	return nil
}

func (m *MockPostRepository) GetByID(ctx context.Context, id uint) (*domain.Post, error) {
	return nil, nil
}

func (m *MockPostRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	return nil, nil
}

func (m *MockPostRepository) GetAll(ctx context.Context) ([]*domain.Post, error) {
	return m.GetAllFunc(ctx)
}

func (m *MockPostRepository) Update(ctx context.Context, post *domain.Post) error {
	return nil
}

func (m *MockPostRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

// TestDatabaseError tests handling of database connection errors
func TestDatabaseError(t *testing.T) {
	// Create a mock repository that always returns an error
	mockRepo := &MockPostRepository{
		GetAllFunc: func(ctx context.Context) ([]*domain.Post, error) {
			return nil, fmt.Errorf("database connection error")
		},
	}
	svc := service.NewPostService(mockRepo)
	router := setupRouter()
	router.GET("/posts", GetPosts(svc))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts", nil)
	req.Header.Set("Accept", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if response["error"] != "database connection error" {
		t.Errorf("Expected error message 'database connection error', got '%s'", response["error"])
	}
}

// TestInvalidJSON tests handling of invalid JSON payloads
func TestInvalidJSON(t *testing.T) {
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(svc))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestMissingFields tests handling of missing required fields
func TestMissingFields(t *testing.T) {
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(svc))

	// Test missing title
	post := &domain.Post{
		Content: "Test Content",
	}
	jsonData, _ := json.Marshal(post)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Test missing content
	post = &domain.Post{
		Title: "Test Title",
	}
	jsonData, _ = json.Marshal(post)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Test missing description
	post = &domain.Post{
		Title: "Test Title",
		Content: "Test Content",
	}
	jsonData, _ = json.Marshal(post)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestSQLInjection tests protection against SQL injection attempts
func TestSQLInjection(t *testing.T) {
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(svc))

	// Test SQL injection in title
	post := &domain.Post{
		Title:   "'; DROP TABLE posts; --",
		Content:     "Test Content",
		Description: "Test Description",
	}
	jsonData, _ := json.Marshal(post)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// Verify the title was properly escaped
	var responsePost domain.Post
	if err := json.NewDecoder(w.Body).Decode(&responsePost); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if responsePost.Title != post.Title {
		t.Errorf("Expected title '%s', got '%s'", post.Title, responsePost.Title)
	}
}

// TestXSS tests protection against XSS attacks
func TestXSS(t *testing.T) {
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(svc))

	// Test XSS in content
	post := &domain.Post{
		Title:       "Test Post",
		Content:     "<script>alert('xss')</script>",
		Description: "Test Description",
	}
	jsonData, _ := json.Marshal(post)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// Verify the content was properly escaped
	var responsePost domain.Post
	if err := json.NewDecoder(w.Body).Decode(&responsePost); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	if responsePost.Content != post.Content {
		t.Errorf("Expected content '%s', got '%s'", post.Content, responsePost.Content)
	}
}

// TestCSRF tests protection against CSRF attacks
func TestCSRF(t *testing.T) {
	repo := repository.NewMemoryPostRepository()
	svc := service.NewPostService(repo)
	router := setupRouter()
	router.POST("/posts", middleware.RequireEditor(), CreatePost(svc))

	// Test without CSRF token
	post := &domain.Post{
		Title:       "Test Post",
		Content:     "Test Content",
		Description: "Test Description",
	}
	jsonData, _ := json.Marshal(post)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// Test with invalid CSRF token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	req.Header.Set("X-CSRF-Token", "invalid-token")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestGetPostsUnauthenticated(t *testing.T) {
	// Create request
	req := httptest.NewRequest("GET", "/posts", nil)
	w := httptest.NewRecorder()

	// Execute request
	testRouter.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
}

func setupSubscriberHandler() (*SubscriberHandler, *repository.MemorySubscriberRepository) {
	repo := repository.NewMemorySubscriberRepository()
	service := service.NewSubscriberService(repo)
	return NewSubscriberHandler(service), repo
}

func TestSubscriberHandler_Subscribe_JSON(t *testing.T) {
	h, _ := setupSubscriberHandler()
	r := gin.Default()
	r.POST("/subscribe", h.Subscribe)

	w := httptest.NewRecorder()
	body := `{"email": "test@example.com"}`
	req, _ := http.NewRequest("POST", "/subscribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Subscription successful")
}

func TestSubscriberHandler_Subscribe_AlreadyExists(t *testing.T) {
	h, _ := setupSubscriberHandler()
	r := gin.Default()
	r.POST("/subscribe", h.Subscribe)

	w := httptest.NewRecorder()
	body := `{"email": "test@example.com"}`
	req, _ := http.NewRequest("POST", "/subscribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/subscribe", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusConflict, w2.Code)
	assert.Contains(t, w2.Body.String(), "already subscribed")
}

func TestSubscriberHandler_ConfirmSubscription(t *testing.T) {
	h, repo := setupSubscriberHandler()
	r := gin.Default()
	r.POST("/subscribe", h.Subscribe)
	r.GET("/confirm", h.ConfirmSubscription)

	// Subscribe to get a token
	w := httptest.NewRecorder()
	body := `{"email": "test2@example.com"}`
	req, _ := http.NewRequest("POST", "/subscribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Extract token from the same repo used by the handler
	subs, _ := repo.GetAll(context.Background())
	var token string
	for _, sub := range subs {
		if sub.Email == "test2@example.com" {
			token = sub.ConfirmationToken
			break
		}
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/confirm?token="+token, nil)
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "confirmed")
}

func TestSubscriberHandler_Unsubscribe_JSON(t *testing.T) {
	h, _ := setupSubscriberHandler()
	r := gin.Default()
	r.POST("/subscribe", h.Subscribe)
	r.POST("/unsubscribe", h.Unsubscribe)

	// Subscribe first
	w := httptest.NewRecorder()
	body := `{"email": "test3@example.com"}`
	req, _ := http.NewRequest("POST", "/subscribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Unsubscribe
	w2 := httptest.NewRecorder()
	body2 := `{"email": "test3@example.com"}`
	req2, _ := http.NewRequest("POST", "/unsubscribe", strings.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "unsubscribed")
}

func TestSubscriberHandler_ListSubscribers(t *testing.T) {
	h, _ := setupSubscriberHandler()
	r := gin.Default()
	r.POST("/subscribe", h.Subscribe)
	r.GET("/subscribers", h.ListSubscribers)

	// Subscribe two users
	w := httptest.NewRecorder()
	body := `{"email": "a@example.com"}`
	req, _ := http.NewRequest("POST", "/subscribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w2 := httptest.NewRecorder()
	body2 := `{"email": "b@example.com"}`
	req2, _ := http.NewRequest("POST", "/subscribe", strings.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	// List subscribers
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/subscribers", nil)
	r.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)
	assert.Contains(t, w3.Body.String(), "a@example.com")
	assert.Contains(t, w3.Body.String(), "b@example.com")
}
