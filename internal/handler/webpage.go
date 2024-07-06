package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/service"
)

// NewWebPageHandler creates a new WebPageHandler
func NewWebPageHandler(service service.WebPageService) *WebPageHandler {
	return &WebPageHandler{service: service}
}

// WebPageHandler is a struct that holds a reference to the service layer
type WebPageHandler struct {
	service service.WebPageService
}

// CreateWebPageHandler handles creating a new webpage
func (h *WebPageHandler) CreateWebPageHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("handler.CreateWebPageHandler")

	var req model.WebPage

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", "error", err, "req", req)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	wp, err := h.service.CreateWebPage(r.Context(), req.Title, req.HTML)
	if err != nil {
		slog.Error("Failed to create webpage", "error", err, "req", req)
		respondWithError(w, http.StatusInternalServerError, "Failed to create webpage")
		return
	}

	slog.Info("webpage created", "Title", req.Title)
	respondJson(w, wp)
}

// GetWebPageHandler handles retrieving a webpage by ID
func (h *WebPageHandler) GetWebPageHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("handler.GetWebPageHandler")
	id := chi.URLParam(r, "id") // Extract ID from URL parameters

	wp, err := h.service.GetWebPage(r.Context(), id)
	if err != nil {
		slog.Error("Failed to get webpage", "error", err, "id", id)
		respondWithError(w, http.StatusInternalServerError, "Failed to get webpage")
		return
	}

	if wp == nil {
		slog.Error("Webpage not found", "id", id)
		respondWithError(w, http.StatusNotFound, "Webpage not found")
		return
	}

	respondJson(w, wp)
}

// UpdateWebPageHandler handles updating an existing webpage
func (h *WebPageHandler) UpdateWebPageHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("handler.UpdateWebPageHandler")

	id := chi.URLParam(r, "id")

	var req model.WebPage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode request", "error", err, "req", req)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	wp, err := h.service.UpdateWebPage(r.Context(), id, req.Title, req.HTML)
	if err != nil {
		slog.Error("Failed to update webpage", "error", err, "id", id, "req", req)
		respondWithError(w, http.StatusInternalServerError, "Failed to update webpage")
		return
	}

	slog.Info("webpage updated", "id", id)
	respondJson(w, wp)
}

// DeleteWebPageHandler handles deleting a webpage by ID
func (h *WebPageHandler) DeleteWebPageHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("handler.DeleteWebPageHandler")
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteWebPage(r.Context(), id); err != nil {
		slog.Error("Failed to delete webpage", "error", err, "id", id)
		respondWithError(w, http.StatusInternalServerError, "Failed to delete webpage")
		return
	}

	slog.Info("webpage deleted", "id", id)
	respondJson(w, jsonResponse{Status: "success"})
}

// ListWebPagesHandler handles listing all webpages
func (h *WebPageHandler) ListWebPagesHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("handler.ListWebPagesHandler")

	webpages, err := h.service.ListWebPages(r.Context())
	if err != nil {
		slog.Error("Failed to list webpages", "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to list webpages")
		return
	}

	respondJson(w, webpages)
}
