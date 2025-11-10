package config

import (
	"fmt"
	"log"
	"os"
)

var (
	// FrontendURL is the base URL for the frontend application
	FrontendURL string

	// PolarAccessToken is the Polar API access token
	PolarAccessToken string

	// PolarWebhookSecret is the secret for verifying Polar webhook signatures
	PolarWebhookSecret string

	// PolarEnvironment determines which Polar server to use (sandbox or production)
	PolarEnvironment string

	// AppEnv is the application environment (development, production, etc.)
	AppEnv string
)

// Init loads and validates configuration from environment variables
func Init() error {
	// Load FrontendURL with default for development
	FrontendURL = getEnv("FRONTEND_URL", "http://localhost:3000")
	if FrontendURL == "" {
		log.Printf("Warning: FRONTEND_URL not set, using default: http://localhost:3000")
		FrontendURL = "http://localhost:3000"
	}

	// Load Polar configuration
	PolarAccessToken = os.Getenv("POLAR_ACCESS_TOKEN")
	if PolarAccessToken == "" {
		log.Printf("Warning: POLAR_ACCESS_TOKEN not set")
	}

	PolarWebhookSecret = os.Getenv("POLAR_WEBHOOK_SECRET")
	if PolarWebhookSecret == "" {
		log.Printf("Warning: POLAR_WEBHOOK_SECRET not set")
	}

	PolarEnvironment = getEnv("POLAR_ENVIRONMENT", "sandbox")
	AppEnv = getEnv("APP_ENV", "development")

	return nil
}

// GetPolarServer returns the Polar server name based on environment configuration
func GetPolarServer() string {
	if PolarEnvironment == "production" || AppEnv == "production" {
		return "production"
	}
	return "sandbox"
}

// ValidateRequired checks that required configuration values are set
// Returns an error if any required values are missing
func ValidateRequired() error {
	required := map[string]string{
		"POLAR_ACCESS_TOKEN": PolarAccessToken,
	}

	for name, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", name)
		}
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}




