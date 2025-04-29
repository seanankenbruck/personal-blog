package handler

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/seanankenbruck/blog/internal/domain"
)

// GetAuthToken generates a JWT token for testing
func GetAuthToken(username string, role domain.Role) string {
	claims := jwt.MapClaims{
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("your-secret-key"))
	return tokenString
}