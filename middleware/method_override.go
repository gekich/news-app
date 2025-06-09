package middleware

import (
	"net/http"
	"strings"
)

// MethodOverride checks for _method form field or X-HTTP-Method-Override header
// to support PUT/DELETE in browsers that don't support these methods in forms
func MethodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			if err := r.ParseForm(); err == nil {
				method := r.PostForm.Get("_method")
				if method != "" {
					method = strings.ToUpper(method)
					if method == "PUT" || method == "DELETE" {
						r.Method = method
					}
				}
			}

			method := r.Header.Get("X-HTTP-Method-Override")
			if method != "" {
				method = strings.ToUpper(method)
				if method == "PUT" || method == "DELETE" {
					r.Method = method
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
