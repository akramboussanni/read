package middleware

import (
	"net/http"

	"github.com/akramboussanni/gocode/internal/api"
	"github.com/akramboussanni/gocode/internal/utils"
	"github.com/go-chi/chi/v5"
)

// RequireAdmin middleware ensures the user is an admin
func RequireAdmin(r chi.Router) {
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := utils.UserFromContext(r.Context())
			if !ok {
				api.WriteUnauthorized(w)
				return
			}

			if !user.IsAdmin {
				api.WriteMessage(w, 403, "error", "Admin access required")
				return
			}

			next.ServeHTTP(w, r)
		})
	})
}
