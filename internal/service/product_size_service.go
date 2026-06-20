package service

import (
	"context"

	"art-backend/internal/model"
	"art-backend/internal/repository"
)

type ProductSizeService struct {
	repository *repository.ProductSizeRepository
}

func NewProductSizeService(repository *repository.ProductSizeRepository) *ProductSizeService {
	return &ProductSizeService{repository: repository}
}

func (s *ProductSizeService) GetAll(ctx context.Context) ([]model.ProductSize, error) {
	return s.repository.FindAll(ctx)
}
