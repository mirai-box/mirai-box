package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"

	"github.com/mirai-box/mirai-box/internal/service"
)

type UserHandler struct {
	service service.UserService
	store   *sessions.CookieStore
}

// LoginRequest represents the expected request body for login.
type LoginRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	KeepSignedIn bool   `json:"keepSignedIn"`
}

func NewUserHandler(service service.UserService, sessionKey string) *UserHandler {
	store := sessions.NewCookieStore([]byte(sessionKey))
	return &UserHandler{
		service: service,
		store:   store,
	}
}

// AuthCheckHandler checks if the user is authenticated.
func (h *UserHandler) AuthCheckHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "session-name")
	userID, ok := session.Values["user_id"]
	if !ok {
		slog.Error("AuthCheckHandler: unauthorized session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	slog.Debug("auth check successful", "user_id", userID)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		slog.Error("error writing response")
	}
}

// LoginHandler handles user login and session creation.
func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		slog.Error("LoginHandler: empty credentials")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	user, err := h.service.Authenticate(req.Username, req.Password)
	if err != nil {
		slog.Error("LoginHandler: invalid credentials", "error", err, "username", req.Username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	session, _ := h.store.Get(r, "session-name")
	session.Values["user_id"] = user.ID

	if req.KeepSignedIn {
		session.Options.MaxAge = 7 * 24 * 60 * 60 // 1 week
	} else {
		session.Options.MaxAge = 24 * 60 * 60 // 1 day
	}

	session.Save(r, w)

	slog.Info("login successful", "user", req.Username, "user_id", user.ID)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		slog.Error("error writing response")
	}
}

// AuthMiddleware checks for a valid session and user role before allowing access to admin routes.
func (h *UserHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := h.store.Get(r, "session-name")
		userID, ok := session.Values["user_id"]
		if !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		user, err := h.service.FindByID(userID.(string))
		if err != nil || user.Role != "admin" {
			slog.Error("LoginHandler: invalid credentials", "error", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		slog.Debug("auth successful", "user_id", user.ID)

		next.ServeHTTP(w, r)
	})
}
