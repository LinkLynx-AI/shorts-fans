package config

import "testing"

func TestLoadReadsProcessEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("API_ADDR", ":9090")
	t.Setenv("POSTGRES_DSN", "postgres://example")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("AWS_REGION", "ap-northeast-1")
	t.Setenv("MEDIA_JOBS_QUEUE_URL", "https://example.com/media-queue")
	t.Setenv("MEDIA_RAW_BUCKET_NAME", "raw-bucket")
	t.Setenv("MEDIA_SHORT_PUBLIC_BUCKET_NAME", "short-bucket")
	t.Setenv("MEDIA_SHORT_PUBLIC_BASE_URL", "https://example.com/shorts")
	t.Setenv("MEDIA_MAIN_PRIVATE_BUCKET_NAME", "main-bucket")
	t.Setenv("MEDIACONVERT_SERVICE_ROLE_ARN", "arn:aws:iam::123456789012:role/media-role")

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
	if cfg.MediaJobsQueueURL != "https://example.com/media-queue" {
		t.Fatalf("Load() media jobs queue url got %q want %q", cfg.MediaJobsQueueURL, "https://example.com/media-queue")
	}
	if cfg.MediaRawBucketName != "raw-bucket" {
		t.Fatalf("Load() media raw bucket got %q want %q", cfg.MediaRawBucketName, "raw-bucket")
	}
	if cfg.MediaShortPublicBucketName != "short-bucket" {
		t.Fatalf("Load() media short public bucket got %q want %q", cfg.MediaShortPublicBucketName, "short-bucket")
	}
	if cfg.MediaShortPublicBaseURL != "https://example.com/shorts" {
		t.Fatalf("Load() media short public base url got %q want %q", cfg.MediaShortPublicBaseURL, "https://example.com/shorts")
	}
	if cfg.MediaMainPrivateBucketName != "main-bucket" {
		t.Fatalf("Load() media main private bucket got %q want %q", cfg.MediaMainPrivateBucketName, "main-bucket")
	}
	if cfg.MediaConvertServiceRoleARN != "arn:aws:iam::123456789012:role/media-role" {
		t.Fatalf("Load() mediaconvert role arn got %q want %q", cfg.MediaConvertServiceRoleARN, "arn:aws:iam::123456789012:role/media-role")
	}
}

func TestLoadPrefersMediaJobsQueueURLOverLegacyAlias(t *testing.T) {
	t.Setenv("MEDIA_JOBS_QUEUE_URL", "https://example.com/media-queue")
	t.Setenv("SQS_QUEUE_URL", "https://example.com/legacy-queue")

	cfg := Load()
	if cfg.MediaJobsQueueURL != "https://example.com/media-queue" {
		t.Fatalf("Load() media jobs queue url got %q want %q", cfg.MediaJobsQueueURL, "https://example.com/media-queue")
	}
}

func TestLoadFallsBackToLegacySQSQueueURL(t *testing.T) {
	t.Setenv("SQS_QUEUE_URL", "https://example.com/legacy-queue")

	cfg := Load()
	if cfg.MediaJobsQueueURL != "https://example.com/legacy-queue" {
		t.Fatalf("Load() media jobs queue url got %q want %q", cfg.MediaJobsQueueURL, "https://example.com/legacy-queue")
	}
}
