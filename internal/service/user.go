package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repo"
)

//go:generate go run github.com/vektra/mockery/v2@v2 --name=UserService --filename=user_service.go --output=../../mocks/
type UserService interface {
	Authenticate(ctx context.Context, username, password string) (*model.User, error)
	CreateUser(ctx context.Context, username, password, role string) (*model.User, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
	GetStorageUsage(ctx context.Context, userID string) (*model.StorageUsage, error)
	UpdateStorageUsage(ctx context.Context, storageUsage *model.StorageUsage) error
	GetStashByUserID(ctx context.Context, userID string) (*model.Stash, error)
}

type userService struct {
	userRepo repo.UserRepository
}

func NewUserService(repo repo.UserRepository) UserService {
	return &userService{userRepo: repo}
}

func (s *userService) Authenticate(ctx context.Context, username, password string) (*model.User, error) {
	logger := slog.With("method", "Authenticate", "username", username)

	if username == "" || password == "" {
		logger.Warn("Invalid input parameters")
		return nil, model.ErrInvalidInput
	}

	user, err := s.userRepo.FindUserByUsername(ctx, username)
	if err != nil {
		logger.Error("Failed to find user", "error", err)
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logger.Warn("Invalid credentials")
		return nil, model.ErrInvalidCredentials
	}

	logger.Info("User authenticated successfully")
	return user, nil
}

func (s *userService) GetUser(ctx context.Context, id string) (*model.User, error) {
	logger := slog.With("method", "GetUser", "userID", id)

	if id == "" {
		logger.Warn("Invalid input: empty id")
		return nil, model.ErrInvalidInput
	}

	user, err := s.userRepo.FindUserByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find user", "error", err)
		return nil, err
	}

	logger.Info("User found successfully")
	return user, nil
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	logger := slog.With("method", "GetUserByUsername", "username", username)

	if username == "" {
		logger.Warn("Invalid input: empty username")
		return nil, model.ErrInvalidInput
	}

	user, err := s.userRepo.FindUserByUsername(ctx, username)
	if err != nil {
		logger.Error("Failed to find user", "error", err)
		return nil, err
	}

	logger.Info("User found successfully")
	return user, nil
}

func (s *userService) CreateUser(ctx context.Context, username, password, role string) (*model.User, error) {
	logger := slog.With("method", "CreateUser", "username", username, "role", role)

	if username == "" || password == "" || role == "" {
		logger.Warn("Invalid input parameters")
		return nil, model.ErrInvalidInput
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		logger.Error("Failed to hash password", "error", err)
		return nil, err
	}

	user := &model.User{
		ID:       uuid.New(),
		Username: username,
		Password: hashedPassword,
		Role:     role,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		logger.Error("Failed to create user", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.createStash(ctx, user.ID); err != nil {
		logger.Error("Failed to create stash", "error", err)
		return nil, model.ErrStashCreationFailed
	}

	logger.Info("User created successfully", "userID", user.ID)
	return user, nil
}

func (s *userService) createStash(ctx context.Context, userID uuid.UUID) error {
	logger := slog.With("method", "createStash", "userID", userID)

	stash := &model.Stash{
		ID:     uuid.New(),
		UserID: userID,
	}

	if err := s.userRepo.CreateStash(ctx, stash); err != nil {
		logger.Error("Failed to create stash", "error", err)
		return err
	}

	logger.Info("Stash created successfully", "stashID", stash.ID)
	return nil
}

func (s *userService) UpdateUser(ctx context.Context, user *model.User) error {
	logger := slog.With("method", "UpdateUser", "userID", user.ID)

	if user.ID == uuid.Nil {
		logger.Warn("Invalid input parameters")
		return model.ErrInvalidInput
	}

	existingUser, err := s.userRepo.FindUserByID(ctx, user.ID.String())
	if err != nil {
		logger.Error("Failed to find user", "error", err)
		return err
	}

	existingUser.Username = user.Username
	existingUser.Role = user.Role

	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("Failed to hash password", "error", err)
			return fmt.Errorf("failed to hash password: %w", err)
		}
		existingUser.Password = string(hashedPassword)
	}

	if err := s.userRepo.UpdateUser(ctx, existingUser); err != nil {
		logger.Error("Failed to update user", "error", err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info("User updated successfully")
	return nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	logger := slog.With("method", "DeleteUser", "userID", id)

	if id == "" {
		logger.Warn("Invalid input: empty id")
		return model.ErrInvalidInput
	}

	if err := s.userRepo.DeleteUser(ctx, id); err != nil {
		logger.Error("Failed to delete user", "error", err)
		return err
	}

	logger.Info("User deleted successfully")
	return nil
}

func (s *userService) GetStorageUsage(ctx context.Context, userID string) (*model.StorageUsage, error) {
	logger := slog.With("method", "GetStorageUsage", "userID", userID)

	if userID == "" {
		logger.Warn("Invalid input: empty userID")
		return nil, model.ErrInvalidInput
	}

	usage, err := s.userRepo.GetStorageUsage(ctx, userID)
	if err != nil {
		logger.Error("Failed to get storage usage", "error", err)
		return nil, err
	}

	logger.Info("Storage usage retrieved successfully")
	return usage, nil
}

func (s *userService) UpdateStorageUsage(ctx context.Context, storageUsage *model.StorageUsage) error {
	logger := slog.With("method", "UpdateStorageUsage", "userID", storageUsage.UserID)

	if storageUsage.UserID == uuid.Nil {
		logger.Warn("Invalid input parameters")
		return model.ErrInvalidInput
	}

	if err := s.userRepo.UpdateStorageUsage(ctx, storageUsage); err != nil {
		logger.Error("Failed to update storage usage", "error", err)
		return fmt.Errorf("failed to update storage usage: %w", err)
	}

	logger.Info("Storage usage updated successfully")
	return nil
}

func (s *userService) GetStashByUserID(ctx context.Context, userID string) (*model.Stash, error) {
	logger := slog.With("method", "GetStashByUserID", "userID", userID)

	if userID == "" {
		logger.Warn("Invalid input: empty userID")
		return nil, model.ErrInvalidInput
	}

	stash, err := s.userRepo.GetStashByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get stash", "error", err)
		return nil, err
	}

	logger.Info("Stash retrieved successfully")
	return stash, nil
}
