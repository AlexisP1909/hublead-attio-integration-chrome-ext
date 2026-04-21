package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	AttioAPIToken           string
	AttioBaseURL            string
	DatabaseURL             string
	Port                    string
	AllowedExtensionOrigins []string
}

func Load() (Config, error) {
	cfg := Config{
		AttioAPIToken:           strings.TrimSpace(os.Getenv("ATTIO_API_TOKEN")),
		AttioBaseURL:            strings.TrimRight(firstNonEmpty(os.Getenv("ATTIO_BASE_URL"), "https://api.attio.com"), "/"),
		DatabaseURL:             strings.TrimSpace(os.Getenv("DATABASE_URL")),
		Port:                    firstNonEmpty(os.Getenv("PORT"), "8080"),
		AllowedExtensionOrigins: splitCSV(os.Getenv("ALLOWED_EXTENSION_ORIGINS")),
	}

	if cfg.AttioAPIToken == "" {
		return Config{}, errors.New("ATTIO_API_TOKEN is required")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}
