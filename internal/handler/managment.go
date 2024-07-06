package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mirai-box/mirai-box/internal/service"
)

// PictureManagementHandler is a struct that holds a reference to the service layer
type PictureManagementHandler struct {
	service service.PictureManagementService
}

// NewPictureManagementHandler creates a new PictureManagementHandler
func NewPictureManagementHandler(svc service.PictureManagementService) *PictureManagementHandler {
	return &PictureManagementHandler{
		service: svc,
	}
}

// ListPicturesHandler handles listing all pictures
func (h *PictureManagementHandler) ListPicturesHandler(w http.ResponseWriter, r *http.Request) {
	pictures, err := h.service.ListAllPictures(r.Context())
	if err != nil {
		slog.Error("Failed to list latest revisions of all files", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to list latest revisions")
		return
	}

	respondJson(w, pictures)
}

// ListRevisionHandler handles listing all revisions for a picture
func (h *PictureManagementHandler) ListRevisionHandler(w http.ResponseWriter, r *http.Request) {
	pictureID := chi.URLParam(r, "pictureID")

	revisions, err := h.service.ListAllRevisions(r.Context(), pictureID)
	if err != nil {
		slog.Error("Failed to list all revisions for a picture", "error", err, "pictureID", pictureID)
		respondWithError(w, http.StatusInternalServerError, "Failed to list all revisions for a picture")
		return
	}

	respondJson(w, revisions)
}
