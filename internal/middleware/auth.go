// Package middleware holds chi-compatible net/http middleware: JWT
// authentication and request logging.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/mbeltran93/taskflow-go/internal/auth"
)

type contextKey int

const userIDKey contextKey = iota

// RequireAuth rejects requests without a valid "Authorization: Bearer <token>"
// header and otherwise stores the authenticated user id in the request context.
func RequireAuth(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(header, prefix) {
				writeUnauthorized(w)
				return
			}

			claims, err := jwtManager.Verify(strings.TrimPrefix(header, prefix))
			if err != nil {
				writeUnauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext returns the authenticated user id set by RequireAuth.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"missing or invalid bearer token"}`))
}
