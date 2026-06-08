package repository

import (
	"context"
	"database/sql"

	"art-backend/internal/model"
)

type ContactRepository struct {
	db *sql.DB
}

func NewContactRepository(db *sql.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) Create(ctx context.Context, request model.ContactRequest) (model.ContactMessage, error) {
	var contact model.ContactMessage

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO contact_messages (name, email, message)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, message, created_at
	`, request.Name, request.Email, request.Message).Scan(
		&contact.ID,
		&contact.Name,
		&contact.Email,
		&contact.Message,
		&contact.CreatedAt,
	)

	return contact, err
}
