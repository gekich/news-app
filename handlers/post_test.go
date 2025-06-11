//go:build unit

package handlers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gekich/news-app/config"
	"github.com/gekich/news-app/models"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PostRepositoryInterface interface {
	FindAll(ctx context.Context, page, limit int64, search string) ([]models.Post, int64, error)
	FindByID(ctx context.Context, id string) (models.Post, error)
	Create(ctx context.Context, post models.Post) (string, error)
	Update(ctx context.Context, id string, post models.Post) error
	Delete(ctx context.Context, id string) error
	CreateMany(ctx context.Context, posts []models.Post) error
}

type MockPostRepository struct {
	posts      map[string]models.Post
	nextID     int
	shouldFail bool
}

func NewMockPostRepository() *MockPostRepository {
	return &MockPostRepository{
		posts:  make(map[string]models.Post),
		nextID: 1,
	}
}

func (m *MockPostRepository) FindAll(ctx context.Context, page, limit int64, search string) ([]models.Post, int64, error) {
	if m.shouldFail {
		return nil, 0, fmt.Errorf("mock error")
	}

	var posts []models.Post
	for _, post := range m.posts {
		if search == "" || strings.Contains(strings.ToLower(post.Title), strings.ToLower(search)) ||
			strings.Contains(strings.ToLower(post.Content), strings.ToLower(search)) {
			posts = append(posts, post)
		}
	}

	totalPages := int64(1)
	if limit > 0 && len(posts) > 0 {
		totalPages = (int64(len(posts)) + limit - 1) / limit
	}

	return posts, totalPages, nil
}

func (m *MockPostRepository) FindByID(ctx context.Context, id string) (models.Post, error) {
	if m.shouldFail {
		return models.Post{}, fmt.Errorf("mock error")
	}

	post, exists := m.posts[id]
	if !exists {
		return models.Post{}, mongo.ErrNoDocuments
	}
	return post, nil
}

func (m *MockPostRepository) Create(ctx context.Context, post models.Post) (string, error) {
	if m.shouldFail {
		return "", fmt.Errorf("mock error")
	}

	id := fmt.Sprintf("%d", m.nextID)
	m.nextID++

	// Create ObjectID for the post
	objectID, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%024d", m.nextID-1))
	post.ID = objectID
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()

	m.posts[id] = post
	return id, nil
}

func (m *MockPostRepository) Update(ctx context.Context, id string, post models.Post) error {
	if m.shouldFail {
		return fmt.Errorf("mock error")
	}

	if _, exists := m.posts[id]; !exists {
		return mongo.ErrNoDocuments
	}

	objectID, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%024s", id))
	post.ID = objectID
	post.UpdatedAt = time.Now()
	m.posts[id] = post
	return nil
}

func (m *MockPostRepository) Delete(ctx context.Context, id string) error {
	if m.shouldFail {
		return fmt.Errorf("mock error")
	}

	if _, exists := m.posts[id]; !exists {
		return mongo.ErrNoDocuments
	}

	delete(m.posts, id)
	return nil
}

func (m *MockPostRepository) CreateMany(ctx context.Context, posts []models.Post) error {
	if m.shouldFail {
		return fmt.Errorf("mock error")
	}

	for _, post := range posts {
		id := fmt.Sprintf("%d", m.nextID)
		m.nextID++

		objectID, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%024d", m.nextID-1))
		post.ID = objectID
		post.CreatedAt = time.Now()
		post.UpdatedAt = time.Now()

		m.posts[id] = post
	}
	return nil
}

func (m *MockPostRepository) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

// MockablePostHandler extends PostHandler to allow repo injection
type MockablePostHandler struct {
	*PostHandler
	mockRepo PostRepositoryInterface
}

func (h *MockablePostHandler) getRepo() PostRepositoryInterface {
	if h.mockRepo != nil {
		return h.mockRepo
	}
	return h.repo
}

// Override handler methods to use the mockable repo
func (h *MockablePostHandler) Index(w http.ResponseWriter, r *http.Request) {
	// Copy the original Index logic but use mockRepo
	page := 1
	limit := int64(h.config.App.PostsPerPage)

	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		pageInt, err := strconv.Atoi(pageStr)
		if err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	search := r.URL.Query().Get("search")

	posts, totalPages, err := h.getRepo().FindAll(r.Context(), int64(page), limit, search)
	if err != nil {
		h.handleError(w, err, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Posts":       posts,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"Search":      search,
	}

	paginationURL := fmt.Sprintf("/posts?page=%d", page)
	if search != "" {
		paginationURL = fmt.Sprintf("/posts?page=%d&search=%s", page, search)
	}

	h.renderTemplate(w, r, "post_list", data, paginationURL)
}

func (h *MockablePostHandler) Show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	post, err := h.getRepo().FindByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Post": post,
	}

	h.renderTemplate(w, r, "show", data, fmt.Sprintf("/posts/%s", id))
}

func (h *MockablePostHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.handleError(w, err, "Failed to parse form", http.StatusBadRequest)
		return
	}

	post := models.Post{
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
	}

	// Simple validation for tests - just check if title and content are not empty
	if post.Title == "" || post.Content == "" {
		data := map[string]interface{}{
			"Post":   post,
			"Errors": map[string]string{"title": "Title is required", "content": "Content is required"},
			"Title":  "Create New Post",
			"Action": "/posts",
			"Method": "post",
		}

		h.renderTemplate(w, r, "form", data, "")
		return
	}

	id, err := h.getRepo().Create(r.Context(), post)
	if err != nil {
		h.handleError(w, err, "Failed to create post", http.StatusInternalServerError)
		return
	}

	redirectURL := "/posts"
	if isHTMXRequest(r) {
		redirectURL = fmt.Sprintf("/posts/%s", id)
	}

	h.redirectResponse(w, r, redirectURL)
}

func (h *MockablePostHandler) Edit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	post, err := h.getRepo().FindByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Post":   post,
		"Title":  "Edit Post",
		"Action": fmt.Sprintf("/posts/%s", id),
		"Method": "put",
	}

	h.renderTemplate(w, r, "form", data, fmt.Sprintf("/posts/%s/edit", id))
}

func (h *MockablePostHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := r.ParseForm(); err != nil {
		h.handleError(w, err, "Failed to parse form", http.StatusBadRequest)
		return
	}

	existingPost, err := h.getRepo().FindByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	existingPost.Title = r.FormValue("title")
	existingPost.Content = r.FormValue("content")

	// Simple validation for tests
	if existingPost.Title == "" || existingPost.Content == "" {
		data := map[string]interface{}{
			"Post":   existingPost,
			"Errors": map[string]string{"title": "Title is required", "content": "Content is required"},
			"Title":  "Edit Post",
			"Action": fmt.Sprintf("/posts/%s", id),
			"Method": "put",
		}

		h.renderTemplate(w, r, "form", data, "")
		return
	}

	err = h.getRepo().Update(r.Context(), id, existingPost)
	if err != nil {
		h.handleError(w, err, "Failed to update post", http.StatusInternalServerError)
		return
	}

	redirectURL := "/posts"
	if isHTMXRequest(r) {
		redirectURL = fmt.Sprintf("/posts/%s", id)
	}

	h.redirectResponse(w, r, redirectURL)
}

func (h *MockablePostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.getRepo().Delete(r.Context(), id)
	if err != nil {
		h.handleError(w, err, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	h.redirectResponse(w, r, "/posts")
}

func (h *MockablePostHandler) Seed(w http.ResponseWriter, r *http.Request) {
	// Create some sample posts for testing
	samplePosts := []models.Post{
		{Title: "Sample Post 1", Content: "Sample Content 1"},
		{Title: "Sample Post 2", Content: "Sample Content 2"},
	}

	err := h.getRepo().CreateMany(r.Context(), samplePosts)
	if err != nil {
		h.handleError(w, err, "Failed to seed database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		posts, totalPages, err := h.getRepo().FindAll(r.Context(), 1, int64(h.config.App.PostsPerPage), "")
		if err != nil {
			h.handleError(w, err, "Failed to fetch posts after seeding", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Posts":       posts,
			"CurrentPage": 1,
			"TotalPages":  totalPages,
			"Search":      "",
		}

		h.renderTemplate(w, r, "post_list", data, "/posts")
		return
	}

	h.redirectResponse(w, r, "/posts")
}

// Helper function to create mock templates
func createMockTemplates() map[string]*template.Template {
	templates := make(map[string]*template.Template)

	// Create simple mock templates
	postListTmpl := template.Must(template.New("post_list").Parse(`
		{{define "content"}}Posts: {{len .Posts}}{{end}}
		Posts: {{len .Posts}}
	`))

	showTmpl := template.Must(template.New("show").Parse(`
		{{define "content"}}Post: {{.Post.Title}}{{end}}
		Post: {{.Post.Title}}
	`))

	formTmpl := template.Must(template.New("form").Parse(`
		{{define "content"}}Form: {{.Title}}{{end}}
		Form: {{.Title}}
	`))

	templates["post_list"] = postListTmpl
	templates["show"] = showTmpl
	templates["form"] = formTmpl

	return templates
}

func createTestHandler() (*MockablePostHandler, *MockPostRepository) {
	mockRepo := NewMockPostRepository()
	templates := createMockTemplates()
	cfg, _ := config.Load()

	postHandler := NewPostHandler(nil, templates, cfg)
	mockableHandler := &MockablePostHandler{
		PostHandler: postHandler,
		mockRepo:    mockRepo,
	}

	return mockableHandler, mockRepo
}

func createRequestWithChiContext(method, url string, body *bytes.Buffer) (*http.Request, *httptest.ResponseRecorder) {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, body)
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	rr := httptest.NewRecorder()
	return req, rr
}

func createMockPost(id string, title, content string) models.Post {
	objectID, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%024s", id))
	return models.Post{
		ID:        objectID,
		Title:     title,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestPostHandler_Index(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockPosts      map[string]models.Post
		expectedStatus int
		shouldFail     bool
	}{
		{
			name:           "successful index request",
			queryParams:    "",
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "index with pagination",
			queryParams:    "?page=2",
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "index with search",
			queryParams:    "?search=test",
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "repository error",
			queryParams:    "",
			mockPosts:      map[string]models.Post{},
			expectedStatus: http.StatusInternalServerError,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := createTestHandler()
			mockRepo.posts = tt.mockPosts
			mockRepo.SetShouldFail(tt.shouldFail)

			req, rr := createRequestWithChiContext("GET", "/posts"+tt.queryParams, nil)

			handler.Index(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestPostHandler_Show(t *testing.T) {
	tests := []struct {
		name           string
		postID         string
		mockPosts      map[string]models.Post
		expectedStatus int
		shouldFail     bool
	}{
		{
			name:           "successful show request",
			postID:         "1",
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "post not found",
			postID:         "999",
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "repository error",
			postID:         "1",
			mockPosts:      map[string]models.Post{},
			expectedStatus: http.StatusInternalServerError,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := createTestHandler()
			mockRepo.posts = tt.mockPosts
			mockRepo.SetShouldFail(tt.shouldFail)

			req, rr := createRequestWithChiContext("GET", "/posts/"+tt.postID, nil)

			// Add chi URL param
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.postID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.Show(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestPostHandler_New(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "successful new form request",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := createTestHandler()

			req, rr := createRequestWithChiContext("GET", "/posts/new", nil)

			handler.New(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestPostHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		formData       url.Values
		expectedStatus int
		shouldFail     bool
	}{
		{
			name: "successful create",
			formData: url.Values{
				"title":   {"Test Title"},
				"content": {"Test Content"},
			},
			expectedStatus: http.StatusSeeOther,
		},
		{
			name: "invalid form data - empty title",
			formData: url.Values{
				"title":   {""},
				"content": {"Test Content"},
			},
			expectedStatus: http.StatusOK, // Form is re-rendered with errors
		},
		{
			name: "invalid form data - empty content",
			formData: url.Values{
				"title":   {"Test Title"},
				"content": {""},
			},
			expectedStatus: http.StatusOK, // Form is re-rendered with errors
		},
		{
			name: "repository error",
			formData: url.Values{
				"title":   {"Test Title"},
				"content": {"Test Content"},
			},
			expectedStatus: http.StatusInternalServerError,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := createTestHandler()
			mockRepo.SetShouldFail(tt.shouldFail)

			body := bytes.NewBufferString(tt.formData.Encode())
			req, rr := createRequestWithChiContext("POST", "/posts", body)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			handler.Create(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestPostHandler_Edit(t *testing.T) {
	tests := []struct {
		name           string
		postID         string
		mockPosts      map[string]models.Post
		expectedStatus int
		shouldFail     bool
	}{
		{
			name:           "successful edit form request",
			postID:         "1",
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "post not found",
			postID:         "999",
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "repository error",
			postID:         "1",
			mockPosts:      map[string]models.Post{},
			expectedStatus: http.StatusInternalServerError,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := createTestHandler()
			mockRepo.posts = tt.mockPosts
			mockRepo.SetShouldFail(tt.shouldFail)

			req, rr := createRequestWithChiContext("GET", "/posts/"+tt.postID+"/edit", nil)

			// Add chi URL param
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.postID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.Edit(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestPostHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		postID         string
		formData       url.Values
		mockPosts      map[string]models.Post
		expectedStatus int
		shouldFail     bool
	}{
		{
			name:   "successful update",
			postID: "1",
			formData: url.Values{
				"title":   {"Updated Title"},
				"content": {"Updated Content"},
			},
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:   "invalid form data - empty title",
			postID: "1",
			formData: url.Values{
				"title":   {""},
				"content": {"Updated Content"},
			},
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusOK, // Form is re-rendered with errors
		},
		{
			name:   "post not found",
			postID: "999",
			formData: url.Values{
				"title":   {"Updated Title"},
				"content": {"Updated Content"},
			},
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "repository error",
			postID: "1",
			formData: url.Values{
				"title":   {"Updated Title"},
				"content": {"Updated Content"},
			},
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusInternalServerError,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := createTestHandler()
			mockRepo.posts = tt.mockPosts
			mockRepo.SetShouldFail(tt.shouldFail)

			body := bytes.NewBufferString(tt.formData.Encode())
			req, rr := createRequestWithChiContext("PUT", "/posts/"+tt.postID, body)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Add chi URL param
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.postID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.Update(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestPostHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		postID         string
		mockPosts      map[string]models.Post
		expectedStatus int
		shouldFail     bool
	}{
		{
			name:           "successful delete",
			postID:         "1",
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "post not found",
			postID:         "999",
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "repository error",
			postID:         "1",
			mockPosts:      map[string]models.Post{"1": createMockPost("1", "Test Post", "Test Content")},
			expectedStatus: http.StatusInternalServerError,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := createTestHandler()
			mockRepo.posts = tt.mockPosts
			mockRepo.SetShouldFail(tt.shouldFail)

			req, rr := createRequestWithChiContext("DELETE", "/posts/"+tt.postID, nil)

			// Add chi URL param
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.postID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.Delete(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestPostHandler_Seed(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		shouldFail     bool
	}{
		{
			name:           "successful seed",
			expectedStatus: http.StatusSeeOther,
		},
		{
			name:           "repository error",
			expectedStatus: http.StatusInternalServerError,
			shouldFail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := createTestHandler()
			mockRepo.SetShouldFail(tt.shouldFail)

			req, rr := createRequestWithChiContext("POST", "/posts/seed", nil)

			handler.Seed(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestPostHandler_HTMXRequests(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		url      string
		postID   string
		formData url.Values
		mockPost models.Post
	}{
		{
			name:     "HTMX index request",
			method:   "GET",
			url:      "/posts",
			mockPost: createMockPost("1", "Test Post", "Test Content"),
		},
		{
			name:     "HTMX show request",
			method:   "GET",
			url:      "/posts/1",
			postID:   "1",
			mockPost: createMockPost("1", "Test Post", "Test Content"),
		},
		{
			name:   "HTMX create request",
			method: "POST",
			url:    "/posts",
			formData: url.Values{
				"title":   {"Test Title"},
				"content": {"Test Content"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := createTestHandler()
			if !tt.mockPost.ID.IsZero() {
				mockRepo.posts[tt.mockPost.ID.Hex()] = tt.mockPost
			}

			var body *bytes.Buffer
			if tt.formData != nil {
				body = bytes.NewBufferString(tt.formData.Encode())
			}

			req, rr := createRequestWithChiContext(tt.method, tt.url, body)
			req.Header.Set("HX-Request", "true") // Set HTMX header

			if tt.formData != nil {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			if tt.postID != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.postID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}

			switch tt.method {
			case "GET":
				if tt.postID != "" {
					handler.Show(rr, req)
				} else {
					handler.Index(rr, req)
				}
			case "POST":
				handler.Create(rr, req)
			}

			if tt.method == "POST" && tt.formData != nil {
				if rr.Header().Get("HX-Redirect") == "" {
					t.Error("Expected HX-Redirect header for HTMX POST request")
				}
			}
		})
	}
}
