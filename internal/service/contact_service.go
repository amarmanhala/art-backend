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
	ErrInvalidContact  = errors.New("invalid contact message")
	ErrContactNotFound = errors.New("contact message not found")
)

type ContactService struct {
	contacts *repository.ContactRepository
}

func NewContactService(contacts *repository.ContactRepository) *ContactService {
	return &ContactService{contacts: contacts}
}

func (s *ContactService) Create(ctx context.Context, request model.ContactRequest) (model.ContactMessage, error) {
	request = normalizeContactRequest(request)

	if request.Name == "" || request.Email == "" || request.Message == "" || !strings.Contains(request.Email, "@") {
		return model.ContactMessage{}, ErrInvalidContact
	}

	return s.contacts.Create(ctx, request)
}

func (s *ContactService) GetAll(ctx context.Context) ([]model.ContactMessage, error) {
	return s.contacts.FindAll(ctx)
}

func (s *ContactService) GetByID(ctx context.Context, id int64) (model.ContactMessage, error) {
	if id <= 0 {
		return model.ContactMessage{}, ErrContactNotFound
	}

	contact, err := s.contacts.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ContactMessage{}, ErrContactNotFound
	}

	return contact, err
}

func (s *ContactService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrContactNotFound
	}

	deleted, err := s.contacts.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrContactNotFound
	}

	return nil
}

func normalizeContactRequest(request model.ContactRequest) model.ContactRequest {
	request.Name = strings.TrimSpace(request.Name)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Message = strings.TrimSpace(request.Message)
	if request.OrderNumber != nil {
		value := strings.TrimSpace(*request.OrderNumber)
		if value == "" {
			request.OrderNumber = nil
		} else {
			request.OrderNumber = &value
		}
	}

	return request
}
