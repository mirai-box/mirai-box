package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/service"
)

// ArtProjectManagementHandler handles art project-related requests
type ArtProjectManagementHandler struct {
	managementService service.ArtProjectManagementServiceInterface
	artService        service.ArtProjectServiceInterface
}

// NewArtProjectManagementHandler creates a new ArtProjectManagementHandler
func NewArtProjectManagementHandler(
	managementService service.ArtProjectManagementServiceInterface,
	artService service.ArtProjectServiceInterface,
) *ArtProjectManagementHandler {
	return &ArtProjectManagementHandler{
		managementService: managementService,
		artService:        artService,
	}
}

// CreateArtProject handles the creation of a new art project
func (h *ArtProjectManagementHandler) CreateArtProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized user attempt to create art project")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	title := r.FormValue("title")
	if title == "" {
		slog.WarnContext(ctx, "Attempt to create art project with empty title", "userID", user.ID)
		SendErrorResponse(w, http.StatusBadRequest, "Title is required")
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		slog.ErrorContext(ctx, "Invalid file upload", "error", err, "userID", user.ID)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid file upload")
		return
	}
	defer file.Close()

	slog.InfoContext(ctx, "Creating new art project", "title", title, "userID", user.ID, "filename", handler.Filename)

	artProject, err := h.managementService.CreateArtProjectAndRevision(ctx, user.ID.String(), file, title, handler.Filename)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create art project", "error", err, "userID", user.ID, "title", title)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create art project")
		return
	}

	response := convertToArtProjectResponse(artProject)

	slog.InfoContext(ctx, "Art project created successfully", "artProjectID", artProject.ID, "userID", user.ID)
	SendJSONResponse(w, http.StatusCreated, response)
}

// AddRevision handles adding a new revision to an existing art project
func (h *ArtProjectManagementHandler) AddRevision(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized user attempt to add revision")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	artProjectID := chi.URLParam(r, "id")
	comment := r.FormValue("comment")

	file, handler, err := r.FormFile("file")
	if err != nil {
		slog.ErrorContext(ctx, "Invalid file upload for revision", "error", err, "userID", user.ID, "artProjectID", artProjectID)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid file upload")
		return
	}
	defer file.Close()

	slog.InfoContext(ctx, "Adding new revision", "artProjectID", artProjectID, "userID", user.ID, "filename", handler.Filename)

	revision, err := h.managementService.AddRevision(ctx, user.ID.String(), artProjectID, file, comment, handler.Filename)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to add revision", "error", err, "userID", user.ID, "artProjectID", artProjectID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to add revision")
		return
	}

	response := convertToRevisionResponse(revision)

	slog.InfoContext(ctx, "Revision added successfully", "revisionID", revision.ID, "artProjectID", artProjectID, "userID", user.ID)
	SendJSONResponse(w, http.StatusCreated, response)
}

// ListRevisions handles listing all revisions for an art project
func (h *ArtProjectManagementHandler) ListRevisions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized user attempt to list revisions")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	artID := chi.URLParam(r, "id")

	slog.InfoContext(ctx, "Listing revisions for art project", "artID", artID, "userID", user.ID)

	revisions, err := h.managementService.ListAllRevisions(ctx, artID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find revisions for art project", "error", err, "artID", artID, "userID", user.ID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to find revisions for art project")
		return
	}

	if len(revisions) == 0 {
		slog.InfoContext(ctx, "No revisions found for art project", "artID", artID, "userID", user.ID)
		SendJSONResponse(w, http.StatusOK, []models.RevisionResponse{})
		return
	}

	if revisions[0].ArtProject.UserID != user.ID {
		slog.WarnContext(ctx, "User not authorized to view revisions", "userID", user.ID, "artProjectUserID", revisions[0].ArtProject.UserID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	response := make([]models.RevisionResponse, len(revisions))
	for i, rev := range revisions {
		response[i] = convertToRevisionResponse(&rev)
	}

	slog.InfoContext(ctx, "Successfully retrieved revisions", "artID", artID, "userID", user.ID, "revisionCount", len(revisions))
	SendJSONResponse(w, http.StatusOK, response)
}

// MyArtProjects handles listing all art projects for the authenticated user
func (h *ArtProjectManagementHandler) MyArtProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized user attempt to list art projects")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	slog.InfoContext(ctx, "Listing art projects for user", "userID", user.ID)

	artProjects, err := h.artService.FindByUserID(ctx, user.ID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list user's art projects", "error", err, "userID", user.ID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to list user's art projects")
		return
	}

	response := make([]models.ArtProjectResponse, len(artProjects))
	for i, project := range artProjects {
		response[i] = convertToArtProjectResponse(&project)
	}

	slog.InfoContext(ctx, "Successfully retrieved art projects", "userID", user.ID, "projectCount", len(artProjects))
	SendJSONResponse(w, http.StatusOK, response)
}

// MyArtProjectByID handles retrieving a specific art project for the authenticated user
func (h *ArtProjectManagementHandler) MyArtProjectByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized user attempt to retrieve art project")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	artID := chi.URLParam(r, "id")

	slog.InfoContext(ctx, "Retrieving art project", "artID", artID, "userID", user.ID)

	artProject, err := h.artService.FindByID(ctx, artID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find user's art project", "error", err, "artID", artID, "userID", user.ID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to find user's art project")
		return
	}

	if artProject.UserID != user.ID {
		slog.WarnContext(ctx, "User not authorized to view art project", "userID", user.ID, "artProjectUserID", artProject.UserID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	response := convertToArtProjectResponse(artProject)

	slog.InfoContext(ctx, "Successfully retrieved art project", "artID", artID, "userID", user.ID)
	SendJSONResponse(w, http.StatusOK, response)
}

func convertToArtProjectResponse(artProject *models.ArtProject) models.ArtProjectResponse {
	response := models.ArtProjectResponse{
		ID:                  artProject.ID,
		Title:               artProject.Title,
		CreatedAt:           artProject.CreatedAt,
		UpdatedAt:           artProject.UpdatedAt,
		ContentType:         artProject.ContentType,
		Filename:            artProject.Filename,
		Public:              artProject.Public,
		LatestRevisionID:    artProject.LatestRevisionID,
		PublishedRevisionID: artProject.PublishedRevisionID,
		Tags:                make([]models.TagResponse, len(artProject.Tags)),
		StashID:             artProject.StashID,
		UserID:              artProject.UserID,
	}

	for i, tag := range artProject.Tags {
		response.Tags[i] = models.TagResponse{
			ID:   tag.ID,
			Name: tag.Name,
		}
	}

	return response
}

func convertToRevisionResponse(revision *models.Revision) models.RevisionResponse {
	return models.RevisionResponse{
		ID:           revision.ID,
		ArtID:        revision.ArtID,
		Version:      revision.Version,
		CreatedAt:    revision.CreatedAt,
		Comment:      revision.Comment,
		Size:         revision.Size,
		ArtProjectID: revision.ArtProjectID,
		UserID:       revision.UserID,
	}
}
