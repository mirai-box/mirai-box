package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"

	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/service"
)

type UserHandler struct {
	userService  service.UserServiceInterface
	stashService service.StashServiceInterface
	store        *sessions.CookieStore
}

func NewUserHandler(
	userService service.UserServiceInterface,
	stashService service.StashServiceInterface,
	cookieStore *sessions.CookieStore,
) *UserHandler {
	return &UserHandler{
		store:        cookieStore,
		userService:  userService,
		stashService: stashService,
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		slog.Error("Failed decode user json", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdUser, err := h.userService.CreateUser(r.Context(), user.Username, user.Password, user.Role)
	if err != nil {
		slog.Error("Failed to create user", "error", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// don't return hashed password back
	createdUser.Password = ""

	slog.Info("user is created", "userID", createdUser.ID, "username", createdUser.Username)

	// Note: in the current version one use can only have one stash storage, so we are creating one now
	stash, err := h.stashService.CreateStash(r.Context(), createdUser.ID.String())
	if err != nil {
		slog.Error("Failed to create stash for user", "error", err, "userID", createdUser.ID)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	slog.Info("stash if created", "stashID", stash.ID, "userID", stash.UserID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		KeepSignedIn bool   `json:"keepSignedIn"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Authenticate(r.Context(), loginRequest.Username, loginRequest.Password)
	if err != nil {
		slog.Error("Authentication failed", "error", err, "username", loginRequest.Username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	session, err := h.store.Get(r, models.SessionCookieName)
	if err != nil {
		slog.Error("Failed to get session from context", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session.Values[models.SessionUserIDKey] = user.ID.String()
	if err := session.Save(r, w); err != nil {
		slog.Error("Failed to save session", "error", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	if loginRequest.KeepSignedIn {
		session.Options.MaxAge = 7 * 24 * 60 * 60 // 1 week
	} else {
		session.Options.MaxAge = 24 * 60 * 60 // 1 day
	}

	// don't return hashed password back
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Check if the requesting user has permission to view this user's data
	requestingUserID, _ := middleware.GetUserIDFromContext(r.Context())
	if requestingUserID != userID {
		// You might want to check for admin role here as well
		respondWithError(w, http.StatusForbidden, "You don't have permission to view this user's data")
		return
	}

	user, err := h.userService.FindByID(r.Context(), userID)
	if err != nil {
		slog.Error("Failed to find user", "error", err, "userID", userID)
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Ensure password is not sent in the response
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var updatedUser models.User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure the user can only update their own profile
	sessionUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || sessionUserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionUserUUID, err := uuid.Parse(sessionUserID)
	if err != nil {
		slog.Error("Failed to parse UUID in the session", "error", err, "sessionUserID", sessionUserID)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	updatedUser.ID = sessionUserUUID // Ensure we're updating the correct user
	user, err := h.userService.UpdateUser(r.Context(), &updatedUser)
	if err != nil {
		slog.Error("Failed to update user", "error", err, "userID", userID)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Ensure the user can only delete their own account
	sessionUserID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok || sessionUserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.userService.DeleteUser(r.Context(), userID)
	if err != nil {
		slog.Error("Failed to delete user", "error", err, "userID", userID)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) LoginCheck(w http.ResponseWriter, r *http.Request) {
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
