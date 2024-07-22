package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"

	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/service"
)

type contextKey string

const (
	SessionKey contextKey = "session"
	UserIDKey  contextKey = "userID"
	UserKey    contextKey = "user"
)

type Middleware struct {
	store       sessions.Store
	userService service.UserService
}

func NewMiddleware(store sessions.Store, userService service.UserService) *Middleware {
	return &Middleware{
		store:       store,
		userService: userService,
	}
}

func (m *Middleware) SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.store.Get(r, model.SessionCookieName)
		if err != nil {
			slog.Error("SessionMiddleware: failed to get session", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), SessionKey, session)
		if userID, ok := session.Values[model.SessionUserIDKey].(string); ok {
			ctx = context.WithValue(ctx, UserIDKey, userID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MockAuthMiddleware simulates authentication for testing purposes
func (m *Middleware) MockAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for the presence of an Authorization header
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Create a mock user
		user := &model.User{
			ID:       uuid.MustParse(userID),
			Username: "testuser",
			Role:     "user",
		}

		// Add the user to the request context
		ctx := context.WithValue(r.Context(), UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := m.store.Get(r, model.SessionCookieName)
		userID, ok := session.Values[model.SessionUserIDKey]
		if !ok {
			slog.Error("AuthMiddleware: no user ID in session")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := m.userService.GetUser(r.Context(), userID.(string))
		if err != nil {
			slog.Error("AuthMiddleware: failed to find user", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserKey).(*model.User)
			if !ok {
				slog.Error("RequireRole: user not found in context")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			for _, role := range roles {
				if role == "any" || role == "self" || user.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			slog.Error("RequireRole: user does not have required role",
				"required", roles,
				"actual", user.Role,
				"userID", user.ID,
			)
			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}

func GetUserFromContext(ctx context.Context) (*model.User, bool) {
	user, ok := ctx.Value(UserKey).(*model.User)
	return user, ok
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

func GetSessionFromContext(ctx context.Context) (*sessions.Session, bool) {
	session, ok := ctx.Value(SessionKey).(*sessions.Session)
	return session, ok
}
