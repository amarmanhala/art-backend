package repository

import (
	"context"
	"database/sql"

	"art-backend/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

const userColumns = `
	id, first_name, last_name, email, password_hash, role, is_verified, created_at, updated_at
`

func (r *UserRepository) Create(ctx context.Context, request model.RegisterRequest, passwordHash string) (model.User, error) {
	var user model.User

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (first_name, last_name, email, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+userColumns+`
	`,
		request.FirstName,
		request.LastName,
		request.Email,
		passwordHash,
		model.RoleCustomer,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User

	err := r.db.QueryRowContext(ctx, `
		SELECT `+userColumns+`
		FROM users
		WHERE email = $1
	`, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (model.User, error) {
	var user model.User

	err := r.db.QueryRowContext(ctx, `
		SELECT `+userColumns+`
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID int64, request model.UpdateProfileRequest) (model.User, error) {
	var user model.User

	err := r.db.QueryRowContext(ctx, `
		UPDATE users
		SET first_name = $1, last_name = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING `+userColumns+`
	`, request.FirstName, request.LastName, userID).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}
