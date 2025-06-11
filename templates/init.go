package templates

import (
	"fmt"
	"html/template"

	"github.com/gekich/news-app/templates/functions"
)

// NewTemplateFuncs creates and returns a new template.FuncMap
func NewTemplateFuncs() template.FuncMap {
	// Initialize template functions map
	funcs := template.FuncMap{}

	// Add basic utility functions
	basicFuncs := functions.BasicFuncs()
	for name, fn := range basicFuncs {
		funcs[name] = fn
	}

	// Add pagination functions
	paginationFuncs := functions.PaginationFuncs()
	for name, fn := range paginationFuncs {
		funcs[name] = fn
	}
	return funcs
}

// PostTemplates initializes and returns templates for post handling
func PostTemplates() map[string]*template.Template {
	return NewPostTemplates("templates")
}

// NewPostTemplates initializes and returns templates for post handling with a custom base path
func NewPostTemplates(basePath string) map[string]*template.Template {
	layout := fmt.Sprintf("%s/layout.html", basePath)
	partials := []string{
		fmt.Sprintf("%s/partials/back_button.html", basePath),
		fmt.Sprintf("%s/partials/post_actions.html", basePath),
		fmt.Sprintf("%s/partials/pagination.html", basePath),
	}

	tmpl := map[string]*template.Template{
		"post_list": template.Must(template.New("layout.html").Funcs(NewTemplateFuncs()).ParseFiles(
			append([]string{layout, fmt.Sprintf("%s/posts/post_list.html", basePath)}, partials...)...)),
		"show": template.Must(template.New("layout.html").Funcs(NewTemplateFuncs()).ParseFiles(
			append([]string{layout, fmt.Sprintf("%s/posts/show.html", basePath)}, partials...)...)),
		"form": template.Must(template.New("layout.html").Funcs(NewTemplateFuncs()).ParseFiles(
			append([]string{layout, fmt.Sprintf("%s/posts/form.html", basePath)}, partials...)...)),
	}

	return tmpl
}
