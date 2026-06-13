package config

import "os"

type Config struct {
	AppPort                 string
	CORSAllowedOrigins      string
	DBHost                  string
	DBPort                  string
	DBUser                  string
	DBPass                  string
	DBName                  string
	DBSSLMode               string
	AzureStorageAccountName string
	AzureStorageAccountKey  string
	AzureStorageContainer   string
	AzureStoragePublicURL   string
}

func Load() Config {
	return Config{
		AppPort:                 getEnv("APP_PORT", "8080"),
		CORSAllowedOrigins:      getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173,http://localhost:5174,http://localhost:5175,http://127.0.0.1:3000,http://127.0.0.1:5173,http://127.0.0.1:5174,http://127.0.0.1:5175"),
		DBHost:                  getEnv("DB_HOST", "localhost"),
		DBPort:                  getEnv("DB_PORT", "5432"),
		DBUser:                  getEnv("DB_USER", "postgres"),
		DBPass:                  getEnv("DB_PASSWORD", "postgres"),
		DBName:                  getEnv("DB_NAME", "art_backend"),
		DBSSLMode:               getEnv("DB_SSLMODE", "disable"),
		AzureStorageAccountName: getEnv("AZURE_STORAGE_ACCOUNT_NAME", ""),
		AzureStorageAccountKey:  getEnv("AZURE_STORAGE_ACCOUNT_KEY", ""),
		AzureStorageContainer:   getEnv("AZURE_STORAGE_CONTAINER", ""),
		AzureStoragePublicURL:   getEnv("AZURE_STORAGE_PUBLIC_URL", ""),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
