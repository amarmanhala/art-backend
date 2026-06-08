package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"art-backend/internal/model"
	"art-backend/internal/response"
	"art-backend/internal/service"
)

type ProfileController struct {
	service *service.UserService
}

func NewProfileController(service *service.UserService) *ProfileController {
	return &ProfileController{service: service}
}

func (c *ProfileController) Get(w http.ResponseWriter, r *http.Request) {
	user, err := c.service.GetProfile(r.Context(), CurrentUserID(r))
	if errors.Is(err, service.ErrUserNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "user not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch profile", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "profile fetched successfully", user)
}

func (c *ProfileController) Update(w http.ResponseWriter, r *http.Request) {
	var request model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	user, err := c.service.UpdateProfile(r.Context(), CurrentUserID(r), request)
	if errors.Is(err, service.ErrInvalidProfile) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "first_name and last_name are required")
		return
	}
	if errors.Is(err, service.ErrUserNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "user not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not update profile", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "profile updated successfully", user)
}
