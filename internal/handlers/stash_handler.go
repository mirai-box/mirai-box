package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/service"
)

type StashHandler struct {
	stashService service.StashServiceInterface
}

func NewStashHandler(stashService service.StashServiceInterface) *StashHandler {
	return &StashHandler{stashService: stashService}
}

func (h *StashHandler) MyStash(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized attempt to access stash")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stash, err := h.stashService.FindByUserID(ctx, user.ID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve stash", "error", err, "userID", user.ID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve stash")
		return
	}

	response := models.StashResponse{
		ID:          stash.ID,
		UserID:      stash.UserID,
		ArtProjects: stash.ArtProjects,
		Files:       stash.Files,
		UsedSpace:   stash.UsedSpace,
		CreatedAt:   stash.CreatedAt,
		UpdatedAt:   stash.UpdatedAt,
	}

	slog.InfoContext(ctx, "Stash retrieved successfully", "userID", user.ID)
	SendJSONResponse(w, http.StatusOK, response)
}

func (h *StashHandler) CreateStash(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized attempt to create stash")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stash, err := h.stashService.CreateStash(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create stash", "error", err, "userID", userID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create stash")
		return
	}

	response := models.StashResponse{
		ID:          stash.ID,
		UserID:      stash.UserID,
		ArtProjects: stash.ArtProjects,
		Files:       stash.Files,
		UsedSpace:   stash.UsedSpace,
		CreatedAt:   stash.CreatedAt,
		UpdatedAt:   stash.UpdatedAt,
	}

	slog.InfoContext(ctx, "Stash created successfully", "stashID", stash.ID, "userID", userID)
	SendJSONResponse(w, http.StatusCreated, response)
}

func (h *StashHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stashID := chi.URLParam(r, "id")
	if stashID == "" {
		slog.WarnContext(ctx, "Attempt to find stash without ID")
		SendErrorResponse(w, http.StatusBadRequest, "Stash ID is required")
		return
	}

	stash, err := h.stashService.FindByID(ctx, stashID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find stash", "error", err, "stashID", stashID)
		SendErrorResponse(w, http.StatusNotFound, "Stash not found")
		return
	}

	response := models.StashResponse{
		ID:          stash.ID,
		UserID:      stash.UserID,
		ArtProjects: stash.ArtProjects,
		Files:       stash.Files,
		UsedSpace:   stash.UsedSpace,
		CreatedAt:   stash.CreatedAt,
		UpdatedAt:   stash.UpdatedAt,
	}

	slog.InfoContext(ctx, "Stash found successfully", "stashID", stashID)
	SendJSONResponse(w, http.StatusOK, response)
}

func (h *StashHandler) FindByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		slog.WarnContext(ctx, "Attempt to find stash without user ID")
		SendErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	stash, err := h.stashService.FindByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find stash for user", "error", err, "userID", userID)
		SendErrorResponse(w, http.StatusNotFound, "Stash not found")
		return
	}

	response := models.StashResponse{
		ID:          stash.ID,
		UserID:      stash.UserID,
		ArtProjects: stash.ArtProjects,
		Files:       stash.Files,
		UsedSpace:   stash.UsedSpace,
		CreatedAt:   stash.CreatedAt,
		UpdatedAt:   stash.UpdatedAt,
	}

	slog.InfoContext(ctx, "Stash found successfully for user", "userID", userID)
	SendJSONResponse(w, http.StatusOK, response)
}
