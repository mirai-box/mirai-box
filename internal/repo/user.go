package repo

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/model"
)

// UserRepository defines the interface for user related database operations.
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	FindUserByID(ctx context.Context, id string) (*model.User, error)
	FindUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
	GetStorageUsage(ctx context.Context, userID string) (*model.StorageUsage, error)
	UpdateStorageUsage(ctx context.Context, storageUsage *model.StorageUsage) error
	CreateStash(ctx context.Context, stash *model.Stash) error
	UpdateStash(ctx context.Context, stash *model.Stash) error
	GetStashByUserID(ctx context.Context, id string) (*model.Stash, error)
}

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// CreateUser adds a new user to the database.
func (r *userRepo) CreateUser(ctx context.Context, user *model.User) error {
	logger := slog.With("method", "CreateUser", "userID", user.ID)

	if err := r.db.Create(user).Error; err != nil {
		logger.Error("Failed to create user", "error", err)
		return err
	}

	logger.Info("User created successfully")
	return nil
}

// FindUserByID retrieves a user by their ID.
func (r *userRepo) FindUserByID(ctx context.Context, id string) (*model.User, error) {
	logger := slog.With("method", "FindUserByID", "userID", id)

	var user model.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("User not found")
			return nil, model.ErrUserNotFound
		}
		logger.Error("Failed to find user", "error", err)
		return nil, err
	}

	logger.Info("User found successfully")
	return &user, nil
}

// FindUserByUsername retrieves a user by their username.
func (r *userRepo) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	logger := slog.With("method", "FindUserByUsername", "username", username)

	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("User not found")
			return nil, model.ErrUserNotFound
		}
		logger.Error("Failed to find user", "error", err)
		return nil, err
	}

	logger.Info("User found successfully")
	return &user, nil
}

// UpdateUser updates an existing user in the database.
func (r *userRepo) UpdateUser(ctx context.Context, user *model.User) error {
	logger := slog.With("method", "UpdateUser", "userID", user.ID)

	result := r.db.Save(user)
	if result.Error != nil {
		logger.Error("Failed to update user", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Info("User not found for update")
		return model.ErrUserNotFound
	}

	logger.Info("User updated successfully")
	return nil
}

// DeleteUser removes a user from the database.
func (r *userRepo) DeleteUser(ctx context.Context, id string) error {
	logger := slog.With("method", "DeleteUser", "userID", id)

	result := r.db.Delete(&model.User{}, "id = ?", id)
	if result.Error != nil {
		logger.Error("Failed to delete user", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Info("User not found for deletion")
		return model.ErrUserNotFound
	}

	logger.Info("User deleted successfully")
	return nil
}

// GetStorageUsage retrieves the storage usage for a specific user.
func (r *userRepo) GetStorageUsage(ctx context.Context, userID string) (*model.StorageUsage, error) {
	logger := slog.With("method", "GetStorageUsage", "userID", userID)

	var storageUsage model.StorageUsage
	if err := r.db.Where("user_id = ?", userID).First(&storageUsage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("Storage usage not found")
			return nil, model.ErrUserNotFound
		}
		logger.Error("Failed to get storage usage", "error", err)
		return nil, err
	}

	logger.Info("Storage usage retrieved successfully")
	return &storageUsage, nil
}

// UpdateStorageUsage updates the storage usage for a specific user.
func (r *userRepo) UpdateStorageUsage(ctx context.Context, storageUsage *model.StorageUsage) error {
	logger := slog.With("method", "UpdateStorageUsage", "userID", storageUsage.UserID)

	result := r.db.Save(storageUsage)
	if result.Error != nil {
		logger.Error("Failed to update storage usage", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Info("Storage usage not found for update")
		return model.ErrUserNotFound
	}

	logger.Info("Storage usage updated successfully")
	return nil
}

// CreateStash adds a new stash to the database
func (r *userRepo) CreateStash(ctx context.Context, stash *model.Stash) error {
	logger := slog.With("method", "CreateStash", "stashID", stash.ID, "userID", stash.UserID)

	if err := r.db.Create(stash).Error; err != nil {
		logger.ErrorContext(ctx, "Failed to create stash for a user", "error", err)
		return err
	}

	logger.InfoContext(ctx, "Stash created successfully")
	return nil
}

// GetStashByUserID retrieves a stash by user ID
func (r *userRepo) GetStashByUserID(ctx context.Context, userID string) (*model.Stash, error) {
	logger := slog.With("method", "GetStashByUserID", "userID", userID)

	var stash model.Stash

	if err := r.db.Preload("User").First(&stash, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.InfoContext(ctx, "Stash not found for user")
			return nil, model.ErrUserNotFound
		}
		logger.ErrorContext(ctx, "Failed to find stash by user ID", "error", err)
		return nil, err
	}

	logger.InfoContext(ctx, "Stash found successfully", "stashID", stash.ID)
	return &stash, nil
}

func (r *userRepo) UpdateStash(ctx context.Context, stash *model.Stash) error {
	logger := slog.With("method", "UpdateStash", "stashID", stash.ID, "userID", stash.UserID)

	result := r.db.Save(stash)
	if result.Error != nil {
		logger.Error("Failed to update stash", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Info("Stash not found for update")
		return model.ErrUserNotFound
	}

	logger.Info("Stash updated successfully")
	return nil
}
