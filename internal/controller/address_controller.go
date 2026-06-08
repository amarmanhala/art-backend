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

type AddressController struct {
	service *service.AddressService
}

func NewAddressController(service *service.AddressService) *AddressController {
	return &AddressController{service: service}
}

func (c *AddressController) GetAll(w http.ResponseWriter, r *http.Request) {
	addresses, err := c.service.GetAll(r.Context(), CurrentUserID(r))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch addresses", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "addresses fetched successfully", addresses)
}

func (c *AddressController) Create(w http.ResponseWriter, r *http.Request) {
	var request model.AddressRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	address, err := c.service.Create(r.Context(), CurrentUserID(r), request)
	if errors.Is(err, service.ErrInvalidAddress) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "full_name, address_line_1, city, province, postal_code, and country are required")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create address", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "address created successfully", address)
}

func (c *AddressController) Update(w http.ResponseWriter, r *http.Request) {
	addressID, ok := readAddressID(w, r)
	if !ok {
		return
	}

	var request model.AddressRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	address, err := c.service.Update(r.Context(), CurrentUserID(r), addressID, request)
	if errors.Is(err, service.ErrInvalidAddress) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "address id and required address fields are required")
		return
	}
	if errors.Is(err, service.ErrAddressNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "address not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not update address", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "address updated successfully", address)
}

func (c *AddressController) Delete(w http.ResponseWriter, r *http.Request) {
	addressID, ok := readAddressID(w, r)
	if !ok {
		return
	}

	err := c.service.Delete(r.Context(), CurrentUserID(r), addressID)
	if errors.Is(err, service.ErrInvalidAddress) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid address id", "id must be a positive number")
		return
	}
	if errors.Is(err, service.ErrAddressNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "address not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not delete address", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "address deleted successfully", nil)
}

func readAddressID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid address id", "id must be a positive number")
		return 0, false
	}

	return id, true
}
