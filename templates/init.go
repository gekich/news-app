package templates

import (
	"html/template"

	"github.com/gekich/news-app/templates/functions"
)

// TemplateFuncs combines all template functions
var TemplateFuncs template.FuncMap

func init() {
	// Initialize template functions map
	TemplateFuncs = template.FuncMap{}

	// Add basic utility functions
	basicFuncs := functions.BasicFuncs()
	for name, fn := range basicFuncs {
		TemplateFuncs[name] = fn
	}

	// Add pagination functions
	paginationFuncs := functions.PaginationFuncs()
	for name, fn := range paginationFuncs {
		TemplateFuncs[name] = fn
	}
}

// PostTemplates initializes and returns templates for post handling
func PostTemplates() map[string]*template.Template {
	layout := "templates/layout.html"
	partials := []string{
		"templates/partials/back_button.html",
		"templates/partials/post_actions.html",
		"templates/partials/pagination.html",
	}

	tmpl := map[string]*template.Template{
		"post_list": template.Must(template.New("layout.html").Funcs(TemplateFuncs).ParseFiles(
			append([]string{layout, "templates/posts/post_list.html"}, partials...)...)),
		"show": template.Must(template.New("layout.html").Funcs(TemplateFuncs).ParseFiles(
			append([]string{layout, "templates/posts/show.html"}, partials...)...)),
		"form": template.Must(template.New("layout.html").Funcs(TemplateFuncs).ParseFiles(
			append([]string{layout, "templates/posts/form.html"}, partials...)...)),
	}

	return tmpl
}
