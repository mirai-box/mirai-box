package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/service"
)

type CollectionHandler struct {
	collectionService service.CollectionService
}

func NewCollectionHandler(cs service.CollectionService) *CollectionHandler {
	return &CollectionHandler{collectionService: cs}
}

func (h *CollectionHandler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "CreateCollection")

	var req struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("User not found in context")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	collection, err := h.collectionService.CreateCollection(ctx, user.ID.String(), req.Title)
	if err != nil {
		logger.Error("Failed to create collection", "error", err)
		if err == model.ErrInvalidInput {
			SendErrorResponse(w, http.StatusBadRequest, "Invalid input")
		} else {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to create collection")
		}
		return
	}

	logger.Info("Collection created successfully", "collectionID", collection.ID)
	SendJSONResponse(w, http.StatusCreated, collection)
}

func (h *CollectionHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collectionID := chi.URLParam(r, "id")
	logger := slog.With("handler", "GetCollection", "collectionID", collectionID)

	collection, err := h.collectionService.FindByID(ctx, collectionID)
	if err != nil {
		logger.Error("Failed to get collection", "error", err)
		if err == model.ErrCollectionNotFound {
			SendErrorResponse(w, http.StatusNotFound, "Collection not found")
		} else {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to get collection")
		}
		return
	}

	logger.Info("Collection retrieved successfully")
	SendJSONResponse(w, http.StatusOK, collection)
}

func (h *CollectionHandler) GetUserCollections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "GetUserCollections")

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		logger.Warn("User not found in context")
		SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	collections, err := h.collectionService.FindByUserID(ctx, user.ID.String())
	if err != nil {
		logger.Error("Failed to get user collections", "error", err, "userID", user.ID)
		if err == model.ErrInvalidInput {
			SendErrorResponse(w, http.StatusBadRequest, "Invalid input")
		} else {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to get user collections")
		}
		return
	}

	logger.Info("User collections retrieved successfully", "userID", user.ID, "count", len(collections))
	SendJSONResponse(w, http.StatusOK, collections)
}

func (h *CollectionHandler) UpdateCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "UpdateCollection")

	var collection model.Collection
	if err := json.NewDecoder(r.Body).Decode(&collection); err != nil {
		logger.Error("Failed to decode request body", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	collectionID := chi.URLParam(r, "id")
	collection.ID = uuid.MustParse(collectionID)

	if err := h.collectionService.UpdateCollection(ctx, &collection); err != nil {
		logger.Error("Failed to update collection", "error", err, "collectionID", collectionID)
		if err == model.ErrCollectionNotFound {
			SendErrorResponse(w, http.StatusNotFound, "Collection not found")
		} else {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to update collection")
		}
		return
	}

	logger.Info("Collection updated successfully", "collectionID", collectionID)
	SendJSONResponse(w, http.StatusOK, collection)
}

func (h *CollectionHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collectionID := chi.URLParam(r, "id")

	logger := slog.With("handler", "DeleteCollection", "collectionID", collectionID)

	err := h.collectionService.DeleteCollection(ctx, collectionID)
	if err != nil {
		logger.Error("Failed to delete collection", "error", err)
		if err == model.ErrCollectionNotFound {
			SendErrorResponse(w, http.StatusNotFound, "Collection not found")
		} else {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to delete collection")
		}
		return
	}

	logger.Info("Collection deleted successfully")
	SendJSONResponse(w, http.StatusOK, map[string]string{"message": "Collection deleted successfully"})
}

func (h *CollectionHandler) AddRevisionToCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collectionID := chi.URLParam(r, "id")
	logger := slog.With("handler", "AddRevisionToCollection", "collectionID", collectionID)

	var req struct {
		RevisionID string `json:"revisionID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request body", "error", err)
		SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.collectionService.AddRevisionToCollection(ctx, collectionID, req.RevisionID); err != nil {
		logger.Error("Failed to add art project to collection", "error", err, "revisionID", req.RevisionID)
		if errors.Is(err, model.ErrCollectionNotFound) {
			SendErrorResponse(w, http.StatusNotFound, "Collection not found")
			return
		}
		SendErrorResponse(w, http.StatusInternalServerError, "Failed to add art project to collection")
		return
	}

	logger.Info("Art revision added to collection successfully", "collectionID", collectionID, "revisionID", req.RevisionID)
	SendJSONResponse(w, http.StatusOK, map[string]string{"message": "Art project added to collection successfully"})
}

func (h *CollectionHandler) ListRevisions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collectionID := chi.URLParam(r, "id")
	logger := slog.With("handler", "ListRevisions", "collectionID", collectionID)

	revisions, err := h.collectionService.GetRevisionsByCollectionID(ctx, collectionID)
	if err != nil {
		logger.Error("Failed to list revisions", "error", err)
		if err == model.ErrCollectionNotFound {
			SendErrorResponse(w, http.StatusNotFound, "Revisions not found in collection")
		} else {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed list all revisions from collection")
		}
		return
	}

	logger.Info("Listed revisions for the collection")
	SendJSONResponse(w, http.StatusOK, revisions)
}

func (h *CollectionHandler) ListPublicRevisions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collectionID := chi.URLParam(r, "id")
	logger := slog.With("handler", "ListPublicRevisions", "collectionID", collectionID)

	revisions, err := h.collectionService.GetRevisionsByPublicCollectionID(ctx, collectionID)
	if err != nil {
		logger.Error("Failed to list revisions", "error", err)
		if err == model.ErrCollectionNotFound {
			SendErrorResponse(w, http.StatusNotFound, "Revisions not found in collection")
		} else {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed list all revisions from collection")
		}
		return
	}

	response := make([]model.PublicRevisionResponse, len(revisions))
	for i, rev := range revisions {
		// Note: we can make additional check here if the revision is public or something
		response[i] = convertToPublicRevisionResponse(&rev)
	}

	logger.Info("Listed revisions for the collection")
	SendJSONResponse(w, http.StatusOK, response)
}

func (h *CollectionHandler) RemoveRevisionFromCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "RemoveArtProjectFromCollection")

	collectionID := chi.URLParam(r, "id")
	revisionID := chi.URLParam(r, "revisionID")

	err := h.collectionService.RemoveRevisionFromCollection(ctx, collectionID, revisionID)
	if err != nil {
		logger.Error("Failed to remove art project from collection", "error", err, "collectionID", collectionID, "revisionID", revisionID)
		if err == model.ErrCollectionNotFound {
			SendErrorResponse(w, http.StatusNotFound, "Revision not found in collection")
		} else {
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to remove art project from collection")
		}
		return
	}

	logger.Info("Art revision removed from collection successfully", "collectionID", collectionID, "revisionID", revisionID)
	SendJSONResponse(w, http.StatusOK, map[string]string{"message": "Art project removed from collection successfully"})
}
