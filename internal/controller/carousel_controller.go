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

type CarouselController struct {
	service *service.CarouselService
}

func NewCarouselController(service *service.CarouselService) *CarouselController {
	return &CarouselController{service: service}
}

func (c *CarouselController) GetActive(w http.ResponseWriter, r *http.Request) {
	items, err := c.service.GetActive(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch carousel items", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "carousel items fetched successfully", items)
}

func (c *CarouselController) GetAll(w http.ResponseWriter, r *http.Request) {
	items, err := c.service.GetAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch carousel items", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "carousel items fetched successfully", items)
}

func (c *CarouselController) Create(w http.ResponseWriter, r *http.Request) {
	var request model.SaveCarouselItemRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	item, err := c.service.Create(r.Context(), request)
	if errors.Is(err, service.ErrInvalidCarouselItems) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "image_url is required")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create carousel item", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "carousel item created successfully", item)
}

func (c *CarouselController) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := readCarouselID(w, r)
	if !ok {
		return
	}

	var request model.SaveCarouselItemRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	item, err := c.service.Update(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidCarouselItems) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "image_url is required")
		return
	}
	if errors.Is(err, service.ErrCarouselItemNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "carousel item not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not update carousel item", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "carousel item updated successfully", item)
}

func (c *CarouselController) SetActive(w http.ResponseWriter, r *http.Request) {
	id, ok := readCarouselID(w, r)
	if !ok {
		return
	}

	var request struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	item, err := c.service.SetActive(r.Context(), id, request.IsActive)
	if errors.Is(err, service.ErrCarouselItemNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "carousel item not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not update carousel status", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "carousel item status updated successfully", item)
}

func (c *CarouselController) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := readCarouselID(w, r)
	if !ok {
		return
	}

	err := c.service.Delete(r.Context(), id)
	if errors.Is(err, service.ErrCarouselItemNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "carousel item not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not delete carousel item", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "carousel item deleted successfully", nil)
}

func (c *CarouselController) ReplaceAll(w http.ResponseWriter, r *http.Request) {
	var request model.SaveCarouselItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	items, err := c.service.ReplaceAll(r.Context(), request)
	if errors.Is(err, service.ErrInvalidCarouselItems) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "at least one carousel item with image_url is required")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not save carousel items", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "carousel items saved successfully", items)
}

func readCarouselID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid carousel item id", "id must be a positive number")
		return 0, false
	}

	return id, true
}
