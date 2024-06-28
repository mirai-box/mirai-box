package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mirai-box/mirai-box/internal/model"
)

func (h *WebPageHandler) CreateWebPageHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("handler.CreateWebPageHandler")

	var req model.WebPage

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode  request", "error", err, "req", req)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	wp, err := h.service.CreateWebPage(req.Title, req.HTML)
	if err != nil {
		slog.Error("Failed to create gallery", "error", err, "req", req)
		http.Error(w, "Failed to create gallery", http.StatusInternalServerError)
		return
	}

	slog.Info("gallery created", "Title", req.Title)

	respondWithJSON(w, http.StatusCreated, jsonResponse{Status: "success", Data: wp})
}

func (h *WebPageHandler) GetWebPageHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("handler.GetWebPageHandler")
	id := r.URL.Query().Get("id") // Extract ID from query parameters

	wp, err := h.service.GetWebPage(id)
	if err != nil {
		slog.Error("Failed to get webpage", "error", err, "id", id)
		http.Error(w, "Failed to get webpage", http.StatusInternalServerError)
		return
	}

	if wp == nil {
		slog.Error("Webpage not found", "id", id)
		http.Error(w, "Webpage not found", http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, jsonResponse{Status: "success", Data: wp})
}
