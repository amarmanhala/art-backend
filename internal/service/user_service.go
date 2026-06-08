package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"art-backend/internal/model"
	"art-backend/internal/repository"
)

var ErrInvalidProfile = errors.New("invalid profile")

type UserService struct {
	users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (model.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}

	return user, err
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, request model.UpdateProfileRequest) (model.User, error) {
	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)

	if request.FirstName == "" || request.LastName == "" {
		return model.User{}, ErrInvalidProfile
	}

	user, err := s.users.UpdateProfile(ctx, userID, request)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}

	return user, err
}
