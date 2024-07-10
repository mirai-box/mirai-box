package repos

import (
	"context"
	"log/slog"

	"github.com/mirai-box/mirai-box/internal/models"

	"gorm.io/gorm"
)

// SaleRepository implements the SaleRepositoryInterface
type SaleRepository struct {
	DB *gorm.DB
}

// NewSaleRepository creates a new instance of SaleRepository
func NewSaleRepository(db *gorm.DB) SaleRepositoryInterface {
	return &SaleRepository{DB: db}
}

// Create adds a new sale to the database
func (r *SaleRepository) Create(ctx context.Context, sale *models.Sale) error {
	slog.InfoContext(ctx, "Creating new sale", "saleID", sale.ID, "userID", sale.UserID, "artProjectID", sale.ArtProjectID)
	if err := r.DB.Create(sale).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create sale", "error", err, "saleID", sale.ID)
		return err
	}
	slog.InfoContext(ctx, "Sale created successfully", "saleID", sale.ID)
	return nil
}

// FindByID retrieves a sale by its ID
func (r *SaleRepository) FindByID(ctx context.Context, id string) (*models.Sale, error) {
	slog.InfoContext(ctx, "Finding sale by ID", "saleID", id)
	var sale models.Sale
	err := r.DB.Preload("ArtProject").Preload("User").First(&sale, "id = ?", id).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find sale by ID", "error", err, "saleID", id)
		return nil, err
	}
	slog.InfoContext(ctx, "Sale found successfully", "saleID", id)
	return &sale, nil
}

// FindByUserID retrieves all sales for a specific user
func (r *SaleRepository) FindByUserID(ctx context.Context, userID string) ([]models.Sale, error) {
	slog.InfoContext(ctx, "Finding sales by user ID", "userID", userID)
	var sales []models.Sale
	err := r.DB.Preload("ArtProject").Preload("User").Where("user_id = ?", userID).Find(&sales).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find sales by user ID", "error", err, "userID", userID)
		return nil, err
	}
	slog.InfoContext(ctx, "Sales found successfully", "userID", userID, "count", len(sales))
	return sales, nil
}

// FindByArtProjectID retrieves all sales for a specific art project
func (r *SaleRepository) FindByArtProjectID(ctx context.Context, artProjectID string) ([]models.Sale, error) {
	slog.InfoContext(ctx, "Finding sales by art project ID", "artProjectID", artProjectID)
	var sales []models.Sale
	err := r.DB.Preload("ArtProject").Preload("User").Where("art_project_id = ?", artProjectID).Find(&sales).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find sales by art project ID", "error", err, "artProjectID", artProjectID)
		return nil, err
	}
	slog.InfoContext(ctx, "Sales found successfully", "artProjectID", artProjectID, "count", len(sales))
	return sales, nil
}

// Update modifies an existing sale in the database
func (r *SaleRepository) Update(ctx context.Context, sale *models.Sale) error {
	slog.InfoContext(ctx, "Updating sale", "saleID", sale.ID)
	if err := r.DB.Save(sale).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to update sale", "error", err, "saleID", sale.ID)
		return err
	}
	slog.InfoContext(ctx, "Sale updated successfully", "saleID", sale.ID)
	return nil
}

// Delete removes a sale from the database
func (r *SaleRepository) Delete(ctx context.Context, id string) error {
	slog.InfoContext(ctx, "Deleting sale", "saleID", id)
	if err := r.DB.Delete(&models.Sale{}, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to delete sale", "error", err, "saleID", id)
		return err
	}
	slog.InfoContext(ctx, "Sale deleted successfully", "saleID", id)
	return nil
}

// Ensure SaleRepository implements SaleRepositoryInterface
var _ SaleRepositoryInterface = (*SaleRepository)(nil)
