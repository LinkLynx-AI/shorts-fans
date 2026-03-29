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

// Config は backend コマンドの実行時設定を保持します。
type Config struct {
	AppEnv      string
	APIAddr     string
	PostgresDSN string
	RedisAddr   string
	AWSRegion   string
	SQSQueueURL string
}

// Load はプロセス環境変数から設定を読み込み、既定値を適用します。
func Load() Config {
	return LoadFromEnv(os.Getenv)
}

// LoadFromEnv は任意の lookup 関数から Config を構築します。
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

// ValidateAPI は API サーバー設定が不足なく与えられているか検証します。
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

// ValidateWorker は worker 設定の整合性を検証します。
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
