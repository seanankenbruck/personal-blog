package handler

import (
	"net/http"
	"strings"
	"time"
	"context"
	// "strconv"

	//"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/seanankenbruck/blog/internal/domain"
	"github.com/seanankenbruck/blog/internal/auth"
	"log"
	//"encoding/json"
	"html/template"
	"path/filepath"
	"runtime"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// SetupTemplates configures the template engine with custom functions
func SetupTemplates(r *gin.Engine) error {
	// Set up template engine with custom functions
	r.SetFuncMap(template.FuncMap{
		"safeHTML": func(text string) template.HTML {
			return template.HTML(text)
		},
	})

	// Get the absolute path to the templates directory
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	templatesDir := filepath.Join(rootDir, "templates")

	// Load templates
	r.LoadHTMLGlob(filepath.Join(templatesDir, "*.html"))
	return nil
}

func GetPosts(svc domain.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		posts, err := svc.GetAllPosts(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Render markdown content for each post
		for i := range posts {
			extensions := parser.CommonExtensions | parser.AutoHeadingIDs
			p := parser.NewWithExtensions(extensions)
			doc := p.Parse([]byte(posts[i].Content))

			htmlFlags := html.CommonFlags | html.HrefTargetBlank
			opts := html.RendererOptions{Flags: htmlFlags}
			renderer := html.NewRenderer(opts)

			renderedContent := markdown.Render(doc, renderer)
			posts[i].Content = string(renderedContent)
		}

		// Check the Accept header to determine response format
		accept := c.GetHeader("Accept")
		if accept == "application/json" {
			c.JSON(http.StatusOK, posts)
			return
		}

		// Default to HTML response
		userVal, _ := c.Get("user")
		user, _ := userVal.(*domain.User)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title": "All Posts",
			"Year":  time.Now().Year(),
			"Posts": posts,
			"User":  user,
		})
	}
}

func CreatePost(svc domain.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var post domain.Post
		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate the post
		if err := post.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set creation and update dates
		now := time.Now()
		post.CreatedAt = now
		post.UpdatedAt = now

		// Create the post (slug will be generated in the service)
		if err := svc.CreatePost(c, &post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return the created post with the generated slug
		c.JSON(http.StatusCreated, post)
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

		// Render markdown content to HTML
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs
		p := parser.NewWithExtensions(extensions)
		doc := p.Parse([]byte(post.Content))

		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		renderer := html.NewRenderer(opts)

		renderedContent := markdown.Render(doc, renderer)
		post.Content = string(renderedContent)

		// Check the Accept header to determine response format
		accept := c.GetHeader("Accept")
		if accept == "application/json" {
			c.JSON(http.StatusOK, post)
			return
		}

		// Default to HTML response
		userVal, _ := c.Get("user")
		user, _ := userVal.(*domain.User)
		c.HTML(http.StatusOK, "post.html", gin.H{
			"Post": post,
			"User": user,
		})
	}
}

func UpdatePost(svc domain.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		slug := c.Param("slug")

		var updateData domain.Post
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Retrieve existing
		post, err := svc.GetPostBySlug(ctx, slug)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if post == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

		// Apply updates
		post.Title = updateData.Title
		post.Content = updateData.Content
		if err := svc.UpdatePost(ctx, post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, post)
	}
}

func DeletePost(svc domain.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		slug := c.Param("slug")
		post, err := svc.GetPostBySlug(ctx, slug)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if post == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		if err := svc.DeletePost(ctx, post.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
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
		userVal, _ := c.Get("user")
		user, _ := userVal.(*domain.User)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title": "Home",
			"Year": time.Now().Year(),
			"Posts": recentPosts,
			"User":  user,
		})
	}
}

func AboutPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, _ := c.Get("user")
		user, _ := userVal.(*domain.User)
		c.HTML(http.StatusOK, "about.html", gin.H{
			"Title": "About",
			"Year": time.Now().Year(),
			"User":  user,
		})
	}
}

func PortfolioPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, _ := c.Get("user")
		user, _ := userVal.(*domain.User)
		c.HTML(http.StatusOK, "portfolio.html", gin.H{
			"Title": "Portfolio",
			"Year": time.Now().Year(),
			"User":  user,
		})
	}
}

func ContactPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, _ := c.Get("user")
		user, _ := userVal.(*domain.User)
		c.HTML(http.StatusOK, "contact.html", gin.H{
			"Title": "Contact",
			"Year": time.Now().Year(),
			"User":  user,
		})
	}
}

// SubmitContact handles the contact form submission
func SubmitContact() gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.PostForm("email")
		job := c.PostForm("job")
		message := c.PostForm("message")
		// TODO: send email notification here
		log.Printf("Contact form submitted: email=%s, job=%s, message=%s", email, job, message)

		userVal, _ := c.Get("user")
		user, _ := userVal.(*domain.User)
		c.HTML(http.StatusOK, "contact.html", gin.H{
			"Title":   "Contact",
			"Year":    time.Now().Year(),
			"Success": "Thank you for your message! I will get back to you soon.",
			"User":    user,
		})
	}
}

// LoginPage renders the login form
func LoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Title": "Login",
			"Year":  time.Now().Year(),
		})
	}
}

// Login handles user authentication and returns a JWT token
func Login(userRepo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.HTML(http.StatusOK, "login.html", gin.H{})
			return
		}

		username := c.PostForm("username")
		password := c.PostForm("password")

		user, err := userRepo.Authenticate(c.Request.Context(), username, password)
		if err != nil {
			c.HTML(http.StatusUnauthorized, "login.html", gin.H{
				"Error": "Invalid username or password",
			})
			return
		}

		// Generate JWT token
		token, err := auth.GenerateToken(user.Username, user.Role)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "login.html", gin.H{
				"Error": "Failed to generate token",
			})
			return
		}

		// Set the token in a cookie for browser clients
		c.SetCookie("jwt", token, 86400, "/", "", false, true)
		c.Redirect(http.StatusFound, "/")
	}
}

// Logout clears the JWT cookie
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clear the JWT cookie
		c.SetCookie("jwt", "", -1, "/", "", false, true)

		// Determine response format
		accept := c.GetHeader("Accept")
		if accept == "" || strings.Contains(accept, "application/json") {
			c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
		} else {
			c.Redirect(http.StatusFound, "/login")
		}
	}
}

// NewPostPage renders the new post form
func NewPostPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, _ := c.Get("user")
		user, _ := userVal.(*domain.User)
		c.HTML(http.StatusOK, "new.html", gin.H{
			"Title": "Create New Post",
			"Year":  time.Now().Year(),
			"User":  user,
		})
	}
}

// EditPostPage renders the edit post form
func EditPostPage(svc domain.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		post, err := svc.GetPostBySlug(c, slug)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if post == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

		userVal, _ := c.Get("user")
		user, _ := userVal.(*domain.User)
		c.HTML(http.StatusOK, "edit.html", gin.H{
			"Title": "Edit Post",
			"Year":  time.Now().Year(),
			"Post":  post,
			"User":  user,
		})
	}
}

// PreviewMarkdown handles markdown preview requests
func PreviewMarkdown() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Create a new parser with common extensions
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs
		p := parser.NewWithExtensions(extensions)

		// Parse the markdown content
		doc := p.Parse([]byte(req.Content))

		// Create HTML renderer with common flags
		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		renderer := html.NewRenderer(opts)

		// Render the markdown to HTML
		html := markdown.Render(doc, renderer)

		// Return the HTML as a raw string
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{"html": template.HTML(html)})
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

func (h *PostHandler) NewPostPage(c *gin.Context) {
	NewPostPage()(c)
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	CreatePost(h.postService)(c)
}

func (h *PostHandler) EditPostPage(c *gin.Context) {
	EditPostPage(h.postService)(c)
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	UpdatePost(h.postService)(c)
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	DeletePost(h.postService)(c)
}

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService domain.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService domain.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Login(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "login.html", gin.H{})
		return
	}

	username := c.PostForm("username")
	password := c.PostForm("password")

	user, err := h.userService.AuthenticateUser(c.Request.Context(), username, password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Error": "Invalid username or password",
		})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.Username, user.Role)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"Error": "Failed to generate token",
		})
		return
	}

	// Set the token in a cookie for browser clients
	c.SetCookie("jwt", token, 86400, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}
