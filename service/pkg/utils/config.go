package utils

import (
	"hivem/pkg/api/auth"

	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port   string
	Host   string
	DBPath string
	Auth   auth.AuthConfig
}

func LoadConfig() Config {
	// load any .env file
	godotenv.Load()

	// read envs
	port := os.Getenv("HIVEM_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("HIVEM_SERVER_HOST")

	dbPath := os.Getenv("HIVEM_DB_PATH")
	if dbPath == "" {
		dbPath = "./database"
	}

	corsOrigins := os.Getenv("HIVEM_CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*"
	}

	ipWhitelist := os.Getenv("HIVEM_IP_WHITELIST")
	frontendKeys := os.Getenv("HIVEM_FRONTEND_KEYS")
	backendKeys := os.Getenv("HIVEM_BACKEND_KEYS")
	signingKeys := os.Getenv("HIVEM_SIGNING_KEYS")
	rateLimitMax := os.Getenv("HIVEM_RATE_LIMIT_MAX")
	max := 100
	if rateLimitMax != "" {
		if val, err := strconv.Atoi(rateLimitMax); err == nil {
			max = val
		}
	}

	rateLimitDuration := os.Getenv("HIVEM_RATE_LIMIT_DURATION")
	duration := 60 * time.Second
	if rateLimitDuration != "" {
		if val, err := time.ParseDuration(rateLimitDuration); err == nil {
			duration = val
		}
	}

	return Config{
		Port:   port,
		Host:   host,
		DBPath: dbPath,
		Auth: auth.AuthConfig{
			AllowedOrigins:    parseCommaSeparatedList(corsOrigins),
			IPWhitelist:       parseCommaSeparatedList(ipWhitelist),
			FrontendKeys:      parseKeyMap(frontendKeys),
			BackendKeys:       parseKeyMap(backendKeys),
			SigningKeys:       parseKeyMap(signingKeys),
			RateLimitMax:      max,
			RateLimitDuration: duration,
		},
	}
}

func parseCommaSeparatedList(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseKeyMap(value string) map[string]bool {
	result := make(map[string]bool)
	if value == "" {
		return result
	}
	parts := strings.Split(value, ",")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result[trimmed] = true
		}
	}
	return result
}
