package config

import "testing"

func TestLoadReadsProcessEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("API_ADDR", ":9090")
	t.Setenv("POSTGRES_DSN", "postgres://example")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("AWS_REGION", "ap-northeast-1")
	t.Setenv("COGNITO_USER_POOL_ID", "ap-northeast-1_example")
	t.Setenv("COGNITO_USER_POOL_CLIENT_ID", "exampleclientid")
	t.Setenv("MEDIA_JOBS_QUEUE_URL", "https://example.com/media-queue")
	t.Setenv("MEDIA_RAW_BUCKET_NAME", "raw-bucket")
	t.Setenv("MEDIA_SHORT_PUBLIC_BUCKET_NAME", "short-bucket")
	t.Setenv("MEDIA_SHORT_PUBLIC_BASE_URL", "https://example.com/shorts")
	t.Setenv("MEDIA_MAIN_PRIVATE_BUCKET_NAME", "main-bucket")
	t.Setenv("MEDIACONVERT_SERVICE_ROLE_ARN", "arn:aws:iam::123456789012:role/media-role")
	t.Setenv("CREATOR_AVATAR_UPLOAD_BUCKET_NAME", "avatar-upload-bucket")
	t.Setenv("CREATOR_AVATAR_DELIVERY_BUCKET_NAME", "avatar-delivery-bucket")
	t.Setenv("CREATOR_AVATAR_BASE_URL", "https://example.com/avatar")

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
	if cfg.CognitoUserPoolID != "ap-northeast-1_example" {
		t.Fatalf("Load() cognito user pool id got %q want %q", cfg.CognitoUserPoolID, "ap-northeast-1_example")
	}
	if cfg.CognitoUserPoolClientID != "exampleclientid" {
		t.Fatalf("Load() cognito user pool client id got %q want %q", cfg.CognitoUserPoolClientID, "exampleclientid")
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
	if cfg.CreatorAvatarUploadBucketName != "avatar-upload-bucket" {
		t.Fatalf("Load() creator avatar upload bucket got %q want %q", cfg.CreatorAvatarUploadBucketName, "avatar-upload-bucket")
	}
	if cfg.CreatorAvatarDeliveryBucketName != "avatar-delivery-bucket" {
		t.Fatalf("Load() creator avatar delivery bucket got %q want %q", cfg.CreatorAvatarDeliveryBucketName, "avatar-delivery-bucket")
	}
	if cfg.CreatorAvatarBaseURL != "https://example.com/avatar" {
		t.Fatalf("Load() creator avatar base url got %q want %q", cfg.CreatorAvatarBaseURL, "https://example.com/avatar")
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

func TestValidateFanAuthRequiresCognitoRuntimeValues(t *testing.T) {
	cfg := Config{
		AWSRegion:               "ap-northeast-1",
		CognitoUserPoolID:       "ap-northeast-1_example",
		CognitoUserPoolClientID: "exampleclientid",
	}

	if err := cfg.ValidateFanAuth(); err != nil {
		t.Fatalf("ValidateFanAuth() error = %v, want nil", err)
	}
}

func TestValidateFanAuthReportsMissingValues(t *testing.T) {
	cfg := Config{
		AWSRegion: "ap-northeast-1",
	}

	err := cfg.ValidateFanAuth()
	if err == nil {
		t.Fatal("ValidateFanAuth() error = nil, want missing env error")
	}

	want := "missing required environment variables: COGNITO_USER_POOL_ID, COGNITO_USER_POOL_CLIENT_ID"
	if err.Error() != want {
		t.Fatalf("ValidateFanAuth() error got %q want %q", err.Error(), want)
	}
}
