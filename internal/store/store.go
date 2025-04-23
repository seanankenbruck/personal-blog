package store

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Store struct {
	posts  []Post
	nextID int64
	mu     sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		posts:  make([]Post, 0),
		nextID: 1,
	}
}

func generateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove all non-alphanumeric characters except hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}

func (s *Store) Create(post Post) Post {
	s.mu.Lock()
	defer s.mu.Unlock()

	post.ID = s.nextID
	s.nextID++
	post.Slug = generateSlug(post.Title)
	post.CreatedAt = time.Now()
	post.UpdatedAt = post.CreatedAt

	s.posts = append(s.posts, post)
	return post
}

func (s *Store) Get(id int64) (Post, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, post := range s.posts {
		if post.ID == id {
			return post, true
		}
	}
	return Post{}, false
}

func (s *Store) GetBySlug(slug string) (Post, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, post := range s.posts {
		if post.Slug == slug {
			return post, true
		}
	}
	return Post{}, false
}

func (s *Store) Update(id int64, updated Post) (Post, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, post := range s.posts {
		if post.ID == id {
			updated.ID = id
			updated.Slug = generateSlug(updated.Title)
			updated.CreatedAt = post.CreatedAt
			updated.UpdatedAt = time.Now()
			s.posts[i] = updated
			return updated, true
		}
	}
	return Post{}, false
}

func (s *Store) DeleteBySlug(slug string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, post := range s.posts {
		if post.Slug == slug {
			s.posts = append(s.posts[:i], s.posts[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) GetAll() []Post {
	s.mu.RLock()
	defer s.mu.RUnlock()

	posts := make([]Post, len(s.posts))
	copy(posts, s.posts)
	return posts
}

func (s *Store) UpdateBySlug(slug string, updated Post) (Post, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, post := range s.posts {
		if post.Slug == slug {
			updated.ID = post.ID
			updated.Slug = generateSlug(updated.Title)
			updated.CreatedAt = post.CreatedAt
			updated.UpdatedAt = time.Now()
			s.posts[i] = updated
			return updated, true
		}
	}
	return Post{}, false
}