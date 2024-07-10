package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/service"
)

// UserHandler is a struct that holds references to the user service and session store.
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

// NewUserHandler creates a new UserHandler.
func NewUserHandler(service service.UserService, sessionKey string) *UserHandler {
	store := sessions.NewCookieStore([]byte(sessionKey))
	return &UserHandler{
		service: service,
		store:   store,
	}
}

// AuthCheckHandler checks if the user is authenticated.
func (h *UserHandler) AuthCheckHandler(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, models.SessionCookieName)
	if err != nil {
		slog.Error("AuthCheckHandler: failed to get session", "error", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, ok := session.Values[models.SessionUserIDKey]
	if !ok {
		slog.Error("AuthCheckHandler: unauthorized session")
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	slog.Debug("auth check successful", "user_id", userID)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		slog.Error("AuthCheckHandler: error writing response", "error", err)
	}
}

// LoginHandler handles user login and session creation.
func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("LoginHandler: failed to decode request", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		slog.Error("LoginHandler: empty credentials")
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	user, err := h.service.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		slog.Error("LoginHandler: invalid credentials", "error", err, "username", req.Username)
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	session, err := h.store.Get(r, models.SessionCookieName)
	if err != nil {
		slog.Error("LoginHandler: failed to get session", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	session.Values[models.SessionUserIDKey] = user.ID
	if req.KeepSignedIn {
		session.Options.MaxAge = 7 * 24 * 60 * 60 // 1 week
	} else {
		session.Options.MaxAge = 24 * 60 * 60 // 1 day
	}

	if err := session.Save(r, w); err != nil {
		slog.Error("LoginHandler: failed to save session", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	slog.Info("login successful", "user", req.Username, "user_id", user.ID)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		slog.Error("LoginHandler: error writing response", "error", err)
	}
}

// AuthMiddleware checks for a valid session and user role before allowing access to admin routes.
func (h *UserHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := h.store.Get(r, models.SessionCookieName)
		if err != nil {
			slog.Error("AuthMiddleware: failed to get session", "error", err)
			respondWithError(w, http.StatusForbidden, "Forbidden")
			return
		}

		userID, ok := session.Values[models.SessionUserIDKey]
		if !ok {
			respondWithError(w, http.StatusForbidden, "Forbidden")
			return
		}

		user, err := h.service.FindByID(r.Context(), userID.(string))
		if err != nil || user.Role != "admin" {
			slog.Error("AuthMiddleware: invalid credentials or role", "error", err)
			respondWithError(w, http.StatusForbidden, "Forbidden")
			return
		}

		slog.Debug("auth successful", "user_id", user.ID)

		next.ServeHTTP(w, r)
	})
}
