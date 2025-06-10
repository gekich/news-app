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
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Posts":       posts,
		"CurrentPage": page,
		"TotalPages":  totalPages,
	}

	if isHTMXRequest(r) {
		w.Header().Set("HX-Push-Url", fmt.Sprintf("/posts?page=%d", page))
		if err := h.tmpl["post_list"].ExecuteTemplate(w, "content", data); err != nil {
			http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if err := h.tmpl["post_list"].Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func (h *PostHandler) Show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	post, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
		}
		return
	}

	data := map[string]interface{}{
		"Post": post,
	}

	if isHTMXRequest(r) {
		w.Header().Set("HX-Push-Url", fmt.Sprintf("/posts/%s", id))
		if err := h.tmpl["show"].ExecuteTemplate(w, "content", data); err != nil {
			http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if err := h.tmpl["show"].Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func (h *PostHandler) New(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Post":   models.Post{},
		"Title":  "Create New Post",
		"Action": "/posts",
		"Method": "post",
	}

	if isHTMXRequest(r) {
		w.Header().Set("HX-Push-Url", "/posts/new")
		if err := h.tmpl["form"].ExecuteTemplate(w, "content", data); err != nil {
			http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if err := h.tmpl["form"].Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	post := models.Post{
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
	}

	isHTMX := isHTMXRequest(r)
	errors, valid := validation.ValidatePost(post)
	if !valid {
		data := map[string]interface{}{
			"Post":   post,
			"Errors": errors,
			"Title":  "Create New Post",
			"Action": "/posts",
			"Method": "post",
		}

		if isHTMX {
			if err := h.tmpl["form"].ExecuteTemplate(w, "content", data); err != nil {
				http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
			}
			return
		}

		if err := h.tmpl["form"].Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
		return
	}

	id, err := h.repo.Create(r.Context(), post)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	if isHTMX {
		http.Redirect(w, r, fmt.Sprintf("/posts/%s", id), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func (h *PostHandler) Edit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	post, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
		}
		return
	}

	data := map[string]interface{}{
		"Post":   post,
		"Title":  "Edit Post",
		"Action": fmt.Sprintf("/posts/%s", id),
		"Method": "put",
	}

	if isHTMXRequest(r) {
		w.Header().Set("HX-Push-Url", fmt.Sprintf("/posts/%s/edit", id))
		if err := h.tmpl["form"].ExecuteTemplate(w, "content", data); err != nil {
			http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if err := h.tmpl["form"].Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	existingPost, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
		}
		return
	}

	existingPost.Title = r.FormValue("title")
	existingPost.Content = r.FormValue("content")

	isHTMX := isHTMXRequest(r)
	errors, valid := validation.ValidatePost(existingPost)
	if !valid {
		data := map[string]interface{}{
			"Post":   existingPost,
			"Errors": errors,
			"Title":  "Edit Post",
			"Action": fmt.Sprintf("/posts/%s", id),
			"Method": "put",
		}

		if isHTMX {
			if err := h.tmpl["form"].ExecuteTemplate(w, "content", data); err != nil {
				http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
			}
			return
		}

		if err := h.tmpl["form"].Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
		return
	}

	err = h.repo.Update(r.Context(), id, existingPost)
	if err != nil {
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
		return
	}

	if isHTMX {
		http.Redirect(w, r, fmt.Sprintf("/posts/%s", id), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.repo.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		w.Header().Set("HX-Redirect", "/posts")
		return
	}

	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func (h *PostHandler) SeedHandler(w http.ResponseWriter, r *http.Request) {
	samplePosts := seeder.GenerateSamplePosts(10)

	err := h.repo.CreateMany(r.Context(), samplePosts)
	if err != nil {
		http.Error(w, "Failed to seed database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		posts, totalPages, err := h.repo.FindAll(r.Context(), 1, int64(h.config.App.PostsPerPage))
		if err != nil {
			http.Error(w, "Failed to fetch posts after seeding", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Push-Url", "/posts")

		data := map[string]interface{}{
			"Posts":       posts,
			"CurrentPage": 1,
			"TotalPages":  totalPages,
		}

		if err := h.tmpl["post_list"].ExecuteTemplate(w, "content", data); err != nil {
			http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}
