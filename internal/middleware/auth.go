package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/seanankenbruck/blog/internal/auth"
	"github.com/seanankenbruck/blog/internal/domain"
)

// wantsJSON returns true if the request expects a JSON response by default or content type is JSON
func wantsJSON(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	ct := c.GetHeader("Content-Type")
	if accept == "" || strings.Contains(accept, "application/json") {
		return true
	}
	if strings.Contains(ct, "application/json") {
		return true
	}
	return false
}

// AuthMiddleware extracts and validates the JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Check Authorization header
		authHeader := c.GetHeader("Authorization")
		if parts := strings.Split(authHeader, " "); len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		}

		// If no token in header, check cookie
		if tokenString == "" {
			if cookie, err := c.Cookie("jwt"); err == nil {
				tokenString = cookie
			}
		}

		// No token, skip
		if tokenString == "" {
			c.Next()
			return
		}

		// Validate token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			if wantsJSON(c) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			} else {
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		// Store claims and user in context
		c.Set("claims", claims)
		c.Set("user", &domain.User{Username: claims.Username, Role: claims.Role})

		c.Next()
	}
}

// RequireEditor ensures the user has editor role
func RequireEditor() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			// Not authenticated
			if wantsJSON(c) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			} else {
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.Claims)
		if !ok || !auth.RequireRole(userClaims, domain.Editor) {
			// Not authorized
			if wantsJSON(c) {
				c.JSON(http.StatusForbidden, gin.H{"error": "editor access required"})
			} else {
				c.HTML(http.StatusForbidden, "403.html", gin.H{"Title": "Forbidden", "Year": 0})
			}
			c.Abort()
			return
		}

		c.Next()
	}
}