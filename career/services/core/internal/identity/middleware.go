package identity

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "auth_user"

func GetUserFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(UserContextKey).(*User)
	return u, ok
}

func AuthMiddleware(authService *AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				next.ServeHTTP(w, r)
				return
			}

			user, err := authService.ValidateToken(r.Context(), parts[1])
			if err == nil && user != nil {
				ctx := context.WithValue(r.Context(), UserContextKey, user)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok || user == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			})
			return
		}
		next(w, r)
	}
}

func RequireSuperAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok || user == nil || user.PlatformRole != PlatformRoleSuperAdmin {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"code":    "FORBIDDEN",
				"message": "Super Admin privileges required",
			})
			return
		}
		next(w, r)
	}
}