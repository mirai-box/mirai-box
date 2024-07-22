package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ValidateUUID is a middleware that checks if the URL parameter with the given name is a valid UUID.
func ValidateUUID(paramName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			param := chi.URLParam(r, paramName)
			if param != "" {
				_, err := uuid.Parse(param)
				if err != nil {
					http.Error(w, "Invalid UUID", http.StatusBadRequest)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
