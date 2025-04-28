package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seanankenbruck/blog/internal/auth"
	"github.com/seanankenbruck/blog/internal/domain"
)

// AuthMiddleware extracts and validates the JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// First check Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// If no token in header, check cookie
		if tokenString == "" {
			cookie, err := c.Cookie("jwt")
			if err == nil {
				tokenString = cookie
			}
		}

		// If no token found anywhere, continue
		if tokenString == "" {
			c.Next()
			return
		}

		// Validate the token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			respondWithError(c, http.StatusUnauthorized, err.Error())
			return
		}

		// Store claims in context for later use
		c.Set("claims", claims)

		// Also set user in context for templates
		c.Set("user", &domain.User{
			Username: claims.Username,
			Role:     claims.Role,
		})

		c.Next()
	}
}

// RequireEditor ensures the user has editor role
func RequireEditor() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			// If Accept header is application/json, return JSON response
			if c.GetHeader("Accept") == "application/json" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			} else {
				// Otherwise redirect to login page
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.Claims)
		if !ok || !auth.RequireRole(userClaims, domain.Editor) {
			// If Accept header is application/json, return JSON response
			if c.GetHeader("Accept") == "application/json" {
				c.JSON(http.StatusForbidden, gin.H{"error": "editor access required"})
			} else {
				// Otherwise render the 403 page
				c.HTML(http.StatusForbidden, "403.html", gin.H{
					"Title": "Forbidden",
					"Year":  time.Now().Year(),
				})
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper function to handle error responses
func respondWithError(c *gin.Context, status int, message string) {
	if c.GetHeader("Accept") == "application/json" {
		c.JSON(status, gin.H{"error": message})
	} else {
		if status == http.StatusUnauthorized {
			c.Redirect(http.StatusFound, "/login")
		} else {
			c.HTML(status, "403.html", gin.H{
				"Title": "Forbidden",
				"Year":  time.Now().Year(),
			})
		}
	}
	c.Abort()
}