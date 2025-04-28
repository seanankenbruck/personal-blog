package handler

import (
	"net/http"
	"time"
	// "strconv"

	//"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/seanankenbruck/blog/internal/domain"
	"github.com/seanankenbruck/blog/internal/store"
	"log"
	"github.com/seanankenbruck/blog/internal/auth"
)

func GetPosts(s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		posts := s.GetAll()

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
			"Year": time.Now().Year(),
			"Posts": posts,
			"User":  user,
		})
	}
}

func CreatePost(s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var post store.Post
		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		createdPost := s.Create(post)
		c.JSON(http.StatusCreated, createdPost)
	}
}

func GetPost(s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		post, exists := s.GetBySlug(slug)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

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
			"Title": post.Title,
			"Year": time.Now().Year(),
			"Post": post,
			"User": user,
		})
	}
}

func UpdatePost(s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		var post store.Post
		if err := c.ShouldBindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updated, exists := s.UpdateBySlug(slug, post)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

		c.JSON(http.StatusOK, updated)
	}
}

func DeletePost(s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if !s.DeleteBySlug(slug) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func HomePage(s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		posts := s.GetAll()
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
func Login(userStore *store.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		user, ok := userStore.Authenticate(username, password)
		if !ok {
			// Check if client expects JSON
			if c.GetHeader("Accept") == "application/json" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			} else {
				c.HTML(http.StatusUnauthorized, "login.html", gin.H{
					"Title": "Login",
					"Year":  time.Now().Year(),
					"Error": "Invalid username or password",
				})
			}
			return
		}

		// Generate JWT token
		token, err := auth.GenerateToken(user.Username, user.Role)
		if err != nil {
			if c.GetHeader("Accept") == "application/json" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			} else {
				c.HTML(http.StatusInternalServerError, "login.html", gin.H{
					"Title": "Login",
					"Year":  time.Now().Year(),
					"Error": "An error occurred during login",
				})
			}
			return
		}

		// Check if client expects JSON
		if c.GetHeader("Accept") == "application/json" {
			c.JSON(http.StatusOK, gin.H{
				"token": token,
				"user": gin.H{
					"username": user.Username,
					"role":     user.Role,
				},
			})
		} else {
			// Set the token in a cookie for browser clients
			c.SetCookie("jwt", token, 86400, "/", "", false, true)
			c.Redirect(http.StatusFound, "/")
		}
	}
}

// Logout clears the JWT cookie
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clear the JWT cookie
		c.SetCookie("jwt", "", -1, "/", "", false, true)

		// If client expects JSON, return JSON response
		if c.GetHeader("Accept") == "application/json" {
			c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
		} else {
			// Otherwise redirect to login page
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
func EditPostPage(s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		post, exists := s.GetBySlug(slug)
		if !exists {
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
