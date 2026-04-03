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

// ReadinessChecker は依存先がリクエストを処理可能か検証します。
type ReadinessChecker interface {
	CheckReadiness(ctx context.Context) error
}

// Dependency は readiness check の対象依存先を表します。
type Dependency struct {
	Name    string
	Checker ReadinessChecker
}

// Config は HTTP サーバーの実行設定を表します。
type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
}

// FanServices は fan read surface 用の domain service を束ねます。
type FanServices struct {
	Feed FanFeedService
}

// Server は Gin の起動と graceful shutdown を管理します。
type Server struct {
	config     Config
	httpServer *http.Server
	logger     *slog.Logger
}

// NewHandler は API サーバー用の Gin router を構築します。
func NewHandler(dependencies []Dependency, fanServices FanServices) *gin.Engine {
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
	registerFanRoutes(router, fanServices)

	return router
}

// New は実行設定と依存先から Server を構築します。
func New(cfg Config, logger *slog.Logger, dependencies []Dependency, fanServices FanServices) *Server {
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 10 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}

	handler := NewHandler(dependencies, fanServices)

	return &Server{
		config: cfg,
		httpServer: &http.Server{
			Addr:    cfg.Addr,
			Handler: handler,
		},
		logger: logger,
	}
}

// Run は HTTP サーバーを起動し、ctx が終了したら停止します。
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
