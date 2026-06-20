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

type ArtStyleController struct {
	service *service.ArtStyleService
}

func NewArtStyleController(service *service.ArtStyleService) *ArtStyleController {
	return &ArtStyleController{service: service}
}

func (c *ArtStyleController) GetAll(w http.ResponseWriter, r *http.Request) {
	items, err := c.service.GetAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch art styles", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "art styles fetched successfully", items)
}

func (c *ArtStyleController) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid art style id", "id must be a positive number")
		return
	}

	item, err := c.service.GetByID(r.Context(), id)
	if errors.Is(err, service.ErrArtStyleNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "art style not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch art style", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "art style fetched successfully", item)
}

func (c *ArtStyleController) Create(w http.ResponseWriter, r *http.Request) {
	var request model.SaveArtStyleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	item, err := c.service.Create(r.Context(), request)
	if errors.Is(err, service.ErrInvalidArtStyle) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "origin, style, and image_url are required; tags is optional")
		return
	}
	if errors.Is(err, service.ErrDuplicateArtStyleName) {
		response.Error(w, http.StatusConflict, "CONFLICT", "art style already exists", "use a unique style name")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create art style", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "art style created successfully", item)
}

func (c *ArtStyleController) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid art style id", "id must be a positive number")
		return
	}

	var request model.SaveArtStyleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	item, err := c.service.Update(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidArtStyle) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "origin, style, and image_url are required; tags is optional")
		return
	}
	if errors.Is(err, service.ErrArtStyleNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "art style not found", "")
		return
	}
	if errors.Is(err, service.ErrDuplicateArtStyleName) {
		response.Error(w, http.StatusConflict, "CONFLICT", "art style already exists", "use a unique style name")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not update art style", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "art style updated successfully", item)
}

func (c *ArtStyleController) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid art style id", "id must be a positive number")
		return
	}

	err = c.service.Delete(r.Context(), id)
	if errors.Is(err, service.ErrArtStyleNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "art style not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not delete art style", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "art style deleted successfully", nil)
}
