package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	defaultAPIAddr = ":8080"
	defaultAppEnv  = "development"
)

// Config contains runtime configuration for backend commands.
type Config struct {
	AppEnv      string
	APIAddr     string
	PostgresDSN string
	RedisAddr   string
	AWSRegion   string
	SQSQueueURL string
}

// Load reads configuration from process environment and applies defaults.
func Load() Config {
	return LoadFromEnv(os.Getenv)
}

// LoadFromEnv builds Config from an arbitrary lookup function.
func LoadFromEnv(lookup func(string) string) Config {
	cfg := Config{
		AppEnv:      strings.TrimSpace(lookup("APP_ENV")),
		APIAddr:     strings.TrimSpace(lookup("API_ADDR")),
		PostgresDSN: strings.TrimSpace(lookup("POSTGRES_DSN")),
		RedisAddr:   strings.TrimSpace(lookup("REDIS_ADDR")),
		AWSRegion:   strings.TrimSpace(lookup("AWS_REGION")),
		SQSQueueURL: strings.TrimSpace(lookup("SQS_QUEUE_URL")),
	}

	if cfg.AppEnv == "" {
		cfg.AppEnv = defaultAppEnv
	}
	if cfg.APIAddr == "" {
		cfg.APIAddr = defaultAPIAddr
	}

	return cfg
}

// ValidateAPI verifies that API server configuration is complete.
func (c Config) ValidateAPI() error {
	var missing []string
	if c.PostgresDSN == "" {
		missing = append(missing, "POSTGRES_DSN")
	}
	if c.RedisAddr == "" {
		missing = append(missing, "REDIS_ADDR")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// ValidateWorker verifies that worker configuration is internally consistent.
func (c Config) ValidateWorker() error {
	switch {
	case c.AWSRegion == "" && c.SQSQueueURL == "":
		return nil
	case c.AWSRegion == "":
		return fmt.Errorf("AWS_REGION is required when SQS_QUEUE_URL is set")
	case c.SQSQueueURL == "":
		return fmt.Errorf("SQS_QUEUE_URL is required when AWS_REGION is set")
	default:
		return nil
	}
}
