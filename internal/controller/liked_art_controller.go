package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"art-backend/internal/model"
	"art-backend/internal/response"
	"art-backend/internal/service"
)

type LikedArtController struct {
	service *service.LikedArtService
}

func NewLikedArtController(service *service.LikedArtService) *LikedArtController {
	return &LikedArtController{service: service}
}

func (c *LikedArtController) GetAll(w http.ResponseWriter, r *http.Request) {
	likedArts, err := c.service.GetLikedArts(r.Context(), CurrentUserID(r))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch liked arts", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "liked arts fetched successfully", likedArts)
}

func (c *LikedArtController) Save(w http.ResponseWriter, r *http.Request) {
	var request model.SaveLikedArtRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	likedArt, err := c.service.SetLikedArt(r.Context(), CurrentUserID(r), request)
	if errors.Is(err, service.ErrInvalidLikedArt) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "product_id must be valid and action must be liked or disliked")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not save liked art", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "liked art saved successfully", likedArt)
}
