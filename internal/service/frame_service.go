package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"art-backend/internal/model"
	"art-backend/internal/repository"

	"github.com/lib/pq"
)

var (
	ErrFrameNotFound           = errors.New("frame not found")
	ErrInvalidFrame            = errors.New("invalid frame")
	ErrDuplicateFrameArticleNo = errors.New("duplicate frame article number")
)

type FrameService struct {
	repository *repository.FrameRepository
}

func NewFrameService(repository *repository.FrameRepository) *FrameService {
	return &FrameService{repository: repository}
}

func (s *FrameService) GetAll(ctx context.Context) ([]model.Frame, error) {
	return s.repository.FindAll(ctx)
}

func (s *FrameService) GetByID(ctx context.Context, id int64) (model.Frame, error) {
	if id <= 0 {
		return model.Frame{}, ErrFrameNotFound
	}

	frame, err := s.repository.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Frame{}, ErrFrameNotFound
	}

	return frame, err
}

func (s *FrameService) Create(ctx context.Context, request model.SaveFrameRequest) (model.Frame, error) {
	request = sanitizeFrameRequest(request)
	if !isValidFrameRequest(request) {
		return model.Frame{}, ErrInvalidFrame
	}

	frame, err := s.repository.Create(ctx, request)
	if isUniqueFrameArticleNumberError(err) {
		return model.Frame{}, ErrDuplicateFrameArticleNo
	}

	return frame, err
}

func (s *FrameService) Update(ctx context.Context, id int64, request model.SaveFrameRequest) (model.Frame, error) {
	if id <= 0 {
		return model.Frame{}, ErrFrameNotFound
	}

	request = sanitizeFrameRequest(request)
	if !isValidFrameRequest(request) {
		return model.Frame{}, ErrInvalidFrame
	}

	frame, err := s.repository.Update(ctx, id, request)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Frame{}, ErrFrameNotFound
	}
	if isUniqueFrameArticleNumberError(err) {
		return model.Frame{}, ErrDuplicateFrameArticleNo
	}

	return frame, err
}

func (s *FrameService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrFrameNotFound
	}

	deleted, err := s.repository.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrFrameNotFound
	}

	return nil
}

func sanitizeFrameRequest(request model.SaveFrameRequest) model.SaveFrameRequest {
	request.VendorName = strings.TrimSpace(request.VendorName)
	request.FrameName = strings.TrimSpace(request.FrameName)
	request.Color = strings.TrimSpace(request.Color)
	request.Description = strings.TrimSpace(request.Description)
	request.ArticleNumber = strings.TrimSpace(request.ArticleNumber)
	request.ProductDetail = strings.TrimSpace(request.ProductDetail)
	request.MaterialDescription = strings.TrimSpace(request.MaterialDescription)
	request.Care = strings.TrimSpace(request.Care)

	return request
}

func isValidFrameRequest(request model.SaveFrameRequest) bool {
	if request.VendorName == "" ||
		request.FrameName == "" ||
		request.Color == "" ||
		request.ArticleNumber == "" ||
		request.Price <= 0 {
		return false
	}

	return validateFrameMeasurements(request.Measurements)
}

func validateFrameMeasurements(measurements model.FrameMeasurements) bool {
	values := []float64{
		measurements.PictureWidthCm,
		measurements.PictureWidthIn,
		measurements.PictureHeightCm,
		measurements.PictureHeightIn,
		measurements.FrameWidthCm,
		measurements.FrameWidthIn,
		measurements.FrameHeightCm,
		measurements.FrameHeightIn,
		measurements.FrameDepthCm,
		measurements.FrameDepthIn,
	}

	for _, value := range values {
		if value <= 0 {
			return false
		}
	}

	return true
}

func isUniqueFrameArticleNumberError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "frames_vendor_name_article_number_key"
}
