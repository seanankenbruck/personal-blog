package handler

import (
	"net/http"
	"time"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/seanankenbruck/blog/internal/domain"
	"html/template"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"os"
	"io"
)

// SetupTemplates configures the template engine with custom functions
func SetupTemplates(r *gin.Engine) error {
	// Set up template engine with custom functions
	r.SetFuncMap(template.FuncMap{
		"safeHTML": func(text string) template.HTML {
			return template.HTML(text)
		},
	})

	// Try container path first, then fall back to local development path
	templatesPath := "/templates/*.html"
	if _, err := os.Stat("/templates"); os.IsNotExist(err) {
		// We're in local development, use relative path
		templatesPath = "templates/*.html"
	}

	// Load templates
	r.LoadHTMLGlob(templatesPath)
	return nil
}

func GetPosts(svc domain.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		posts, err := svc.GetAllPosts(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Content is already rendered to HTML by the file loader

		// Check the Accept header to determine response format
		accept := c.GetHeader("Accept")
		if accept == "application/json" {
			c.JSON(http.StatusOK, posts)
			return
		}

		// Default to HTML response
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title": "All Posts",
			"Year":  time.Now().Year(),
			"Posts": posts,
		})
	}
}


func GetPost(svc domain.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		post, err := svc.GetPostBySlug(ctx, slug)
		if err != nil {
			if err == domain.ErrPostNotFound {
				c.HTML(http.StatusNotFound, "404.html", nil)
			} else {
				c.HTML(http.StatusInternalServerError, "500.html", nil)
			}
			return
		}

		// Content is already rendered to HTML by the file loader

		// Check the Accept header to determine response format
		accept := c.GetHeader("Accept")
		if accept == "application/json" {
			c.JSON(http.StatusOK, post)
			return
		}

		// Default to HTML response
		c.HTML(http.StatusOK, "post.html", gin.H{
			"Post": post,
		})
	}
}



func HomePage(svc domain.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		posts, err := svc.GetAllPosts(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Get the 3 most recent posts
		recentPosts := posts
		if len(posts) > 3 {
			recentPosts = posts[len(posts)-3:]
		}

		// Default to HTML response
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title": "Home",
			"Year": time.Now().Year(),
			"Posts": recentPosts,
		})
	}
}

func AboutPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.html", gin.H{
			"Title": "About",
			"Year": time.Now().Year(),
		})
	}
}

func PortfolioPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "portfolio.html", gin.H{
			"Title": "Portfolio",
			"Year": time.Now().Year(),
		})
	}
}




// PreviewMarkdown returns a handler function that renders markdown to HTML
func (h *PostHandler) PreviewMarkdown() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the markdown from the request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Create a markdown parser with extensions
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.SuperSubscript
		p := parser.NewWithExtensions(extensions)

		// Parse the markdown
		doc := p.Parse(body)

		// Create a renderer with HTML options
		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		renderer := html.NewRenderer(opts)

		// Render to HTML
		html := markdown.Render(doc, renderer)

		// Set the content type and return the HTML
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, string(html))
	}
}

// PostHandler handles post-related HTTP requests
type PostHandler struct {
	postService domain.PostService
}

// NewPostHandler creates a new PostHandler
func NewPostHandler(postService domain.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

func (h *PostHandler) HomePage(c *gin.Context) {
	HomePage(h.postService)(c)
}

func (h *PostHandler) GetPosts(c *gin.Context) {
	GetPosts(h.postService)(c)
}

func (h *PostHandler) GetPost(c *gin.Context) {
	GetPost(h.postService)(c)
}





