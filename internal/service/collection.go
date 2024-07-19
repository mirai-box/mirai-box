package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repo"
)

//go:generate go run github.com/vektra/mockery/v2@v2 --name=CollectionService --filename=collection_service.go --output=../../mocks/
type CollectionService interface {
	CreateCollection(ctx context.Context, userID, title string) (*model.Collection, error)
	FindByID(ctx context.Context, id string) (*model.Collection, error)
	FindByUserID(ctx context.Context, userID string) ([]model.Collection, error)
}

// CollectionService implements the CollectionService
type collectionService struct {
	repo repo.CollectionRepository
}

// NewCollectionService creates a new instance of CollectionService
func NewCollectionService(repo repo.CollectionRepository) CollectionService {
	return &collectionService{repo: repo}
}

// CreateCollection creates a new collection
func (s *collectionService) CreateCollection(ctx context.Context, userID, title string) (*model.Collection, error) {
	collection := &model.Collection{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(userID),
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	slog.InfoContext(ctx, "Creating new collection",
		"collectionID", collection.ID,
		"userID", userID,
		"title", title,
	)

	if err := s.repo.CreateCollection(ctx, collection); err != nil {
		slog.ErrorContext(ctx, "Failed to create collection",
			"error", err,
			"collectionID", collection.ID,
			"userID", userID,
		)
		return nil, err
	}

	slog.InfoContext(ctx, "Collection created successfully",
		"collectionID", collection.ID,
		"userID", userID,
	)

	return collection, nil
}

// FindByID finds a collection by its ID
func (s *collectionService) FindByID(ctx context.Context, id string) (*model.Collection, error) {
	slog.InfoContext(ctx, "Finding collection by ID", "collectionID", id)

	collection, err := s.repo.FindCollectionByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find collection by ID",
			"error", err,
			"collectionID", id,
		)
		return nil, err
	}

	slog.InfoContext(ctx, "Collection found successfully", "collectionID", id)
	return collection, nil
}

// FindByUserID finds all collections by a user ID
func (s *collectionService) FindByUserID(ctx context.Context, userID string) ([]model.Collection, error) {
	slog.InfoContext(ctx, "Finding collections by user ID", "userID", userID)

	collections, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find collections by user ID",
			"error", err,
			"userID", userID,
		)
		return nil, err
	}

	slog.InfoContext(ctx, "Collections found successfully",
		"userID", userID,
		"count", len(collections),
	)
	return collections, nil
}
