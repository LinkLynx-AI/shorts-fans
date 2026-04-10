package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/config"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatoravatar"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatorupload"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/fanprofile"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/httpserver"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/redis"
	medias3 "github.com/LinkLynx-AI/shorts-fans/backend/internal/s3"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := newLogger()
	cfg := config.Load()
	if err := cfg.ValidateAPI(); err != nil {
		logger.Error("invalid api configuration", "error", err)
		os.Exit(1)
	}

	setGinMode(cfg.AppEnv)

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Error("failed to initialize postgres pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	redisClient, err := redis.NewClient(ctx, cfg.RedisAddr)
	if err != nil {
		logger.Error("failed to initialize redis client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("failed to close redis client", "error", err)
		}
	}()

	s3Client, err := medias3.NewClient(ctx, medias3.Config{Region: cfg.AWSRegion})
	if err != nil {
		logger.Error("failed to initialize s3 client", "error", err)
		os.Exit(1)
	}

	creatorRepository := creator.NewRepository(pool)
	creatorUploadRepository := creatorupload.NewRepository(pool)
	fanProfileRepository := fanprofile.NewRepository(pool)
	authRepository := auth.NewRepository(pool)
	viewerBootstrapReader := auth.NewReader(authRepository)
	authLifecycle := auth.NewLifecycle(authRepository)
	modeSwitcher := auth.NewModeSwitcher(authRepository)
	creatorUploadService, err := creatorupload.NewService(
		creatorupload.ServiceConfig{
			RawBucketName: cfg.MediaRawBucketName,
		},
		s3Client,
		creatorupload.NewRedisPackageStore(redisClient),
		creatorUploadRepository,
	)
	if err != nil {
		logger.Error("failed to initialize creator upload service", "error", err)
		os.Exit(1)
	}
	creatorAvatarService, err := creatoravatar.NewService(
		creatoravatar.ServiceConfig{
			DeliveryBaseURL:    cfg.CreatorAvatarBaseURL,
			DeliveryBucketName: cfg.CreatorAvatarDeliveryBucketName,
			UploadBucketName:   cfg.CreatorAvatarUploadBucketName,
		},
		s3Client,
		creatoravatar.NewRedisUploadStore(redisClient),
	)
	if err != nil {
		logger.Error("failed to initialize creator avatar service", "error", err)
		os.Exit(1)
	}

	server := httpserver.New(
		httpserver.Config{
			Addr:            cfg.APIAddr,
			ShutdownTimeout: 10 * time.Second,
		},
		logger,
		httpserver.HandlerConfig{
			AppEnv:               cfg.AppEnv,
			CreatorSearch:        creatorRepository,
			CreatorWorkspace:     creatorRepository,
			CreatorUpload:        creatorUploadService,
			CreatorProfile:       creatorRepository,
			CreatorProfileShorts: creatorRepository,
			CreatorFollow:        creatorRepository,
			CreatorAvatarUpload:  creatorAvatarService,
			CreatorRegistration:  creatorRepository,
			FanProfileOverview:   fanProfileRepository,
			FanProfileFollowing:  fanProfileRepository,
			FanAuth:              authLifecycle,
			AuthCookie:           httpserver.AuthCookieConfig{Secure: cfg.AppEnv == "production"},
			ViewerActiveMode:     modeSwitcher,
			ViewerBootstrap:      viewerBootstrapReader,
			Dependencies: []httpserver.Dependency{
				{Name: "postgres", Checker: postgres.NewReadinessChecker(pool)},
				{Name: "redis", Checker: redis.NewReadinessChecker(redisClient)},
			},
		},
	)

	logger.Info("api server starting", "addr", cfg.APIAddr, "app_env", cfg.AppEnv)
	if err := server.Run(ctx); err != nil {
		logger.Error("api server stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("api server stopped")
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func setGinMode(appEnv string) {
	if appEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
		return
	}

	gin.SetMode(gin.DebugMode)
}
