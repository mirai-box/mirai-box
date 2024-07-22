package repo

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/model"
)

// WebPageRepository defines the interface for webpage related database operations.
type WebPageRepository interface {
	CreateWebPage(ctx context.Context, webPage *model.WebPage) error
	FindWebPageByID(ctx context.Context, id string) (*model.WebPage, error)
	FindAllWebPages(ctx context.Context) ([]model.WebPage, error)
	FindWebPagesByType(ctx context.Context, pageType string) ([]model.WebPage, error)
	FindWebPagesByUserID(ctx context.Context, userID string) ([]model.WebPage, error)
	UpdateWebPage(ctx context.Context, webPage *model.WebPage) error
	DeleteWebPage(ctx context.Context, id string) error
}

type webPageRepo struct {
	db *gorm.DB
}

// NewWebPageRepository creates a new instance of WebPageRepository.
func NewWebPageRepository(db *gorm.DB) WebPageRepository {
	return &webPageRepo{db: db}
}

// CreateWebPage adds a new webpage to the database.
func (r *webPageRepo) CreateWebPage(ctx context.Context, webPage *model.WebPage) error {
	logger := slog.With("method", "CreateWebPage", "webPageID", webPage.ID)

	if err := r.db.Create(webPage).Error; err != nil {
		logger.Error("Failed to create webpage", "error", err)
		return err
	}

	logger.Info("Webpage created successfully")
	return nil
}

// FindWebPageByID retrieves a webpage by its ID.
func (r *webPageRepo) FindWebPageByID(ctx context.Context, id string) (*model.WebPage, error) {
	logger := slog.With("method", "FindWebPageByID", "webPageID", id)

	var webPage model.WebPage
	if err := r.db.First(&webPage, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("Webpage not found")
			return nil, model.ErrWebPageNotFound
		}
		logger.Error("Failed to find webpage", "error", err)
		return nil, err
	}

	logger.Info("Webpage found successfully")
	return &webPage, nil
}

// FindAllWebPages retrieves all webpages from the database.
func (r *webPageRepo) FindAllWebPages(ctx context.Context) ([]model.WebPage, error) {
	logger := slog.With("method", "FindAllWebPages")

	var webPages []model.WebPage
	if err := r.db.Find(&webPages).Error; err != nil {
		logger.Error("Failed to find all webpages", "error", err)
		return nil, err
	}

	if len(webPages) == 0 {
		return nil, model.ErrWebPageNotFound
	}

	logger.Info("All webpages retrieved successfully", "count", len(webPages))
	return webPages, nil
}

// FindWebPagesByType retrieves all webpages of a specific type.
func (r *webPageRepo) FindWebPagesByType(ctx context.Context, pageType string) ([]model.WebPage, error) {
	logger := slog.With("method", "FindWebPagesByType", "pageType", pageType)

	var webPages []model.WebPage
	if err := r.db.Where("page_type = ?", pageType).Find(&webPages).Error; err != nil {
		logger.Error("Failed to find webpages by type", "error", err)
		return nil, err
	}

	if len(webPages) == 0 {
		return nil, model.ErrWebPageNotFound
	}

	logger.Info("Webpages found successfully", "count", len(webPages))
	return webPages, nil
}

// FindWebPagesByUserID retrieves all webpages for a specific user.
func (r *webPageRepo) FindWebPagesByUserID(ctx context.Context, userID string) ([]model.WebPage, error) {
	logger := slog.With("method", "FindWebPagesByUserID", "userID", userID)

	var webPages []model.WebPage
	if err := r.db.Where("user_id = ?", userID).Find(&webPages).Error; err != nil {
		logger.Error("Failed to find webpages by user ID", "error", err)
		return nil, err
	}

	if len(webPages) == 0 {
		return nil, model.ErrWebPageNotFound
	}

	logger.Info("Webpages found successfully", "count", len(webPages))
	return webPages, nil
}

// UpdateWebPage updates an existing webpage in the database.
func (r *webPageRepo) UpdateWebPage(ctx context.Context, webPage *model.WebPage) error {
	logger := slog.With("method", "UpdateWebPage", "webPageID", webPage.ID)

	result := r.db.Save(webPage)
	if result.Error != nil {
		logger.Error("Failed to update webpage", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Info("Webpage not found for update")
		return model.ErrWebPageNotFound
	}

	logger.Info("Webpage updated successfully")
	return nil
}

// DeleteWebPage removes a webpage from the database.
func (r *webPageRepo) DeleteWebPage(ctx context.Context, id string) error {
	logger := slog.With("method", "DeleteWebPage", "webPageID", id)

	result := r.db.Delete(&model.WebPage{}, "id = ?", id)
	if result.Error != nil {
		logger.Error("Failed to delete webpage", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Info("Webpage not found for deletion")
		return model.ErrWebPageNotFound
	}

	logger.Info("Webpage deleted successfully")
	return nil
}