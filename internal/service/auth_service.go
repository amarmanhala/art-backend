package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"art-backend/internal/model"
	"art-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidRegister = errors.New("invalid register request")
	ErrInvalidLogin    = errors.New("invalid login")
	ErrUserNotFound    = errors.New("user not found")
)

type AuthService struct {
	users  *repository.UserRepository
	tokens *TokenStore
}

func NewAuthService(users *repository.UserRepository, tokens *TokenStore) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

func (s *AuthService) Register(ctx context.Context, request model.RegisterRequest) (model.AuthResponse, error) {
	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))

	if request.FirstName == "" || request.LastName == "" || request.Email == "" || len(request.Password) < 6 {
		return model.AuthResponse{}, ErrInvalidRegister
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.AuthResponse{}, err
	}

	user, err := s.users.Create(ctx, request, string(passwordHash))
	if err != nil {
		return model.AuthResponse{}, err
	}

	token, err := s.tokens.Create(user.ID)
	if err != nil {
		return model.AuthResponse{}, err
	}

	return model.AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, request model.LoginRequest) (model.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(request.Email))
	if email == "" || request.Password == "" {
		return model.AuthResponse{}, ErrInvalidLogin
	}

	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return model.AuthResponse{}, ErrInvalidLogin
	}
	if err != nil {
		return model.AuthResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		return model.AuthResponse{}, ErrInvalidLogin
	}

	token, err := s.tokens.Create(user.ID)
	if err != nil {
		return model.AuthResponse{}, err
	}

	return model.AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) Logout(token string) {
	s.tokens.Delete(token)
}

func (s *AuthService) GetUserIDByToken(token string) (int64, bool) {
	return s.tokens.GetUserID(token)
}
