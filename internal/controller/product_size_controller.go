package controller

import (
	"net/http"

	"art-backend/internal/response"
	"art-backend/internal/service"
)

type ProductSizeController struct {
	service *service.ProductSizeService
}

func NewProductSizeController(service *service.ProductSizeService) *ProductSizeController {
	return &ProductSizeController{service: service}
}

func (c *ProductSizeController) GetAll(w http.ResponseWriter, r *http.Request) {
	items, err := c.service.GetAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch product sizes", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "product sizes fetched successfully", items)
}
