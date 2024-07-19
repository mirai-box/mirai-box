package repo

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/model"
)

// ArtLinkRepository defines the interface for art link related database operations.
type ArtLinkRepository interface {
	CreateArtLink(ctx context.Context, artLink *model.ArtLink) error
	UpdateArtLink(ctx context.Context, artLink *model.ArtLink) error
	GetArtLinkByToken(ctx context.Context, token string) (*model.ArtLink, error)
}

type artLinkRepo struct {
	db *gorm.DB
}

// NewArtLinkRepository creates a new instance of ArtLinkRepository.
func NewArtLinkRepository(db *gorm.DB) ArtLinkRepository {
	return &artLinkRepo{db: db}
}

// CreateArtLink adds a new art link to the database.
func (r *artLinkRepo) CreateArtLink(ctx context.Context, artLink *model.ArtLink) error {
	logger := slog.With("method", "CreateArtLink", "token", artLink.Token)

	if err := r.db.Create(artLink).Error; err != nil {
		logger.Error("Failed to create art link", "error", err)
		return err
	}

	logger.Info("Art link created successfully")
	return nil
}

// UpdateArtLink updates an existing art link in the database.
func (r *artLinkRepo) UpdateArtLink(ctx context.Context, artLink *model.ArtLink) error {
	logger := slog.With("method", "UpdateArtLink", "token", artLink.Token)

	if err := r.db.Save(artLink).Error; err != nil {
		logger.Error("Failed to update art link", "error", err)
		return err
	}

	logger.Info("Art link updated successfully")
	return nil
}

// GetArtLinkByToken retrieves an art link by its token.
func (r *artLinkRepo) GetArtLinkByToken(ctx context.Context, token string) (*model.ArtLink, error) {
	logger := slog.With("method", "GetArtLinkByToken", "token", token)

	var artLink model.ArtLink
	if err := r.db.Where("token = ?", token).First(&artLink).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("Art link not found")
			return nil, model.ErrArtLinkNotFound
		}

		logger.Error("Failed to get art link by token", "error", err)
		return nil, err
	}

	logger.Info("Art link retrieved successfully")
	return &artLink, nil
}
