package sqs

import "testing"

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cfg       Config
		wantErr   bool
		wantReady bool
	}{
		{
			name:      "empty config is allowed",
			cfg:       Config{},
			wantErr:   false,
			wantReady: false,
		},
		{
			name: "complete config is enabled",
			cfg: Config{
				Region:   "ap-northeast-1",
				QueueURL: "https://example.com/queue",
			},
			wantErr:   false,
			wantReady: true,
		},
		{
			name: "queue url without region is rejected",
			cfg: Config{
				QueueURL: "https://example.com/queue",
			},
			wantErr:   true,
			wantReady: false,
		},
		{
			name: "region without queue url is rejected",
			cfg: Config{
				Region: "ap-northeast-1",
			},
			wantErr:   true,
			wantReady: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cfg.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
			if got := tt.cfg.Enabled(); got != tt.wantReady {
				t.Fatalf("Enabled() got %t want %t", got, tt.wantReady)
			}
		})
	}
}
