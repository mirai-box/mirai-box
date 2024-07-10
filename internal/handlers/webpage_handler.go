package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/service"
)

type WebPageHandler struct {
	webPageService service.WebPageServiceInterface
}

func NewWebPageHandler(webPageService service.WebPageServiceInterface) *WebPageHandler {
	return &WebPageHandler{webPageService: webPageService}
}

func (h *WebPageHandler) CreateWebPage(w http.ResponseWriter, r *http.Request) {
	var webPage models.WebPage
	if err := json.NewDecoder(r.Body).Decode(&webPage); err != nil {
		slog.Error("Invalid json in request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	slog.Info("session user found", "user", user.Username)

	webPage.UserID = user.ID

	createdWebPage, err := h.webPageService.CreateWebPage(r.Context(), &webPage)
	if err != nil {
		slog.Error("Failed to create web page", "error", err)
		http.Error(w, "Failed to create web page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdWebPage)
}

func (h *WebPageHandler) GetWebPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	webPageID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid web page ID", http.StatusBadRequest)
		return
	}

	webPage, err := h.webPageService.GetWebPage(r.Context(), webPageID.String())
	if err != nil {
		slog.Error("Failed to get web page", "error", err, "webPageID", webPageID)
		http.Error(w, "Failed to get web page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webPage)
}

func (h *WebPageHandler) UpdateWebPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	webPageID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid web page ID", http.StatusBadRequest)
		return
	}

	var updatedWebPage models.WebPage
	if err := json.NewDecoder(r.Body).Decode(&updatedWebPage); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedWebPage.ID = webPageID

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the web page belongs to the user
	existingWebPage, err := h.webPageService.GetWebPage(r.Context(), webPageID.String())
	if err != nil || existingWebPage.UserID != user.ID {
		slog.Error("User is not the owner of the web page",
			"existingWebPage.UserID", existingWebPage.UserID,
			"webPageID", webPageID,
			"userID", user.ID,
		)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	updatedWebPage.CreatedAt = existingWebPage.CreatedAt
	if updatedWebPage.Title == "" {
		updatedWebPage.Title = existingWebPage.Title
	}

	if updatedWebPage.Html == "" {
		updatedWebPage.Html = existingWebPage.Html
	}

	if updatedWebPage.PageType == "" {
		updatedWebPage.PageType = existingWebPage.PageType
	}

	updatedWebPage.User = existingWebPage.User
	updatedWebPage.UserID = existingWebPage.UserID

	webPage, err := h.webPageService.UpdateWebPage(r.Context(), &updatedWebPage)
	if err != nil {
		slog.Error("Failed to update web page", "error", err, "webPageID", webPageID)
		http.Error(w, "Failed to update web page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webPage)
}

func (h *WebPageHandler) DeleteWebPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	webPageID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid web page ID", http.StatusBadRequest)
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the web page belongs to the user
	existingWebPage, err := h.webPageService.GetWebPage(r.Context(), webPageID.String())
	if err != nil || existingWebPage.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.webPageService.DeleteWebPage(r.Context(), webPageID.String()); err != nil {
		slog.Error("Failed to delete web page", "error", err, "webPageID", webPageID)
		http.Error(w, "Failed to delete web page", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WebPageHandler) ListWebPages(w http.ResponseWriter, r *http.Request) {
	webPages, err := h.webPageService.ListWebPages(r.Context())
	if err != nil {
		slog.Error("Failed to list web pages", "error", err)
		http.Error(w, "Failed to list web pages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webPages)
}

func (h *WebPageHandler) ListUserWebPages(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	webPages, err := h.webPageService.ListUserWebPages(r.Context(), userUUID.String())
	if err != nil {
		slog.Error("Failed to list user web pages", "error", err, "userID", userID)
		http.Error(w, "Failed to list user web pages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webPages)
}

func (h *WebPageHandler) MyWebPages(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	webPages, err := h.webPageService.ListUserWebPages(r.Context(), user.ID.String())
	if err != nil {
		slog.Error("Failed to list user's web pages", "error", err, "userID", user.ID)
		http.Error(w, "Failed to list user's web pages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webPages)
}

func (h *WebPageHandler) MyWebPageByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	webPageID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid web page ID", http.StatusBadRequest)
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	webPage, err := h.webPageService.GetWebPage(r.Context(), webPageID.String())
	if err != nil {
		slog.Error("Failed to get web page", "error", err, "webPageID", webPageID)
		http.Error(w, "Failed to get web page", http.StatusInternalServerError)
		return
	}

	if webPage.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webPage)
}
