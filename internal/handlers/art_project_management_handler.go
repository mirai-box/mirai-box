package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/service"
)

// ArtProjectHandler
type ArtProjectManagementHandler struct {
	managmentService service.ArtProjectManagementServiceInterface
	artService       service.ArtProjectServiceInterface
}

func NewArtProjectManagementHandler(
	managmentService service.ArtProjectManagementServiceInterface,
	artService service.ArtProjectServiceInterface,
) *ArtProjectManagementHandler {
	return &ArtProjectManagementHandler{
		managmentService: managmentService,
		artService:       artService,
	}
}

func (h *ArtProjectManagementHandler) CreateArtProject(w http.ResponseWriter, r *http.Request) {
	h.handleFileUpload(w, r, func(file io.Reader, handler *multipart.FileHeader) (interface{}, error) {
		title := r.FormValue("title")
		if title == "" {
			return nil, fmt.Errorf("Title is required")
		}

		user, ok := middleware.GetUserFromContext(r.Context())
		if !ok {
			return nil, fmt.Errorf("Unauthorized")
		}

		slog.Info("upload new art project", "title", title, "userID", user.ID)

		return h.managmentService.CreateArtProjectAndRevision(r.Context(), user.ID.String(), file, title, handler.Filename)
	})
}

func (p *ArtProjectManagementHandler) AddRevision(w http.ResponseWriter, r *http.Request) {
	p.handleFileUpload(w, r, func(file io.Reader, handler *multipart.FileHeader) (interface{}, error) {
		comment := r.FormValue("comment")
		artProjectID := r.PathValue("id")

		user, ok := middleware.GetUserFromContext(r.Context())
		if !ok {
			return nil, fmt.Errorf("Unauthorized")
		}

		slog.Debug("add new revision for a picture", "artProjectID", artProjectID, "userID", user.ID)

		return p.managmentService.AddRevision(r.Context(), user.ID.String(), artProjectID, file, comment, handler.Filename)
	})
}

func (h *ArtProjectManagementHandler) ListRevisions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.DebugContext(ctx, "ListRevisions is called")

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.ErrorContext(ctx, "unauthorized user")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	artID := chi.URLParam(r, "id")

	slog.InfoContext(ctx, "list all revisions for an art project", "artID", artID, "userID", user.ID)

	revisions, err := h.managmentService.ListAllRevisions(ctx, artID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find all revisions for art project", "error", err, "artID", artID)
		http.Error(w, "Failed to find all revisions for art project", http.StatusInternalServerError)
		return
	}

	if revisions[0].ArtProject.UserID != user.ID {
		slog.ErrorContext(ctx, "user and ArtProject user are not the same",
			"userID", user.ID,
			"artProject.Stash.UserID", revisions[0].ArtProject.UserID,
		)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(revisions)
}

func (h *ArtProjectManagementHandler) handleFileUpload(w http.ResponseWriter, r *http.Request, processFile func(io.Reader, *multipart.FileHeader) (interface{}, error)) {
	r.ParseMultipartForm(10 << 20) // Limit file size to 10MB

	file, handler, err := r.FormFile("file")
	if err != nil {
		slog.Error("Invalid file upload", "error", err)
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	result, err := processFile(file, handler)
	if err != nil {
		var notFoundErr *models.ErrNotFound
		if errors.As(err, &notFoundErr) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Failed to process file upload", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		slog.Error("Failed to encode result as json", "error", err, "result", result)
	}
}

func (h *ArtProjectManagementHandler) MyArtProjects(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	artProjects, err := h.artService.FindByUserID(r.Context(), user.ID.String())
	if err != nil {
		slog.Error("Failed to list user's art projects", "error", err, "userID", user.ID)
		http.Error(w, "Failed to list user's web pages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artProjects)
}

func (h *ArtProjectManagementHandler) MyArtProjectByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.DebugContext(ctx, "MyArtProjectByID is called")

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.ErrorContext(ctx, "unauthorized user")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	artID := chi.URLParam(r, "id")

	slog.InfoContext(ctx, "getting art project", "artID", artID, "userID", user.ID)

	artProject, err := h.artService.FindByID(ctx, artID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find user's art project", "error", err, "artID", artID)
		http.Error(w, "Failed to find user's art project", http.StatusInternalServerError)
		return
	}

	slog.Info("found art project", "artID", artProject.ID, "userID", user.ID)

	if artProject.UserID != user.ID {
		slog.ErrorContext(ctx, "user and artProject user are not the same",
			"userID", user.ID,
			"artProject.UserID", artProject.Stash.UserID,
		)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artProject)
}
