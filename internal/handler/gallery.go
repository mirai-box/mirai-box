package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mirai-box/mirai-box/internal/service"
)

type GalleryHandler struct {
	service service.GalleryService
}

func NewGalleryHandler(service service.GalleryService) *GalleryHandler {
	return &GalleryHandler{service: service}
}

// CreateGallery handles the creation of a new gallery
func (h *GalleryHandler) CreateGallery(w http.ResponseWriter, r *http.Request) {
	slog.Debug("handler.CreateGallery")

	var req struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to decode request")
		return
	}

	gallery, err := h.service.CreateGallery(req.Title)
	if err != nil {
		slog.Error("Failed to create gallery", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create gallery")
		return
	}

	slog.Info("Gallery created", "Title", req.Title)
	respondJson(w, gallery)
}

// AddImageToGallery handles adding an image to a gallery
func (h *GalleryHandler) AddImageToGallery(w http.ResponseWriter, r *http.Request) {
	galleryID := chi.URLParam(r, "galleryID")

	var req struct {
		ArtID      string `json:"artID"`
		RevisionID string `json:"revisionID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", "error", err, "galleryID", galleryID)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	gallery, err := h.service.AddImageToGallery(galleryID, req.RevisionID)
	if err != nil {
		slog.Error("Failed to add image to gallery", "error", err, "galleryID", galleryID, "RevisionID", req.RevisionID)
		respondWithError(w, http.StatusInternalServerError, "Failed to add image to gallery")
		return
	}

	slog.Info("Image added to gallery", "galleryID", galleryID, "RevisionID", req.RevisionID)
	respondJson(w, gallery)
}

// PublishGallery handles publishing a gallery
func (h *GalleryHandler) PublishGallery(w http.ResponseWriter, r *http.Request) {
	galleryID := chi.URLParam(r, "galleryID")

	if err := h.service.PublishGallery(galleryID); err != nil {
		slog.Error("Failed to publish gallery", "error", err, "galleryID", galleryID)
		respondWithError(w, http.StatusInternalServerError, "Failed to publish gallery")
		return
	}

	slog.Info("Gallery published", "galleryID", galleryID)
	w.WriteHeader(http.StatusOK)
}

// ListGalleries handles listing all galleries
func (h *GalleryHandler) ListGalleries(w http.ResponseWriter, r *http.Request) {
	galleries, err := h.service.ListGalleries()
	if err != nil {
		slog.Error("Failed to list galleries", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to list galleries")
		return
	}

	slog.Debug("list galleries", "galleries", galleries)

	respondJson(w, galleries)
}

// GetMainGallery handles retrieving the main gallery
func (h *GalleryHandler) GetMainGallery(w http.ResponseWriter, r *http.Request) {
	gallery, err := h.service.GetMainGallery()
	if err != nil {
		slog.Error("Failed to get main gallery images", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get main gallery images")
		return
	}

	respondJson(w, gallery)
}

// GetGalleryByID handles retrieving a gallery by its ID
func (h *GalleryHandler) GetGalleryByID(w http.ResponseWriter, r *http.Request) {
	galleryID := chi.URLParam(r, "galleryID")
	gallery, err := h.service.GetGalleryByID(galleryID)
	if err != nil {
		slog.Error("Failed to get gallery", "error", err, "galleryID", galleryID)
		respondWithError(w, http.StatusInternalServerError, "Failed to get gallery")
		return
	}

	slog.Info("Gallery found", "galleryID", galleryID)
	respondJson(w, gallery)
}

// GetImagesByGalleryIDHandler handles retrieving images by gallery ID
func (h *GalleryHandler) GetImagesByGalleryIDHandler(w http.ResponseWriter, r *http.Request) {
	galleryID := chi.URLParam(r, "galleryID")

	images, err := h.service.GetImagesByGalleryID(galleryID)
	if err != nil {
		slog.Error("Failed to get images", "error", err, "galleryID", galleryID)
		respondWithError(w, http.StatusInternalServerError, "Failed to get images")
		return
	}

	slog.Info("Images found", "galleryID", galleryID)
	respondJson(w, images)
}
