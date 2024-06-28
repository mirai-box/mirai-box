package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *GalleryHandler) CreateGallery(w http.ResponseWriter, r *http.Request) {
	slog.Debug("handler.CreateGallery")

	var req struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode  request", "error", err, "req", req)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	gallery, err := h.service.CreateGallery(req.Title)
	if err != nil {
		slog.Error("Failed to create gallery", "error", err, "req", req)
		http.Error(w, "Failed to create gallery", http.StatusInternalServerError)
		return
	}

	slog.Info("gallery created", "Title", req.Title)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(gallery)
}

func (h *GalleryHandler) AddImageToGallery(w http.ResponseWriter, r *http.Request) {
	galleryID := chi.URLParam(r, "galleryID")
	var req struct {
		ArtID      string `json:"artID"`
		RevisionID string `json:"revisionID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode  request", "error", err, "galleryID", galleryID)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	gallery, err := h.service.AddImageToGallery(galleryID, req.RevisionID)
	if err != nil {
		slog.Error("Failed to add image to gallery", "error", err, "galleryID", galleryID, "RevisionID", req.RevisionID)
		respondWithError(w, http.StatusInternalServerError, "Failed to add image to gallery")
		return
	}

	slog.Info("image added to gallery", "galleryID", galleryID, "RevisionID", req.RevisionID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(gallery)
}

func (h *GalleryHandler) PublishGallery(w http.ResponseWriter, r *http.Request) {
	galleryID := chi.URLParam(r, "galleryID")

	if err := h.service.PublishGallery(galleryID); err != nil {
		msg := "Failed to publish gallery"
		slog.Error(msg, "error", err, "galleryID", galleryID)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	slog.Info("gallery published", "galleryID", galleryID)

	w.WriteHeader(http.StatusOK)
}

func (h *GalleryHandler) ListGalleries(w http.ResponseWriter, r *http.Request) {
	galleries, err := h.service.ListGalleries()
	if err != nil {
		slog.Error("Failed to list galleries", "error", err)
		http.Error(w, "Failed to list galleries", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(galleries)
}

func (h *GalleryHandler) GetGalleryByID(w http.ResponseWriter, r *http.Request) {
	galleryID := chi.URLParam(r, "galleryID")
	gallery, err := h.service.GetGalleryByID(galleryID)
	if err != nil {
		msg := "Failed to get gallery"
		slog.Error(msg, "error", err, "galleryID", galleryID)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	slog.Info("gallery found", "galleryID", galleryID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(gallery)
}

func (h *GalleryHandler) GetImagesByGalleryIDHandler(w http.ResponseWriter, r *http.Request) {
	galleryID := chi.URLParam(r, "galleryID")

	images, err := h.service.GetImagesByGalleryID(galleryID)
	if err != nil {
		msg := "Failed to get images"
		slog.Error(msg, "error", err, "galleryID", galleryID)
		http.Error(w, msg, http.StatusInternalServerError)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	slog.Info("images found", "galleryID", galleryID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}
