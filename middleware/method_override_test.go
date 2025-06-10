//go:build unit

package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethodOverride(t *testing.T) {
	tests := []struct {
		name           string
		initialMethod  string
		formMethod     string
		headerMethod   string
		expectedMethod string
	}{
		{
			name:           "no method override",
			initialMethod:  http.MethodPost,
			expectedMethod: http.MethodPost,
		},
		{
			name:           "method override via form field - PUT",
			initialMethod:  http.MethodPost,
			formMethod:     "PUT",
			expectedMethod: http.MethodPut,
		},
		{
			name:           "method override via form field - DELETE",
			initialMethod:  http.MethodPost,
			formMethod:     "DELETE",
			expectedMethod: http.MethodDelete,
		},
		{
			name:           "method override via header - PUT",
			initialMethod:  http.MethodPost,
			headerMethod:   "PUT",
			expectedMethod: http.MethodPut,
		},
		{
			name:           "method override via header - DELETE",
			initialMethod:  http.MethodPost,
			headerMethod:   "DELETE",
			expectedMethod: http.MethodDelete,
		},
		{
			name:           "method override with invalid method",
			initialMethod:  http.MethodPost,
			formMethod:     "INVALID",
			expectedMethod: http.MethodPost,
		},
		{
			name:           "method override with lowercase method",
			initialMethod:  http.MethodPost,
			formMethod:     "put",
			expectedMethod: http.MethodPut,
		},
		{
			name:           "non-POST request",
			initialMethod:  http.MethodGet,
			headerMethod:   "PUT",
			expectedMethod: http.MethodGet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request

			if tt.formMethod != "" {
				form := url.Values{}
				form.Add("_method", tt.formMethod)
				req = httptest.NewRequest(tt.initialMethod, "/test", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req = httptest.NewRequest(tt.initialMethod, "/test", nil)
			}

			if tt.headerMethod != "" {
				req.Header.Set("X-HTTP-Method-Override", tt.headerMethod)
			}

			rr := httptest.NewRecorder()

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedMethod, r.Method)
			})

			handler := MethodOverride(testHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}
}
