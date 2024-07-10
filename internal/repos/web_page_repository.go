package repos

import (
	"context"
	"log/slog"

	"github.com/mirai-box/mirai-box/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebPageRepository implements the WebPageRepositoryInterface
type WebPageRepository struct {
	DB *gorm.DB
}

// NewWebPageRepository creates a new instance of WebPageRepository
func NewWebPageRepository(db *gorm.DB) WebPageRepositoryInterface {
	return &WebPageRepository{DB: db}
}

// Create adds a new web page to the database
func (r *WebPageRepository) Create(ctx context.Context, webPage *models.WebPage) error {
	slog.InfoContext(ctx, "Creating new web page", "pageID", webPage.ID)
	if err := r.DB.Create(webPage).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create web page", "error", err, "pageID", webPage.ID)
		return err
	}
	slog.InfoContext(ctx, "Web page created successfully", "pageID", webPage.ID)
	return nil
}

// Update modifies an existing web page in the database
func (r *WebPageRepository) Update(ctx context.Context, webPage *models.WebPage) error {
	slog.InfoContext(ctx, "db: Updating web page",
		"pageID", webPage.ID,
		"userID", webPage.UserID,
	)

	if err := r.DB.Save(webPage).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to update web page", "error", err, "pageID", webPage.ID)
		return err
	}

	slog.InfoContext(ctx, "Web page updated successfully", "pageID", webPage.ID)
	return nil
}

// Delete removes a web page from the database
func (r *WebPageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	slog.InfoContext(ctx, "Deleting web page", "pageID", id)
	if err := r.DB.Delete(&models.WebPage{}, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to delete web page", "error", err, "pageID", id)
		return err
	}
	slog.InfoContext(ctx, "Web page deleted successfully", "pageID", id)
	return nil
}

// FindByID retrieves a web page by its ID
func (r *WebPageRepository) FindByID(ctx context.Context, id string) (*models.WebPage, error) {
	slog.InfoContext(ctx, "Finding web page by ID", "pageID", id)
	var webPage models.WebPage
	if err := r.DB.First(&webPage, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to find web page by ID", "error", err, "pageID", id)
		return nil, err
	}
	slog.InfoContext(ctx, "Web page found successfully", "pageID", id)
	return &webPage, nil
}

// FindAll retrieves all web pages from the database
func (r *WebPageRepository) FindAll(ctx context.Context) ([]models.WebPage, error) {
	slog.InfoContext(ctx, "Finding all web pages")
	var webPages []models.WebPage
	if err := r.DB.Find(&webPages).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to find all web pages", "error", err)
		return nil, err
	}
	slog.InfoContext(ctx, "All web pages retrieved successfully", "count", len(webPages))
	return webPages, nil
}

// FindByType retrieves all web pages of a specific type
func (r *WebPageRepository) FindByType(ctx context.Context, pageType string) ([]models.WebPage, error) {
	slog.InfoContext(ctx, "Finding web pages by type", "pageType", pageType)

	var webPages []models.WebPage
	if err := r.DB.Where("page_type = ?", pageType).Find(&webPages).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to find web pages by type", "error", err, "pageType", pageType)
		return nil, err
	}

	slog.InfoContext(ctx, "Web pages found successfully", "pageType", pageType, "count", len(webPages))
	return webPages, nil
}

// FindByType retrieves all web pages of a specific user
func (r *WebPageRepository) FindByUserID(ctx context.Context, userID string) ([]models.WebPage, error) {
	slog.InfoContext(ctx, "Finding web pages by user", "userID", userID)

	var webPages []models.WebPage
	if err := r.DB.Where("user_id = ?", userID).Find(&webPages).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to find web pages by user", "error", err, "userID", userID)
		return nil, err
	}

	slog.InfoContext(ctx, "Web pages found successfully", "userID", userID, "count", len(webPages))
	return webPages, nil
}

// Ensure WebPageRepository implements WebPageRepositoryInterface
var _ WebPageRepositoryInterface = (*WebPageRepository)(nil)
