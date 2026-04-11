package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/config"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/mediaconvert"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	medias3 "github.com/LinkLynx-AI/shorts-fans/backend/internal/s3"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/sqs"
)

type wakeQueueAdapter struct {
	queue *sqs.Queue
}

func (a wakeQueueAdapter) ReceiveWakeMessages(ctx context.Context) ([]media.WakeMessage, error) {
	messages, err := a.queue.ReceiveWakeMessages(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]media.WakeMessage, 0, len(messages))
	for _, message := range messages {
		result = append(result, media.WakeMessage{
			MediaAssetID:  message.MediaAssetID,
			ReceiptHandle: message.ReceiptHandle,
		})
	}

	return result, nil
}

func (a wakeQueueAdapter) DeleteMessage(ctx context.Context, receiptHandle string) error {
	return a.queue.DeleteMessage(ctx, receiptHandle)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()
	if err := cfg.ValidateWorker(); err != nil {
		logger.Error("invalid worker configuration", "error", err)
		os.Exit(1)
	}
	if !cfg.MediaSandboxEnabled() {
		logger.Info("worker is idle because media sandbox configuration is not enabled")
		<-ctx.Done()
		return
	}

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Error("failed to initialize postgres pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	s3Client, err := medias3.NewClient(ctx, medias3.Config{Region: cfg.AWSRegion})
	if err != nil {
		logger.Error("failed to initialize s3 client", "error", err)
		os.Exit(1)
	}
	mediaConvertClient, err := mediaconvert.NewClient(ctx, mediaconvert.Config{Region: cfg.AWSRegion})
	if err != nil {
		logger.Error("failed to initialize mediaconvert client", "error", err)
		os.Exit(1)
	}
	sqsClient, err := sqs.NewClient(ctx, sqs.Config{
		Region:   cfg.AWSRegion,
		QueueURL: cfg.MediaJobsQueueURL,
	})
	if err != nil {
		logger.Error("failed to initialize sqs client", "error", err)
		os.Exit(1)
	}
	mediaJobsQueue, err := sqs.NewQueue(sqsClient, cfg.MediaJobsQueueURL)
	if err != nil {
		logger.Error("failed to initialize media jobs queue", "error", err)
		os.Exit(1)
	}

	delivery, err := media.NewDelivery(media.DeliveryConfig{
		ShortPublicBaseURL:    cfg.MediaShortPublicBaseURL,
		ShortPublicBucketName: cfg.MediaShortPublicBucketName,
		MainPrivateBucketName: cfg.MediaMainPrivateBucketName,
	}, s3Client)
	if err != nil {
		logger.Error("failed to initialize media delivery", "error", err)
		os.Exit(1)
	}
	materializer, err := media.NewMaterializer(media.MaterializerConfig{
		ShortPublicBucketName:      cfg.MediaShortPublicBucketName,
		MainPrivateBucketName:      cfg.MediaMainPrivateBucketName,
		MediaConvertServiceRoleARN: cfg.MediaConvertServiceRoleARN,
	}, mediaConvertClient, delivery, s3Client)
	if err != nil {
		logger.Error("failed to initialize media materializer", "error", err)
		os.Exit(1)
	}
	processor, err := media.NewProcessor(pool, materializer)
	if err != nil {
		logger.Error("failed to initialize media processor", "error", err)
		os.Exit(1)
	}
	worker, err := media.NewWorker(media.WorkerConfig{}, wakeQueueAdapter{queue: mediaJobsQueue}, processor)
	if err != nil {
		logger.Error("failed to initialize media worker", "error", err)
		os.Exit(1)
	}

	logger.Info(
		"media worker starting",
		"region", cfg.AWSRegion,
		"queue_url", cfg.MediaJobsQueueURL,
		"raw_bucket", cfg.MediaRawBucketName,
		"short_public_bucket", cfg.MediaShortPublicBucketName,
		"main_private_bucket", cfg.MediaMainPrivateBucketName,
	)
	if err := worker.Run(ctx); err != nil {
		logger.Error("media worker stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("media worker stopped")
}
