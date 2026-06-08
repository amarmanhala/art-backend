package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"art-backend/internal/model"
	"art-backend/internal/response"
	"art-backend/internal/service"
)

type ContactController struct {
	service *service.ContactService
}

func NewContactController(service *service.ContactService) *ContactController {
	return &ContactController{service: service}
}

func (c *ContactController) Create(w http.ResponseWriter, r *http.Request) {
	var request model.ContactRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	contact, err := c.service.Create(r.Context(), request)
	if errors.Is(err, service.ErrInvalidContact) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "name, valid email, and message are required")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not send contact message", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "contact message sent successfully", contact)
}
