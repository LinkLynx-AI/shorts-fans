package main

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/config"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/mediaconvert"
	medias3 "github.com/LinkLynx-AI/shorts-fans/backend/internal/s3"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/sqs"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()
	if err := cfg.ValidateMediaSmoke(); err != nil {
		logger.Error("invalid media smoke configuration", "error", err)
		os.Exit(1)
	}

	s3Client, err := medias3.NewClient(ctx, medias3.Config{Region: cfg.AWSRegion})
	if err != nil {
		logger.Error("failed to initialize s3 client", "error", err)
		os.Exit(1)
	}

	queueChecker, err := sqs.NewAccessChecker(ctx, sqs.Config{
		Region:   cfg.AWSRegion,
		QueueURL: cfg.MediaJobsQueueURL,
	})
	if err != nil {
		logger.Error("failed to initialize sqs access checker", "error", err)
		os.Exit(1)
	}

	mediaConvertClient, err := mediaconvert.NewClient(ctx, mediaconvert.Config{Region: cfg.AWSRegion})
	if err != nil {
		logger.Error("failed to initialize mediaconvert client", "error", err)
		os.Exit(1)
	}

	delivery, err := media.NewDelivery(media.DeliveryConfig{
		ShortPublicBaseURL:    cfg.MediaShortPublicBaseURL,
		ShortPublicBucketName: cfg.MediaShortPublicBucketName,
		MainPrivateBucketName: cfg.MediaMainPrivateBucketName,
	}, s3Client)
	if err != nil {
		logger.Error("failed to initialize media delivery helper", "error", err)
		os.Exit(1)
	}

	probeRunner, err := media.NewProbeRunner(media.ProbeConfig{
		ShortPublicBucketName: cfg.MediaShortPublicBucketName,
		MainPrivateBucketName: cfg.MediaMainPrivateBucketName,
	}, delivery, s3Client, queueChecker, mediaConvertClient, nil)
	if err != nil {
		logger.Error("failed to initialize media smoke runner", "error", err)
		os.Exit(1)
	}

	result, err := probeRunner.Run(ctx)
	if err != nil {
		logger.Error("media smoke failed", "error", err)
		os.Exit(1)
	}

	logger.Info(
		"media smoke succeeded",
		"queue_arn", result.QueueARN,
		"mediaconvert_queue", result.MediaConvertQueueName,
		"short_public_url", result.ShortPublicURL,
		"short_object_key", result.ShortObjectKey,
		"main_object_key", result.MainObjectKey,
		"main_signed_url_host", signedURLHost(result.MainSignedURL),
	)
}

func signedURLHost(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return parsedURL.Host
}
