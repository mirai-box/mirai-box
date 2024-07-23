package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/service"
)

// ArtProjectHandler handles HTTP requests related to art projects.
type ArtProjectHandler struct {
	artProjectService service.ArtProjectService
}

// NewArtProjectHandler creates a new ArtProjectHandler instance.
func NewArtProjectHandler(aps service.ArtProjectService) *ArtProjectHandler {
	return &ArtProjectHandler{
		artProjectService: aps,
	}
}

// CreateArtProject handles the creation of a new art project.
func (h *ArtProjectHandler) CreateArtProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "CreateArtProject")

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("Unauthorized user attempt to create art project")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	title := r.FormValue("title")
	if title == "" {
		logger.Warn("Attempt to create art project with empty title", "userID", user.ID)
		SendErrorResponse(w, http.StatusBadRequest, "Title is required")
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		logger.Error("Invalid file upload", "error", err, "userID", user.ID)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid file upload")
		return
	}
	defer file.Close()

	logger = logger.With("title", title, "userID", user.ID, "filename", handler.Filename)
	logger.Info("Creating new art project")

	contentType := detectContentType(file)
	fileData, err := prepareFileData(file)
	if err != nil {
		logger.Error("Failed to prepare file data", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to process file")
		return
	}

	logger.Info("content type is detected as", "contentType", contentType)

	artProject := &model.ArtProject{
		ID:          uuid.New(),
		Title:       title,
		Filename:    handler.Filename,
		UserID:      user.ID,
		ContentType: contentType,
	}

	if err := h.artProjectService.CreateArtProject(ctx, artProject); err != nil {
		logger.Error("Failed to create art project", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create art project")
		return
	}

	revision := &model.Revision{
		ID:           uuid.New(),
		ArtProjectID: artProject.ID,
		UserID:       user.ID,
		Comment:      title,
		CreatedAt:    time.Now(),
	}

	if err := h.artProjectService.AddRevision(ctx, revision, fileData); err != nil {
		logger.Error("Failed to add revision", "error", err, "artProjectID", artProject.ID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to add revision")
		return
	}

	response := convertToArtProjectResponse(artProject)
	logger.Info("Art project created successfully", "artProjectID", artProject.ID)
	SendJSONResponse(w, http.StatusCreated, response)
}

// AddRevision handles adding a new revision to an existing art project.
func (h *ArtProjectHandler) AddRevision(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	artProjectID := chi.URLParam(r, "artID")
	logger := slog.With("handler", "AddRevision", "artProjectID", artProjectID)

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("Unauthorized user attempt to add revision")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	comment := r.FormValue("comment")
	file, handler, err := r.FormFile("file")
	if err != nil {
		logger.Error("Invalid file upload for revision", "error", err, "userID", user.ID, "artProjectID", artProjectID)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid file upload")
		return
	}
	defer file.Close()

	logger = logger.With("userID", user.ID, "filename", handler.Filename)

	revision := &model.Revision{
		ID:           uuid.New(),
		ArtProjectID: uuid.MustParse(artProjectID),
		UserID:       user.ID,
		Comment:      comment,
	}

	if err := h.artProjectService.AddRevision(ctx, revision, file); err != nil {
		logger.Error("Failed to add revision", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to add revision")
		return
	}

	response := convertToRevisionResponse(revision)
	logger.Info("Revision added successfully", "revisionID", revision.ID)
	SendJSONResponse(w, http.StatusCreated, response)
}

// ListRevisions handles listing all revisions for an art project.
func (h *ArtProjectHandler) ListRevisions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "ListRevisions")

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok || user == nil {
		logger.Warn("Unauthorized user attempt to list revisions")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	artID := chi.URLParam(r, "artID")
	logger = logger.With("artID", artID, "userID", user.ID)
	logger.Info("Listing revisions for art project")

	revisions, err := h.artProjectService.ListRevisions(ctx, artID)
	if err != nil {
		logger.Error("Failed to find revisions for art project", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to find revisions for art project")
		return
	}

	if len(revisions) == 0 {
		logger.Info("No revisions found for art project")
		SendJSONResponse(w, http.StatusOK, []model.RevisionResponse{})
		return
	}

	if revisions[0].UserID != user.ID {
		logger.Warn("User not authorized to view revisions", "artProjectUserID", revisions[0].UserID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	response := make([]model.RevisionResponse, len(revisions))
	for i, rev := range revisions {
		response[i] = convertToRevisionResponse(&rev)
	}

	logger.Info("Successfully retrieved revisions", "revisionCount", len(revisions))
	SendJSONResponse(w, http.StatusOK, response)
}

// MyArtProjects handles listing all art projects for the authenticated user.
func (h *ArtProjectHandler) MyArtProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "MyArtProjects")

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok || user == nil {
		logger.Warn("Unauthorized user attempt to list art projects")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	logger = logger.With("userID", user.ID)
	logger.Info("Listing art projects for user")

	artProjects, err := h.artProjectService.FindByUserID(ctx, user.ID.String())
	if err != nil {
		if errors.Is(err, model.ErrArtProjectNotFound) {
			logger.Info("No art projects found for user")
			SendJSONResponse(w, http.StatusOK, []model.ArtProjectResponse{})
			return
		}

		logger.Error("Failed to list user's art projects", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to list user's art projects")
		return
	}

	response := make([]model.ArtProjectResponse, len(artProjects))
	for i, project := range artProjects {
		response[i] = convertToArtProjectResponse(&project)
	}

	logger.Info("Successfully retrieved art projects", "projectCount", len(artProjects))
	SendJSONResponse(w, http.StatusOK, response)
}

// MyArtProjectByID handles retrieving a specific art project for the authenticated user.
func (h *ArtProjectHandler) MyArtProjectByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "MyArtProjectByID")

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("Unauthorized user attempt to retrieve art project")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	artID := chi.URLParam(r, "artID")
	logger = logger.With("artID", artID, "userID", user.ID)
	logger.Info("Retrieving art project")

	artProject, err := h.artProjectService.FindByID(ctx, artID)
	if err != nil {
		if errors.Is(err, model.ErrArtProjectNotFound) {
			logger.Info("Art project not found")
			SendErrorResponse(w, http.StatusNotFound, "Art project not found")
			return
		}

		logger.Error("Failed to find user's art project", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to find user's art project")
		return
	}

	if artProject.UserID != user.ID {
		logger.Warn("User not authorized to view art project", "artProjectUserID", artProject.UserID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	response := convertToArtProjectResponse(artProject)
	logger.Info("Successfully retrieved art project")
	SendJSONResponse(w, http.StatusOK, response)
}

// RevisionDownload handles the download of a specific revision.
func (h *ArtProjectHandler) RevisionDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "RevisionDownload")

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("Unauthorized attempt to download revision")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	revisionID := chi.URLParam(r, "revisionID")
	artID := chi.URLParam(r, "artID")
	logger = logger.With("revisionID", revisionID, "artID", artID, "userID", user.ID)
	logger.Info("Retrieving revision for download")

	h.handleDownload(w, r, func() (io.ReadCloser, *model.ArtProject, error) {
		return h.artProjectService.GetArtProjectByRevision(ctx, user.ID.String(), artID, revisionID)
	}, revisionID)
}

// handleDownload is a helper function to handle file downloads.
func (h *ArtProjectHandler) handleDownload(w http.ResponseWriter, _ *http.Request, fetch func() (io.ReadCloser, *model.ArtProject, error), id string) {
	logger := slog.With("handler", "handleDownload", "ID", id)

	fh, pic, err := fetch()
	if err != nil {
		if errors.Is(err, model.ErrArtProjectNotFound) {
			logger.Warn("Art not found", "error", err)
			SendErrorResponse(w, http.StatusNotFound, "Art not found")
			return
		}
		logger.Error("Failed to get picture", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to get picture")
		return
	}
	defer fh.Close()

	w.Header().Set("Content-Type", pic.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", pic.Filename))

	if _, err = io.Copy(w, fh); err != nil {
		logger.Error("Error streaming file", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Error streaming file")
		return
	}

	logger.Info("File downloaded successfully", "filename", pic.Filename)
}

// GetArtByID handles retrieving art by its ID.
func (h *ArtProjectHandler) GetArtByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	artID := chi.URLParam(r, "artID")
	logger := slog.With("handler", "GetArtByID", "artID", artID)
	logger.Info("Retrieving art by ID")

	rev, err := h.artProjectService.GetRevisionByArtID(ctx, artID)
	if err != nil {
		if errors.Is(err, model.ErrArtProjectNotFound) {
			logger.Warn("Art not found", "error", err)
			SendErrorResponse(w, http.StatusNotFound, "Art not found")
			return
		}

		logger.Error("Failed to retrieve art", "error", err)
		SendErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	logger.Info("Art retrieved successfully", "revisionID", rev.ID)
	http.ServeFile(w, r, rev.FilePath)
}

// Helper functions

func detectContentType(file io.Reader) string {
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "application/octet-stream"
	}
	return http.DetectContentType(buffer)
}

func prepareFileData(file io.Reader) (io.Reader, error) {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return io.MultiReader(bytes.NewReader(buffer[:n]), file), nil
}

func convertToArtProjectResponse(artProject *model.ArtProject) model.ArtProjectResponse {
	response := model.ArtProjectResponse{
		ID:                  artProject.ID,
		Title:               artProject.Title,
		CreatedAt:           artProject.CreatedAt,
		UpdatedAt:           artProject.UpdatedAt,
		ContentType:         artProject.ContentType,
		Filename:            artProject.Filename,
		Public:              artProject.Public,
		LatestRevisionID:    artProject.LatestRevisionID,
		PublishedRevisionID: artProject.PublishedRevisionID,
		Tags:                make([]model.TagResponse, len(artProject.Tags)),
		StashID:             artProject.StashID,
		UserID:              artProject.UserID,
	}

	return response
}

func convertToRevisionResponse(revision *model.Revision) model.RevisionResponse {
	return model.RevisionResponse{
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

func convertToPublicRevisionResponse(revision *model.Revision) model.PublicRevisionResponse {
	return model.PublicRevisionResponse{
		ArtID:     revision.ArtID,
		CreatedAt: revision.CreatedAt,
		Comment:   revision.Comment,
		Size:      revision.Size,
	}
}
