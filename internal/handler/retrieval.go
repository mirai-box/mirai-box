package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/mirai-box/mirai-box/internal/model"
)

func (h *PictureRetrievalHandler) SharedPictureHandler(w http.ResponseWriter, r *http.Request) {
	artID := r.PathValue("artID")
	slog.Debug("get shared picture by art id", "artID", artID)

	h.handleDownload(w, r, func() (io.ReadCloser, *model.Picture, error) {
		return h.service.GetSharedPicture(artID)
	}, artID)
}

func (h *PictureRetrievalHandler) FleRevisionDownloadHandler(w http.ResponseWriter, r *http.Request) {
	revisionID := r.PathValue("revisionID")
	pictureID := r.PathValue("pictureID")
	slog.Debug("get picture by id and revision", "revisionID", revisionID, "pictureID", pictureID)

	h.handleDownload(w, r, func() (io.ReadCloser, *model.Picture, error) {
		return h.service.GetPictureByRevision(pictureID, revisionID)
	}, revisionID)
}

func (h *PictureRetrievalHandler) LatestFileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	pictureID := r.PathValue("pictureID")
	slog.Debug("get latest version of the picture by id", "pictureID", pictureID)

	h.handleDownload(w, r, func() (io.ReadCloser, *model.Picture, error) {
		return h.service.GetPictureByID(pictureID)
	}, pictureID)
}

func (h *PictureRetrievalHandler) handleDownload(w http.ResponseWriter, _ *http.Request, fetch func() (io.ReadCloser, *model.Picture, error), id string) {
	fh, pic, err := fetch()
	if err != nil {
		slog.Error("can't get picture", "error", err, "ID", id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer fh.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", pic.Filename))

	if _, err = io.Copy(w, fh); err != nil {
		slog.Error("error streaming file", "error", err, "ID", id)
		http.Error(w, "Error streaming file", http.StatusInternalServerError)
		return
	}
}
