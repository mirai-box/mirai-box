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
	ctx := r.Context()
	var webPageRequest models.WebPage
	if err := json.NewDecoder(r.Body).Decode(&webPageRequest); err != nil {
		slog.ErrorContext(ctx, "Invalid json in request body", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized attempt to create web page")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	slog.InfoContext(ctx, "Creating web page", "user", user.Username)

	webPageRequest.UserID = user.ID

	createdWebPage, err := h.webPageService.CreateWebPage(ctx, &webPageRequest)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create web page", "error", err, "userID", user.ID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create web page")
		return
	}

	response := models.WebPageResponse{
		ID:        createdWebPage.ID,
		UserID:    createdWebPage.UserID,
		Title:     createdWebPage.Title,
		Html:      createdWebPage.Html,
		PageType:  createdWebPage.PageType,
		Public:    createdWebPage.Public,
		CreatedAt: createdWebPage.CreatedAt,
		UpdatedAt: createdWebPage.UpdatedAt,
	}

	slog.InfoContext(ctx, "Web page created successfully", "webPageID", createdWebPage.ID, "userID", user.ID)
	SendJSONResponse(w, http.StatusCreated, response)
}

func (h *WebPageHandler) GetWebPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	webPageID, err := uuid.Parse(id)
	if err != nil {
		slog.WarnContext(ctx, "Invalid web page ID", "id", id)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid web page ID")
		return
	}

	webPage, err := h.webPageService.GetWebPage(ctx, webPageID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get web page", "error", err, "webPageID", webPageID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to get web page")
		return
	}

	response := models.WebPageResponse{
		ID:        webPage.ID,
		UserID:    webPage.UserID,
		Title:     webPage.Title,
		Html:      webPage.Html,
		PageType:  webPage.PageType,
		Public:    webPage.Public,
		CreatedAt: webPage.CreatedAt,
		UpdatedAt: webPage.UpdatedAt,
	}

	slog.InfoContext(ctx, "Web page retrieved successfully", "webPageID", webPageID)
	SendJSONResponse(w, http.StatusOK, response)
}

func (h *WebPageHandler) UpdateWebPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	webPageID, err := uuid.Parse(id)
	if err != nil {
		slog.WarnContext(ctx, "Invalid web page ID", "id", id)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid web page ID")
		return
	}

	var updatedWebPage models.WebPage
	if err := json.NewDecoder(r.Body).Decode(&updatedWebPage); err != nil {
		slog.ErrorContext(ctx, "Invalid json in request body", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedWebPage.ID = webPageID

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized attempt to update web page", "webPageID", webPageID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	existingWebPage, err := h.webPageService.GetWebPage(ctx, webPageID.String())
	if err != nil || existingWebPage.UserID != user.ID {
		slog.ErrorContext(ctx, "User is not the owner of the web page",
			"existingWebPage.UserID", existingWebPage.UserID,
			"webPageID", webPageID,
			"userID", user.ID,
		)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Merge existing and updated web page data
	updatedWebPage = mergeWebPageData(*existingWebPage, updatedWebPage)

	webPage, err := h.webPageService.UpdateWebPage(ctx, &updatedWebPage)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update web page", "error", err, "webPageID", webPageID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to update web page")
		return
	}

	response := models.WebPageResponse{
		ID:        webPage.ID,
		UserID:    webPage.UserID,
		Title:     webPage.Title,
		Html:      webPage.Html,
		PageType:  webPage.PageType,
		Public:    webPage.Public,
		CreatedAt: webPage.CreatedAt,
		UpdatedAt: webPage.UpdatedAt,
	}

	slog.InfoContext(ctx, "Web page updated successfully", "webPageID", webPageID, "userID", user.ID)
	SendJSONResponse(w, http.StatusOK, response)
}

func (h *WebPageHandler) DeleteWebPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	webPageID, err := uuid.Parse(id)
	if err != nil {
		slog.WarnContext(ctx, "Invalid web page ID", "id", id)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid web page ID")
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized attempt to delete web page", "webPageID", webPageID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	existingWebPage, err := h.webPageService.GetWebPage(ctx, webPageID.String())
	if err != nil || existingWebPage.UserID != user.ID {
		slog.ErrorContext(ctx, "User is not the owner of the web page",
			"existingWebPage.UserID", existingWebPage.UserID,
			"webPageID", webPageID,
			"userID", user.ID,
		)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.webPageService.DeleteWebPage(ctx, webPageID.String()); err != nil {
		slog.ErrorContext(ctx, "Failed to delete web page", "error", err, "webPageID", webPageID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to delete web page")
		return
	}

	slog.InfoContext(ctx, "Web page deleted successfully", "webPageID", webPageID, "userID", user.ID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *WebPageHandler) ListWebPages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	webPages, err := h.webPageService.ListWebPages(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list web pages", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to list web pages")
		return
	}

	response := make([]models.WebPageResponse, len(webPages))
	for i, webPage := range webPages {
		response[i] = models.WebPageResponse{
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

	slog.InfoContext(ctx, "Web pages listed successfully", "count", len(webPages))
	SendJSONResponse(w, http.StatusOK, response)
}

func (h *WebPageHandler) ListUserWebPages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userId")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.WarnContext(ctx, "Invalid user ID", "userID", userID)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	webPages, err := h.webPageService.ListUserWebPages(ctx, userUUID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list user web pages", "error", err, "userID", userID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to list user web pages")
		return
	}

	response := make([]models.WebPageResponse, len(webPages))
	for i, webPage := range webPages {
		response[i] = models.WebPageResponse{
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

	slog.InfoContext(ctx, "User web pages listed successfully", "userID", userID, "count", len(webPages))
	SendJSONResponse(w, http.StatusOK, response)
}

func (h *WebPageHandler) MyWebPages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized attempt to list user's web pages")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	webPages, err := h.webPageService.ListUserWebPages(ctx, user.ID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list user's web pages", "error", err, "userID", user.ID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to list user's web pages")
		return
	}

	response := make([]models.WebPageResponse, len(webPages))
	for i, webPage := range webPages {
		response[i] = models.WebPageResponse{
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

	slog.InfoContext(ctx, "User's web pages listed successfully", "userID", user.ID, "count", len(webPages))
	SendJSONResponse(w, http.StatusOK, response)
}

func (h *WebPageHandler) MyWebPageByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	webPageID, err := uuid.Parse(id)
	if err != nil {
		slog.WarnContext(ctx, "Invalid web page ID", "id", id)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid web page ID")
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized attempt to get user's web page", "webPageID", webPageID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	webPage, err := h.webPageService.GetWebPage(ctx, webPageID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get web page", "error", err, "webPageID", webPageID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to get web page")
		return
	}

	if webPage.UserID != user.ID {
		slog.WarnContext(ctx, "User attempted to access unauthorized web page", "webPageID", webPageID, "userID", user.ID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	response := models.WebPageResponse{
		ID:        webPage.ID,
		UserID:    webPage.UserID,
		Title:     webPage.Title,
		Html:      webPage.Html,
		PageType:  webPage.PageType,
		Public:    webPage.Public,
		CreatedAt: webPage.CreatedAt,
		UpdatedAt: webPage.UpdatedAt,
	}

	slog.InfoContext(ctx, "User's web page retrieved successfully", "webPageID", webPageID, "userID", user.ID)
	SendJSONResponse(w, http.StatusOK, response)
}

func mergeWebPageData(existing, updated models.WebPage) models.WebPage {
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
