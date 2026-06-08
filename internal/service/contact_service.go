package service

import (
	"context"
	"errors"
	"strings"

	"art-backend/internal/model"
	"art-backend/internal/repository"
)

var ErrInvalidContact = errors.New("invalid contact message")

type ContactService struct {
	contacts *repository.ContactRepository
}

func NewContactService(contacts *repository.ContactRepository) *ContactService {
	return &ContactService{contacts: contacts}
}

func (s *ContactService) Create(ctx context.Context, request model.ContactRequest) (model.ContactMessage, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Message = strings.TrimSpace(request.Message)

	if request.Name == "" || request.Email == "" || request.Message == "" || !strings.Contains(request.Email, "@") {
		return model.ContactMessage{}, ErrInvalidContact
	}

	return s.contacts.Create(ctx, request)
}
