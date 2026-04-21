package main

import (
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/config"
	"github.com/gin-gonic/gin"
)

func TestAuthCookieConfigForUsesNormalizedProduction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfg        config.Config
		wantSecure bool
	}{
		{
			name:       "production with mixed case and spaces",
			cfg:        config.Config{AppEnv: " Production ", AppEnvExplicitlySet: true},
			wantSecure: true,
		},
		{
			name:       "development",
			cfg:        config.Config{AppEnv: "development", AppEnvExplicitlySet: true},
			wantSecure: false,
		},
		{
			name:       "default development",
			cfg:        config.Config{},
			wantSecure: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := authCookieConfigFor(tt.cfg).Secure; got != tt.wantSecure {
				t.Fatalf("authCookieConfigFor() Secure got %t want %t", got, tt.wantSecure)
			}
		})
	}
}

func TestSetGinModeUsesNormalizedProduction(t *testing.T) {
	previousMode := gin.Mode()
	t.Cleanup(func() {
		gin.SetMode(previousMode)
	})

	setGinMode(config.Config{AppEnv: " Production ", AppEnvExplicitlySet: true})
	if got := gin.Mode(); got != gin.ReleaseMode {
		t.Fatalf("setGinMode() mode got %q want %q", got, gin.ReleaseMode)
	}

	setGinMode(config.Config{AppEnv: "development", AppEnvExplicitlySet: true})
	if got := gin.Mode(); got != gin.DebugMode {
		t.Fatalf("setGinMode() mode got %q want %q", got, gin.DebugMode)
	}
}
