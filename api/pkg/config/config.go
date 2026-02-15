package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	CassandraHosts    []string
	CassandraKeyspace string
	JWTSecret         string
	AllowedOrigins    []string
}

func Load() *Config {
	// Load .env file if present (won't override existing env vars)
	_ = godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" || len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET environment variable must be set with at least 32 characters")
	}

	allowedOrigins := []string{}
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		for _, o := range strings.Split(origins, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	return &Config{
		Port:              getEnv("PORT", "8080"),
		CassandraHosts:    strings.Split(getEnv("CASSANDRA_HOSTS", "localhost"), ","),
		CassandraKeyspace: getEnv("CASSANDRA_KEYSPACE", "newcord"),
		JWTSecret:         jwtSecret,
		AllowedOrigins:    allowedOrigins,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
