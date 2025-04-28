package main

import (
	"html/template"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/seanankenbruck/blog/internal/handler"
	"github.com/seanankenbruck/blog/internal/middleware"
	"github.com/seanankenbruck/blog/internal/store"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Initialize stores
	postStore := store.NewStore()
	userStore := store.NewUserStore()

	// Add some test posts
	postStore.Create(store.Post{
		Title:   "Welcome to My Blog",
		Content: "This is my first blog post. More content coming soon!",
		Author:  "Admin",
		Slug:    "welcome-to-my-blog",
	})

	postStore.Create(store.Post{
		Title:   "Getting Started with Go",
		Content: "Go is a powerful programming language that makes it easy to build simple, reliable, and efficient software.",
		Author:  "Admin",
		Slug:    "getting-started-with-go",
	})

	// Set up template engine with custom functions
	r.SetFuncMap(map[string]interface{}{
		"safeHTML": func(text string) template.HTML {
			extensions := parser.CommonExtensions | parser.AutoHeadingIDs
			p := parser.NewWithExtensions(extensions)
			htmlFlags := html.CommonFlags | html.HrefTargetBlank
			opts := html.RendererOptions{Flags: htmlFlags}
			renderer := html.NewRenderer(opts)
			doc := p.Parse([]byte(text))
			return template.HTML(markdown.Render(doc, renderer))
		},
	})

	// Load templates
	r.LoadHTMLGlob("templates/*.html")

	// Static files
	r.Static("/static", "./static")

	// Add preview endpoint
	r.POST("/preview", func(c *gin.Context) {
		var req struct {
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create a new parser with common extensions
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs
		p := parser.NewWithExtensions(extensions)

		// Create HTML renderer with common extensions
		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		renderer := html.NewRenderer(opts)

		// Convert markdown to HTML
		doc := p.Parse([]byte(req.Content))
		html := markdown.Render(doc, renderer)

		c.String(http.StatusOK, string(html))
	})

	// Auth routes (before middleware)
	r.GET("/login", handler.LoginPage())
	r.POST("/login", handler.Login(userStore))
	r.GET("/logout", handler.Logout())

	// Apply authentication middleware
	r.Use(middleware.AuthMiddleware())

	// Public routes
	r.GET("/", handler.HomePage(postStore))
	r.GET("/about", handler.AboutPage())
	r.GET("/portfolio", handler.PortfolioPage())
	r.GET("/contact", handler.ContactPage())
	r.GET("/posts", handler.GetPosts(postStore))
	r.GET("/posts/:slug", handler.GetPost(postStore))

	// Protected routes
	authorized := r.Group("/")
	authorized.Use(middleware.RequireEditor())
	{
		authorized.GET("/posts/new", handler.NewPostPage())
		authorized.POST("/posts", handler.CreatePost(postStore))
		authorized.GET("/posts/:slug/edit", handler.EditPostPage(postStore))
		authorized.PUT("/posts/:slug", handler.UpdatePost(postStore))
		authorized.DELETE("/posts/:slug", handler.DeletePost(postStore))
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}