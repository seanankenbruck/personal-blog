package middleware

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/seanankenbruck/blog/internal/domain"
    "github.com/gin-contrib/sessions"
)

// AuthMiddleware populates a User from session if it exists
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Try to get user info from session
        session := sessions.Default(c)
        username := session.Get("username")
        role := session.Get("role")

        // If no session exists, user is not logged in
        if username == nil || role == nil {
            c.Set("user", nil)
            c.Next()
            return
        }

        // Create user from session data
        user := &domain.User{
            Username: username.(string),
            Role:     domain.Role(role.(string)),
        }
        c.Set("user", user)
        c.Next()
    }
}

// RequireEditor redirects to login if not logged in as editor
func RequireEditor() gin.HandlerFunc {
    return func(c *gin.Context) {
        user, exists := c.Get("user")
        if !exists || user == nil {
            c.Redirect(http.StatusFound, "/login")
            c.Abort()
            return
        }

        if u, ok := user.(*domain.User); !ok || u.Role != domain.Editor {
            c.HTML(http.StatusForbidden, "403.html", gin.H{
                "Title": "Forbidden",
                "Year":  time.Now().Year(),
            })
            c.Abort()
            return
        }
        c.Next()
    }
}