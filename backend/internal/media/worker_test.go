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

func TestNewWorkerValidatesDependenciesAndDefaultsIdleDelay(t *testing.T) {
	t.Parallel()

	processor := stubWorkerProcessor{
		processAsset: func(context.Context, uuid.UUID) error { return nil },
		processNextQueued: func(context.Context) (bool, error) {
			return false, nil
		},
	}
	queue := stubWakeQueue{
		receive: func(context.Context) ([]WakeMessage, error) { return nil, nil },
		delete:  func(context.Context, string) error { return nil },
	}

	if _, err := NewWorker(WorkerConfig{}, nil, processor); err == nil {
		t.Fatal("NewWorker() error = nil, want queue validation error")
	}
	if _, err := NewWorker(WorkerConfig{}, queue, nil); err == nil {
		t.Fatal("NewWorker() error = nil, want processor validation error")
	}

	worker, err := NewWorker(WorkerConfig{}, queue, processor)
	if err != nil {
		t.Fatalf("NewWorker() error = %v, want nil", err)
	}
	if got, want := worker.idleDelay, defaultWorkerIdleDelay; got != want {
		t.Fatalf("NewWorker() idleDelay got %s want %s", got, want)
	}
}

func TestWorkerRunRejectsNilWorker(t *testing.T) {
	t.Parallel()

	var worker *Worker
	if err := worker.Run(context.Background()); err == nil {
		t.Fatal("Run() error = nil, want nil receiver error")
	}
}

func TestWorkerRunPropagatesReceiveError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("receive failed")
	worker, err := NewWorker(WorkerConfig{IdleDelay: time.Millisecond}, stubWakeQueue{
		receive: func(context.Context) ([]WakeMessage, error) { return nil, wantErr },
		delete:  func(context.Context, string) error { return nil },
	}, stubWorkerProcessor{
		processAsset: func(context.Context, uuid.UUID) error { return nil },
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

func TestWorkerRunPropagatesDeleteError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("delete failed")
	worker, err := NewWorker(WorkerConfig{IdleDelay: time.Millisecond}, stubWakeQueue{
		receive: func(context.Context) ([]WakeMessage, error) {
			return []WakeMessage{{
				MediaAssetID:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				ReceiptHandle: "receipt-1",
			}}, nil
		},
		delete: func(context.Context, string) error { return wantErr },
	}, stubWorkerProcessor{
		processAsset: func(context.Context, uuid.UUID) error { return nil },
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

func TestWorkerRunReconcilesWhenQueueIsEmpty(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	receiveCalls := 0
	reconcileCalls := 0
	worker, err := NewWorker(WorkerConfig{IdleDelay: time.Hour}, stubWakeQueue{
		receive: func(context.Context) ([]WakeMessage, error) {
			receiveCalls++
			return nil, nil
		},
		delete: func(context.Context, string) error { return nil },
	}, stubWorkerProcessor{
		processAsset: func(context.Context, uuid.UUID) error {
			t.Fatal("ProcessAsset() should not be called when queue is empty")
			return nil
		},
		processNextQueued: func(context.Context) (bool, error) {
			reconcileCalls++
			if reconcileCalls == 2 {
				cancel()
			}
			return false, nil
		},
	})
	if err != nil {
		t.Fatalf("NewWorker() error = %v, want nil", err)
	}

	if err := worker.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if receiveCalls != 1 {
		t.Fatalf("Run() receiveCalls got %d want 1", receiveCalls)
	}
	if reconcileCalls != 2 {
		t.Fatalf("Run() reconcileCalls got %d want 2", reconcileCalls)
	}
}

func TestWorkerRunPropagatesReconcileError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("reconcile failed")
	worker, err := NewWorker(WorkerConfig{IdleDelay: time.Millisecond}, stubWakeQueue{
		receive: func(context.Context) ([]WakeMessage, error) { return nil, nil },
		delete:  func(context.Context, string) error { return nil },
	}, stubWorkerProcessor{
		processAsset: func(context.Context, uuid.UUID) error { return nil },
		processNextQueued: func(context.Context) (bool, error) {
			return false, wantErr
		},
	})
	if err != nil {
		t.Fatalf("NewWorker() error = %v, want nil", err)
	}

	if err := worker.Run(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("Run() error got %v want %v", err, wantErr)
	}
}

func TestWaitForIdle(t *testing.T) {
	t.Parallel()

	if err := waitForIdle(context.Background(), time.Nanosecond); err != nil {
		t.Fatalf("waitForIdle() error = %v, want nil", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitForIdle(ctx, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("waitForIdle() error got %v want %v", err, context.Canceled)
	}
}
