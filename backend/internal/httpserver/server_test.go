package httpserver

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestNewAppliesDefaults(t *testing.T) {
	t.Parallel()

	server := New(Config{Addr: "127.0.0.1:0"}, nil, nil, nil)

	if server.config.ShutdownTimeout != 10*time.Second {
		t.Fatalf("New() shutdown timeout got %s want %s", server.config.ShutdownTimeout, 10*time.Second)
	}
	if server.logger == nil {
		t.Fatal("New() logger = nil, want non-nil")
	}
	if server.httpServer == nil {
		t.Fatal("New() httpServer = nil, want non-nil")
	}
}

func TestRunShutsDownOnContextCancel(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := New(Config{
		Addr:            "127.0.0.1:0",
		ShutdownTimeout: time.Second,
	}, logger, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run() timed out waiting for graceful shutdown")
	}
}

func TestRunReturnsListenError(t *testing.T) {
	t.Parallel()

	server := &Server{
		config: Config{ShutdownTimeout: time.Second},
		httpServer: &http.Server{
			Addr: "bad address",
		},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	if err := server.Run(context.Background()); err == nil {
		t.Fatal("Run() error = nil, want listen error")
	}
}
