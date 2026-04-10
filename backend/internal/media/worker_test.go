package media

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stubWakeQueue struct {
	receive func(context.Context) ([]WakeMessage, error)
	delete  func(context.Context, string) error
}

func (s stubWakeQueue) ReceiveWakeMessages(ctx context.Context) ([]WakeMessage, error) {
	return s.receive(ctx)
}

func (s stubWakeQueue) DeleteMessage(ctx context.Context, receiptHandle string) error {
	return s.delete(ctx, receiptHandle)
}

type stubWorkerProcessor struct {
	processAsset      func(context.Context, uuid.UUID) error
	processNextQueued func(context.Context) (bool, error)
}

func (s stubWorkerProcessor) ProcessAsset(ctx context.Context, mediaAssetID uuid.UUID) error {
	return s.processAsset(ctx, mediaAssetID)
}

func (s stubWorkerProcessor) ProcessNextQueued(ctx context.Context) (bool, error) {
	return s.processNextQueued(ctx)
}

func TestWorkerProcessesWakeMessages(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mediaAssetID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	reconcileCalls := 0
	processCalls := 0
	deleteCalls := 0

	worker, err := NewWorker(WorkerConfig{IdleDelay: time.Millisecond}, stubWakeQueue{
		receive: func(context.Context) ([]WakeMessage, error) {
			cancel()
			return []WakeMessage{{
				MediaAssetID:  mediaAssetID,
				ReceiptHandle: "receipt-1",
			}}, nil
		},
		delete: func(_ context.Context, receiptHandle string) error {
			deleteCalls++
			if receiptHandle != "receipt-1" {
				t.Fatalf("DeleteMessage() receipt handle got %q want %q", receiptHandle, "receipt-1")
			}
			return nil
		},
	}, stubWorkerProcessor{
		processAsset: func(_ context.Context, got uuid.UUID) error {
			processCalls++
			if got != mediaAssetID {
				t.Fatalf("ProcessAsset() id got %s want %s", got, mediaAssetID)
			}
			return nil
		},
		processNextQueued: func(context.Context) (bool, error) {
			reconcileCalls++
			return false, nil
		},
	})
	if err != nil {
		t.Fatalf("NewWorker() error = %v, want nil", err)
	}

	if err := worker.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if reconcileCalls == 0 {
		t.Fatal("Run() reconcileCalls = 0, want startup reconciliation")
	}
	if processCalls != 1 {
		t.Fatalf("Run() processCalls got %d want 1", processCalls)
	}
	if deleteCalls != 1 {
		t.Fatalf("Run() deleteCalls got %d want 1", deleteCalls)
	}
}

func TestWorkerPropagatesProcessorError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("process failed")
	worker, err := NewWorker(WorkerConfig{IdleDelay: time.Millisecond}, stubWakeQueue{
		receive: func(context.Context) ([]WakeMessage, error) {
			return []WakeMessage{{
				MediaAssetID:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				ReceiptHandle: "receipt-1",
			}}, nil
		},
		delete: func(context.Context, string) error { return nil },
	}, stubWorkerProcessor{
		processAsset: func(context.Context, uuid.UUID) error { return wantErr },
		processNextQueued: func(context.Context) (bool, error) {
			return false, nil
		},
	})
	if err != nil {
		t.Fatalf("NewWorker() error = %v, want nil", err)
	}

	if err := worker.Run(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("Run() error got %v want %v", err, wantErr)
	}
}
