package service

import (
	"errors"
	"log/slog"

	"golang.org/x/crypto/bcrypt"

	"github.com/mirai-box/mirai-box/internal/repository"
	"github.com/mirai-box/mirai-box/internal/model"
)

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *userService {
	return &userService{repo: repo}
}

func (s *userService) Authenticate(username, password string) (*model.User, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		slog.Error("can't find user", "error", err, "user", user)
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		slog.Error("bcrypt: invalid credentials", "error", err, "user", user)
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *userService) FindByID(id string) (*model.User, error) {
	return s.repo.FindByID(id)
}

// HashPassword hashes the password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
