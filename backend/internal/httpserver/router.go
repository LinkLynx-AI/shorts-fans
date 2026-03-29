package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const readinessTimeout = 2 * time.Second

// ReadinessChecker validates that a dependency can serve requests.
type ReadinessChecker interface {
	CheckReadiness(ctx context.Context) error
}

// Dependency names a readiness check dependency.
type Dependency struct {
	Name    string
	Checker ReadinessChecker
}

// Config configures the HTTP server runtime.
type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
}

// Server manages Gin startup and graceful shutdown.
type Server struct {
	config     Config
	httpServer *http.Server
	logger     *slog.Logger
}

// NewHandler builds the Gin router for the API server.
func NewHandler(dependencies []Dependency) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/readyz", func(c *gin.Context) {
		var failed []string
		for _, dependency := range dependencies {
			if dependency.Checker == nil {
				failed = append(failed, dependency.Name)
				continue
			}

			readinessCtx, cancel := context.WithTimeout(c.Request.Context(), readinessTimeout)
			err := dependency.Checker.CheckReadiness(readinessCtx)
			cancel()
			if err != nil {
				failed = append(failed, dependency.Name)
			}
		}

		if len(failed) > 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"failed": failed,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	return router
}

// New constructs a Server from runtime config and dependencies.
func New(cfg Config, logger *slog.Logger, dependencies []Dependency) *Server {
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 10 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}

	handler := NewHandler(dependencies)

	return &Server{
		config: cfg,
		httpServer: &http.Server{
			Addr:    cfg.Addr,
			Handler: handler,
		},
		logger: logger,
	}
}

// Run starts the HTTP server and shuts it down when ctx is canceled.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()

		s.logger.Info("shutting down api server")
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}

		if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	}
}
