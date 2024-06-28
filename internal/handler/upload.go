package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
)

func (h *PictureManagementHandler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	h.handleFileUpload(w, r, func(file io.Reader, handler *multipart.FileHeader) (interface{}, error) {
		title := r.FormValue("title")
		if title == "" {
			return nil, fmt.Errorf("Title is required")
		}

		slog.Debug("upload new picture", "title", title)

		return h.service.CreatePictureAndRevision(file, title, handler.Filename)
	})
}

func (p *PictureManagementHandler) AddRevisionHandler(w http.ResponseWriter, r *http.Request) {
	p.handleFileUpload(w, r, func(file io.Reader, handler *multipart.FileHeader) (interface{}, error) {
		comment := r.FormValue("comment")
		pictureID := r.PathValue("pictureID")
		
		slog.Debug("add new revision for a picture", "pictureID", pictureID)
		
		return p.service.AddRevision(pictureID, file, comment, handler.Filename)
	})
}

func (h *PictureManagementHandler) handleFileUpload(w http.ResponseWriter, r *http.Request, processFile func(io.Reader, *multipart.FileHeader) (interface{}, error)) {
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
