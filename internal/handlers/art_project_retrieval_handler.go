package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/service"
)

type CreateArtLinkRequest struct {
	RevisionID uuid.UUID `json:"revision_id"`
	Duration   int       `json:"duration"` // in minutes
	OneTime    bool      `json:"one_time"`
	Unlimited  bool      `json:"unlimited,omitempty"`
}

type CreateArtLinkResponse struct {
	Link string `json:"link"`
}

type ArtProjectRetrievalHandler struct {
	service         service.ArtProjectRetrievalServiceInterface
	revisionService service.RevisionServiceInterface
}

func NewArtProjectRetrievalHandler(
	service service.ArtProjectRetrievalServiceInterface,
	revisionService service.RevisionServiceInterface,
) *ArtProjectRetrievalHandler {
	return &ArtProjectRetrievalHandler{
		service:         service,
		revisionService: revisionService,
	}
}

func (h *ArtProjectRetrievalHandler) GetArtByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	artID := chi.URLParam(r, "artID")

	slog.InfoContext(ctx, "Retrieving art by ID", "artID", artID)

	rev, err := h.service.GetRevisionByArtID(ctx, artID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "Art not found", "artID", artID)
			SendErrorResponse(w, http.StatusNotFound, "Art not found")
			return
		}
		slog.ErrorContext(ctx, "Failed to retrieve art", "error", err, "artID", artID)
		SendErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	slog.InfoContext(ctx, "Art retrieved successfully", "artID", artID, "revisionID", rev.ID)
	http.ServeFile(w, r, rev.FilePath)
}

func (h *ArtProjectRetrievalHandler) CreateArtLinkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized attempt to create art link")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateArtLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.WarnContext(ctx, "Invalid request payload", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	rev, err := h.revisionService.FindByID(ctx, req.RevisionID.String())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find revision", "error", err, "revisionID", req.RevisionID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to find revision")
		return
	}
	if user.ID != rev.UserID {
		slog.WarnContext(ctx, "Unauthorized attempt to create art link", "userID", user.ID, "revisionUserID", rev.UserID)
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	duration := time.Duration(req.Duration) * time.Minute
	if req.Unlimited {
		duration = 0 // Will be handled in service layer
	}

	link, err := h.service.CreateArtLink(ctx, req.RevisionID, duration, req.OneTime, req.Unlimited)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create art link", "error", err, "revisionID", req.RevisionID)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to create art link")
		return
	}

	res := CreateArtLinkResponse{Link: link}
	slog.InfoContext(ctx, "Art link created successfully", "revisionID", req.RevisionID)
	SendJSONResponse(w, http.StatusCreated, res)
}

func (h *ArtProjectRetrievalHandler) RevisionDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Unauthorized attempt to download revision")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	revisionID := chi.URLParam(r, "revisionID")
	artID := chi.URLParam(r, "artID")
	slog.InfoContext(ctx, "Retrieving revision for download", "revisionID", revisionID, "artID", artID, "userID", user.ID)

	h.handleDownload(w, r, func() (io.ReadCloser, *models.ArtProject, error) {
		return h.service.GetArtProjectByRevision(ctx, user.ID.String(), artID, revisionID)
	}, revisionID)
}

func (h *ArtProjectRetrievalHandler) handleDownload(w http.ResponseWriter, r *http.Request, fetch func() (io.ReadCloser, *models.ArtProject, error), id string) {
	ctx := r.Context()
	fh, pic, err := fetch()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get picture", "error", err, "ID", id)
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to get picture")
		return
	}
	defer fh.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", pic.Filename))

	if _, err = io.Copy(w, fh); err != nil {
		slog.ErrorContext(ctx, "Error streaming file", "error", err, "ID", id)
		SendErrorResponse(w, http.StatusInternalServerError, "Error streaming file")
		return
	}

	slog.InfoContext(ctx, "File downloaded successfully", "ID", id, "filename", pic.Filename)
}
