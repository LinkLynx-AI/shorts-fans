package media

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const defaultWorkerIdleDelay = 2 * time.Second

type wakeQueue interface {
	ReceiveWakeMessages(ctx context.Context) ([]WakeMessage, error)
	DeleteMessage(ctx context.Context, receiptHandle string) error
}

type workerProcessor interface {
	ProcessAsset(ctx context.Context, mediaAssetID uuid.UUID) error
	ProcessNextQueued(ctx context.Context) (bool, error)
}

// WakeMessage は queue から受け取る media processing wake-up payload です。
type WakeMessage struct {
	MediaAssetID  uuid.UUID
	ReceiptHandle string
}

// WorkerConfig は media worker loop の実行設定です。
type WorkerConfig struct {
	IdleDelay time.Duration
}

// Worker は SQS wake-up と queued job reconciliation を実行します。
type Worker struct {
	queue     wakeQueue
	processor workerProcessor
	idleDelay time.Duration
}

// NewWorker は media worker loop を構築します。
func NewWorker(cfg WorkerConfig, queue wakeQueue, processor workerProcessor) (*Worker, error) {
	switch {
	case queue == nil:
		return nil, fmt.Errorf("wake queue is required")
	case processor == nil:
		return nil, fmt.Errorf("worker processor is required")
	}
	if cfg.IdleDelay <= 0 {
		cfg.IdleDelay = defaultWorkerIdleDelay
	}

	return &Worker{
		queue:     queue,
		processor: processor,
		idleDelay: cfg.IdleDelay,
	}, nil
}

// Run は SQS wake-up を処理し、queue が空のときは DB queued job を回収します。
func (w *Worker) Run(ctx context.Context) error {
	if w == nil {
		return fmt.Errorf("media worker is nil")
	}

	if err := w.reconcileOnce(ctx); err != nil {
		return err
	}

	for {
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return err
		}

		messages, err := w.queue.ReceiveWakeMessages(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("receive wake messages: %w", err)
		}
		if len(messages) == 0 {
			if err := w.reconcileOnce(ctx); err != nil {
				return err
			}
			if err := waitForIdle(ctx, w.idleDelay); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return nil
				}
				return err
			}
			continue
		}

		for _, message := range messages {
			if err := w.processor.ProcessAsset(ctx, message.MediaAssetID); err != nil {
				return fmt.Errorf("process media asset id=%s: %w", message.MediaAssetID, err)
			}
			if err := w.queue.DeleteMessage(ctx, message.ReceiptHandle); err != nil {
				return fmt.Errorf("delete wake message media_asset_id=%s: %w", message.MediaAssetID, err)
			}
		}
	}
}

func (w *Worker) reconcileOnce(ctx context.Context) error {
	_, err := w.processor.ProcessNextQueued(ctx)
	if err != nil {
		return fmt.Errorf("reconcile queued media processing job: %w", err)
	}

	return nil
}

func waitForIdle(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
