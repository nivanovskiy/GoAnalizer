package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	AIModelURL  string
	Port        string
}

func New() *Config {
	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		AIModelURL:  os.Getenv("AI_MODEL_URL"),
		Port:        getEnvOrDefault("PORT", "8000"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDatabaseURL constructs database URL from individual components if DATABASE_URL is not set
func (c *Config) GetDatabaseURL() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}

	host := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	dbname := os.Getenv("PGDATABASE")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}

	return "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=disable"
}

func (c *Config) GetAIModelURL() string {
	if c.AIModelURL != "" {
		return c.AIModelURL
	}
	return "http://localhost:1234"
}
