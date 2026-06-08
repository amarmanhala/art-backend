package repository

import (
	"context"
	"database/sql"

	"art-backend/internal/model"
)

type AddressRepository struct {
	db *sql.DB
}

func NewAddressRepository(db *sql.DB) *AddressRepository {
	return &AddressRepository{db: db}
}

const addressColumns = `
	id, user_id, full_name, phone, address_line_1, address_line_2,
	city, province, postal_code, country, is_default, created_at, updated_at
`

func (r *AddressRepository) FindByUserID(ctx context.Context, userID int64) ([]model.Address, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+addressColumns+`
		FROM addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addresses := make([]model.Address, 0)
	for rows.Next() {
		address, err := scanAddress(rows)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, address)
	}

	return addresses, rows.Err()
}

func (r *AddressRepository) Create(ctx context.Context, userID int64, request model.AddressRequest) (model.Address, error) {
	if request.IsDefault {
		if err := r.clearDefault(ctx, userID); err != nil {
			return model.Address{}, err
		}
	}

	var address model.Address
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO addresses (
			user_id, full_name, phone, address_line_1, address_line_2,
			city, province, postal_code, country, is_default
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING `+addressColumns+`
	`,
		userID,
		request.FullName,
		request.Phone,
		request.AddressLine1,
		request.AddressLine2,
		request.City,
		request.Province,
		request.PostalCode,
		request.Country,
		request.IsDefault,
	).Scan(
		&address.ID,
		&address.UserID,
		&address.FullName,
		&address.Phone,
		&address.AddressLine1,
		&address.AddressLine2,
		&address.City,
		&address.Province,
		&address.PostalCode,
		&address.Country,
		&address.IsDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	return address, err
}

func (r *AddressRepository) Update(ctx context.Context, userID int64, addressID int64, request model.AddressRequest) (model.Address, error) {
	if request.IsDefault {
		if err := r.clearDefault(ctx, userID); err != nil {
			return model.Address{}, err
		}
	}

	var address model.Address
	err := r.db.QueryRowContext(ctx, `
		UPDATE addresses
		SET full_name = $1,
			phone = $2,
			address_line_1 = $3,
			address_line_2 = $4,
			city = $5,
			province = $6,
			postal_code = $7,
			country = $8,
			is_default = $9,
			updated_at = NOW()
		WHERE id = $10 AND user_id = $11
		RETURNING `+addressColumns+`
	`,
		request.FullName,
		request.Phone,
		request.AddressLine1,
		request.AddressLine2,
		request.City,
		request.Province,
		request.PostalCode,
		request.Country,
		request.IsDefault,
		addressID,
		userID,
	).Scan(
		&address.ID,
		&address.UserID,
		&address.FullName,
		&address.Phone,
		&address.AddressLine1,
		&address.AddressLine2,
		&address.City,
		&address.Province,
		&address.PostalCode,
		&address.Country,
		&address.IsDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	return address, err
}

func (r *AddressRepository) Delete(ctx context.Context, userID int64, addressID int64) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM addresses
		WHERE id = $1 AND user_id = $2
	`, addressID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *AddressRepository) clearDefault(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE addresses
		SET is_default = FALSE
		WHERE user_id = $1
	`, userID)

	return err
}

func scanAddress(rows *sql.Rows) (model.Address, error) {
	var address model.Address
	err := rows.Scan(
		&address.ID,
		&address.UserID,
		&address.FullName,
		&address.Phone,
		&address.AddressLine1,
		&address.AddressLine2,
		&address.City,
		&address.Province,
		&address.PostalCode,
		&address.Country,
		&address.IsDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	return address, err
}
