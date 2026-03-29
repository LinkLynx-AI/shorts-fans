package config

import "testing"

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
}

func TestValidateAPI(t *testing.T) {
	t.Parallel()

	cfg := Config{
		PostgresDSN: "postgres://example",
		RedisAddr:   "localhost:6379",
	}

	if err := cfg.ValidateAPI(); err != nil {
		t.Fatalf("ValidateAPI() unexpected error: %v", err)
	}
}

func TestValidateAPIRequiresDependencies(t *testing.T) {
	t.Parallel()

	err := (Config{}).ValidateAPI()
	if err == nil {
		t.Fatal("ValidateAPI() error = nil, want error")
	}
}

func TestValidateWorker(t *testing.T) {
	t.Parallel()

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
			name: "complete sqs config is allowed",
			cfg: Config{
				AWSRegion:   "ap-northeast-1",
				SQSQueueURL: "https://example.com/queue",
			},
			wantErr: false,
		},
		{
			name: "missing region is rejected",
			cfg: Config{
				SQSQueueURL: "https://example.com/queue",
			},
			wantErr: true,
		},
		{
			name: "missing queue url is rejected",
			cfg: Config{
				AWSRegion: "ap-northeast-1",
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
