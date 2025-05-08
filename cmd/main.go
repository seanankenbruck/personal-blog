package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seanankenbruck/blog/internal/db"
	"github.com/seanankenbruck/blog/internal/handler"
	"github.com/seanankenbruck/blog/internal/middleware"
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

	// Initialize database connection
	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	postRepo := repository.NewPostgresPostRepository(db.DB)
	userRepo := repository.NewMemoryUserRepository()
	subscriberRepo := repository.NewMemorySubscriberRepository()

	// Initialize services
	postService := service.NewPostService(postRepo)
	userService := service.NewUserService(userRepo)
	subscriberService := service.NewSubscriberService(subscriberRepo)

	// Initialize handlers
	postHandler := handler.NewPostHandler(postService)
	userHandler := handler.NewUserHandler(userService)
	subscriberHandler := handler.NewSubscriberHandler(subscriberService)

	// Set up routes
	setupRoutes(r, postHandler, userHandler, subscriberHandler)

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(r *gin.Engine, postHandler *handler.PostHandler, userHandler *handler.UserHandler, subscriberHandler *handler.SubscriberHandler) {
	// Add context timeout middleware
	r.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	// Apply authentication middleware
	r.Use(middleware.AuthMiddleware())

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
		public.GET("/", postHandler.HomePage)
		public.GET("/posts", postHandler.GetPosts)
		public.GET("/posts/:slug", postHandler.GetPost)
		public.GET("/about", handler.AboutPage())
		public.GET("/portfolio", handler.PortfolioPage())
		public.GET("/contact", handler.ContactPage())
		public.POST("/contact", handler.SubmitContact())
		public.GET("/login", handler.LoginPage())
		public.POST("/login", userHandler.Login)
		public.GET("/logout", userHandler.Logout)
		public.POST("/preview", postHandler.PreviewMarkdown())
		public.POST("/subscribe", subscriberHandler.Subscribe)
		public.GET("/confirm", subscriberHandler.ConfirmSubscription)
		public.POST("/unsubscribe", subscriberHandler.Unsubscribe)
	}

	// Protected routes
	editor := r.Group("/")
	editor.Use(middleware.RequireEditor())
	{
		editor.GET("/posts/new", postHandler.NewPostPage)
		editor.POST("/posts", postHandler.CreatePost)
		editor.GET("/posts/:slug/edit", postHandler.EditPostPage)
		editor.PUT("/posts/:slug", postHandler.UpdatePost)
		editor.DELETE("/posts/:slug", postHandler.DeletePost)
		editor.POST("/upload", postHandler.UploadImage)
		editor.GET("/subscribers", subscriberHandler.ListSubscribers)
	}
}