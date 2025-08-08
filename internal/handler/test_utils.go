package handler

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/seanankenbruck/blog/internal/domain"
)

// GetAuthToken generates a JWT token for testing
func GetAuthToken(username string, role domain.Role) string {
	// Get JWT secret from environment, fallback to test secret for tests
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-jwt-secret-key-for-testing-only"
	}
	
	claims := jwt.MapClaims{
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtSecret))
	return tokenString
}