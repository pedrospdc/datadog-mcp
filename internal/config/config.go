package config

import (
	"errors"
	"os"
)

// Config holds the Datadog API configuration.
type Config struct {
	APIKey string
	AppKey string
	Site   string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	apiKey := os.Getenv("DD_API_KEY")
	if apiKey == "" {
		return nil, errors.New("DD_API_KEY environment variable is required")
	}

	appKey := os.Getenv("DD_APP_KEY")
	if appKey == "" {
		return nil, errors.New("DD_APP_KEY environment variable is required")
	}

	site := os.Getenv("DD_SITE")
	if site == "" {
		site = "datadoghq.com"
	}

	return &Config{
		APIKey: apiKey,
		AppKey: appKey,
		Site:   site,
	}, nil
}
