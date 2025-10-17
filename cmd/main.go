package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seanankenbruck/blog/internal/content"
	"github.com/seanankenbruck/blog/internal/handler"
	"github.com/seanankenbruck/blog/internal/repository"
	"github.com/seanankenbruck/blog/internal/service"
)

func main() {
	// Initialize router
	r := gin.Default()

	// Set up static file serving
	r.Static("/static", "./static")

	// Set up templates
	if err := handler.SetupTemplates(r); err != nil {
		log.Fatalf("Failed to set up templates: %v", err)
	}

	// Determine content directory path
	contentDir := os.Getenv("CONTENT_DIR")
	if contentDir == "" {
		// Try container path first, then fall back to local development path
		if _, err := os.Stat("/content/posts"); err == nil {
			contentDir = "/content/posts"
		} else {
			contentDir = "content/posts"
		}
	}

	// Determine if we're in development mode
	isDev := gin.Mode() == gin.DebugMode

	// Initialize content loader
	content.Init(contentDir, isDev)
	if err := content.LoadPosts(); err != nil {
		log.Fatalf("Failed to load posts: %v", err)
	}
	log.Printf("Loaded posts from %s", contentDir)


	// Initialize repositories
	postRepo := repository.NewFilePostRepository()

	// Initialize services
	postService := service.NewPostService(postRepo)

	// Initialize handlers
	postHandler := handler.NewPostHandler(postService)

	// Set up routes
	setupRoutes(r, postHandler)

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(r *gin.Engine, postHandler *handler.PostHandler) {
	// Add context timeout middleware
	r.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})


	// Custom 404 handler for nonexistent routes
	r.NoRoute(func(c *gin.Context) {
		accept := c.GetHeader("Accept")
		if accept == "" || strings.Contains(accept, "application/json") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		} else {
			c.HTML(http.StatusNotFound, "404.html", gin.H{"Title": "404 - Page Not Found", "Year": time.Now().Year()})
		}
	})

	// Public routes
	public := r.Group("/")
	{
		public.GET("/", handler.PortfolioPage())
		public.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})
		public.GET("/posts", postHandler.GetPosts)
		public.GET("/posts/:slug", postHandler.GetPost)
		public.GET("/portfolio", handler.PortfolioPage())
		public.POST("/preview", postHandler.PreviewMarkdown())
	}

}