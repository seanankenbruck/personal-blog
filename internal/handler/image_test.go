package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/seanankenbruck/blog/internal/domain"
	"github.com/seanankenbruck/blog/internal/middleware"
	"github.com/seanankenbruck/blog/internal/repository"
	"github.com/seanankenbruck/blog/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestImageUpload(t *testing.T) {
	// Setup
	postRepo := repository.NewMemoryPostRepository()
	postSvc := service.NewPostService(postRepo)
	handler := NewPostHandler(postSvc)

	// Create a test router
	router := gin.Default()
	router.Use(middleware.AuthMiddleware())
	router.POST("/upload", middleware.RequireEditor(), handler.UploadImage)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "image-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create uploads directory if it doesn't exist
	uploadsDir := filepath.Join("static", "uploads")
	err = os.MkdirAll(uploadsDir, 0755)
	assert.NoError(t, err)

	// Register cleanup for the uploads directory
	t.Cleanup(func() {
		// Clean up the static/uploads directory and its parent if empty
		if err := os.RemoveAll(uploadsDir); err != nil {
			t.Logf("Warning: failed to clean up uploads directory %s: %v", uploadsDir, err)
		}
		// Try to remove the static directory if it's empty
		if err := os.Remove(filepath.Dir(uploadsDir)); err != nil && !os.IsNotExist(err) {
			t.Logf("Note: could not remove static directory (it might not be empty): %v", err)
		}
	})

	// Create a minimal valid JPEG file
	// JPEG file format starts with FF D8 FF and ends with FF D9
	jpegData := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, // JPEG SOI and APP0 marker
		0x00, 0x10, // APP0 segment length
		0x4A, 0x46, 0x49, 0x46, 0x00, // JFIF identifier
		0x01, 0x01, // JFIF version
		0x00, // units
		0x00, 0x01, 0x00, 0x01, // density
		0x00, 0x00, // thumbnail
		0xFF, 0xD9, // JPEG EOI marker
	}

	imagePath := filepath.Join(tempDir, "test.jpg")
	err = os.WriteFile(imagePath, jpegData, 0644)
	assert.NoError(t, err)

	// Create a multipart form with the test image
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create the file part with Content-Type header
	part, err := writer.CreatePart(map[string][]string{
		"Content-Type": {"image/jpeg"},
		"Content-Disposition": {`form-data; name="image"; filename="test.jpg"`},
	})
	assert.NoError(t, err)

	file, err := os.Open(imagePath)
	assert.NoError(t, err)
	_, err = io.Copy(part, file)
	assert.NoError(t, err)
	file.Close()
	writer.Close()

	// Create request
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "url")
	assert.True(t, strings.HasPrefix(response["url"], "/static/uploads/"))
}

func TestImageUploadInvalidFile(t *testing.T) {
	// Setup
	postRepo := repository.NewMemoryPostRepository()
	postSvc := service.NewPostService(postRepo)
	handler := NewPostHandler(postSvc)

	// Create a test router
	router := gin.Default()
	router.Use(middleware.AuthMiddleware())
	router.POST("/upload", middleware.RequireEditor(), handler.UploadImage)

	// Create a multipart form with invalid file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test.txt")
	assert.NoError(t, err)
	part.Write([]byte("not an image"))
	writer.Close()

	// Create request
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+GetAuthToken("editor", domain.Editor))
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMarkdownPreviewWithImage(t *testing.T) {
	// Setup
	postRepo := repository.NewMemoryPostRepository()
	postSvc := service.NewPostService(postRepo)
	handler := NewPostHandler(postSvc)

	// Create a test router
	router := gin.Default()
	router.POST("/preview", handler.PreviewMarkdown())

	// Create test markdown with image
	markdown := `# Test Post
![Test Image](/static/uploads/test.jpg)`

	// Create request
	req := httptest.NewRequest("POST", "/preview", strings.NewReader(markdown))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "<img")
	assert.Contains(t, w.Body.String(), "/static/uploads/test.jpg")
}

func TestImageUploadUnauthorized(t *testing.T) {
	// Setup
	postRepo := repository.NewMemoryPostRepository()
	postSvc := service.NewPostService(postRepo)
	handler := NewPostHandler(postSvc)

	// Create a test router with auth middleware
	router := gin.Default()
	router.Use(middleware.AuthMiddleware())
	router.POST("/upload", middleware.RequireEditor(), handler.UploadImage)

	// Create a test image file
	imagePath := filepath.Join("testdata", "test.jpg")
	err := os.MkdirAll("testdata", 0755)
	assert.NoError(t, err)

	file, err := os.Create(imagePath)
	assert.NoError(t, err)
	file.Write([]byte("fake image content"))
	file.Close()
	defer os.RemoveAll("testdata")

	// Create a multipart form with the test image
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test.jpg")
	assert.NoError(t, err)

	file, err = os.Open(imagePath)
	assert.NoError(t, err)
	_, err = io.Copy(part, file)
	assert.NoError(t, err)
	file.Close()
	writer.Close()

	// Create request without auth token
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}