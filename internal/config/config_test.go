package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set up test environment variables
	os.Setenv("DB_HOST", "test_host")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("DB_NAME", "test_db")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("OTLP_ENDPOINT", "http://test:4318")

	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("OTLP_ENDPOINT")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test that environment variables are correctly loaded
	if config.DBHost != "test_host" {
		t.Errorf("DBHost = %v, want %v", config.DBHost, "test_host")
	}
	if config.DBPort != 5433 {
		t.Errorf("DBPort = %v, want %v", config.DBPort, 5433)
	}
	if config.DBUser != "test_user" {
		t.Errorf("DBUser = %v, want %v", config.DBUser, "test_user")
	}
	if config.DBPassword != "test_password" {
		t.Errorf("DBPassword = %v, want %v", config.DBPassword, "test_password")
	}
	if config.DBName != "test_db" {
		t.Errorf("DBName = %v, want %v", config.DBName, "test_db")
	}
	if config.ServerPort != "9090" {
		t.Errorf("ServerPort = %v, want %v", config.ServerPort, "9090")
	}
	if config.OTLPEndpoint != "http://test:4318" {
		t.Errorf("OTLPEndpoint = %v, want %v", config.OTLPEndpoint, "http://test:4318")
	}
}

func TestLoad_Defaults(t *testing.T) {
	// Ensure no environment variables are set
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("OTLP_ENDPOINT")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test that default values are used
	if config.DBHost != "localhost" {
		t.Errorf("DBHost = %v, want %v", config.DBHost, "localhost")
	}
	if config.DBPort != 5432 {
		t.Errorf("DBPort = %v, want %v", config.DBPort, 5432)
	}
	if config.DBUser != "postgres" {
		t.Errorf("DBUser = %v, want %v", config.DBUser, "postgres")
	}
	if config.DBPassword != "postgres" {
		t.Errorf("DBPassword = %v, want %v", config.DBPassword, "postgres")
	}
	if config.DBName != "blog" {
		t.Errorf("DBName = %v, want %v", config.DBName, "blog")
	}
	if config.ServerPort != "8080" {
		t.Errorf("ServerPort = %v, want %v", config.ServerPort, "8080")
	}
	if config.OTLPEndpoint != "http://localhost:4318" {
		t.Errorf("OTLPEndpoint = %v, want %v", config.OTLPEndpoint, "http://localhost:4318")
	}
}