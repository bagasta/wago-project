package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort        string
	DatabaseURL    string
	JWTSecret      string
	WhatsappData   string
	AllowedOrigins []string
	LogLevel       string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	return &Config{
		AppPort:        getEnv("APP_PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/wago?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-secret"),
		WhatsappData:   getEnv("WHATSAPP_DATA_DIR", "whatsapp-sessions"),
		AllowedOrigins: parseCSV(getEnv("ALLOWED_ORIGINS", "*")),
		LogLevel:       strings.ToUpper(getEnv("LOG_LEVEL", "INFO")),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}
