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

type ProductController struct {
	service *service.ProductService
}

func NewProductController(service *service.ProductService) *ProductController {
	return &ProductController{service: service}
}

func (c *ProductController) GetAll(w http.ResponseWriter, r *http.Request) {
	page, ok := readIntQuery(r, "page", 0)
	if !ok {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid page", "page must be a number")
		return
	}

	size, ok := readIntQuery(r, "size", 10)
	if !ok {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid size", "size must be a number")
		return
	}

	filter, ok := readProductFilter(r)
	if !ok {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid filters", "min_price and max_price must be valid numbers")
		return
	}

	products, err := c.service.GetAll(r.Context(), filter, page, size)
	if errors.Is(err, service.ErrInvalidPage) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid pagination", "page must be 0 or greater, size must be between 1 and 100")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch products", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "products fetched successfully", products)
}

func (c *ProductController) GetBySlug(w http.ResponseWriter, r *http.Request) {
	product, err := c.service.GetBySlug(r.Context(), r.PathValue("slug"))
	if errors.Is(err, service.ErrProductNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "product not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch product", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "product fetched successfully", product)
}

func (c *ProductController) GetFeatured(w http.ResponseWriter, r *http.Request) {
	products, err := c.service.GetFeatured(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch featured products", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "featured products fetched successfully", products)
}

func (c *ProductController) Search(w http.ResponseWriter, r *http.Request) {
	page, ok := readIntQuery(r, "page", 0)
	if !ok {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid page", "page must be a number")
		return
	}

	size, ok := readIntQuery(r, "size", 10)
	if !ok {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid size", "size must be a number")
		return
	}

	products, err := c.service.Search(r.Context(), r.URL.Query().Get("q"), page, size)
	if errors.Is(err, service.ErrInvalidSearch) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid search", "q is required")
		return
	}
	if errors.Is(err, service.ErrInvalidPage) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid pagination", "page must be 0 or greater, size must be between 1 and 100")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not search products", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "products searched successfully", products)
}

func (c *ProductController) GetCategories(w http.ResponseWriter, r *http.Request) {
	values, err := c.service.GetCategories(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch categories", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "categories fetched successfully", values)
}

func (c *ProductController) GetStyles(w http.ResponseWriter, r *http.Request) {
	values, err := c.service.GetStyles(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch styles", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "styles fetched successfully", values)
}

func (c *ProductController) GetThemes(w http.ResponseWriter, r *http.Request) {
	values, err := c.service.GetThemes(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch themes", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "themes fetched successfully", values)
}

func readIntQuery(r *http.Request, name string, defaultValue int) (int, bool) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue, true
	}

	number, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}

	return number, true
}

func readFloatQuery(r *http.Request, name string) (*float64, bool) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil, true
	}

	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, false
	}

	return &number, true
}

func readProductFilter(r *http.Request) (model.ProductFilter, bool) {
	minPrice, ok := readFloatQuery(r, "min_price")
	if !ok {
		return model.ProductFilter{}, false
	}

	maxPrice, ok := readFloatQuery(r, "max_price")
	if !ok {
		return model.ProductFilter{}, false
	}

	query := r.URL.Query()
	return model.ProductFilter{
		Category:    query.Get("category"),
		Style:       query.Get("style"),
		Theme:       query.Get("theme"),
		Orientation: query.Get("orientation"),
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
	}, true
}

func (c *ProductController) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid product id", "id must be a positive number")
		return
	}

	product, err := c.service.GetByID(r.Context(), id)
	if errors.Is(err, service.ErrProductNotFound) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "product not found", "")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch product", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "product fetched successfully", product)
}

func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) {
	var request model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", err.Error())
		return
	}

	product, err := c.service.Create(r.Context(), request)
	if errors.Is(err, service.ErrInvalidProduct) {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed", "title, slug, category, and price greater than 0 are required")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not create product", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "product created successfully", product)
}
