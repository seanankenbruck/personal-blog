package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/seanankenbruck/blog/internal/domain"
)

var (
	// JWT key loaded from environment variable
	jwtKey []byte

	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrMissingJWTSecret = errors.New("JWT_SECRET environment variable is required")
)

// InitJWT initializes the JWT key from environment variable
func InitJWT() error {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return ErrMissingJWTSecret
	}
	jwtKey = []byte(secret)
	return nil
}

type Claims struct {
	Username string      `json:"username"`
	Role     domain.Role `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(username string, role domain.Role) (string, error) {
	if jwtKey == nil {
		return "", ErrMissingJWTSecret
	}
	
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateToken validates the JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	if jwtKey == nil {
		return nil, ErrMissingJWTSecret
	}
	
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RequireRole checks if the token has the required role
func RequireRole(claims *Claims, requiredRole domain.Role) bool {
	return claims.Role == requiredRole
}