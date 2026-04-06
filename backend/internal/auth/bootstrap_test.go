package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type bootstrapRepositoryStub struct {
	touchSessionLastSeen func(context.Context, string, time.Time) (SessionRecord, error)
	getCurrentViewer     func(context.Context, string) (CurrentViewer, error)
}

func (s bootstrapRepositoryStub) TouchSessionLastSeenByTokenHash(
	ctx context.Context,
	sessionTokenHash string,
	lastSeenAt time.Time,
) (SessionRecord, error) {
	return s.touchSessionLastSeen(ctx, sessionTokenHash, lastSeenAt)
}

func (s bootstrapRepositoryStub) GetCurrentViewerBySessionTokenHash(
	ctx context.Context,
	sessionTokenHash string,
) (CurrentViewer, error) {
	return s.getCurrentViewer(ctx, sessionTokenHash)
}

func TestReadCurrentViewerReturnsEmptyStateForBlankToken(t *testing.T) {
	t.Parallel()

	called := false
	reader := NewReader(bootstrapRepositoryStub{
		touchSessionLastSeen: func(context.Context, string, time.Time) (SessionRecord, error) {
			called = true
			return SessionRecord{}, nil
		},
		getCurrentViewer: func(context.Context, string) (CurrentViewer, error) {
			called = true
			return CurrentViewer{}, nil
		},
	})

	got, err := reader.ReadCurrentViewer(context.Background(), "   ")
	if err != nil {
		t.Fatalf("ReadCurrentViewer() error = %v, want nil", err)
	}
	if got.CurrentViewer != nil {
		t.Fatalf("ReadCurrentViewer() current viewer got %#v want nil", got.CurrentViewer)
	}
	if called {
		t.Fatal("ReadCurrentViewer() repository was called for blank token")
	}
}

func TestReadCurrentViewerReturnsEmptyStateForMissingSession(t *testing.T) {
	t.Parallel()

	reader := NewReader(bootstrapRepositoryStub{
		touchSessionLastSeen: func(context.Context, string, time.Time) (SessionRecord, error) {
			return SessionRecord{}, ErrSessionNotFound
		},
		getCurrentViewer: func(context.Context, string) (CurrentViewer, error) {
			return CurrentViewer{}, ErrCurrentViewerNotFound
		},
	})

	got, err := reader.ReadCurrentViewer(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("ReadCurrentViewer() error = %v, want nil", err)
	}
	if got.CurrentViewer != nil {
		t.Fatalf("ReadCurrentViewer() current viewer got %#v want nil", got.CurrentViewer)
	}
}

func TestReadCurrentViewerNormalizesInvalidCreatorMode(t *testing.T) {
	t.Parallel()

	reader := NewReader(bootstrapRepositoryStub{
		touchSessionLastSeen: func(context.Context, string, time.Time) (SessionRecord, error) {
			return SessionRecord{}, nil
		},
		getCurrentViewer: func(context.Context, string) (CurrentViewer, error) {
			return CurrentViewer{
				ID:                   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				ActiveMode:           ActiveModeCreator,
				CanAccessCreatorMode: false,
			}, nil
		},
	})

	got, err := reader.ReadCurrentViewer(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("ReadCurrentViewer() error = %v, want nil", err)
	}
	if got.CurrentViewer == nil {
		t.Fatal("ReadCurrentViewer() current viewer = nil, want viewer")
	}
	if got.CurrentViewer.ActiveMode != ActiveModeFan {
		t.Fatalf("ReadCurrentViewer() active mode got %q want %q", got.CurrentViewer.ActiveMode, ActiveModeFan)
	}
}

func TestReadCurrentViewerWrapsRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("query failed")
	reader := NewReader(bootstrapRepositoryStub{
		touchSessionLastSeen: func(context.Context, string, time.Time) (SessionRecord, error) {
			return SessionRecord{}, nil
		},
		getCurrentViewer: func(context.Context, string) (CurrentViewer, error) {
			return CurrentViewer{}, expectedErr
		},
	})

	if _, err := reader.ReadCurrentViewer(context.Background(), "raw-token"); !errors.Is(err, expectedErr) {
		t.Fatalf("ReadCurrentViewer() error got %v want wrapped %v", err, expectedErr)
	}
}
