package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/config"
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

	queueConfig := sqs.Config{
		Region:   cfg.AWSRegion,
		QueueURL: cfg.SQSQueueURL,
	}

	if queueConfig.Enabled() {
		if _, err := sqs.NewClient(ctx, queueConfig); err != nil {
			logger.Error("failed to initialize sqs client", "error", err)
			os.Exit(1)
		}

		logger.Info("worker skeleton started with sqs configuration", "queue_url", queueConfig.QueueURL, "region", queueConfig.Region)
	} else {
		logger.Info("worker skeleton started without sqs configuration")
	}

	<-ctx.Done()
	logger.Info("worker skeleton stopped")
}
