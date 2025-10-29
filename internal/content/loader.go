package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v3"
)

// Post represents a blog post loaded from a markdown file
type Post struct {
	Title       string    `yaml:"title"`
	Slug        string    `yaml:"slug"`
	Date        time.Time `yaml:"date"`
	Tags        []string  `yaml:"tags"`
	Description string    `yaml:"description"`
	Published   bool      `yaml:"published"` // Controls whether post is visible
	Content     string    `yaml:"-"`         // Raw markdown content
	HTMLContent string    `yaml:"-"`         // Rendered HTML
}

// FrontMatter represents the YAML front matter in a markdown file
type FrontMatter struct {
	Title       string    `yaml:"title"`
	Slug        string    `yaml:"slug"`
	Date        time.Time `yaml:"date"`
	Tags        []string  `yaml:"tags"`
	Description string    `yaml:"description"`
	Published   bool      `yaml:"published"` // Controls whether post is visible
}

var (
	posts     []*Post
	postsMap  map[string]*Post
	isLoaded  bool
	isDev     bool
	contentDir string
)

// Init initializes the content loader with the content directory path
func Init(dir string, devMode bool) {
	contentDir = dir
	isDev = devMode
}

// LoadPosts loads all markdown posts from the content directory
func LoadPosts() error {
	if contentDir == "" {
		contentDir = "content/posts"
	}

	// Check if content directory exists
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return fmt.Errorf("content directory does not exist: %s", contentDir)
	}

	posts = make([]*Post, 0)
	postsMap = make(map[string]*Post)

	// Walk the content directory and load all .md files
	err := filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		// Load the post from the file
		post, err := loadPostFromFile(path)
		if err != nil {
			return fmt.Errorf("error loading %s: %w", path, err)
		}

		// Only include published posts
		if !post.Published {
			return nil // Skip unpublished posts in production
		}

		posts = append(posts, post)
		postsMap[post.Slug] = post

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to load posts: %w", err)
	}

	// Sort posts by date (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	isLoaded = true
	return nil
}

// loadPostFromFile loads a post from a markdown file with YAML front matter
func loadPostFromFile(path string) (*Post, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse front matter and content
	frontMatter, markdown, err := parseFrontMatter(string(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing front matter: %w", err)
	}

	// Render markdown to HTML
	htmlContent := renderMarkdown(markdown)

	post := &Post{
		Title:       frontMatter.Title,
		Slug:        frontMatter.Slug,
		Date:        frontMatter.Date,
		Tags:        frontMatter.Tags,
		Description: frontMatter.Description,
		Published:   frontMatter.Published,
		Content:     markdown,
		HTMLContent: htmlContent,
	}

	// If slug is empty, generate it from the filename
	if post.Slug == "" {
		post.Slug = generateSlugFromFilename(path)
	}

	return post, nil
}

// parseFrontMatter extracts YAML front matter and markdown content from a file
func parseFrontMatter(content string) (*FrontMatter, string, error) {
	// Split by front matter delimiter (---)
	parts := strings.Split(content, "---")
	if len(parts) < 3 {
		return nil, "", fmt.Errorf("invalid front matter format: expected at least 3 parts separated by '---'")
	}

	// Parse YAML front matter (parts[1])
	var fm FrontMatter
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return nil, "", fmt.Errorf("error parsing YAML: %w", err)
	}

	// The rest is the markdown content
	markdownContent := strings.Join(parts[2:], "---")
	markdownContent = strings.TrimSpace(markdownContent)

	return &fm, markdownContent, nil
}

// renderMarkdown converts markdown to HTML
func renderMarkdown(md string) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}

// generateSlugFromFilename extracts a slug from a filename
// Expected format: YYYY-MM-DD-slug.md or slug.md
func generateSlugFromFilename(path string) string {
	filename := filepath.Base(path)
	// Remove .md extension
	filename = strings.TrimSuffix(filename, ".md")
	// Remove date prefix if present (YYYY-MM-DD-)
	parts := strings.SplitN(filename, "-", 4)
	if len(parts) >= 4 {
		// Has date prefix, return the slug part
		return parts[3]
	}
	// No date prefix, return as-is
	return filename
}

// GetPostBySlug returns a post by its slug
func GetPostBySlug(slug string) (*Post, error) {
	if !isLoaded {
		if err := LoadPosts(); err != nil {
			return nil, err
		}
	}

	post, ok := postsMap[slug]
	if !ok {
		return nil, fmt.Errorf("post not found: %s", slug)
	}

	return post, nil
}

// GetAllPosts returns all loaded posts
func GetAllPosts() ([]*Post, error) {
	if !isLoaded {
		if err := LoadPosts(); err != nil {
			return nil, err
		}
	}

	return posts, nil
}

// GetRecentPosts returns the N most recent posts
func GetRecentPosts(limit int) ([]*Post, error) {
	if !isLoaded {
		if err := LoadPosts(); err != nil {
			return nil, err
		}
	}

	if limit > len(posts) {
		limit = len(posts)
	}

	return posts[:limit], nil
}

// Reload reloads all posts from disk (useful for hot-reload in development)
func Reload() error {
	isLoaded = false
	return LoadPosts()
}
