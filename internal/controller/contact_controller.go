package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "name, valid email, and message are required; order_number is optional")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not send contact message", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "contact message sent successfully", contact)
}

func (c *ContactController) GetAll(w http.ResponseWriter, r *http.Request) {
	contacts, err := c.service.GetAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch contact messages", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "contact messages fetched successfully", contacts)
}

func (c *ContactController) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid contact message id", "id must be a positive number")
		return
	}

	contact, err := c.service.GetByID(r.Context(), id)
	if errors.Is(err, service.ErrContactNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "contact message not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch contact message", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "contact message fetched successfully", contact)
}

func (c *ContactController) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid contact message id", "id must be a positive number")
		return
	}

	err = c.service.Delete(r.Context(), id)
	if errors.Is(err, service.ErrContactNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "contact message not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not delete contact message", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "contact message deleted successfully", nil)
}
