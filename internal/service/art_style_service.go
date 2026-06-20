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
	ErrArtStyleNotFound      = errors.New("art style not found")
	ErrInvalidArtStyle       = errors.New("invalid art style")
	ErrDuplicateArtStyleName = errors.New("duplicate art style")
)

type ArtStyleService struct {
	repository *repository.ArtStyleRepository
}

func NewArtStyleService(repository *repository.ArtStyleRepository) *ArtStyleService {
	return &ArtStyleService{repository: repository}
}

func (s *ArtStyleService) GetAll(ctx context.Context) ([]model.ArtStyle, error) {
	return s.repository.FindAll(ctx)
}

func (s *ArtStyleService) GetByID(ctx context.Context, id int64) (model.ArtStyle, error) {
	if id <= 0 {
		return model.ArtStyle{}, ErrArtStyleNotFound
	}

	item, err := s.repository.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ArtStyle{}, ErrArtStyleNotFound
	}

	return item, err
}

func (s *ArtStyleService) Create(ctx context.Context, request model.SaveArtStyleRequest) (model.ArtStyle, error) {
	request = normalizeArtStyleRequest(request)
	if request.Origin == "" || request.Style == "" || request.ImageURL == "" {
		return model.ArtStyle{}, ErrInvalidArtStyle
	}

	item, err := s.repository.Create(ctx, request)
	if isUniqueArtStyleError(err) {
		return model.ArtStyle{}, ErrDuplicateArtStyleName
	}

	return item, err
}

func (s *ArtStyleService) Update(ctx context.Context, id int64, request model.SaveArtStyleRequest) (model.ArtStyle, error) {
	if id <= 0 {
		return model.ArtStyle{}, ErrArtStyleNotFound
	}

	request = normalizeArtStyleRequest(request)
	if request.Origin == "" || request.Style == "" || request.ImageURL == "" {
		return model.ArtStyle{}, ErrInvalidArtStyle
	}

	item, err := s.repository.Update(ctx, id, request)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ArtStyle{}, ErrArtStyleNotFound
	}
	if isUniqueArtStyleError(err) {
		return model.ArtStyle{}, ErrDuplicateArtStyleName
	}

	return item, err
}

func (s *ArtStyleService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrArtStyleNotFound
	}

	deleted, err := s.repository.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrArtStyleNotFound
	}

	return nil
}

func normalizeArtStyleRequest(request model.SaveArtStyleRequest) model.SaveArtStyleRequest {
	request.Origin = strings.ToLower(strings.TrimSpace(request.Origin))
	request.Style = strings.TrimSpace(request.Style)
	request.Description = strings.TrimSpace(request.Description)
	request.Tags = normalizeTags(request.Tags)
	request.ImageURL = strings.TrimSpace(request.ImageURL)
	request.BlobName = strings.TrimSpace(request.BlobName)
	return request
}

func normalizeTags(tags []string) []string {
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		normalized = append(normalized, tag)
	}

	return normalized
}

func isUniqueArtStyleError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "art_styles_origin_style_key"
}
