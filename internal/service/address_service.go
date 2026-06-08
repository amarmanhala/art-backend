package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"art-backend/internal/model"
	"art-backend/internal/repository"
)

var (
	ErrInvalidAddress  = errors.New("invalid address")
	ErrAddressNotFound = errors.New("address not found")
)

type AddressService struct {
	addresses *repository.AddressRepository
}

func NewAddressService(addresses *repository.AddressRepository) *AddressService {
	return &AddressService{addresses: addresses}
}

func (s *AddressService) GetAll(ctx context.Context, userID int64) ([]model.Address, error) {
	return s.addresses.FindByUserID(ctx, userID)
}

func (s *AddressService) Create(ctx context.Context, userID int64, request model.AddressRequest) (model.Address, error) {
	request = cleanAddress(request)
	if !validAddress(request) {
		return model.Address{}, ErrInvalidAddress
	}

	return s.addresses.Create(ctx, userID, request)
}

func (s *AddressService) Update(ctx context.Context, userID int64, addressID int64, request model.AddressRequest) (model.Address, error) {
	request = cleanAddress(request)
	if addressID <= 0 || !validAddress(request) {
		return model.Address{}, ErrInvalidAddress
	}

	address, err := s.addresses.Update(ctx, userID, addressID, request)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Address{}, ErrAddressNotFound
	}

	return address, err
}

func (s *AddressService) Delete(ctx context.Context, userID int64, addressID int64) error {
	if addressID <= 0 {
		return ErrInvalidAddress
	}

	err := s.addresses.Delete(ctx, userID, addressID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAddressNotFound
	}

	return err
}

func cleanAddress(request model.AddressRequest) model.AddressRequest {
	request.FullName = strings.TrimSpace(request.FullName)
	request.Phone = strings.TrimSpace(request.Phone)
	request.AddressLine1 = strings.TrimSpace(request.AddressLine1)
	request.AddressLine2 = strings.TrimSpace(request.AddressLine2)
	request.City = strings.TrimSpace(request.City)
	request.Province = strings.TrimSpace(request.Province)
	request.PostalCode = strings.TrimSpace(request.PostalCode)
	request.Country = strings.TrimSpace(request.Country)

	return request
}

func validAddress(request model.AddressRequest) bool {
	return request.FullName != "" &&
		request.AddressLine1 != "" &&
		request.City != "" &&
		request.Province != "" &&
		request.PostalCode != "" &&
		request.Country != ""
}
