package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/service"
)

// WebPageHandler handles HTTP requests related to web page operations.
type WebPageHandler struct {
	webPageService service.WebPageService
}

// NewWebPageHandler creates a new instance of WebPageHandler.
func NewWebPageHandler(webPageService service.WebPageService) *WebPageHandler {
	return &WebPageHandler{webPageService: webPageService}
}

// CreateWebPage handles the creation of a new web page.
func (h *WebPageHandler) CreateWebPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "CreateWebPage")

	var webPageRequest model.WebPage
	if err := json.NewDecoder(r.Body).Decode(&webPageRequest); err != nil {
		logger.Error("Invalid json in request body", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("Unauthorized attempt to create web page")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	logger = logger.With("userID", user.ID)
	logger.Info("Creating web page")

	webPageRequest.UserID = user.ID

	createdWebPage, err := h.webPageService.CreateWebPage(ctx, &webPageRequest)
	if err != nil {
		logger.Error("Failed to create web page", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create web page")
		return
	}

	logger.Info("Web page created successfully", "webPageID", createdWebPage.ID)
	SendJSONResponse(w, http.StatusCreated, convertToWebPageResponse(createdWebPage))
}

// GetWebPage retrieves a specific web page.
func (h *WebPageHandler) GetWebPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "GetWebPage")

	webPageID := chi.URLParam(r, "id")

	webPage, err := h.webPageService.GetWebPage(ctx, webPageID)
	if err != nil {
		if errors.Is(err, model.ErrWebPageNotFound) {
			logger.Info("Web page not found", "webPageID", webPageID)
			SendErrorResponse(w, http.StatusNotFound, "Web page not found")
		} else {
			logger.Error("Failed to get web page", "error", err, "webPageID", webPageID)
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to get web page")
		}
		return
	}

	logger.Info("Web page retrieved successfully", "webPageID", webPageID)
	SendJSONResponse(w, http.StatusOK, convertToWebPageResponse(webPage))
}

// UpdateWebPage handles updating an existing web page.
func (h *WebPageHandler) UpdateWebPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "UpdateWebPage")

	webPageID := chi.URLParam(r, "id")

	var updatedWebPage model.WebPage
	if err := json.NewDecoder(r.Body).Decode(&updatedWebPage); err != nil {
		logger.Error("Invalid json in request body", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedWebPage.ID = uuid.MustParse(webPageID)

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("Unauthorized attempt to update web page", "webPageID", webPageID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	existingWebPage, err := h.webPageService.GetWebPage(ctx, webPageID)
	if err != nil {
		if errors.Is(err, model.ErrWebPageNotFound) {
			logger.Info("Web page not found", "webPageID", webPageID)
			SendErrorResponse(w, http.StatusNotFound, "Web page not found")
		} else {
			logger.Error("Error retrieving web page", "error", err)
			SendErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	if existingWebPage.UserID != user.ID {
		logger.Error("User is not the owner of the web page", "existingWebPage.UserID", existingWebPage.UserID, "userID", user.ID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	updatedWebPage = mergeWebPageData(*existingWebPage, updatedWebPage)

	webPage, err := h.webPageService.UpdateWebPage(ctx, &updatedWebPage)
	if err != nil {
		logger.Error("Failed to update web page", "error", err, "webPageID", webPageID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to update web page")
		return
	}

	logger.Info("Web page updated successfully", "webPageID", webPageID, "userID", user.ID)
	SendJSONResponse(w, http.StatusOK, convertToWebPageResponse(webPage))
}

// DeleteWebPage handles the deletion of a web page.
func (h *WebPageHandler) DeleteWebPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "DeleteWebPage")

	id := chi.URLParam(r, "id")
	webPageID, err := uuid.Parse(id)
	if err != nil {
		logger.Warn("Invalid web page ID", "id", id)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid web page ID")
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("Unauthorized attempt to delete web page", "webPageID", webPageID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	existingWebPage, err := h.webPageService.GetWebPage(ctx, webPageID.String())
	if err != nil {
		if errors.Is(err, model.ErrWebPageNotFound) {
			logger.Info("Web page not found", "webPageID", webPageID)
			SendErrorResponse(w, http.StatusNotFound, "Web page not found")
		} else {
			logger.Error("Error retrieving web page", "error", err)
			SendErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	if existingWebPage.UserID != user.ID {
		logger.Error("User is not the owner of the web page", "existingWebPage.UserID", existingWebPage.UserID, "userID", user.ID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.webPageService.DeleteWebPage(ctx, webPageID.String()); err != nil {
		logger.Error("Failed to delete web page", "error", err, "webPageID", webPageID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to delete web page")
		return
	}

	logger.Info("Web page deleted successfully", "webPageID", webPageID, "userID", user.ID)
	w.WriteHeader(http.StatusNoContent)
}

// ListWebPages handles listing all web pages.
func (h *WebPageHandler) ListWebPages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "ListWebPages")

	webPages, err := h.webPageService.ListWebPages(ctx)
	if err != nil {
		if errors.Is(err, model.ErrWebPageNotFound) {
			logger.Info("No web pages found")
			SendErrorResponse(w, http.StatusNotFound, "No web pages found")
		} else {
			logger.Error("Failed to list web pages", "error", err)
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to list web pages")
		}
		return
	}

	response := make([]model.WebPageResponse, len(webPages))
	for i, webPage := range webPages {
		response[i] = convertToWebPageResponse(&webPage)
	}

	logger.Info("Web pages listed successfully", "count", len(webPages))
	SendJSONResponse(w, http.StatusOK, response)
}

// ListUserWebPages handles listing web pages for a specific user.
func (h *WebPageHandler) ListUserWebPages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "ListUserWebPages")

	userID := chi.URLParam(r, "userId")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		logger.Warn("Invalid user ID", "userID", userID)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	webPages, err := h.webPageService.ListUserWebPages(ctx, userUUID.String())
	if err != nil {
		if errors.Is(err, model.ErrWebPageNotFound) {
			logger.Info("No web pages found for user", "userID", userID)
			SendErrorResponse(w, http.StatusNotFound, "No web pages found for this user")
		} else {
			logger.Error("Failed to list user web pages", "error", err, "userID", userID)
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to list user web pages")
		}
		return
	}

	response := make([]model.WebPageResponse, len(webPages))
	for i, webPage := range webPages {
		response[i] = convertToWebPageResponse(&webPage)
	}

	logger.Info("User web pages listed successfully", "userID", userID, "count", len(webPages))
	SendJSONResponse(w, http.StatusOK, response)
}

// MyWebPages handles listing web pages for the authenticated user.
func (h *WebPageHandler) MyWebPages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "MyWebPages")

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("Unauthorized attempt to list user's web pages")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	webPages, err := h.webPageService.ListUserWebPages(ctx, user.ID.String())
	if err != nil {
		if errors.Is(err, model.ErrWebPageNotFound) {
			logger.Info("No web pages found for user", "userID", user.ID)
			SendErrorResponse(w, http.StatusNotFound, "No web pages found for this user")
		} else {
			logger.Error("Failed to list user's web pages", "error", err, "userID", user.ID)
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to list user's web pages")
		}
		return
	}

	response := make([]model.WebPageResponse, len(webPages))
	for i, webPage := range webPages {
		response[i] = convertToWebPageResponse(&webPage)
	}

	logger.Info("User's web pages listed successfully", "userID", user.ID, "count", len(webPages))
	SendJSONResponse(w, http.StatusOK, response)
}

// MyWebPageByID handles retrieving a specific web page for the authenticated user.
func (h *WebPageHandler) MyWebPageByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "MyWebPageByID")

	id := chi.URLParam(r, "id")
	webPageID, err := uuid.Parse(id)
	if err != nil {
		logger.Warn("Invalid web page ID", "id", id)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid web page ID")
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("Unauthorized attempt to get user's web page", "webPageID", webPageID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	webPage, err := h.webPageService.GetWebPage(ctx, webPageID.String())
	if err != nil {
		if errors.Is(err, model.ErrWebPageNotFound) {
			logger.Info("Web page not found", "webPageID", webPageID)
			SendErrorResponse(w, http.StatusNotFound, "Web page not found")
		} else {
			logger.Error("Failed to get web page", "error", err, "webPageID", webPageID)
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to get web page")
		}
		return
	}

	if webPage.UserID != user.ID {
		logger.Warn("User attempted to access unauthorized web page", "webPageID", webPageID, "userID", user.ID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	logger.Info("User's web page retrieved successfully", "webPageID", webPageID, "userID", user.ID)
	SendJSONResponse(w, http.StatusOK, convertToWebPageResponse(webPage))
}

// Helper functions

func mergeWebPageData(existing, updated model.WebPage) model.WebPage {
	if updated.Title != "" {
		existing.Title = updated.Title
	}
	if updated.Html != "" {
		existing.Html = updated.Html
	}
	if updated.PageType != "" {
		existing.PageType = updated.PageType
	}
	existing.Public = updated.Public
	return existing
}

func convertToWebPageResponse(webPage *model.WebPage) model.WebPageResponse {
	return model.WebPageResponse{
		ID:        webPage.ID,
		UserID:    webPage.UserID,
		Title:     webPage.Title,
		Html:      webPage.Html,
		PageType:  webPage.PageType,
		Public:    webPage.Public,
		CreatedAt: webPage.CreatedAt,
		UpdatedAt: webPage.UpdatedAt,
	}
}
