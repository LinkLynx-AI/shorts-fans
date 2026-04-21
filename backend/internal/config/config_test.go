package config

import (
	"strings"
	"testing"
)

func TestLoadFromEnvDefaults(t *testing.T) {
	t.Parallel()

	cfg := LoadFromEnv(func(string) string {
		return ""
	})

	if cfg.AppEnv != "development" {
		t.Fatalf("LoadFromEnv() default app env got %q want %q", cfg.AppEnv, "development")
	}
	if cfg.APIAddr != ":8080" {
		t.Fatalf("LoadFromEnv() default api addr got %q want %q", cfg.APIAddr, ":8080")
	}
	if cfg.CCBillBaseURL != "https://api.ccbill.com" {
		t.Fatalf("LoadFromEnv() default ccbill base url got %q want %q", cfg.CCBillBaseURL, "https://api.ccbill.com")
	}
	if cfg.CCBillCurrencyCode != 392 {
		t.Fatalf("LoadFromEnv() default ccbill currency code got %d want %d", cfg.CCBillCurrencyCode, 392)
	}
	if cfg.AppEnvExplicitlySet {
		t.Fatal("LoadFromEnv() AppEnvExplicitlySet = true, want false")
	}
	if cfg.PaymentBypassEnabled() {
		t.Fatal("PaymentBypassEnabled() = true, want false")
	}
}

func TestLoadFromEnvTracksExplicitAppEnvForPaymentBypass(t *testing.T) {
	t.Parallel()

	cfg := LoadFromEnv(func(key string) string {
		if key == "APP_ENV" {
			return " Development "
		}

		return ""
	})

	if cfg.AppEnv != "development" {
		t.Fatalf("LoadFromEnv() explicit app env got %q want %q", cfg.AppEnv, "development")
	}
	if !cfg.AppEnvExplicitlySet {
		t.Fatal("LoadFromEnv() AppEnvExplicitlySet = false, want true")
	}
	if !cfg.PaymentBypassEnabled() {
		t.Fatal("PaymentBypassEnabled() = false, want true")
	}
}

func TestLoadFromEnvNormalizesProductionAppEnv(t *testing.T) {
	t.Parallel()

	cfg := LoadFromEnv(func(key string) string {
		if key == "APP_ENV" {
			return " Production "
		}

		return ""
	})

	if cfg.AppEnv != "production" {
		t.Fatalf("LoadFromEnv() app env got %q want production", cfg.AppEnv)
	}
	if !cfg.AppEnvExplicitlySet {
		t.Fatal("LoadFromEnv() AppEnvExplicitlySet = false, want true")
	}
	if !cfg.IsProduction() {
		t.Fatal("IsProduction() = false, want true")
	}
	if cfg.PaymentBypassEnabled() {
		t.Fatal("PaymentBypassEnabled() = true, want false")
	}
}

func TestAppEnvRuntimeGuardsUseNormalizedValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		cfg               Config
		wantProduction    bool
		wantPaymentBypass bool
	}{
		{
			name:              "manual production with mixed case and spaces",
			cfg:               Config{AppEnv: " Production ", AppEnvExplicitlySet: true},
			wantProduction:    true,
			wantPaymentBypass: false,
		},
		{
			name:              "explicit development with mixed case and spaces",
			cfg:               Config{AppEnv: " Development ", AppEnvExplicitlySet: true},
			wantProduction:    false,
			wantPaymentBypass: true,
		},
		{
			name:              "default development does not enable payment bypass",
			cfg:               Config{},
			wantProduction:    false,
			wantPaymentBypass: false,
		},
		{
			name:              "explicit staging is not production or development bypass",
			cfg:               Config{AppEnv: " Staging ", AppEnvExplicitlySet: true},
			wantProduction:    false,
			wantPaymentBypass: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.cfg.IsProduction(); got != tt.wantProduction {
				t.Fatalf("IsProduction() got %t want %t", got, tt.wantProduction)
			}
			if got := tt.cfg.PaymentBypassEnabled(); got != tt.wantPaymentBypass {
				t.Fatalf("PaymentBypassEnabled() got %t want %t", got, tt.wantPaymentBypass)
			}
		})
	}
}

func TestValidateAPI(t *testing.T) {
	t.Parallel()

	cfg := Config{
		AppEnv:                          "production",
		AppEnvExplicitlySet:             true,
		PostgresDSN:                     "postgres://example",
		RedisAddr:                       "localhost:6379",
		AWSRegion:                       "ap-northeast-1",
		CognitoUserPoolClientID:         "exampleclientid",
		CCBillBackendClientID:           "backend-client-id",
		CCBillBackendClientSecret:       "backend-client-secret",
		CCBillFrontendClientID:          "frontend-client-id",
		CCBillFrontendClientSecret:      "frontend-client-secret",
		CCBillClientAccountNumber:       900100,
		CCBillClientSubAccountNumber:    1,
		CCBillCurrencyCode:              392,
		CCBillInitialPeriodDays:         30,
		MediaJobsQueueURL:               "https://example.com/queue",
		MediaRawBucketName:              "raw-bucket",
		MediaShortPublicBucketName:      "short-bucket",
		MediaShortPublicBaseURL:         "https://example.com/shorts",
		MediaMainPrivateBucketName:      "main-bucket",
		MediaConvertServiceRoleARN:      "arn:aws:iam::123456789012:role/media-role",
		CreatorAvatarUploadBucketName:   "avatar-upload-bucket",
		CreatorAvatarDeliveryBucketName: "avatar-delivery-bucket",
		CreatorAvatarBaseURL:            "https://example.com/avatar",
		CreatorReviewEvidenceBucketName: "creator-review-evidence-bucket",
	}

	if err := cfg.ValidateAPI(); err != nil {
		t.Fatalf("ValidateAPI() unexpected error: %v", err)
	}
}

func TestValidateAPIRequiresDependencies(t *testing.T) {
	t.Parallel()

	err := (Config{AppEnv: "production", AppEnvExplicitlySet: true}).ValidateAPI()
	if err == nil {
		t.Fatal("ValidateAPI() error = nil, want error")
	}
}

func TestValidateAPIRequiresExplicitSupportedAppEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "missing app env",
			cfg: Config{
				AppEnv: "development",
			},
			wantErr: "APP_ENV must be explicitly set",
		},
		{
			name: "blank explicit app env",
			cfg: Config{
				AppEnv:              " ",
				AppEnvExplicitlySet: true,
			},
			wantErr: "APP_ENV must be explicitly set",
		},
		{
			name: "unsupported app env",
			cfg: Config{
				AppEnv:              " Staging ",
				AppEnvExplicitlySet: true,
			},
			wantErr: "unsupported APP_ENV \"staging\"",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cfg.ValidateAPI()
			if err == nil {
				t.Fatal("ValidateAPI() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ValidateAPI() error got %q want contains %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateAPIRequiresMediaSandboxConfig(t *testing.T) {
	t.Parallel()

	cfg := Config{
		AppEnv:                       "production",
		AppEnvExplicitlySet:          true,
		PostgresDSN:                  "postgres://example",
		RedisAddr:                    "localhost:6379",
		CCBillBackendClientID:        "backend-client-id",
		CCBillBackendClientSecret:    "backend-client-secret",
		CCBillFrontendClientID:       "frontend-client-id",
		CCBillFrontendClientSecret:   "frontend-client-secret",
		CCBillClientAccountNumber:    900100,
		CCBillClientSubAccountNumber: 1,
		CCBillInitialPeriodDays:      30,
		CognitoUserPoolClientID:      "exampleclientid",
	}

	if err := cfg.ValidateAPI(); err == nil {
		t.Fatal("ValidateAPI() error = nil, want error")
	}
}

func TestValidateAPIAllowsMissingPaymentConfigInDevelopment(t *testing.T) {
	t.Parallel()

	cfg := Config{
		AppEnv:                          "development",
		AppEnvExplicitlySet:             true,
		PostgresDSN:                     "postgres://example",
		RedisAddr:                       "localhost:6379",
		AWSRegion:                       "ap-northeast-1",
		CognitoUserPoolClientID:         "exampleclientid",
		MediaJobsQueueURL:               "https://example.com/queue",
		MediaRawBucketName:              "raw-bucket",
		MediaShortPublicBucketName:      "short-bucket",
		MediaShortPublicBaseURL:         "https://example.com/shorts",
		MediaMainPrivateBucketName:      "main-bucket",
		MediaConvertServiceRoleARN:      "arn:aws:iam::123456789012:role/media-role",
		CreatorAvatarUploadBucketName:   "avatar-upload-bucket",
		CreatorAvatarDeliveryBucketName: "avatar-delivery-bucket",
		CreatorAvatarBaseURL:            "https://example.com/avatar",
		CreatorReviewEvidenceBucketName: "creator-review-evidence-bucket",
	}

	if err := cfg.ValidateAPI(); err != nil {
		t.Fatalf("ValidateAPI() unexpected error in development bypass: %v", err)
	}
	if !cfg.PaymentBypassEnabled() {
		t.Fatal("PaymentBypassEnabled() = false, want true")
	}
}

func TestValidateAPIStillRequiresPaymentConfigOutsideDevelopment(t *testing.T) {
	t.Parallel()

	cfg := Config{
		AppEnv:                          "production",
		AppEnvExplicitlySet:             true,
		PostgresDSN:                     "postgres://example",
		RedisAddr:                       "localhost:6379",
		AWSRegion:                       "ap-northeast-1",
		CognitoUserPoolClientID:         "exampleclientid",
		MediaJobsQueueURL:               "https://example.com/queue",
		MediaRawBucketName:              "raw-bucket",
		MediaShortPublicBucketName:      "short-bucket",
		MediaShortPublicBaseURL:         "https://example.com/shorts",
		MediaMainPrivateBucketName:      "main-bucket",
		MediaConvertServiceRoleARN:      "arn:aws:iam::123456789012:role/media-role",
		CreatorAvatarUploadBucketName:   "avatar-upload-bucket",
		CreatorAvatarDeliveryBucketName: "avatar-delivery-bucket",
		CreatorAvatarBaseURL:            "https://example.com/avatar",
		CreatorReviewEvidenceBucketName: "creator-review-evidence-bucket",
	}

	if err := cfg.ValidateAPI(); err == nil {
		t.Fatal("ValidateAPI() error = nil, want payment config error")
	}
	if cfg.PaymentBypassEnabled() {
		t.Fatal("PaymentBypassEnabled() = true, want false")
	}
}

func TestValidateAPIRequiresCreatorAvatarConfig(t *testing.T) {
	t.Parallel()

	cfg := Config{
		AppEnv:                       "production",
		AppEnvExplicitlySet:          true,
		PostgresDSN:                  "postgres://example",
		RedisAddr:                    "localhost:6379",
		AWSRegion:                    "ap-northeast-1",
		CognitoUserPoolClientID:      "exampleclientid",
		CCBillBackendClientID:        "backend-client-id",
		CCBillBackendClientSecret:    "backend-client-secret",
		CCBillFrontendClientID:       "frontend-client-id",
		CCBillFrontendClientSecret:   "frontend-client-secret",
		CCBillClientAccountNumber:    900100,
		CCBillClientSubAccountNumber: 1,
		CCBillInitialPeriodDays:      30,
		MediaJobsQueueURL:            "https://example.com/queue",
		MediaRawBucketName:           "raw-bucket",
		MediaShortPublicBucketName:   "short-bucket",
		MediaShortPublicBaseURL:      "https://example.com/shorts",
		MediaMainPrivateBucketName:   "main-bucket",
		MediaConvertServiceRoleARN:   "arn:aws:iam::123456789012:role/media-role",
	}

	if err := cfg.ValidateAPI(); err == nil {
		t.Fatal("ValidateAPI() error = nil, want error")
	}
}

func TestValidatePayment(t *testing.T) {
	t.Parallel()

	cfg := Config{
		CCBillBackendClientID:        "backend-client-id",
		CCBillBackendClientSecret:    "backend-client-secret",
		CCBillFrontendClientID:       "frontend-client-id",
		CCBillFrontendClientSecret:   "frontend-client-secret",
		CCBillClientAccountNumber:    900100,
		CCBillClientSubAccountNumber: 1,
		CCBillInitialPeriodDays:      30,
	}

	if err := cfg.ValidatePayment(); err != nil {
		t.Fatalf("ValidatePayment() unexpected error: %v", err)
	}

	if err := (Config{}).ValidatePayment(); err == nil {
		t.Fatal("ValidatePayment() error = nil, want error")
	}
}

func TestValidateAPIRequiresExplicitDevelopmentOptInForPaymentBypass(t *testing.T) {
	t.Parallel()

	cfg := Config{
		AppEnv:                          "development",
		PostgresDSN:                     "postgres://example",
		RedisAddr:                       "localhost:6379",
		AWSRegion:                       "ap-northeast-1",
		CognitoUserPoolClientID:         "exampleclientid",
		MediaJobsQueueURL:               "https://example.com/queue",
		MediaRawBucketName:              "raw-bucket",
		MediaShortPublicBucketName:      "short-bucket",
		MediaShortPublicBaseURL:         "https://example.com/shorts",
		MediaMainPrivateBucketName:      "main-bucket",
		MediaConvertServiceRoleARN:      "arn:aws:iam::123456789012:role/media-role",
		CreatorAvatarUploadBucketName:   "avatar-upload-bucket",
		CreatorAvatarDeliveryBucketName: "avatar-delivery-bucket",
		CreatorAvatarBaseURL:            "https://example.com/avatar",
		CreatorReviewEvidenceBucketName: "creator-review-evidence-bucket",
	}

	if err := cfg.ValidateAPI(); err == nil {
		t.Fatal("ValidateAPI() error = nil, want APP_ENV error when APP_ENV was not explicit")
	}
	if cfg.PaymentBypassEnabled() {
		t.Fatal("PaymentBypassEnabled() = true, want false")
	}
}

func TestValidateWorker(t *testing.T) {
	t.Parallel()

	validMediaConfig := Config{
		PostgresDSN:                "postgres://example",
		AWSRegion:                  "ap-northeast-1",
		MediaJobsQueueURL:          "https://example.com/queue",
		MediaRawBucketName:         "raw-bucket",
		MediaShortPublicBucketName: "short-bucket",
		MediaShortPublicBaseURL:    "https://example.com/shorts",
		MediaMainPrivateBucketName: "main-bucket",
		MediaConvertServiceRoleARN: "arn:aws:iam::123456789012:role/media-role",
	}

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "empty is allowed",
			cfg:     Config{},
			wantErr: false,
		},
		{
			name:    "complete media sandbox config is allowed",
			cfg:     validMediaConfig,
			wantErr: false,
		},
		{
			name: "missing postgres dsn is rejected",
			cfg: Config{
				AWSRegion:                  validMediaConfig.AWSRegion,
				MediaJobsQueueURL:          validMediaConfig.MediaJobsQueueURL,
				MediaRawBucketName:         validMediaConfig.MediaRawBucketName,
				MediaShortPublicBucketName: validMediaConfig.MediaShortPublicBucketName,
				MediaShortPublicBaseURL:    validMediaConfig.MediaShortPublicBaseURL,
				MediaMainPrivateBucketName: validMediaConfig.MediaMainPrivateBucketName,
				MediaConvertServiceRoleARN: validMediaConfig.MediaConvertServiceRoleARN,
			},
			wantErr: true,
		},
		{
			name: "missing region is rejected",
			cfg: Config{
				PostgresDSN:       validMediaConfig.PostgresDSN,
				MediaJobsQueueURL: validMediaConfig.MediaJobsQueueURL,
			},
			wantErr: true,
		},
		{
			name: "missing queue url is rejected",
			cfg: Config{
				PostgresDSN: validMediaConfig.PostgresDSN,
				AWSRegion:   validMediaConfig.AWSRegion,
			},
			wantErr: true,
		},
		{
			name: "partial media sandbox config is rejected",
			cfg: Config{
				PostgresDSN:                validMediaConfig.PostgresDSN,
				AWSRegion:                  validMediaConfig.AWSRegion,
				MediaJobsQueueURL:          validMediaConfig.MediaJobsQueueURL,
				MediaShortPublicBucketName: validMediaConfig.MediaShortPublicBucketName,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cfg.ValidateWorker()
			if tt.wantErr && err == nil {
				t.Fatal("ValidateWorker() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateWorker() error = %v, want nil", err)
			}
		})
	}
}

func TestValidateMediaSmoke(t *testing.T) {
	t.Parallel()

	cfg := Config{
		AWSRegion:                  "ap-northeast-1",
		MediaJobsQueueURL:          "https://example.com/queue",
		MediaRawBucketName:         "raw-bucket",
		MediaShortPublicBucketName: "short-bucket",
		MediaShortPublicBaseURL:    "https://example.com/shorts",
		MediaMainPrivateBucketName: "main-bucket",
		MediaConvertServiceRoleARN: "arn:aws:iam::123456789012:role/media-role",
	}

	if err := cfg.ValidateMediaSmoke(); err != nil {
		t.Fatalf("ValidateMediaSmoke() unexpected error: %v", err)
	}

	cfg.MediaMainPrivateBucketName = ""
	if err := cfg.ValidateMediaSmoke(); err == nil {
		t.Fatal("ValidateMediaSmoke() error = nil, want error")
	}
}
