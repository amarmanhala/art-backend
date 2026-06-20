package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"art-backend/internal/model"
)

func TestLikedArtServiceSetAndGetFlow(t *testing.T) {
	store := newFakeLikedArtStore()
	store.products[1] = model.Product{ID: 1, Title: "First", IsActive: true}
	store.products[2] = model.Product{ID: 2, Title: "Second", IsActive: true}
	service := NewLikedArtService(store)

	if _, err := service.SetLikedArt(context.Background(), 10, model.SaveLikedArtRequest{
		ProductID: 1,
		Action:    model.LikedArtStatusLiked,
	}); err != nil {
		t.Fatalf("SetLikedArt liked returned error: %v", err)
	}

	if _, err := service.SetLikedArt(context.Background(), 10, model.SaveLikedArtRequest{
		ProductID: 1,
		Action:    model.LikedArtStatusLiked,
	}); err != nil {
		t.Fatalf("repeated SetLikedArt liked returned error: %v", err)
	}

	likedArts, err := service.GetLikedArts(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetLikedArts returned error: %v", err)
	}
	if len(likedArts) != 1 || likedArts[0].ProductID != 1 || likedArts[0].Status != model.LikedArtStatusLiked {
		t.Fatalf("expected one liked art for product 1, got %#v", likedArts)
	}

	if _, err := service.SetLikedArt(context.Background(), 10, model.SaveLikedArtRequest{
		ProductID: 1,
		Action:    model.LikedArtStatusDisliked,
	}); err != nil {
		t.Fatalf("SetLikedArt disliked returned error: %v", err)
	}

	if _, err := service.SetLikedArt(context.Background(), 10, model.SaveLikedArtRequest{
		ProductID: 1,
		Action:    model.LikedArtStatusDisliked,
	}); err != nil {
		t.Fatalf("repeated SetLikedArt disliked returned error: %v", err)
	}

	likedArts, err = service.GetLikedArts(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetLikedArts after disliked returned error: %v", err)
	}
	if len(likedArts) != 0 {
		t.Fatalf("expected disliked art to be omitted, got %#v", likedArts)
	}

	if _, err := service.SetLikedArt(context.Background(), 10, model.SaveLikedArtRequest{
		ProductID: 2,
		Action:    model.LikedArtStatusLiked,
	}); err != nil {
		t.Fatalf("SetLikedArt second product returned error: %v", err)
	}

	if _, err := service.SetLikedArt(context.Background(), 20, model.SaveLikedArtRequest{
		ProductID: 1,
		Action:    model.LikedArtStatusLiked,
	}); err != nil {
		t.Fatalf("SetLikedArt other user returned error: %v", err)
	}

	userTenLikedArts, err := service.GetLikedArts(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetLikedArts user 10 returned error: %v", err)
	}
	if len(userTenLikedArts) != 1 || userTenLikedArts[0].ProductID != 2 {
		t.Fatalf("expected user 10 to only have product 2 liked, got %#v", userTenLikedArts)
	}

	userTwentyLikedArts, err := service.GetLikedArts(context.Background(), 20)
	if err != nil {
		t.Fatalf("GetLikedArts user 20 returned error: %v", err)
	}
	if len(userTwentyLikedArts) != 1 || userTwentyLikedArts[0].ProductID != 1 {
		t.Fatalf("expected user 20 to only have product 1 liked, got %#v", userTwentyLikedArts)
	}
}

func TestLikedArtServiceRejectsInvalidRequests(t *testing.T) {
	store := newFakeLikedArtStore()
	store.products[1] = model.Product{ID: 1, IsActive: true}
	service := NewLikedArtService(store)

	tests := []struct {
		name    string
		userID  int64
		request model.SaveLikedArtRequest
	}{
		{
			name:   "invalid user id",
			userID: 0,
			request: model.SaveLikedArtRequest{
				ProductID: 1,
				Action:    model.LikedArtStatusLiked,
			},
		},
		{
			name:   "invalid product id",
			userID: 10,
			request: model.SaveLikedArtRequest{
				ProductID: 0,
				Action:    model.LikedArtStatusLiked,
			},
		},
		{
			name:   "invalid action",
			userID: 10,
			request: model.SaveLikedArtRequest{
				ProductID: 1,
				Action:    "favorite",
			},
		},
		{
			name:   "missing product",
			userID: 10,
			request: model.SaveLikedArtRequest{
				ProductID: 999,
				Action:    model.LikedArtStatusLiked,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.SetLikedArt(context.Background(), tt.userID, tt.request)
			if !errors.Is(err, ErrInvalidLikedArt) {
				t.Fatalf("expected ErrInvalidLikedArt, got %v", err)
			}
		})
	}
}

type fakeLikedArtKey struct {
	userID    int64
	productID int64
}

type fakeLikedArtStore struct {
	nextID   int64
	records  map[fakeLikedArtKey]model.LikedArt
	products map[int64]model.Product
}

func newFakeLikedArtStore() *fakeLikedArtStore {
	return &fakeLikedArtStore{
		nextID:   1,
		records:  make(map[fakeLikedArtKey]model.LikedArt),
		products: make(map[int64]model.Product),
	}
}

func (s *fakeLikedArtStore) FindLikedByUser(_ context.Context, userID int64) ([]model.LikedArt, error) {
	likedArts := make([]model.LikedArt, 0)
	for _, likedArt := range s.records {
		if likedArt.UserID == userID && likedArt.Status == model.LikedArtStatusLiked {
			likedArts = append(likedArts, likedArt)
		}
	}

	return likedArts, nil
}

func (s *fakeLikedArtStore) SetStatus(_ context.Context, userID int64, productID int64, status string) (model.LikedArt, error) {
	product, ok := s.products[productID]
	if !ok || !product.IsActive {
		return model.LikedArt{}, sql.ErrNoRows
	}

	key := fakeLikedArtKey{userID: userID, productID: productID}
	now := time.Now()
	likedArt, ok := s.records[key]
	if !ok {
		likedArt = model.LikedArt{
			ID:        s.nextID,
			UserID:    userID,
			ProductID: productID,
			CreatedAt: now,
		}
		s.nextID++
	}

	likedArt.Status = status
	likedArt.Product = product
	likedArt.UpdatedAt = now
	s.records[key] = likedArt

	return likedArt, nil
}
