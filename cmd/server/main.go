package main

import (
	"log"
	"net/http"

	"art-backend/internal/config"
	"art-backend/internal/controller"
	"art-backend/internal/repository"
	"art-backend/internal/service"
)

func main() {
	appConfig := config.Load()

	db, err := config.OpenDB(appConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	productRepository := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepository)
	productController := controller.NewProductController(productService)
	carouselRepository := repository.NewCarouselRepository(db)
	carouselService := service.NewCarouselService(carouselRepository)
	carouselController := controller.NewCarouselController(carouselService)

	userRepository := repository.NewUserRepository(db)
	addressRepository := repository.NewAddressRepository(db)
	cartRepository := repository.NewCartRepository(db)
	contactRepository := repository.NewContactRepository(db)
	tokenStore := service.NewTokenStore()
	authService := service.NewAuthService(userRepository, tokenStore)
	userService := service.NewUserService(userRepository)
	addressService := service.NewAddressService(addressRepository)
	cartService := service.NewCartService(cartRepository)
	contactService := service.NewContactService(contactRepository)
	authController := controller.NewAuthController(authService)
	profileController := controller.NewProfileController(userService)
	addressController := controller.NewAddressController(addressService)
	cartController := controller.NewCartController(cartService)
	contactController := controller.NewContactController(contactService)
	uploadController := controller.NewUploadController(appConfig)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", controller.Health)

	mux.HandleFunc("GET /api/products", productController.GetAll)
	mux.HandleFunc("POST /api/products", productController.Create)
	mux.HandleFunc("GET /api/products/featured", productController.GetFeatured)
	mux.HandleFunc("GET /api/products/search", productController.Search)
	mux.HandleFunc("GET /api/products/categories", productController.GetCategories)
	mux.HandleFunc("GET /api/products/styles", productController.GetStyles)
	mux.HandleFunc("GET /api/products/themes", productController.GetThemes)
	mux.HandleFunc("GET /api/products/{slug}", productController.GetBySlug)
	mux.HandleFunc("PUT /api/products/{identifier}", productController.UpdateByIdentifier)
	mux.HandleFunc("PATCH /api/products/{identifier}", productController.UpdateByIdentifier)
	mux.HandleFunc("DELETE /api/products/{identifier}", productController.DeleteByIdentifier)
	mux.HandleFunc("GET /api/carousel", carouselController.GetActive)
	mux.HandleFunc("GET /api/admin/carousel", carouselController.GetAll)
	mux.HandleFunc("POST /api/admin/carousel", carouselController.Create)
	mux.HandleFunc("PUT /api/admin/carousel/{id}", carouselController.Update)
	mux.HandleFunc("PATCH /api/admin/carousel/{id}/status", carouselController.SetActive)
	mux.HandleFunc("DELETE /api/admin/carousel/{id}", carouselController.Delete)
	mux.HandleFunc("PUT /api/carousel", carouselController.ReplaceAll)

	mux.HandleFunc("POST /api/contact", contactController.Create)
	mux.HandleFunc("POST /api/uploads/carousel/sas", uploadController.CreateCarouselUploadSAS)

	mux.HandleFunc("POST /api/v1/products", productController.Create)
	mux.HandleFunc("GET /api/v1/products/{id}", productController.GetByID)
	mux.HandleFunc("PUT /api/v1/products/{id}", productController.UpdateByID)
	mux.HandleFunc("PATCH /api/v1/products/{id}", productController.UpdateByID)
	mux.HandleFunc("DELETE /api/v1/products/{id}", productController.DeleteByID)

	mux.HandleFunc("POST /api/auth/register", authController.Register)
	mux.HandleFunc("POST /api/auth/login", authController.Login)
	mux.HandleFunc("POST /api/auth/logout", controller.RequireAuth(authService, authController.Logout))

	mux.HandleFunc("GET /api/profile", controller.RequireAuth(authService, profileController.Get))
	mux.HandleFunc("PUT /api/profile", controller.RequireAuth(authService, profileController.Update))

	mux.HandleFunc("GET /api/addresses", controller.RequireAuth(authService, addressController.GetAll))
	mux.HandleFunc("POST /api/addresses", controller.RequireAuth(authService, addressController.Create))
	mux.HandleFunc("PUT /api/addresses/{id}", controller.RequireAuth(authService, addressController.Update))
	mux.HandleFunc("DELETE /api/addresses/{id}", controller.RequireAuth(authService, addressController.Delete))

	mux.HandleFunc("GET /api/cart", controller.RequireAuth(authService, cartController.Get))
	mux.HandleFunc("DELETE /api/cart", controller.RequireAuth(authService, cartController.Clear))
	mux.HandleFunc("POST /api/cart/items", controller.RequireAuth(authService, cartController.AddItem))
	mux.HandleFunc("PUT /api/cart/items/{id}", controller.RequireAuth(authService, cartController.UpdateItem))
	mux.HandleFunc("DELETE /api/cart/items/{id}", controller.RequireAuth(authService, cartController.DeleteItem))

	server := &http.Server{
		Addr:    ":" + appConfig.AppPort,
		Handler: controller.LoggingMiddleware(controller.CORSMiddleware(appConfig.CORSAllowedOrigins, mux)),
	}

	log.Println("server started on port", appConfig.AppPort)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
