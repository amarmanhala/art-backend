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
	var orderNumber sql.NullString
	var orderNumberValue any
	if request.OrderNumber != nil {
		orderNumberValue = *request.OrderNumber
	}

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO contact_requests (name, email, order_number, message)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, order_number, message, created_at
	`, request.Name, request.Email, orderNumberValue, request.Message).Scan(
		&contact.ID,
		&contact.Name,
		&contact.Email,
		&orderNumber,
		&contact.Message,
		&contact.CreatedAt,
	)
	if err != nil {
		return model.ContactMessage{}, err
	}
	if orderNumber.Valid {
		value := orderNumber.String
		contact.OrderNumber = &value
	}

	return contact, err
}

func (r *ContactRepository) FindAll(ctx context.Context) ([]model.ContactMessage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, email, order_number, message, created_at
		FROM contact_requests
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ContactMessage, 0)
	for rows.Next() {
		item, err := scanContactMessage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *ContactRepository) FindByID(ctx context.Context, id int64) (model.ContactMessage, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, order_number, message, created_at
		FROM contact_requests
		WHERE id = $1
	`, id)

	return scanContactRow(row)
}

func (r *ContactRepository) Delete(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM contact_requests
		WHERE id = $1
	`, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func scanContactMessage(scanner interface {
	Scan(...any) error
}) (model.ContactMessage, error) {
	var contact model.ContactMessage
	var orderNumber sql.NullString
	if err := scanner.Scan(
		&contact.ID,
		&contact.Name,
		&contact.Email,
		&orderNumber,
		&contact.Message,
		&contact.CreatedAt,
	); err != nil {
		return model.ContactMessage{}, err
	}
	if orderNumber.Valid {
		value := orderNumber.String
		contact.OrderNumber = &value
	}

	return contact, nil
}

func scanContactRow(row *sql.Row) (model.ContactMessage, error) {
	var contact model.ContactMessage
	var orderNumber sql.NullString
	if err := row.Scan(
		&contact.ID,
		&contact.Name,
		&contact.Email,
		&orderNumber,
		&contact.Message,
		&contact.CreatedAt,
	); err != nil {
		return model.ContactMessage{}, err
	}
	if orderNumber.Valid {
		value := orderNumber.String
		contact.OrderNumber = &value
	}

	return contact, nil
}
