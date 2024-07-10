// File: internal/handlers/stash_handler.go

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/service"
)

type StashHandler struct {
	stashService service.StashServiceInterface
}

func NewStashHandler(stashService service.StashServiceInterface) *StashHandler {
	return &StashHandler{stashService: stashService}
}

func (h *StashHandler) MyStash(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stash, err := h.stashService.FindByUserID(r.Context(), user.ID.String())
	if err != nil {
		slog.Error("Failed to retrieve stash", "error", err, "userID", user.ID.String())
		http.Error(w, "Failed to retrieve stash", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stash)
}

func (h *StashHandler) CreateStash(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stash, err := h.stashService.CreateStash(r.Context(), userID)
	if err != nil {
		slog.Error("Failed to create stash", "error", err, "userID", userID)
		http.Error(w, "Failed to create stash", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(stash)
}

func (h *StashHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	stashID := chi.URLParam(r, "id")
	if stashID == "" {
		http.Error(w, "Stash ID is required", http.StatusBadRequest)
		return
	}

	stash, err := h.stashService.FindByID(r.Context(), stashID)
	if err != nil {
		slog.Error("Failed to find stash", "error", err, "stashID", stashID)
		http.Error(w, "Stash not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stash)
}

func (h *StashHandler) FindByUserID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	stash, err := h.stashService.FindByUserID(r.Context(), userID)
	if err != nil {
		slog.Error("Failed to find stash for user", "error", err, "userID", userID)
		http.Error(w, "Stash not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stash)
}
