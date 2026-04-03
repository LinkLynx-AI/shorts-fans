package config

import "testing"

func TestLoadReadsProcessEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("API_ADDR", ":9090")
	t.Setenv("POSTGRES_DSN", "postgres://example")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("AWS_REGION", "ap-northeast-1")
	t.Setenv("SQS_QUEUE_URL", "https://example.com/queue")

	cfg := Load()

	if cfg.AppEnv != "production" {
		t.Fatalf("Load() app env got %q want %q", cfg.AppEnv, "production")
	}
	if cfg.APIAddr != ":9090" {
		t.Fatalf("Load() api addr got %q want %q", cfg.APIAddr, ":9090")
	}
	if cfg.PostgresDSN != "postgres://example" {
		t.Fatalf("Load() postgres dsn got %q want %q", cfg.PostgresDSN, "postgres://example")
	}
	if cfg.RedisAddr != "redis:6379" {
		t.Fatalf("Load() redis addr got %q want %q", cfg.RedisAddr, "redis:6379")
	}
	if cfg.AWSRegion != "ap-northeast-1" {
		t.Fatalf("Load() aws region got %q want %q", cfg.AWSRegion, "ap-northeast-1")
	}
	if cfg.SQSQueueURL != "https://example.com/queue" {
		t.Fatalf("Load() sqs queue url got %q want %q", cfg.SQSQueueURL, "https://example.com/queue")
	}
}
