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

type FrameController struct {
	service      *service.FrameService
	imageService *service.FrameImageService
}

func NewFrameController(service *service.FrameService, imageService *service.FrameImageService) *FrameController {
	return &FrameController{service: service, imageService: imageService}
}

func (c *FrameController) GetAll(w http.ResponseWriter, r *http.Request) {
	frames, err := c.service.GetAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch frames", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "frames fetched successfully", frames)
}

func (c *FrameController) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid frame id", "id must be a positive number")
		return
	}

	frame, err := c.service.GetByID(r.Context(), id)
	if errors.Is(err, service.ErrFrameNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "frame not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch frame", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "frame fetched successfully", frame)
}

func (c *FrameController) Create(w http.ResponseWriter, r *http.Request) {
	var request model.SaveFrameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	frame, err := c.service.Create(r.Context(), request)
	if errors.Is(err, service.ErrInvalidFrame) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "vendor_name, frame_name, color, article_number, price, and measurements are required")
		return
	}
	if errors.Is(err, service.ErrDuplicateFrameArticleNo) {
		response.Error(w, http.StatusConflict, "CONFLICT", "frame article number already exists", "use a unique article number for this vendor")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create frame", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "frame created successfully", frame)
}

func (c *FrameController) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid frame id", "id must be a positive number")
		return
	}

	var request model.SaveFrameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	frame, err := c.service.Update(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidFrame) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "vendor_name, frame_name, color, article_number, price, and measurements are required")
		return
	}
	if errors.Is(err, service.ErrFrameNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "frame not found", "")
		return
	}
	if errors.Is(err, service.ErrDuplicateFrameArticleNo) {
		response.Error(w, http.StatusConflict, "CONFLICT", "frame article number already exists", "use a unique article number for this vendor")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not update frame", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "frame updated successfully", frame)
}

func (c *FrameController) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid frame id", "id must be a positive number")
		return
	}

	err = c.service.Delete(r.Context(), id)
	if errors.Is(err, service.ErrFrameNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "frame not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not delete frame", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "frame deleted successfully", nil)
}

func (c *FrameController) ReplaceImages(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid frame id", "id must be a positive number")
		return
	}

	var request model.SaveFrameImagesRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	frame, err := c.imageService.ReplaceImages(r.Context(), id, request)
	if errors.Is(err, service.ErrInvalidFrameImages) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "main_image is required and total images must be at most 10 frame original blob names")
		return
	}
	if errors.Is(err, service.ErrFrameNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "frame not found", "")
		return
	}
	if errors.Is(err, service.ErrFrameImageStorageNotConfigured) {
		response.Error(w, http.StatusInternalServerError, "AZURE_NOT_CONFIGURED", "azure frame image storage is not configured", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not save frame images", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "frame images saved successfully", frame)
}
