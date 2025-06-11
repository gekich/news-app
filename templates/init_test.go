//go:build unit

package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTemplateFuncs(t *testing.T) {
	funcs := NewTemplateFuncs()
	if len(funcs) == 0 {
		t.Error("expected to have some template functions, but got none")
	}
}

func TestNewPostTemplates(t *testing.T) {
	// Create a temporary directory for our test templates
	tmpDir, err := os.MkdirTemp("", "templates")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create dummy template files
	createDummyFile := func(path string, content string) {
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			t.Fatalf("failed to create dir for dummy file: %v", err)
		}
		err = os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write dummy file: %v", err)
		}
	}

	createDummyFile(filepath.Join(tmpDir, "layout.html"), `{{template "content" .}}`)
	createDummyFile(filepath.Join(tmpDir, "partials", "back_button.html"), `back`)
	createDummyFile(filepath.Join(tmpDir, "partials", "post_actions.html"), `actions`)
	createDummyFile(filepath.Join(tmpDir, "partials", "pagination.html"), `pagination`)
	createDummyFile(filepath.Join(tmpDir, "posts", "post_list.html"), `{{define "content"}}post list{{end}}`)
	createDummyFile(filepath.Join(tmpDir, "posts", "show.html"), `{{define "content"}}show post{{end}}`)
	createDummyFile(filepath.Join(tmpDir, "posts", "form.html"), `{{define "content"}}post form{{end}}`)

	templates := NewPostTemplates(tmpDir)

	if templates == nil {
		t.Fatal("expected templates to be initialized, but got nil")
	}

	expectedKeys := []string{"post_list", "show", "form"}
	for _, key := range expectedKeys {
		if _, ok := templates[key]; !ok {
			t.Errorf("expected to find key %q in templates map, but it was not there", key)
		}
	}
}
