package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddr   string
	PublicBase   string
	DatabaseURL  string
}

func Load() Config {
	_ = godotenv.Load()

	c := Config{
		ServerAddr:  getEnv("SERVER_ADDR", ":8080"),
		PublicBase:  getEnv("PUBLIC_BASE_URL", "http://localhost:8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/webhook_inbox?sslmode=disable"),
	}
	if c.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	return c
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
