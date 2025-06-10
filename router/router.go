package router

import (
	"net/http"

	"github.com/gekich/news-app/handlers"
	custom "github.com/gekich/news-app/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SetupRouter configures and returns the application router
func SetupRouter(postHandler *handlers.PostHandler) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(custom.MethodOverride)

	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

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
