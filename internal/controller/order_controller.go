package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"art-backend/internal/model"
	"art-backend/internal/response"
	"art-backend/internal/service"
)

type OrderController struct {
	service *service.OrderService
}

func NewOrderController(service *service.OrderService) *OrderController {
	return &OrderController{service: service}
}

func (c *OrderController) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	session, err := c.service.CreateCheckoutSession(r.Context(), CurrentUserID(r))
	if err != nil {
		switch err {
		case service.ErrInvalidStripeConfig:
			response.Error(w, http.StatusInternalServerError, "STRIPE_NOT_CONFIGURED", "stripe is not configured", "")
		case service.ErrEmptyCart:
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "cart is empty", "add at least one item before checkout")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create checkout session", err.Error())
		}
		return
	}

	response.JSON(w, http.StatusCreated, "checkout session created successfully", session)
}

func (c *OrderController) GetAll(w http.ResponseWriter, r *http.Request) {
	orders, err := c.service.GetAll(r.Context(), CurrentUserID(r))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch orders", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "orders fetched successfully", orders)
}

func (c *OrderController) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid order id", "id must be a positive number")
		return
	}

	order, err := c.service.GetByID(r.Context(), CurrentUserID(r), id)
	if err == service.ErrOrderNotFound {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "order not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch order", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "order fetched successfully", order)
}

func (c *OrderController) GetBySessionID(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("session_id")
	if sessionID == "" {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid stripe session id", "session_id is required")
		return
	}

	order, err := c.service.GetByStripeSessionID(r.Context(), CurrentUserID(r), sessionID)
	if err == service.ErrOrderNotFound {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "order not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch order", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "order fetched successfully", order)
}

func (c *OrderController) Track(w http.ResponseWriter, r *http.Request) {
	var request model.TrackOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	order, err := c.service.TrackOrder(r.Context(), request)
	if err == service.ErrOrderNotFound {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "order not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not track order", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "order fetched successfully", order)
}

func (c *OrderController) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	if err := c.service.HandleStripeWebhook(r.Context(), payload, r.Header.Get("Stripe-Signature")); err != nil {
		switch err {
		case service.ErrInvalidStripeConfig:
			response.Error(w, http.StatusInternalServerError, "STRIPE_NOT_CONFIGURED", "stripe is not configured", "")
		case service.ErrInvalidStripeSignature:
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid stripe signature", "")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not process stripe webhook", err.Error())
		}
		return
	}

	response.JSON(w, http.StatusOK, "stripe webhook processed successfully", nil)
}
