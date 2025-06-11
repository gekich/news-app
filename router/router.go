package router

import (
	"net/http"

	custom "github.com/gekich/news-app/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// PostHandler defines the interface for post-related handlers.
// This allows for mocking in tests. You'll need to ensure your
// 'handlers.PostHandler' struct implements this interface.
type PostHandler interface {
	Index(w http.ResponseWriter, r *http.Request)
	New(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Show(w http.ResponseWriter, r *http.Request)
	Edit(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Seed(w http.ResponseWriter, r *http.Request)
}

// SetupRouter configures and returns the application router.
// It now takes a staticDir parameter to specify the directory for static files.
func SetupRouter(postHandler PostHandler, staticDir string) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(custom.MethodOverride)

	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	})

	r.Route("/posts", func(r chi.Router) {
		r.Get("/", postHandler.Index)
		r.Get("/new", postHandler.New)
		r.Post("/", postHandler.Create)
		r.Get("/{id}", postHandler.Show)
		r.Get("/{id}/edit", postHandler.Edit)
		r.Put("/{id}", postHandler.Update)
		r.Delete("/{id}", postHandler.Delete)
		r.Post("/seed", postHandler.Seed)
	})

	return r
}
