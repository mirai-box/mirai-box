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
	ctx := r.Context()
	var createUserRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&createUserRequest); err != nil {
		slog.ErrorContext(ctx, "Failed to decode user json", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	createdUser, err := h.userService.CreateUser(ctx, createUserRequest.Username, createUserRequest.Password, createUserRequest.Role)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create user", "error", err, "username", createUserRequest.Username)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	slog.InfoContext(ctx, "User created", "userID", createdUser.ID, "username", createdUser.Username)

	stash, err := h.stashService.CreateStash(ctx, createdUser.ID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create stash for user", "error", err, "userID", createdUser.ID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create user stash")
		return
	}

	slog.InfoContext(ctx, "Stash created", "stashID", stash.ID, "userID", stash.UserID)

	response := models.UserResponse{
		ID:        createdUser.ID,
		Username:  createdUser.Username,
		Role:      createdUser.Role,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
	}

	SendJSONResponse(w, http.StatusCreated, response)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var loginRequest struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		KeepSignedIn bool   `json:"keepSignedIn"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		slog.ErrorContext(ctx, "Failed to decode login request", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.userService.Authenticate(ctx, loginRequest.Username, loginRequest.Password)
	if err != nil {
		slog.ErrorContext(ctx, "Authentication failed", "error", err, "username", loginRequest.Username)
		SendErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	session, err := h.store.Get(r, models.SessionCookieName)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get session", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	session.Values[models.SessionUserIDKey] = user.ID.String()
	if loginRequest.KeepSignedIn {
		session.Options.MaxAge = 7 * 24 * 60 * 60 // 1 week
	} else {
		session.Options.MaxAge = 24 * 60 * 60 // 1 day
	}

	if err := session.Save(r, w); err != nil {
		slog.ErrorContext(ctx, "Failed to save session", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	response := models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	slog.InfoContext(ctx, "User logged in", "userID", user.ID, "username", user.Username)
	SendJSONResponse(w, http.StatusOK, response)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")
	if userID == "" {
		slog.WarnContext(ctx, "GetUser called without user ID")
		SendErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	requestingUserID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok || requestingUserID != userID {
		slog.WarnContext(ctx, "Unauthorized attempt to view user data", "requestingUserID", requestingUserID, "targetUserID", userID)
		SendErrorResponse(w, http.StatusForbidden, "You don't have permission to view this user's data")
		return
	}

	user, err := h.userService.FindByID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find user", "error", err, "userID", userID)
		SendErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	response := models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	slog.InfoContext(ctx, "User data retrieved", "userID", user.ID)
	SendJSONResponse(w, http.StatusOK, response)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")
	if userID == "" {
		slog.WarnContext(ctx, "UpdateUser called without user ID")
		SendErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	var updateUserRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateUserRequest); err != nil {
		slog.ErrorContext(ctx, "Failed to decode update user request", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sessionUserID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok || sessionUserID != userID {
		slog.WarnContext(ctx, "Unauthorized attempt to update user", "sessionUserID", sessionUserID, "targetUserID", userID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionUserUUID, err := uuid.Parse(sessionUserID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to parse UUID in the session", "error", err, "sessionUserID", sessionUserID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	updatedUser := &models.User{
		ID:       sessionUserUUID,
		Username: updateUserRequest.Username,
		Password: updateUserRequest.Password,
		Role:     updateUserRequest.Role,
	}

	user, err := h.userService.UpdateUser(ctx, updatedUser)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update user", "error", err, "userID", userID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	response := models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	slog.InfoContext(ctx, "User updated", "userID", user.ID)
	SendJSONResponse(w, http.StatusOK, response)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")
	if userID == "" {
		slog.WarnContext(ctx, "DeleteUser called without user ID")
		SendErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	sessionUserID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok || sessionUserID != userID {
		slog.WarnContext(ctx, "Unauthorized attempt to delete user", "sessionUserID", sessionUserID, "targetUserID", userID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.userService.DeleteUser(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete user", "error", err, "userID", userID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	slog.InfoContext(ctx, "User deleted", "userID", userID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) LoginCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, err := h.store.Get(r, models.SessionCookieName)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get session", "error", err)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, ok := session.Values[models.SessionUserIDKey]
	if !ok {
		slog.WarnContext(ctx, "Unauthorized session")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	slog.InfoContext(ctx, "Auth check successful", "userID", userID)
	SendJSONResponse(w, http.StatusOK, map[string]string{"status": "OK"})
}
