//go:build unit

package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockPostHandler is a mock implementation of the PostHandler interface for testing.
type mockPostHandler struct{}

func (m *mockPostHandler) Index(w http.ResponseWriter, r *http.Request)  { w.Write([]byte("Index")) }
func (m *mockPostHandler) New(w http.ResponseWriter, r *http.Request)    { w.Write([]byte("New")) }
func (m *mockPostHandler) Create(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Create")) }
func (m *mockPostHandler) Show(w http.ResponseWriter, r *http.Request)   { w.Write([]byte("Show")) }
func (m *mockPostHandler) Edit(w http.ResponseWriter, r *http.Request)   { w.Write([]byte("Edit")) }
func (m *mockPostHandler) Update(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Update")) }
func (m *mockPostHandler) Delete(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Delete")) }
func (m *mockPostHandler) Seed(w http.ResponseWriter, r *http.Request)   { w.Write([]byte("Seed")) }

// TestSetupRouter verifies that all routes are correctly configured.
func TestSetupRouter(t *testing.T) {
	tests := []struct {
		method       string
		path         string
		expectedCode int
		expectedBody string
	}{
		{"GET", "/", http.StatusSeeOther, ""},
		{"GET", "/posts", http.StatusOK, "Index"},
		{"GET", "/posts/new", http.StatusOK, "New"},
		{"POST", "/posts", http.StatusOK, "Create"},
		{"GET", "/posts/123", http.StatusOK, "Show"},
		{"GET", "/posts/123/edit", http.StatusOK, "Edit"},
		{"PUT", "/posts/123", http.StatusOK, "Update"},
		{"DELETE", "/posts/123", http.StatusOK, "Delete"},
		{"POST", "/posts/seed", http.StatusOK, "Seed"},
		{"GET", "/non-existent-path", http.StatusNotFound, "404 page not found"},
	}

	// The static directory can be a dummy value since we are not testing static files here.
	router := SetupRouter(&mockPostHandler{}, ".")

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedCode {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tc.expectedCode)
			}

			if tc.expectedCode == http.StatusSeeOther {
				if location := rr.Header().Get("Location"); location != "/posts" {
					t.Errorf("Handler returned wrong redirect location: got %s want /posts", location)
				}
				return
			}

			if tc.expectedBody != "" && !strings.Contains(rr.Body.String(), tc.expectedBody) {
				t.Errorf("Handler returned unexpected body: got %q, want to contain %q", rr.Body.String(), tc.expectedBody)
			}
		})
	}
}

// TestStaticFileServer tests the static file serving functionality.
func TestStaticFileServer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "static_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fileContent := "body { color: blue; }"
	filePath := filepath.Join(tmpDir, "style.css")
	if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("Failed to write dummy static file: %v", err)
	}

	router := SetupRouter(&mockPostHandler{}, tmpDir)

	// Test case for an existing file
	t.Run("existing file", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/static/style.css", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Static file handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		if body := rr.Body.String(); body != fileContent {
			t.Errorf("Static file handler returned unexpected body: got %q want %q", body, fileContent)
		}
	})

	// Test case for a non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/static/nonexistent.js", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("Static file handler returned wrong status code for non-existent file: got %v want %v", status, http.StatusNotFound)
		}
	})
}
