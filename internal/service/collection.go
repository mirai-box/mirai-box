package service

import (
	"context"
	"errors"
	"fmt"
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
	UpdateCollection(ctx context.Context, collection *model.Collection) error
	DeleteCollection(ctx context.Context, id string) error
	AddRevisionToCollection(ctx context.Context, collectionID, revisionID string) error
	RemoveRevisionFromCollection(ctx context.Context, collectionID, revisionID string) error
	GetRevisionsByCollectionID(ctx context.Context, collectionID string) ([]model.Revision, error)
	GetRevisionsByPublicCollectionID(ctx context.Context, collectionPublicID string) ([]model.Revision, error)
}

type collectionService struct {
	collectionRepo repo.CollectionRepository
	artRepo        repo.ArtProjectRepository
	secretKey      []byte
}

func NewCollectionService(cr repo.CollectionRepository, ar repo.ArtProjectRepository, secretKey []byte) CollectionService {
	return &collectionService{
		collectionRepo: cr,
		artRepo:        ar,
		secretKey:      secretKey,
	}
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

	collectionID := uuid.New()
	publicCollectionID, err := GeneratePublicID(collectionID, parsedUserID, s.secretKey)
	if err != nil {
		logger.Error("Failed to generate artID", "error", err)
		return nil, fmt.Errorf("failed to generate artID: %w", err)
	}

	collection := &model.Collection{
		ID:           collectionID,
		CollectionID: publicCollectionID,
		UserID:       parsedUserID,
		Title:        title,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	logger = logger.With("collectionID", collection.ID)
	logger.Info("Creating new collection")

	if err := s.collectionRepo.CreateCollection(ctx, collection); err != nil {
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

	collection, err := s.collectionRepo.FindCollectionByID(ctx, id)
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

	collections, err := s.collectionRepo.FindByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to find collections", "error", err)
		return nil, err
	}

	logger.Info("Collections found successfully", "count", len(collections))
	return collections, nil
}

func (s *collectionService) UpdateCollection(ctx context.Context, collection *model.Collection) error {
	logger := slog.With("method", "UpdateCollection", "collectionID", collection.ID)

	if err := s.collectionRepo.UpdateCollection(ctx, collection); err != nil {
		logger.Error("Failed to update collection", "error", err)
		return err
	}

	logger.Info("Collection updated successfully")
	return nil
}

func (s *collectionService) DeleteCollection(ctx context.Context, id string) error {
	logger := slog.With("method", "DeleteCollection", "collectionID", id)

	if id == "" {
		logger.Warn("Invalid input: empty id")
		return model.ErrInvalidInput
	}

	if err := s.collectionRepo.DeleteCollection(ctx, id); err != nil {
		logger.Error("Failed to delete collection", "error", err)
		return err
	}

	logger.Info("Collection deleted successfully")
	return nil
}

func (s *collectionService) AddRevisionToCollection(ctx context.Context, collectionID, revisionID string) error {
	logger := slog.With("service", "AddRevisionToCollection", "collectionID", collectionID, "revisionID", revisionID)

	if collectionID == "" || revisionID == "" {
		logger.Warn("Invalid input: empty collectionID or revisionID")
		return model.ErrInvalidInput
	}

	rev, err := s.artRepo.FindRevisionByID(ctx, revisionID)
	if err != nil {
		logger.Error("Failed to find revision", "error", err)
		return model.ErrArtProjectNotFound
	}

	col, err := s.FindByID(ctx, collectionID)
	if err != nil {
		logger.Error("Failed to find collection", "error", err)
		return err
	}

	if rev.UserID != col.UserID {
		logger.Error("User is not authorized for the action", "error", err,
			"rev.UserID", rev.UserID, "col.UserID", col.UserID)
		return model.ErrUnauthorized
	}

	if err := s.collectionRepo.AddRevisionToCollection(ctx, collectionID, revisionID); err != nil {
		logger.Error("Failed to add art project to collection", "error", err)
		return err
	}

	logger.Info("Art revision added to collection successfully")
	return nil
}

func (s *collectionService) RemoveRevisionFromCollection(ctx context.Context, collectionID, revisionID string) error {
	logger := slog.With("method", "RemoveRevisionFromCollection", "collectionID", collectionID, "revisionID", revisionID)

	if collectionID == "" || revisionID == "" {
		logger.Warn("Invalid input: empty collectionID or artProjectID")
		return model.ErrInvalidInput
	}

	rev, err := s.artRepo.FindRevisionByID(ctx, revisionID)
	if err != nil {
		logger.Error("Failed to find revision", "error", err)
		return model.ErrArtProjectNotFound
	}

	col, err := s.FindByID(ctx, collectionID)
	if err != nil {
		logger.Error("Failed to find collection", "error", err)
		return err
	}

	if rev.UserID != col.UserID {
		logger.Error("User is not authorized for the action", "error", err,
			"rev.UserID", rev.UserID, "col.UserID", col.UserID)
		return model.ErrUnauthorized
	}

	if err := s.collectionRepo.RemoveRevisionFromCollection(ctx, collectionID, revisionID); err != nil {
		logger.Error("Failed to remove revision from collection", "error", err)
		return err
	}

	logger.Info("Art revision removed from collection successfully")
	return nil
}

func (s *collectionService) GetRevisionsByCollectionID(ctx context.Context, collectionID string) ([]model.Revision, error) {
	logger := slog.With("service", "GetRevisionsByCollectionID", "collectionID", collectionID)

	revisions, err := s.collectionRepo.GetRevisionsByCollectionID(ctx, collectionID)
	if err != nil {
		logger.Error("Failed to get revisions for the collection", "error", err)
		return nil, err
	}

	logger.Info("Listed all revisions for a collection")
	return revisions, nil
}

func (s *collectionService) GetRevisionsByPublicCollectionID(ctx context.Context, collectionPublicID string) ([]model.Revision, error) {
	logger := slog.With("service", "GetRevisionsByPublicCollectionID", "collectionPublicID", collectionPublicID)

	collectionID, userID, err := DecodePublicID(collectionPublicID, s.secretKey)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to decode collectionPublicID", "error", err)
		return nil, err
	}

	revisions, err := s.GetRevisionsByCollectionID(ctx, collectionID)
	if err != nil {
		logger.Error("Failed to get revisions for the collection", "error", err)
		return nil, err
	}

	if len(revisions) == 0 {
		return []model.Revision{}, nil
	}

	if revisions[0].UserID.String() != userID {
		logger.ErrorContext(ctx, "userID does not match the revision's userID", "revision.UserID", revisions[0].UserID, "userID", userID)
		return nil, errors.New("userID does not match the collection userID")
	}

	logger.Info("Listed all revisions for a collection")
	return revisions, nil
}
