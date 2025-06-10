package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gekich/news-app/config"
	"github.com/gekich/news-app/models"
	"github.com/gekich/news-app/repository"
	"github.com/gekich/news-app/seeder"
	"github.com/gekich/news-app/validation"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type PostHandler struct {
	repo   *repository.PostRepository
	tmpl   map[string]*template.Template
	config config.Config
}

func NewPostHandler(repo *repository.PostRepository, tmpl map[string]*template.Template, cfg config.Config) *PostHandler {
	return &PostHandler{
		repo:   repo,
		tmpl:   tmpl,
		config: cfg,
	}
}

// isHTMXRequest checks if the request is from HTMX
func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// renderTemplate renders the specified template with the given data
func (h *PostHandler) renderTemplate(w http.ResponseWriter, r *http.Request, templateName string, data map[string]interface{}, pushURL string) {
	isHTMX := isHTMXRequest(r)

	if isHTMX {
		if pushURL != "" {
			w.Header().Set("HX-Push-Url", pushURL)
		}
		if err := h.tmpl[templateName].ExecuteTemplate(w, "content", data); err != nil {
			http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if err := h.tmpl[templateName].Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// handleError sends an appropriate error response
func (h *PostHandler) handleError(w http.ResponseWriter, err error, message string, status int) {
	if err == mongo.ErrNoDocuments {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	http.Error(w, message, status)
}

// redirectResponse redirects the user to the specified URL
func (h *PostHandler) redirectResponse(w http.ResponseWriter, r *http.Request, url string) {
	if isHTMXRequest(r) {
		if url != "" {
			w.Header().Set("HX-Redirect", url)
		}
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *PostHandler) Index(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := int64(h.config.App.PostsPerPage)

	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		pageInt, err := strconv.Atoi(pageStr)
		if err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	posts, totalPages, err := h.repo.FindAll(r.Context(), int64(page), limit)
	if err != nil {
		h.handleError(w, err, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Posts":       posts,
		"CurrentPage": page,
		"TotalPages":  totalPages,
	}

	h.renderTemplate(w, r, "post_list", data, fmt.Sprintf("/posts?page=%d", page))
}

func (h *PostHandler) Show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	post, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Post": post,
	}

	h.renderTemplate(w, r, "show", data, fmt.Sprintf("/posts/%s", id))
}

func (h *PostHandler) New(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Post":   models.Post{},
		"Title":  "Create New Post",
		"Action": "/posts",
		"Method": "post",
	}

	h.renderTemplate(w, r, "form", data, "/posts/new")
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.handleError(w, err, "Failed to parse form", http.StatusBadRequest)
		return
	}

	post := models.Post{
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
	}

	errors, valid := validation.ValidatePost(post)
	if !valid {
		data := map[string]interface{}{
			"Post":   post,
			"Errors": errors,
			"Title":  "Create New Post",
			"Action": "/posts",
			"Method": "post",
		}

		h.renderTemplate(w, r, "form", data, "")
		return
	}

	id, err := h.repo.Create(r.Context(), post)
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

func (h *PostHandler) Edit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	post, err := h.repo.FindByID(r.Context(), id)
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

func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := r.ParseForm(); err != nil {
		h.handleError(w, err, "Failed to parse form", http.StatusBadRequest)
		return
	}

	existingPost, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	existingPost.Title = r.FormValue("title")
	existingPost.Content = r.FormValue("content")

	errors, valid := validation.ValidatePost(existingPost)
	if !valid {
		data := map[string]interface{}{
			"Post":   existingPost,
			"Errors": errors,
			"Title":  "Edit Post",
			"Action": fmt.Sprintf("/posts/%s", id),
			"Method": "put",
		}

		h.renderTemplate(w, r, "form", data, "")
		return
	}

	err = h.repo.Update(r.Context(), id, existingPost)
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

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.repo.Delete(r.Context(), id)
	if err != nil {
		h.handleError(w, err, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	h.redirectResponse(w, r, "/posts")
}

func (h *PostHandler) Seed(w http.ResponseWriter, r *http.Request) {
	samplePosts := seeder.GenerateSamplePosts(10)

	err := h.repo.CreateMany(r.Context(), samplePosts)
	if err != nil {
		h.handleError(w, err, "Failed to seed database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		posts, totalPages, err := h.repo.FindAll(r.Context(), 1, int64(h.config.App.PostsPerPage))
		if err != nil {
			h.handleError(w, err, "Failed to fetch posts after seeding", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Posts":       posts,
			"CurrentPage": 1,
			"TotalPages":  totalPages,
		}

		h.renderTemplate(w, r, "post_list", data, "/posts")
		return
	}

	h.redirectResponse(w, r, "/posts")
}
