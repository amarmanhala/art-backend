package service

import (
	"context"
	"database/sql"
	"errors"

	"art-backend/internal/model"
	"art-backend/internal/repository"
)

var (
	ErrInvalidCartItem  = errors.New("invalid cart item")
	ErrCartItemNotFound = errors.New("cart item not found")
)

type CartService struct {
	carts *repository.CartRepository
}

func NewCartService(carts *repository.CartRepository) *CartService {
	return &CartService{carts: carts}
}

func (s *CartService) GetCart(ctx context.Context, userID int64) (model.CartResponse, error) {
	cart, err := s.carts.GetOrCreate(ctx, userID)
	if err != nil {
		return model.CartResponse{}, err
	}

	return s.buildCartResponse(ctx, cart)
}

func (s *CartService) AddItem(ctx context.Context, userID int64, request model.AddCartItemRequest) (model.CartResponse, error) {
	if request.ProductVariantID <= 0 || request.Quantity <= 0 {
		return model.CartResponse{}, ErrInvalidCartItem
	}

	cart, err := s.carts.GetOrCreate(ctx, userID)
	if err != nil {
		return model.CartResponse{}, err
	}

	err = s.carts.AddItem(ctx, cart.ID, request)
	if errors.Is(err, sql.ErrNoRows) {
		return model.CartResponse{}, ErrInvalidCartItem
	}
	if err != nil {
		return model.CartResponse{}, err
	}

	return s.buildCartResponse(ctx, cart)
}

func (s *CartService) UpdateItem(ctx context.Context, userID int64, itemID int64, request model.UpdateCartItemRequest) (model.CartResponse, error) {
	if itemID <= 0 || request.Quantity <= 0 {
		return model.CartResponse{}, ErrInvalidCartItem
	}

	cart, err := s.carts.GetOrCreate(ctx, userID)
	if err != nil {
		return model.CartResponse{}, err
	}

	err = s.carts.UpdateItem(ctx, cart.ID, itemID, request.Quantity)
	if errors.Is(err, sql.ErrNoRows) {
		return model.CartResponse{}, ErrCartItemNotFound
	}
	if err != nil {
		return model.CartResponse{}, err
	}

	return s.buildCartResponse(ctx, cart)
}

func (s *CartService) DeleteItem(ctx context.Context, userID int64, itemID int64) (model.CartResponse, error) {
	if itemID <= 0 {
		return model.CartResponse{}, ErrInvalidCartItem
	}

	cart, err := s.carts.GetOrCreate(ctx, userID)
	if err != nil {
		return model.CartResponse{}, err
	}

	err = s.carts.DeleteItem(ctx, cart.ID, itemID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.CartResponse{}, ErrCartItemNotFound
	}
	if err != nil {
		return model.CartResponse{}, err
	}

	return s.buildCartResponse(ctx, cart)
}

func (s *CartService) Clear(ctx context.Context, userID int64) (model.CartResponse, error) {
	cart, err := s.carts.GetOrCreate(ctx, userID)
	if err != nil {
		return model.CartResponse{}, err
	}

	if err := s.carts.Clear(ctx, cart.ID); err != nil {
		return model.CartResponse{}, err
	}

	return s.buildCartResponse(ctx, cart)
}

func (s *CartService) buildCartResponse(ctx context.Context, cart model.Cart) (model.CartResponse, error) {
	items, err := s.carts.FindItems(ctx, cart.ID)
	if err != nil {
		return model.CartResponse{}, err
	}

	totalItems := 0
	totalPrice := 0.0
	for _, item := range items {
		totalItems += item.Quantity
		totalPrice += item.Subtotal
	}

	return model.CartResponse{
		ID:         cart.ID,
		UserID:     cart.UserID,
		Items:      items,
		TotalItems: totalItems,
		TotalPrice: totalPrice,
		CreatedAt:  cart.CreatedAt,
		UpdatedAt:  cart.UpdatedAt,
	}, nil
}
