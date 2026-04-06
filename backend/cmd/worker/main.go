package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/config"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/mediaconvert"
	medias3 "github.com/LinkLynx-AI/shorts-fans/backend/internal/s3"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/sqs"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()
	if err := cfg.ValidateWorker(); err != nil {
		logger.Error("invalid worker configuration", "error", err)
		os.Exit(1)
	}

	if cfg.MediaSandboxEnabled() {
		queueConfig := sqs.Config{
			Region:   cfg.AWSRegion,
			QueueURL: cfg.MediaJobsQueueURL,
		}
		if _, err := sqs.NewClient(ctx, queueConfig); err != nil {
			logger.Error("failed to initialize sqs client", "error", err)
			os.Exit(1)
		}
		if _, err := medias3.NewClient(ctx, medias3.Config{Region: cfg.AWSRegion}); err != nil {
			logger.Error("failed to initialize s3 client", "error", err)
			os.Exit(1)
		}
		if _, err := mediaconvert.NewClient(ctx, mediaconvert.Config{Region: cfg.AWSRegion}); err != nil {
			logger.Error("failed to initialize mediaconvert client", "error", err)
			os.Exit(1)
		}

		logger.Info(
			"worker skeleton started with media sandbox configuration",
			"queue_url", queueConfig.QueueURL,
			"region", queueConfig.Region,
			"raw_bucket", cfg.MediaRawBucketName,
			"short_public_bucket", cfg.MediaShortPublicBucketName,
			"short_public_base_url", cfg.MediaShortPublicBaseURL,
			"main_private_bucket", cfg.MediaMainPrivateBucketName,
			"mediaconvert_service_role_arn", cfg.MediaConvertServiceRoleARN,
		)
	} else {
		logger.Info("worker skeleton started without media sandbox configuration")
	}

	<-ctx.Done()
	logger.Info("worker skeleton stopped")
}
