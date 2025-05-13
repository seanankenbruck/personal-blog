package db

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/seanankenbruck/blog/internal/domain"
)

// DB is the global database connection
var DB *gorm.DB

// Init loads the .env file (if present), connects to the Postgres database,
// and runs AutoMigrate on all domain models.
func Init() error {
	// Load .env for local development (optional)
	_ = godotenv.Load()

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	DB = db

	// AutoMigrate domain models
	if err := DB.AutoMigrate(&domain.User{}, &domain.Post{}, &domain.Subscriber{}); err != nil {
		return fmt.Errorf("failed to auto-migrate models: %w", err)
	}

	return nil
}