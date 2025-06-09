package templates

import (
	"fmt"
	"html/template"
)

// TemplateFuncs defines custom functions available in templates
var TemplateFuncs = template.FuncMap{
	"add":      func(a, b int) int { return a + b },
	"subtract": func(a, b int) int { return a - b },
	"truncate": func(s string, n int) string {
		if len(s) <= n {
			return s
		}
		return s[:n] + "..."
	},
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
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
