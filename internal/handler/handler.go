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
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"fmt"
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
		// Returns true if the given user is an editor
		"isEditor": func(user interface{}) bool {
			u, ok := user.(*domain.User)
			return ok && u != nil && u.Role == domain.Editor
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

		// Set cookie with session expiration (0), secure flag, and SameSite strict
		c.SetCookie("jwt", token, 0, "/", "", true, true)
		c.Redirect(http.StatusFound, "/")
	}
}

// Logout clears the JWT cookie and redirects to login page
func (h *UserHandler) Logout(c *gin.Context) {
	// Clear the JWT cookie by setting it to empty with immediate expiration
	// Use secure cookies only in production
	isSecure := gin.Mode() == gin.ReleaseMode
	c.SetCookie("jwt", "", -1, "/", "", isSecure, true)

	// Clear any user data from the context
	c.Set("user", nil)
	c.Set("claims", nil)

	// Always redirect to login page for web requests
	if strings.Contains(c.Request.Header.Get("Accept"), "text/html") {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}

	// Return JSON response for API requests
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
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

// UploadImage handles image upload requests
func (h *PostHandler) UploadImage(c *gin.Context) {
	// Get the file from the request
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}

	fileType := file.Header.Get("Content-Type")
	if !allowedTypes[fileType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPEG, PNG, and GIF are allowed"})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "static/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filepath := filepath.Join(uploadDir, filename)

	// Save the file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Return the URL
	c.JSON(http.StatusOK, gin.H{
		"url": fmt.Sprintf("/static/uploads/%s", filename),
	})
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

	// Set cookie with session expiration (0) and secure flag based on environment
	isSecure := gin.Mode() == gin.ReleaseMode
	c.SetCookie("jwt", token, 0, "/", "", isSecure, true)
	c.Redirect(http.StatusFound, "/")
}

// UploadImage handles image uploads for blog posts
func UploadImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the file from the request
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
			return
		}

		// Generate a unique filename
		filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), file.Filename)
		filepath := filepath.Join("static", "images", filename)

		// Save the file
		if err := c.SaveUploadedFile(file, filepath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}

		// Return the URL of the saved image
		imageURL := fmt.Sprintf("/static/images/%s", filename)
		c.JSON(http.StatusOK, gin.H{"url": imageURL})
	}
}

// SubscriberHandler handles subscriber-related HTTP requests
type SubscriberHandler struct {
	subscriberService domain.SubscriberService
}

// NewSubscriberHandler creates a new SubscriberHandler
func NewSubscriberHandler(subscriberService domain.SubscriberService) *SubscriberHandler {
	return &SubscriberHandler{subscriberService: subscriberService}
}

// Subscribe handles new subscription requests
func (h *SubscriberHandler) Subscribe(c *gin.Context) {
	ctx := c.Request.Context()
	var email string
	if c.ContentType() == "application/json" {
		var req struct{ Email string `json:"email"` }
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		email = req.Email
	} else {
		email = c.PostForm("email")
	}

	subscriber, err := h.subscriberService.Subscribe(ctx, email)
	if err != nil {
		accept := c.GetHeader("Accept")
		if err == domain.ErrSubscriberExists {
			if strings.Contains(accept, "text/html") || c.ContentType() == "application/x-www-form-urlencoded" {
				c.HTML(http.StatusConflict, "subscribe_exists.html", gin.H{"Email": email})
				return
			}
			c.JSON(http.StatusConflict, gin.H{"error": "You are already subscribed."})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	// TODO: Send confirmation email with subscriber.ConfirmationToken

	// Support both HTML and JSON responses
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "text/html") {
		c.HTML(http.StatusOK, "subscribe_success.html", gin.H{
			"Title": "Subscribed!",
			"Year": time.Now().Year(),
			"Email": subscriber.Email,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Subscription successful. Please check your email to confirm.", "email": subscriber.Email})
	}
}

// ConfirmSubscription handles subscription confirmation
func (h *SubscriberHandler) ConfirmSubscription(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing confirmation token"})
		return
	}
	ctx := c.Request.Context()
	err := h.subscriberService.ConfirmSubscription(ctx, token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "text/html") {
		c.HTML(http.StatusOK, "confirm_success.html", gin.H{
			"Title": "Subscription Confirmed",
			"Year": time.Now().Year(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed."})
	}
}

// Unsubscribe handles unsubscribe requests
func (h *SubscriberHandler) Unsubscribe(c *gin.Context) {
	ctx := c.Request.Context()
	var email string
	if c.ContentType() == "application/json" {
		var req struct{ Email string `json:"email"` }
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		email = req.Email
	} else {
		email = c.PostForm("email")
	}

	err := h.subscriberService.Unsubscribe(ctx, email)
	if err != nil {
		if err == domain.ErrSubscriberNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscriber not found."})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "text/html") {
		c.HTML(http.StatusOK, "unsubscribe_success.html", gin.H{
			"Title": "Unsubscribed",
			"Year": time.Now().Year(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "You have been unsubscribed."})
	}
}

// ListSubscribers handles requests to list all subscribers
func (h *SubscriberHandler) ListSubscribers(c *gin.Context) {
	ctx := c.Request.Context()
	subscribers, err := h.subscriberService.ListSubscribers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "text/html") {
		c.HTML(http.StatusOK, "subscribers.html", gin.H{
			"Title": "Subscribers",
			"Year": time.Now().Year(),
			"Subscribers": subscribers,
		})
	} else {
		c.JSON(http.StatusOK, subscribers)
	}
}
