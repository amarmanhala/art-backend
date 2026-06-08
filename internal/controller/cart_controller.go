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

type CartController struct {
	service *service.CartService
}

func NewCartController(service *service.CartService) *CartController {
	return &CartController{service: service}
}

func (c *CartController) Get(w http.ResponseWriter, r *http.Request) {
	cart, err := c.service.GetCart(r.Context(), CurrentUserID(r))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch cart", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "cart fetched successfully", cart)
}

func (c *CartController) AddItem(w http.ResponseWriter, r *http.Request) {
	var request model.AddCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	cart, err := c.service.AddItem(r.Context(), CurrentUserID(r), request)
	if errors.Is(err, service.ErrInvalidCartItem) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "product_id and quantity greater than 0 are required")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not add cart item", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "cart item added successfully", cart)
}

func (c *CartController) UpdateItem(w http.ResponseWriter, r *http.Request) {
	itemID, ok := readCartItemID(w, r)
	if !ok {
		return
	}

	var request model.UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	cart, err := c.service.UpdateItem(r.Context(), CurrentUserID(r), itemID, request)
	if errors.Is(err, service.ErrInvalidCartItem) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "quantity must be greater than 0")
		return
	}
	if errors.Is(err, service.ErrCartItemNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "cart item not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not update cart item", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "cart item updated successfully", cart)
}

func (c *CartController) DeleteItem(w http.ResponseWriter, r *http.Request) {
	itemID, ok := readCartItemID(w, r)
	if !ok {
		return
	}

	cart, err := c.service.DeleteItem(r.Context(), CurrentUserID(r), itemID)
	if errors.Is(err, service.ErrInvalidCartItem) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid cart item id", "id must be a positive number")
		return
	}
	if errors.Is(err, service.ErrCartItemNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "cart item not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not delete cart item", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "cart item deleted successfully", cart)
}

func (c *CartController) Clear(w http.ResponseWriter, r *http.Request) {
	cart, err := c.service.Clear(r.Context(), CurrentUserID(r))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not clear cart", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "cart cleared successfully", cart)
}

func readCartItemID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid cart item id", "id must be a positive number")
		return 0, false
	}

	return id, true
}
