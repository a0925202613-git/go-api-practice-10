package config

import (
	"os"

	"github.com/joho/godotenv"
)

func Load() error {
	return godotenv.Load()
}

func Get(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func Port() string {
	return Get("PORT", "8085")
}

func DatabaseURL() string {
	return Get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/practice10?sslmode=disable")
}
