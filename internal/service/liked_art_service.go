package service

import (
	"context"
	"database/sql"
	"errors"

	"art-backend/internal/model"
)

var ErrInvalidLikedArt = errors.New("invalid liked art")

type LikedArtStore interface {
	FindLikedByUser(ctx context.Context, userID int64) ([]model.LikedArt, error)
	SetStatus(ctx context.Context, userID int64, productID int64, status string) (model.LikedArt, error)
}

type LikedArtService struct {
	likedArts LikedArtStore
}

func NewLikedArtService(likedArts LikedArtStore) *LikedArtService {
	return &LikedArtService{likedArts: likedArts}
}

func (s *LikedArtService) GetLikedArts(ctx context.Context, userID int64) ([]model.LikedArt, error) {
	return s.likedArts.FindLikedByUser(ctx, userID)
}

func (s *LikedArtService) SetLikedArt(ctx context.Context, userID int64, request model.SaveLikedArtRequest) (model.LikedArt, error) {
	if userID <= 0 || request.ProductID <= 0 || !isValidLikedArtAction(request.Action) {
		return model.LikedArt{}, ErrInvalidLikedArt
	}

	likedArt, err := s.likedArts.SetStatus(ctx, userID, request.ProductID, request.Action)
	if errors.Is(err, sql.ErrNoRows) {
		return model.LikedArt{}, ErrInvalidLikedArt
	}
	if err != nil {
		return model.LikedArt{}, err
	}

	return likedArt, nil
}

func isValidLikedArtAction(action string) bool {
	return action == model.LikedArtStatusLiked || action == model.LikedArtStatusDisliked
}
