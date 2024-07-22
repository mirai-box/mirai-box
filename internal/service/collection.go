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

type collectionService struct {
	repo repo.CollectionRepository
}

func NewCollectionService(repo repo.CollectionRepository) CollectionService {
	return &collectionService{repo: repo}
}

func (s *collectionService) CreateCollection(ctx context.Context, userID, title string) (*model.Collection, error) {
	logger := slog.With("method", "CreateCollection", "userID", userID, "title", title)

	if userID == "" || title == "" {
		logger.Warn("Invalid input parameters")
		return nil, model.ErrInvalidInput
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		logger.Warn("Invalid userID format", "error", err)
		return nil, model.ErrInvalidInput
	}

	collection := &model.Collection{
		ID:        uuid.New(),
		UserID:    parsedUserID,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	logger = logger.With("collectionID", collection.ID)
	logger.Info("Creating new collection")

	if err := s.repo.CreateCollection(ctx, collection); err != nil {
		logger.Error("Failed to create collection", "error", err)
		return nil, err
	}

	logger.Info("Collection created successfully")
	return collection, nil
}

func (s *collectionService) FindByID(ctx context.Context, id string) (*model.Collection, error) {
	logger := slog.With("method", "FindByID", "collectionID", id)

	if id == "" {
		logger.Warn("Invalid input: empty id")
		return nil, model.ErrInvalidInput
	}

	collection, err := s.repo.FindCollectionByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find collection", "error", err)
		return nil, err
	}

	logger.Info("Collection found successfully")
	return collection, nil
}

func (s *collectionService) FindByUserID(ctx context.Context, userID string) ([]model.Collection, error) {
	logger := slog.With("method", "FindByUserID", "userID", userID)

	if userID == "" {
		logger.Warn("Invalid input: empty userID")
		return nil, model.ErrInvalidInput
	}

	collections, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to find collections", "error", err)
		return nil, err
	}

	logger.Info("Collections found successfully", "count", len(collections))
	return collections, nil
}
