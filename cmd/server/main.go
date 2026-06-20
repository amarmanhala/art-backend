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
	productImageStorage := service.NewAzureProductImageStorage(appConfig)
	productImageService := service.NewProductImageService(productRepository, productImageStorage)
	productController := controller.NewProductController(productService, productImageService)
	productSizeRepository := repository.NewProductSizeRepository(db)
	productSizeService := service.NewProductSizeService(productSizeRepository)
	productSizeController := controller.NewProductSizeController(productSizeService)
	frameRepository := repository.NewFrameRepository(db)
	frameService := service.NewFrameService(frameRepository)
	frameImageStorage := service.NewAzureFrameImageStorage(appConfig)
	frameImageService := service.NewFrameImageService(frameRepository, frameImageStorage)
	frameController := controller.NewFrameController(frameService, frameImageService)
	artStyleRepository := repository.NewArtStyleRepository(db)
	artStyleService := service.NewArtStyleService(artStyleRepository)
	artStyleController := controller.NewArtStyleController(artStyleService)
	carouselRepository := repository.NewCarouselRepository(db)
	carouselService := service.NewCarouselService(carouselRepository)
	carouselController := controller.NewCarouselController(carouselService)

	userRepository := repository.NewUserRepository(db)
	addressRepository := repository.NewAddressRepository(db)
	cartRepository := repository.NewCartRepository(db)
	likedArtRepository := repository.NewLikedArtRepository(db)
	orderRepository := repository.NewOrderRepository(db)
	contactRepository := repository.NewContactRepository(db)
	tokenStore := service.NewTokenStore()
	authService := service.NewAuthService(userRepository, tokenStore)
	userService := service.NewUserService(userRepository)
	addressService := service.NewAddressService(addressRepository)
	cartService := service.NewCartService(cartRepository)
	likedArtService := service.NewLikedArtService(likedArtRepository)
	orderService := service.NewOrderService(appConfig, cartRepository, orderRepository)
	contactService := service.NewContactService(contactRepository)
	authController := controller.NewAuthController(authService)
	profileController := controller.NewProfileController(userService)
	addressController := controller.NewAddressController(addressService)
	cartController := controller.NewCartController(cartService)
	likedArtController := controller.NewLikedArtController(likedArtService)
	orderController := controller.NewOrderController(orderService)
	contactController := controller.NewContactController(contactService)
	uploadController := controller.NewUploadController(appConfig)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", controller.Health)

	mux.HandleFunc("GET /api/products", productController.GetAll)
	mux.HandleFunc("POST /api/products", controller.RequireAdmin(authService, productController.Create))
	mux.HandleFunc("GET /api/products/featured", productController.GetFeatured)
	mux.HandleFunc("GET /api/products/search", productController.Search)
	mux.HandleFunc("GET /api/products/categories", productController.GetCategories)
	mux.HandleFunc("GET /api/products/styles", productController.GetStyles)
	mux.HandleFunc("GET /api/products/themes", productController.GetThemes)
	mux.HandleFunc("GET /api/products/sizes", productSizeController.GetAll)
	mux.HandleFunc("GET /api/products/{slug}", productController.GetBySlug)
	mux.HandleFunc("GET /api/product-sizes", productSizeController.GetAll)
	mux.HandleFunc("PUT /api/products/{identifier}", controller.RequireAdmin(authService, productController.UpdateByIdentifier))
	mux.HandleFunc("PATCH /api/products/{identifier}", controller.RequireAdmin(authService, productController.UpdateByIdentifier))
	mux.HandleFunc("DELETE /api/products/{identifier}", controller.RequireAdmin(authService, productController.DeleteByIdentifier))
	mux.HandleFunc("GET /api/frames", frameController.GetAll)
	mux.HandleFunc("GET /api/frames/{id}", frameController.GetByID)
	mux.HandleFunc("POST /api/frames", controller.RequireAdmin(authService, frameController.Create))
	mux.HandleFunc("PUT /api/frames/{id}", controller.RequireAdmin(authService, frameController.Update))
	mux.HandleFunc("PATCH /api/frames/{id}", controller.RequireAdmin(authService, frameController.Update))
	mux.HandleFunc("DELETE /api/frames/{id}", controller.RequireAdmin(authService, frameController.Delete))
	mux.HandleFunc("GET /api/styles", artStyleController.GetAll)
	mux.HandleFunc("GET /api/styles/{id}", artStyleController.GetByID)
	mux.HandleFunc("POST /api/styles", controller.RequireAdmin(authService, artStyleController.Create))
	mux.HandleFunc("PUT /api/styles/{id}", controller.RequireAdmin(authService, artStyleController.Update))
	mux.HandleFunc("PATCH /api/styles/{id}", controller.RequireAdmin(authService, artStyleController.Update))
	mux.HandleFunc("DELETE /api/styles/{id}", controller.RequireAdmin(authService, artStyleController.Delete))
	mux.HandleFunc("GET /api/art-styles", artStyleController.GetAll)
	mux.HandleFunc("GET /api/art-styles/{id}", artStyleController.GetByID)
	mux.HandleFunc("POST /api/art-styles", controller.RequireAdmin(authService, artStyleController.Create))
	mux.HandleFunc("PUT /api/art-styles/{id}", controller.RequireAdmin(authService, artStyleController.Update))
	mux.HandleFunc("PATCH /api/art-styles/{id}", controller.RequireAdmin(authService, artStyleController.Update))
	mux.HandleFunc("DELETE /api/art-styles/{id}", controller.RequireAdmin(authService, artStyleController.Delete))
	mux.HandleFunc("GET /api/carousel", carouselController.GetActive)
	mux.HandleFunc("GET /api/admin/carousel", controller.RequireAdmin(authService, carouselController.GetAll))
	mux.HandleFunc("POST /api/admin/carousel", controller.RequireAdmin(authService, carouselController.Create))
	mux.HandleFunc("PUT /api/admin/carousel/{id}", controller.RequireAdmin(authService, carouselController.Update))
	mux.HandleFunc("PATCH /api/admin/carousel/{id}/status", controller.RequireAdmin(authService, carouselController.SetActive))
	mux.HandleFunc("DELETE /api/admin/carousel/{id}", controller.RequireAdmin(authService, carouselController.Delete))
	mux.HandleFunc("PUT /api/carousel", controller.RequireAdmin(authService, carouselController.ReplaceAll))

	mux.HandleFunc("POST /api/contact", contactController.Create)
	mux.HandleFunc("GET /api/admin/contact-requests", controller.RequireAdmin(authService, contactController.GetAll))
	mux.HandleFunc("GET /api/admin/contact-requests/{id}", controller.RequireAdmin(authService, contactController.GetByID))
	mux.HandleFunc("DELETE /api/admin/contact-requests/{id}", controller.RequireAdmin(authService, contactController.Delete))
	mux.HandleFunc("POST /api/uploads/carousel/sas", controller.RequireAdmin(authService, uploadController.CreateCarouselUploadSAS))
	mux.HandleFunc("POST /api/admin/products/images/sas", controller.RequireAdmin(authService, uploadController.CreateProductImagesUploadSAS))
	mux.HandleFunc("POST /api/admin/frames/images/sas", controller.RequireAdmin(authService, uploadController.CreateFrameImagesUploadSAS))
	mux.HandleFunc("POST /api/admin/styles/image/sas", controller.RequireAdmin(authService, uploadController.CreateArtStyleUploadSAS))
	mux.HandleFunc("POST /api/admin/art-styles/image/sas", controller.RequireAdmin(authService, uploadController.CreateArtStyleUploadSAS))

	mux.HandleFunc("POST /api/v1/products", controller.RequireAdmin(authService, productController.Create))
	mux.HandleFunc("GET /api/v1/products/{id}", productController.GetByID)
	mux.HandleFunc("PUT /api/v1/products/{id}", controller.RequireAdmin(authService, productController.UpdateByID))
	mux.HandleFunc("PATCH /api/v1/products/{id}", controller.RequireAdmin(authService, productController.UpdateByID))
	mux.HandleFunc("DELETE /api/v1/products/{id}", controller.RequireAdmin(authService, productController.DeleteByID))
	mux.HandleFunc("POST /api/admin/products/{id}/images", controller.RequireAdmin(authService, productController.ReplaceImages))
	mux.HandleFunc("PUT /api/admin/products/{id}/images", controller.RequireAdmin(authService, productController.ReplaceImages))
	mux.HandleFunc("POST /api/admin/products/{id}/variants", controller.RequireAdmin(authService, productController.ReplaceVariants))
	mux.HandleFunc("PUT /api/admin/products/{id}/variants", controller.RequireAdmin(authService, productController.ReplaceVariants))
	mux.HandleFunc("POST /api/admin/frames/{id}/images", controller.RequireAdmin(authService, frameController.ReplaceImages))
	mux.HandleFunc("PUT /api/admin/frames/{id}/images", controller.RequireAdmin(authService, frameController.ReplaceImages))

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
	mux.HandleFunc("GET /api/liked-arts", controller.RequireAuth(authService, likedArtController.GetAll))
	mux.HandleFunc("POST /api/liked-arts", controller.RequireAuth(authService, likedArtController.Save))
	mux.HandleFunc("POST /api/checkout/stripe/session", controller.RequireAuth(authService, orderController.CreateCheckoutSession))
	mux.HandleFunc("GET /api/orders", controller.RequireAuth(authService, orderController.GetAll))
	mux.HandleFunc("GET /api/orders/by-session/{session_id}", controller.RequireAuth(authService, orderController.GetBySessionID))
	mux.HandleFunc("POST /api/orders/track", orderController.Track)
	mux.HandleFunc("GET /api/orders/{id}", controller.RequireAuth(authService, orderController.GetByID))
	mux.HandleFunc("POST /api/webhooks/stripe", orderController.StripeWebhook)

	server := &http.Server{
		Addr:    ":" + appConfig.AppPort,
		Handler: controller.LoggingMiddleware(controller.CORSMiddleware(appConfig.CORSAllowedOrigins, mux)),
	}

	log.Println("server started on port", appConfig.AppPort)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
