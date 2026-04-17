package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	defaultAPIAddr             = ":8080"
	defaultAppEnv              = "development"
	legacySQSQueueURLEnv       = "SQS_QUEUE_URL"
	cognitoUserPoolIDEnv       = "COGNITO_USER_POOL_ID"
	cognitoUserPoolClientIDEnv = "COGNITO_USER_POOL_CLIENT_ID"
	mediaJobsQueueURLEnv       = "MEDIA_JOBS_QUEUE_URL"
	mediaRawBucketNameEnv      = "MEDIA_RAW_BUCKET_NAME"
	mediaShortBucketNameEnv    = "MEDIA_SHORT_PUBLIC_BUCKET_NAME"
	mediaShortBaseURLEnv       = "MEDIA_SHORT_PUBLIC_BASE_URL"
	mediaMainBucketNameEnv     = "MEDIA_MAIN_PRIVATE_BUCKET_NAME"
	mediaRoleARNEnv            = "MEDIACONVERT_SERVICE_ROLE_ARN"
	avatarUploadBucketEnv      = "CREATOR_AVATAR_UPLOAD_BUCKET_NAME"
	avatarDeliveryBucketEnv    = "CREATOR_AVATAR_DELIVERY_BUCKET_NAME"
	avatarBaseURLEnv           = "CREATOR_AVATAR_BASE_URL"
	reviewEvidenceBucketEnv    = "CREATOR_REVIEW_EVIDENCE_BUCKET_NAME"
)

// Config は backend コマンドの実行時設定を保持します。
type Config struct {
	AppEnv                          string
	APIAddr                         string
	PostgresDSN                     string
	RedisAddr                       string
	AWSRegion                       string
	CognitoUserPoolID               string
	CognitoUserPoolClientID         string
	MediaJobsQueueURL               string
	MediaRawBucketName              string
	MediaShortPublicBucketName      string
	MediaShortPublicBaseURL         string
	MediaMainPrivateBucketName      string
	MediaConvertServiceRoleARN      string
	CreatorAvatarUploadBucketName   string
	CreatorAvatarDeliveryBucketName string
	CreatorAvatarBaseURL            string
	CreatorReviewEvidenceBucketName string
}

// Load はプロセス環境変数から設定を読み込み、既定値を適用します。
func Load() Config {
	return LoadFromEnv(os.Getenv)
}

// LoadFromEnv は任意の lookup 関数から Config を構築します。
func LoadFromEnv(lookup func(string) string) Config {
	cfg := Config{
		AppEnv:                          trimmedLookup(lookup, "APP_ENV"),
		APIAddr:                         trimmedLookup(lookup, "API_ADDR"),
		PostgresDSN:                     trimmedLookup(lookup, "POSTGRES_DSN"),
		RedisAddr:                       trimmedLookup(lookup, "REDIS_ADDR"),
		AWSRegion:                       trimmedLookup(lookup, "AWS_REGION"),
		CognitoUserPoolID:               trimmedLookup(lookup, cognitoUserPoolIDEnv),
		CognitoUserPoolClientID:         trimmedLookup(lookup, cognitoUserPoolClientIDEnv),
		MediaJobsQueueURL:               firstNonEmpty(trimmedLookup(lookup, mediaJobsQueueURLEnv), trimmedLookup(lookup, legacySQSQueueURLEnv)),
		MediaRawBucketName:              trimmedLookup(lookup, mediaRawBucketNameEnv),
		MediaShortPublicBucketName:      trimmedLookup(lookup, mediaShortBucketNameEnv),
		MediaShortPublicBaseURL:         trimmedLookup(lookup, mediaShortBaseURLEnv),
		MediaMainPrivateBucketName:      trimmedLookup(lookup, mediaMainBucketNameEnv),
		MediaConvertServiceRoleARN:      trimmedLookup(lookup, mediaRoleARNEnv),
		CreatorAvatarUploadBucketName:   trimmedLookup(lookup, avatarUploadBucketEnv),
		CreatorAvatarDeliveryBucketName: trimmedLookup(lookup, avatarDeliveryBucketEnv),
		CreatorAvatarBaseURL:            trimmedLookup(lookup, avatarBaseURLEnv),
		CreatorReviewEvidenceBucketName: trimmedLookup(lookup, reviewEvidenceBucketEnv),
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
	if err := c.ValidateFanAuth(); err != nil {
		return err
	}

	if err := c.validateMediaSandbox(true); err != nil {
		return err
	}

	requiredAvatarFields := []struct {
		name  string
		value string
	}{
		{name: avatarUploadBucketEnv, value: c.CreatorAvatarUploadBucketName},
		{name: avatarDeliveryBucketEnv, value: c.CreatorAvatarDeliveryBucketName},
		{name: avatarBaseURLEnv, value: c.CreatorAvatarBaseURL},
		{name: reviewEvidenceBucketEnv, value: c.CreatorReviewEvidenceBucketName},
	}

	var missingAvatar []string
	for _, field := range requiredAvatarFields {
		if field.value == "" {
			missingAvatar = append(missingAvatar, field.name)
		}
	}
	if len(missingAvatar) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingAvatar, ", "))
	}

	return nil
}

// ValidateWorker は worker 設定の整合性を検証します。
func (c Config) ValidateWorker() error {
	if err := c.validateMediaSandbox(false); err != nil {
		return err
	}
	if c.MediaSandboxEnabled() && c.PostgresDSN == "" {
		return fmt.Errorf("missing required environment variables: POSTGRES_DSN")
	}

	return nil
}

// ValidateMediaSmoke は media smoke 用の設定が不足なく与えられているか検証します。
func (c Config) ValidateMediaSmoke() error {
	return c.validateMediaSandbox(true)
}

// ValidateFanAuth は fan auth runtime が必要とする Cognito 設定を検証します。
func (c Config) ValidateFanAuth() error {
	required := []struct {
		name  string
		value string
	}{
		{name: "AWS_REGION", value: c.AWSRegion},
		{name: cognitoUserPoolClientIDEnv, value: c.CognitoUserPoolClientID},
	}

	var missing []string
	for _, field := range required {
		if field.value == "" {
			missing = append(missing, field.name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// MediaSandboxEnabled は media sandbox 用の env が 1 つ以上投入されているか返します。
func (c Config) MediaSandboxEnabled() bool {
	return c.AWSRegion != "" ||
		c.MediaJobsQueueURL != "" ||
		c.MediaRawBucketName != "" ||
		c.MediaShortPublicBucketName != "" ||
		c.MediaShortPublicBaseURL != "" ||
		c.MediaMainPrivateBucketName != "" ||
		c.MediaConvertServiceRoleARN != ""
}

func (c Config) validateMediaSandbox(requireAll bool) error {
	if !requireAll && !c.MediaSandboxEnabled() {
		return nil
	}

	required := []struct {
		name  string
		value string
	}{
		{name: "AWS_REGION", value: c.AWSRegion},
		{name: mediaJobsQueueURLEnv, value: c.MediaJobsQueueURL},
		{name: mediaRawBucketNameEnv, value: c.MediaRawBucketName},
		{name: mediaShortBucketNameEnv, value: c.MediaShortPublicBucketName},
		{name: mediaShortBaseURLEnv, value: c.MediaShortPublicBaseURL},
		{name: mediaMainBucketNameEnv, value: c.MediaMainPrivateBucketName},
		{name: mediaRoleARNEnv, value: c.MediaConvertServiceRoleARN},
	}

	var missing []string
	for _, field := range required {
		if field.value == "" {
			missing = append(missing, field.name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

func trimmedLookup(lookup func(string) string, key string) string {
	return strings.TrimSpace(lookup(key))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
