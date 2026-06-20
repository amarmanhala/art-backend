package config

import (
	"os"
	"strconv"
)

const DefaultProductThumbnailSize = 400
const DefaultFrameThumbnailSize = 400
const DefaultArtStylesContainer = "japneese-art-styles"

type Config struct {
	AppPort                        string
	CORSAllowedOrigins             string
	DBHost                         string
	DBPort                         string
	DBUser                         string
	DBPass                         string
	DBName                         string
	DBSSLMode                      string
	AzureStorageAccountName        string
	AzureStorageAccountKey         string
	AzureStorageContainer          string
	AzureProductImagesContainer    string
	AzureStoragePublicURL          string
	AzureProductImagesPublicURL    string
	ProductThumbnailSize           int
	AzureFrameImagesContainer      string
	AzureFrameImagesPublicURL      string
	FrameThumbnailSize             int
	AzureArtStylesContainer        string
	AzureArtStylesPublicURL        string
	StripeSecretKey                string
	StripeWebhookSecret            string
	StripeSuccessURL               string
	StripeCancelURL                string
	StripeAllowedShippingCountries string
	StripeProductTaxCode           string
	SMTPHost                       string
	SMTPPort                       string
	SMTPUsername                   string
	SMTPPassword                   string
	SMTPFromEmail                  string
	SMTPFromName                   string
	FrontendOrderTrackingURL       string
}

func Load() Config {
	return Config{
		AppPort:                        getEnv("APP_PORT", "8080"),
		CORSAllowedOrigins:             getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173,http://localhost:5174,http://localhost:5175,http://127.0.0.1:3000,http://127.0.0.1:5173,http://127.0.0.1:5174,http://127.0.0.1:5175"),
		DBHost:                         getEnv("DB_HOST", "localhost"),
		DBPort:                         getEnv("DB_PORT", "5432"),
		DBUser:                         getEnv("DB_USER", "postgres"),
		DBPass:                         getEnv("DB_PASSWORD", "postgres"),
		DBName:                         getEnv("DB_NAME", "art_backend"),
		DBSSLMode:                      getEnv("DB_SSLMODE", "disable"),
		AzureStorageAccountName:        getEnv("AZURE_STORAGE_ACCOUNT_NAME", ""),
		AzureStorageAccountKey:         getEnv("AZURE_STORAGE_ACCOUNT_KEY", ""),
		AzureStorageContainer:          getEnv("AZURE_STORAGE_CONTAINER", ""),
		AzureProductImagesContainer:    getEnv("AZURE_PRODUCT_IMAGES_CONTAINER", "product-images"),
		AzureStoragePublicURL:          getEnv("AZURE_STORAGE_PUBLIC_URL", ""),
		AzureProductImagesPublicURL:    getEnv("AZURE_PRODUCT_IMAGES_PUBLIC_URL", ""),
		ProductThumbnailSize:           getIntEnv("PRODUCT_THUMBNAIL_SIZE", DefaultProductThumbnailSize),
		AzureFrameImagesContainer:      getEnv("AZURE_FRAME_IMAGES_CONTAINER", "frame-images"),
		AzureFrameImagesPublicURL:      getEnv("AZURE_FRAME_IMAGES_PUBLIC_URL", ""),
		FrameThumbnailSize:             getIntEnv("FRAME_THUMBNAIL_SIZE", DefaultFrameThumbnailSize),
		AzureArtStylesContainer:        getEnv("AZURE_ART_STYLES_CONTAINER", DefaultArtStylesContainer),
		AzureArtStylesPublicURL:        getEnv("AZURE_ART_STYLES_PUBLIC_URL", ""),
		StripeSecretKey:                getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:            getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeSuccessURL:               getEnv("STRIPE_SUCCESS_URL", "http://localhost:3000/checkout/success?session_id={CHECKOUT_SESSION_ID}"),
		StripeCancelURL:                getEnv("STRIPE_CANCEL_URL", "http://localhost:3000/cart"),
		StripeAllowedShippingCountries: getEnv("STRIPE_ALLOWED_SHIPPING_COUNTRIES", "US,CA"),
		StripeProductTaxCode:           getEnv("STRIPE_PRODUCT_TAX_CODE", "txcd_99999999"),
		SMTPHost:                       getEnv("SMTP_HOST", ""),
		SMTPPort:                       getEnv("SMTP_PORT", "587"),
		SMTPUsername:                   getEnv("SMTP_USERNAME", ""),
		SMTPPassword:                   getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:                  getEnv("SMTP_FROM_EMAIL", ""),
		SMTPFromName:                   getEnv("SMTP_FROM_NAME", "Art Backend"),
		FrontendOrderTrackingURL:       getEnv("FRONTEND_ORDER_TRACKING_URL", "http://localhost:5174/orders/track"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getIntEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		return fallback
	}

	return number
}
