package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/seanankenbruck/blog/internal/handler"
	"github.com/seanankenbruck/blog/internal/store"
)

func main() {
	// Initialize the store
	s := store.NewStore()

	// Add some test posts
	s.Create(store.Post{
		Title:   "Welcome to My Blog",
		Content: "This is my first blog post. More content coming soon!",
		Author:  "Admin",
		Slug:    "welcome-to-my-blog",
	})

	s.Create(store.Post{
		Title:   "Getting Started with Go",
		Content: "Go is a powerful programming language that makes it easy to build simple, reliable, and efficient software.",
		Author:  "Admin",
		Slug:    "getting-started-with-go",
	})

	// Create a new Gin router
	r := gin.Default()

	// Set up template engine with markdown function
	r.SetFuncMap(template.FuncMap{
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

	// Load HTML templates
	r.LoadHTMLGlob("templates/*.html")

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

	// Register routes
	r.GET("/", handler.HomePage(s))
	r.GET("/posts", handler.GetPosts(s))
	r.GET("/posts/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "new.html", gin.H{
			"Title": "Create New Post",
			"Year": time.Now().Year(),
		})
	})
	r.POST("/posts", handler.CreatePost(s))
	r.GET("/posts/:slug", handler.GetPost(s))
	r.GET("/posts/:slug/edit", func(c *gin.Context) {
		slug := c.Param("slug")
		post, exists := s.GetBySlug(slug)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.HTML(http.StatusOK, "edit.html", gin.H{
			"Title": "Edit Post",
			"Year": time.Now().Year(),
			"Post": post,
		})
	})
	r.PUT("/posts/:slug", handler.UpdatePost(s))
	r.DELETE("/posts/:slug", handler.DeletePost(s))

	// Start the server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server starting on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}