package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/service"
)

// PictureRetrievalHandler is a struct that holds a reference to the service layer
type PictureRetrievalHandler struct {
	service service.PictureRetrievalService
}

func NewPictureRetrievalHandler(svc service.PictureRetrievalService) *PictureRetrievalHandler {
	return &PictureRetrievalHandler{
		service: svc,
	}
}

// SharedPictureHandler handles retrieval of a shared picture by art ID
func (h *PictureRetrievalHandler) SharedPictureHandler(w http.ResponseWriter, r *http.Request) {
	artID := chi.URLParam(r, "artID")
	slog.Debug("get shared picture by art id", "artID", artID)

	h.handleDownload(w, r, func() (io.ReadCloser, *model.Picture, error) {
		return h.service.GetSharedPicture(artID)
	}, artID)
}

// FileRevisionDownloadHandler handles retrieval of a picture by picture ID and revision ID
func (h *PictureRetrievalHandler) FileRevisionDownloadHandler(w http.ResponseWriter, r *http.Request) {
	revisionID := chi.URLParam(r, "revisionID")
	pictureID := chi.URLParam(r, "pictureID")
	slog.Debug("get picture by id and revision", "revisionID", revisionID, "pictureID", pictureID)

	h.handleDownload(w, r, func() (io.ReadCloser, *model.Picture, error) {
		return h.service.GetPictureByRevision(pictureID, revisionID)
	}, revisionID)
}

// LatestFileDownloadHandler handles retrieval of the latest version of a picture by picture ID
func (h *PictureRetrievalHandler) LatestFileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	pictureID := chi.URLParam(r, "pictureID")
	slog.Debug("get latest version of the picture by id", "pictureID", pictureID)

	h.handleDownload(w, r, func() (io.ReadCloser, *model.Picture, error) {
		return h.service.GetPictureByID(pictureID)
	}, pictureID)
}

// handleDownload is a helper function to handle the file download process
func (h *PictureRetrievalHandler) handleDownload(w http.ResponseWriter, _ *http.Request, fetch func() (io.ReadCloser, *model.Picture, error), id string) {
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
