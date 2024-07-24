package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"

	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/service"
)

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

// UserHandler handles HTTP requests related to user operations.
type UserHandler struct {
	userService service.UserService
	store       *sessions.CookieStore
}

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(userService service.UserService, cookieStore *sessions.CookieStore) *UserHandler {
	return &UserHandler{
		userService: userService,
		store:       cookieStore,
	}
}

// CreateUser handles the creation of a new user.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "CreateUser")

	var createUserRequest struct {
		Username string `json:"username"  validate:"required"`
		Password string `json:"password"  validate:"required"`
		Role     string `json:"role"      validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&createUserRequest); err != nil {
		logger.Error("Failed to decode user json", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(createUserRequest); err != nil {
		logger.Error("Invalid input data", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	createdUser, err := h.userService.CreateUser(ctx, createUserRequest.Username, createUserRequest.Password, createUserRequest.Role)
	if err != nil {
		if errors.Is(err, model.ErrDuplicateUsername) {
			SendErrorResponse(w, http.StatusConflict, "Username already exists")
			return
		}

		logger.Error("Failed to create user", "error", err, "username", createUserRequest.Username)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	logger.Info("User created", "userID", createdUser.ID, "username", createdUser.Username)
	SendJSONResponse(w, http.StatusCreated, convertToUserResponse(createdUser))
}

// Login handles user authentication and session creation.
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "Login")

	var loginRequest struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		KeepSignedIn bool   `json:"keepSignedIn"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		logger.Error("Failed to decode login request", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.userService.Authenticate(ctx, loginRequest.Username, loginRequest.Password)
	if err != nil {
		logger.Warn("Authentication failed", "error", err, "username", loginRequest.Username)
		SendErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	session, err := h.store.Get(r, model.SessionCookieName)
	if err != nil {
		logger.Error("Failed to get session", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	session.Values[model.SessionUserIDKey] = user.ID.String()
	session.Options.MaxAge = h.getSessionMaxAge(loginRequest.KeepSignedIn)

	if err := session.Save(r, w); err != nil {
		logger.Error("Failed to save session", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	logger.Info("User logged in", "userID", user.ID, "username", user.Username)
	SendJSONResponse(w, http.StatusOK, convertToUserResponse(user))
}

// GetUser retrieves user information.
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")
	logger := slog.With("handler", "GetUser", "userID", userID)

	requestingUser, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("User not found in context")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Allow access if the requesting user is an admin or if they're requesting their own data
	if requestingUser.Role != "admin" && requestingUser.ID.String() != userID {
		logger.Warn("Unauthorized attempt to view user data",
			"requestingUserID", requestingUser.ID,
			"requestingUserRole", requestingUser.Role,
			"targetUserID", userID)
		SendErrorResponse(w, http.StatusForbidden, "You don't have permission to view this user's data")
		return
	}

	user, err := h.userService.GetUser(ctx, userID)
	if err != nil {
		logger.Error("Failed to find user", "error", err, "userID", userID)
		SendErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	logger.Info("User data retrieved", "userID", user.ID)
	SendJSONResponse(w, http.StatusOK, convertToUserResponse(user))
}

// UpdateUser handles updating user information.
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")
	logger := slog.With("handler", "UpdateUser", "userID", userID)

	var updateUserRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateUserRequest); err != nil {
		logger.Error("Failed to decode update user request", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	requestingUser, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("User not found in context")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Allow access if the requesting user is an admin or if they're requesting their own data
	if requestingUser.Role != "admin" && requestingUser.ID.String() != userID {
		logger.Warn("Unauthorized attempt to update user",
			"requestingUserID", requestingUser.ID,
			"requestingUserRole", requestingUser.Role,
			"targetUserID", userID)
		SendErrorResponse(w, http.StatusForbidden, "You don't have permission to view this user's data")
		return
	}

	sessionUserUUID, err := uuid.Parse(userID)
	if err != nil {
		logger.Error("Failed to parse UUID in the session", "error", err, "sessionUserID", sessionUserUUID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	updatedUser := &model.User{
		ID:       sessionUserUUID,
		Username: updateUserRequest.Username,
		Password: updateUserRequest.Password,
		Role:     updateUserRequest.Role,
	}

	if err := h.userService.UpdateUser(ctx, updatedUser); err != nil {
		logger.Error("Failed to update user", "error", err, "userID", userID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	logger.Info("User updated", "userID", updatedUser.ID)
	SendJSONResponse(w, http.StatusOK, convertToUserResponse(updatedUser))
}

// DeleteUser handles user deletion.
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")
	logger := slog.With("handler", "DeleteUser", "targetUserID", userID)

	sessionUser, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("Unauthorized attempt to delete user", "sessionUserID", sessionUser.ID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if sessionUser.Role != "admin" {
		logger.Warn("Unauthorized attempt to delete user", "sessionUserID", sessionUser.ID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.userService.DeleteUser(ctx, userID)
	if err != nil {
		logger.Error("Failed to delete user", "error", err)
		if errors.Is(err, model.ErrUserNotFound) {
			SendErrorResponse(w, http.StatusNotFound, "User not found")
			return
		}
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	logger.Info("User deleted", "userID", userID)
	w.WriteHeader(http.StatusNoContent)
}

// LoginCheck verifies if the user is currently logged in.
func (h *UserHandler) LoginCheck(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "LoginCheck")

	session, err := h.store.Get(r, model.SessionCookieName)
	if err != nil {
		logger.Error("Failed to get session", "error", err)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, ok := session.Values[model.SessionUserIDKey]
	if !ok {
		logger.Warn("Unauthorized session")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	logger.Info("Auth check successful", "userID", userID)
	SendJSONResponse(w, http.StatusOK, map[string]string{"status": "OK"})
}

// MyStash retrieves the stash information for the authenticated user.
func (h *UserHandler) MyStash(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "MyStash")

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("Unauthorized user attempt to retrieve stash")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stash, err := h.userService.GetStashByUserID(ctx, user.ID.String())
	if err != nil {
		logger.Error("Failed to retrieve stash", "error", err, "userID", user.ID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve stash")
		return
	}

	logger.Info("Stash retrieved successfully", "userID", user.ID)
	SendJSONResponse(w, http.StatusOK, convertToStashResponse(stash))
}

// Helper functions

func (h *UserHandler) getSessionMaxAge(keepSignedIn bool) int {
	if keepSignedIn {
		return 7 * 24 * 60 * 60 // 1 week
	}
	return 24 * 60 * 60 // 1 day
}

func convertToUserResponse(user *model.User) model.UserResponse {
	return model.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func convertToStashResponse(stash *model.Stash) model.StashResponse {
	return model.StashResponse{
		ID:          stash.ID,
		UserID:      stash.UserID,
		ArtProjects: stash.ArtProjects,
		Files:       stash.Files,
		UsedSpace:   stash.UsedSpace,
		CreatedAt:   stash.CreatedAt,
		UpdatedAt:   stash.UpdatedAt,
	}
}
