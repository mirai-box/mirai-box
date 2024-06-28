package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func (h *PictureManagementHandler) ListPicturesHandler(w http.ResponseWriter, r *http.Request) {
	pictures, err := h.service.ListAllPictures()
	if err != nil {
		slog.Error("Failed to list latest revisions of all files", "error", err)
		http.Error(w, "Failed to list latest revisions", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pictures); err != nil {
		slog.Error("Failed to encode result as json", "error", err, "pictures", pictures)
	}
}

func (h *PictureManagementHandler) ListRevisionHandler(w http.ResponseWriter, r *http.Request) {
	pictureID := r.PathValue("pictureID")

	revisions, err := h.service.ListAllRevisions(pictureID)
	if err != nil {
		slog.Error("Failed to list all revisions for a picture", "error", err, "pictureID", pictureID)
		http.Error(w, "Failed to list all revisions for a picture", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(revisions); err != nil {
		slog.Error("Failed to encode result as json", "error", err, "revisions", revisions)
	}
}
