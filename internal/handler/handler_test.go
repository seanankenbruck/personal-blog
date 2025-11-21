package handler

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/seanankenbruck/blog/internal/content"
	"github.com/seanankenbruck/blog/internal/repository"
	"github.com/seanankenbruck/blog/internal/service"
	"github.com/stretchr/testify/assert"
)

func setupTestEnvironment(t *testing.T) (*gin.Engine, string) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create temporary content directory with test posts
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

	// Create repository and service
	repo := repository.NewFilePostRepository()
	svc := service.NewPostService(repo)

	// Setup Gin router
	router := gin.New()

	// Setup templates - need to set FuncMap before loading templates
	router.SetFuncMap(template.FuncMap{
		"safeHTML": func(text string) template.HTML {
			return template.HTML(text)
		},
	})

	// Create temporary template directory
	templateDir := t.TempDir()

	// Create test templates matching actual template structure
	indexTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Blog Posts</title>
</head>
<body>
    <h1>Blog Posts</h1>
    <div class="posts">
        {{range .Posts}}
        <div class="post">
            <h2>{{.Title}}</h2>
            <p>{{.Description}}</p>
            <a href="/posts/{{.Slug}}">Read More</a>
        </div>
        {{end}}
    </div>
</body>
</html>`

	aboutTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>About</title>
</head>
<body>
    <h1>About</h1>
    <p>About page content</p>
</body>
</html>`

	portfolioTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Portfolio</title>
</head>
<body>
    <h1>Portfolio</h1>
    <p>Portfolio page content</p>
</body>
</html>`

	postTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.Post.Title}}</title>
</head>
<body>
    <article class="post">
        <h1>{{.Post.Title}}</h1>
        <h3>{{.Post.Description}}</h3>
        <div>{{ .Post.HTMLContent | safeHTML }}</div>
    </article>
</body>
</html>`

	errorTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>404 Not Found</title>
</head>
<body>
    <h1>404</h1>
    <p>Post not found</p>
</body>
</html>`

	// Write template files
	templates := map[string]string{
		"index.html":     indexTemplate,
		"about.html":     aboutTemplate,
		"portfolio.html": portfolioTemplate,
		"post.html":      postTemplate,
		"404.html":       errorTemplate,
	}

	for name, tmpl := range templates {
		if err := os.WriteFile(filepath.Join(templateDir, name), []byte(tmpl), 0644); err != nil {
			t.Fatalf("Failed to create template file %s: %v", name, err)
		}
	}

	// Load templates
	router.LoadHTMLGlob(filepath.Join(templateDir, "*.html"))

	// Setup routes
	router.GET("/", HomePage(svc))
	router.GET("/portfolio", PortfolioPage())
	router.GET("/posts/:slug", GetPost(svc))

	return router, templateDir
}

func TestHomePageIntegration(t *testing.T) {
	router, _ := setupTestEnvironment(t)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Blog Posts")
	assert.Contains(t, w.Body.String(), "Test Post")
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
}

func TestPortfolioPageIntegration(t *testing.T) {
	router, _ := setupTestEnvironment(t)

	req, err := http.NewRequest(http.MethodGet, "/portfolio", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Portfolio")
	assert.Contains(t, w.Body.String(), "Portfolio page content")
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
}

func TestGetPostHTMLIntegration(t *testing.T) {
	router, _ := setupTestEnvironment(t)

	t.Run("Existing post returns HTML", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/posts/test-post", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Test Post")
		assert.Contains(t, w.Body.String(), "Test description")
		assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	})

	t.Run("Non-existent post returns 404", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/posts/non-existent", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "404")
		assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	})
}

func TestSetupTemplatesIntegration(t *testing.T) {
	t.Skip("SetupTemplates requires actual templates directory - tested through other integration tests")
}

func TestGetPosts(t *testing.T) {
	router, _ := setupTestEnvironment(t)

	req, err := http.NewRequest(http.MethodGet, "/posts", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	w := httptest.NewRecorder()
	router.GET("/posts", GetPosts(service.NewPostService(repository.NewFilePostRepository())))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Post")
}

func TestGetPost(t *testing.T) {
	router, _ := setupTestEnvironment(t)

	t.Run("Get existing post returns JSON", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/posts/test-post", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Accept", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Test Post")
	})
}

func TestPostHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create temporary content directory with test posts
	tempDir := t.TempDir()
	testPost := `---
title: "Handler Test Post"
slug: "handler-test-post"
date: 2024-01-15T10:00:00Z
description: "Test description"
published: true
---

Test content.`

	if err := os.WriteFile(filepath.Join(tempDir, "handler-test.md"), []byte(testPost), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	content.Init(tempDir, false)
	if err := content.LoadPosts(); err != nil {
		t.Fatalf("LoadPosts() failed: %v", err)
	}

	repo := repository.NewFilePostRepository()
	svc := service.NewPostService(repo)
	handler := NewPostHandler(svc)

	assert.NotNil(t, handler)

	// Setup router with PostHandler methods
	router := gin.New()
	router.SetFuncMap(template.FuncMap{
		"safeHTML": func(text string) template.HTML {
			return template.HTML(text)
		},
	})

	templateDir := t.TempDir()
	indexTemplate := `<!DOCTYPE html><html><head><title>Test</title></head><body>{{range .Posts}}<div>{{.Title}}</div>{{end}}</body></html>`
	postTemplate := `<!DOCTYPE html><html><head><title>{{.Post.Title}}</title></head><body><h1>{{.Post.Title}}</h1></body></html>`
	errorTemplate := `<!DOCTYPE html><html><head><title>404</title></head><body><h1>404</h1></body></html>`

	os.WriteFile(filepath.Join(templateDir, "index.html"), []byte(indexTemplate), 0644)
	os.WriteFile(filepath.Join(templateDir, "post.html"), []byte(postTemplate), 0644)
	os.WriteFile(filepath.Join(templateDir, "404.html"), []byte(errorTemplate), 0644)

	router.LoadHTMLGlob(filepath.Join(templateDir, "*.html"))

	router.GET("/", handler.HomePage)
	router.GET("/posts", handler.GetPosts)
	router.GET("/posts/:slug", handler.GetPost)
	router.POST("/preview", handler.PreviewMarkdown())

	t.Run("PostHandler.HomePage", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("PostHandler.GetPosts", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/posts", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("PostHandler.GetPost", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/posts/handler-test-post", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("PreviewMarkdown", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/preview", strings.NewReader("# Hello World"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "<h1")
		assert.Contains(t, w.Body.String(), "Hello World")
	})
}

func TestGetPostsHTMLResponse(t *testing.T) {
	router, _ := setupTestEnvironment(t)
	router.GET("/posts", GetPosts(service.NewPostService(repository.NewFilePostRepository())))

	// Test HTML response (no Accept header)
	req, _ := http.NewRequest(http.MethodGet, "/posts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
}
