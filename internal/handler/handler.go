package handler

import (
	"net/http"
	"time"
	// "strconv"

	"github.com/gin-gonic/gin"
	"github.com/seanankenbruck/blog/internal/store"
	"log"
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
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title": "All Posts",
			"Year": time.Now().Year(),
			"Posts": posts,
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
		c.HTML(http.StatusOK, "post.html", gin.H{
			"Title": post.Title,
			"Year": time.Now().Year(),
			"Post": post,
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

func ContactPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "contact.html", gin.H{
			"Title": "Contact",
			"Year": time.Now().Year(),
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

		c.HTML(http.StatusOK, "contact.html", gin.H{
			"Title":   "Contact",
			"Year":    time.Now().Year(),
			"Success": "Thank you for your message! I will get back to you soon.",
		})
	}
}
