package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"art-backend/internal/model"
	"art-backend/internal/response"
	"art-backend/internal/service"
)

type AuthController struct {
	service *service.AuthService
}

func NewAuthController(service *service.AuthService) *AuthController {
	return &AuthController{service: service}
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var request model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	auth, err := c.service.Register(r.Context(), request)
	if errors.Is(err, service.ErrInvalidRegister) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "first_name, last_name, email, and password with at least 6 characters are required")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not register user", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "user registered successfully", auth)
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var request model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	auth, err := c.service.Login(r.Context(), request)
	if errors.Is(err, service.ErrInvalidLogin) {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid email or password", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not login", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "user logged in successfully", auth)
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r)
	c.service.Logout(token)

	response.JSON(w, http.StatusOK, "user logged out successfully", nil)
}
