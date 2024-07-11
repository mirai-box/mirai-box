package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/service"
)

type ArtProjectRetrievalHandler struct {
	service service.ArtProjectRetrievalServiceInterface
}

func NewArtProjectRetrievalHandler(service service.ArtProjectRetrievalServiceInterface) *ArtProjectRetrievalHandler {
	return &ArtProjectRetrievalHandler{
		service: service,
	}
}

// SharedPictureHandler handles retrieval of a shared picture by art ID
func (h *ArtProjectRetrievalHandler) SharedPictureHandler(w http.ResponseWriter, r *http.Request) {
	artID := chi.URLParam(r, "artID")
	slog.Debug("get shared picture by art id", "artID", artID)

	h.handleDownload(w, r, func() (io.ReadCloser, *models.ArtProject, error) {
		return h.service.GetSharedArtProject(r.Context(), artID)
	}, artID)
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
