package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatoravatar"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatorupload"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/fanprofile"
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

// CreatorSearchReader は creator search 用の read 操作を表します。
type CreatorSearchReader interface {
	ListRecentPublicProfiles(ctx context.Context, cursor *creator.PublicProfileCursor, limit int) ([]creator.Profile, *creator.PublicProfileCursor, error)
	SearchPublicProfiles(ctx context.Context, query string, cursor *creator.PublicProfileCursor, limit int) ([]creator.Profile, *creator.PublicProfileCursor, error)
}

// CreatorProfileReader は creator profile header 用の read 操作を表します。
type CreatorProfileReader interface {
	GetPublicProfileHeader(ctx context.Context, creatorID string, viewerUserID *uuid.UUID) (creator.PublicProfileHeader, error)
}

// CreatorProfileShortsReader は creator profile short grid 用の read 操作を表します。
type CreatorProfileShortsReader interface {
	ListPublicProfileShorts(ctx context.Context, creatorID string, cursor *creator.PublicProfileShortCursor, limit int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error)
}

// CreatorFollowWriter は creator follow mutation を表します。
type CreatorFollowWriter interface {
	FollowPublicCreator(ctx context.Context, viewerUserID uuid.UUID, creatorID string) (creator.FollowMutationResult, error)
	UnfollowPublicCreator(ctx context.Context, viewerUserID uuid.UUID, creatorID string) (creator.FollowMutationResult, error)
}

// ViewerCreatorRegistrationWriter は creator registration mutation を表します。
type ViewerCreatorRegistrationWriter interface {
	RegisterApprovedCreator(ctx context.Context, input creator.SelfServeRegistrationInput) (creator.SelfServeRegistrationResult, error)
}

// ViewerCreatorAvatarUploadHandler は creator registration avatar upload を表します。
type ViewerCreatorAvatarUploadHandler interface {
	CompleteUpload(ctx context.Context, input creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error)
	ConsumeCompletedUpload(ctx context.Context, viewerUserID uuid.UUID, avatarUploadToken string) error
	CreateUpload(ctx context.Context, input creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error)
	ResolveCompletedUpload(ctx context.Context, viewerUserID uuid.UUID, avatarUploadToken string) (creatoravatar.CompletedUpload, error)
}

// ViewerActiveModeSwitcher は viewer active mode 切替を表します。
type ViewerActiveModeSwitcher interface {
	SwitchActiveMode(ctx context.Context, rawSessionToken string, activeMode auth.ActiveMode) error
}

// CreatorUploadHandler は creator-private upload initiation / completion を表します。
type CreatorUploadHandler interface {
	CompletePackage(ctx context.Context, input creatorupload.CompletePackageInput) (creatorupload.CompletePackageResult, error)
	CreatePackage(ctx context.Context, input creatorupload.CreatePackageInput) (creatorupload.CreatePackageResult, error)
}

// FanProfileOverviewReader は fan profile overview 用の read 操作を表します。
type FanProfileOverviewReader interface {
	GetOverview(ctx context.Context, viewerUserID uuid.UUID) (fanprofile.Overview, error)
}

// FanProfileFollowingReader は fan profile following 用の read 操作を表します。
type FanProfileFollowingReader interface {
	ListFollowing(ctx context.Context, viewerID uuid.UUID, cursor *fanprofile.FollowingCursor, limit int) ([]fanprofile.FollowingItem, *fanprofile.FollowingCursor, error)
}

// HandlerConfig は router が依存する read model をまとめます。
type HandlerConfig struct {
	AppEnv               string
	CreatorSearch        CreatorSearchReader
	CreatorWorkspace     CreatorWorkspaceReader
	CreatorUpload        CreatorUploadHandler
	CreatorProfile       CreatorProfileReader
	CreatorProfileShorts CreatorProfileShortsReader
	CreatorFollow        CreatorFollowWriter
	CreatorAvatarUpload  ViewerCreatorAvatarUploadHandler
	CreatorRegistration  ViewerCreatorRegistrationWriter
	FanProfileFollowing  FanProfileFollowingReader
	FanProfileOverview   FanProfileOverviewReader
	FanAuth              FanAuthService
	AuthCookie           AuthCookieConfig
	ViewerActiveMode     ViewerActiveModeSwitcher
	ViewerBootstrap      ViewerBootstrapReader
	Dependencies         []Dependency
}

// Config は HTTP サーバーの実行設定を表します。
type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
}

// Server は Gin の起動と graceful shutdown を管理します。
type Server struct {
	config     Config
	httpServer *http.Server
	logger     *slog.Logger
}

// NewHandler は API サーバー用の Gin router を構築します。
func NewHandler(config HandlerConfig) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(devLoopbackCORS(config.AppEnv))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/readyz", func(c *gin.Context) {
		var failed []string
		for _, dependency := range config.Dependencies {
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

	if config.ViewerBootstrap != nil {
		router.GET("/api/viewer/bootstrap", buildViewerBootstrapHandler(config.ViewerBootstrap))
	}

	registerFanAuthRoutes(router, config.FanAuth, config.AuthCookie)
	registerFanProfileRoutes(router, config.FanProfileOverview, config.FanProfileFollowing, config.ViewerBootstrap)
	registerCreatorWorkspaceRoutes(router, config.CreatorWorkspace, config.ViewerBootstrap)
	registerCreatorUploadRoutes(router, config.CreatorUpload, config.ViewerBootstrap)
	registerCreatorSearchRoutes(router, config.CreatorSearch)
	registerCreatorProfileRoutes(router, config.CreatorProfile, config.CreatorProfileShorts, config.CreatorFollow, config.ViewerBootstrap)
	registerViewerCreatorEntryRoutes(router, config.CreatorRegistration, config.CreatorAvatarUpload, config.ViewerActiveMode, config.ViewerBootstrap)

	return router
}

// New は実行設定と依存先から Server を構築します。
func New(cfg Config, logger *slog.Logger, handlerConfig HandlerConfig) *Server {
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 10 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}

	handler := NewHandler(handlerConfig)

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
