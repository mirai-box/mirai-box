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
	service        service.ArtProjectRetrievalServiceInterface
	revisonService service.RevisionServiceInterface
}

func NewArtProjectRetrievalHandler(
	service service.ArtProjectRetrievalServiceInterface,
	revisonService service.RevisionServiceInterface,
) *ArtProjectRetrievalHandler {
	return &ArtProjectRetrievalHandler{
		service:        service,
		revisonService: revisonService,
	}
}

func (h *ArtProjectRetrievalHandler) GetArtByID(w http.ResponseWriter, r *http.Request) {
	artID := chi.URLParam(r, "artID")
	ctx := r.Context()

	rev, err := h.service.GetRevisionByArtID(ctx, artID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, rev.FilePath)
}

// CreateArtLinkHandler handles the creation of art links
func (h *ArtProjectRetrievalHandler) CreateArtLinkHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateArtLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	rev, err := h.revisonService.FindByID(ctx, req.RevisionID.String())
	if err != nil {
		http.Error(w, "Failed to find revision by id", http.StatusInternalServerError)
		return
	}
	if user.ID.String() != rev.UserID.String() {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	duration := time.Duration(req.Duration) * time.Minute
	if req.Unlimited {
		duration = 0 // Will be handled in service layer
	}

	link, err := h.service.CreateArtLink(ctx, req.RevisionID, duration, req.OneTime, req.Unlimited)
	if err != nil {
		http.Error(w, "Failed to create art link", http.StatusInternalServerError)
		return
	}

	res := CreateArtLinkResponse{Link: link}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// RevisionDownload handles retrieval of a picture by picture ID and revision ID
func (h *ArtProjectRetrievalHandler) RevisionDownload(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	revisionID := chi.URLParam(r, "revisionID")
	artID := chi.URLParam(r, "artID")
	slog.Debug("get revision by id and art-id", "revisionID", revisionID, "artID", artID)

	h.handleDownload(w, r, func() (io.ReadCloser, *models.ArtProject, error) {
		return h.service.GetArtProjectByRevision(r.Context(), user.ID.String(), artID, revisionID)
	}, revisionID)
}

// handleDownload is a helper function to handle the file download process
func (h *ArtProjectRetrievalHandler) handleDownload(w http.ResponseWriter, _ *http.Request, fetch func() (io.ReadCloser, *models.ArtProject, error), id string) {
	fh, pic, err := fetch()
	if err != nil {
		slog.Error("can't get picture", "error", err, "ID", id)
		respondWithError(w, http.StatusInternalServerError, "can't get picture")
		return
	}
	defer fh.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", pic.Filename))

	if _, err = io.Copy(w, fh); err != nil {
		slog.Error("error streaming file", "error", err, "ID", id)
		respondWithError(w, http.StatusInternalServerError, "Error streaming file")
		return
	}
}
