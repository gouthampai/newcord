package config

import (
	"os"
	"strings"
)

type Config struct {
	Port              string
	CassandraHosts    []string
	CassandraKeyspace string
	JWTSecret         string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		CassandraHosts:    strings.Split(getEnv("CASSANDRA_HOSTS", "localhost"), ","),
		CassandraKeyspace: getEnv("CASSANDRA_KEYSPACE", "newcord"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
