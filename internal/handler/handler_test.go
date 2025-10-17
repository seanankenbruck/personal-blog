package handler

import (
	"log"
	"testing"

	"github.com/seanankenbruck/blog/internal/repository"
	"github.com/seanankenbruck/blog/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestPostHandlerCreation(t *testing.T) {
	log.Println("Testing PostHandler creation...")
	
	// Create a test repository and service
	repo := repository.NewFilePostRepository()
	svc := service.NewPostService(repo)
	postHandler := NewPostHandler(svc)

	// Verify the handler was created successfully
	assert.NotNil(t, postHandler)
	assert.NotNil(t, postHandler.postService)
	
	log.Println("PostHandler creation test completed")
}

func TestServiceCreation(t *testing.T) {
	log.Println("Testing service creation...")
	
	// Create a test repository and service
	repo := repository.NewFilePostRepository()
	svc := service.NewPostService(repo)

	// Verify the service was created successfully
	assert.NotNil(t, svc)
	
	log.Println("Service creation test completed")
}

func TestRepositoryCreation(t *testing.T) {
	log.Println("Testing repository creation...")
	
	// Create a test repository
	repo := repository.NewFilePostRepository()

	// Verify the repository was created successfully
	assert.NotNil(t, repo)
	
	log.Println("Repository creation test completed")
}