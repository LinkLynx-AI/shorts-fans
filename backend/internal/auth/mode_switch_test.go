package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type activeModeRepositoryStub struct {
	updateActiveModeByTokenHash func(context.Context, string, ActiveMode) (SessionRecord, error)
}

func (s activeModeRepositoryStub) UpdateActiveModeByTokenHash(
	ctx context.Context,
	sessionTokenHash string,
	activeMode ActiveMode,
) (SessionRecord, error) {
	return s.updateActiveModeByTokenHash(ctx, sessionTokenHash, activeMode)
}

func TestModeSwitcherSwitchActiveMode(t *testing.T) {
	t.Parallel()

	var gotTokenHash string
	var gotActiveMode ActiveMode

	switcher := NewModeSwitcher(activeModeRepositoryStub{
		updateActiveModeByTokenHash: func(_ context.Context, sessionTokenHash string, activeMode ActiveMode) (SessionRecord, error) {
			gotTokenHash = sessionTokenHash
			gotActiveMode = activeMode
			return SessionRecord{}, nil
		},
	})

	if err := switcher.SwitchActiveMode(context.Background(), " raw-session-token ", ActiveModeCreator); err != nil {
		t.Fatalf("SwitchActiveMode() error = %v, want nil", err)
	}
	if gotTokenHash != HashSessionToken("raw-session-token") {
		t.Fatalf("SwitchActiveMode() token hash got %q want %q", gotTokenHash, HashSessionToken("raw-session-token"))
	}
	if gotActiveMode != ActiveModeCreator {
		t.Fatalf("SwitchActiveMode() active mode got %q want %q", gotActiveMode, ActiveModeCreator)
	}
}

func TestModeSwitcherRejectsInvalidMode(t *testing.T) {
	t.Parallel()

	switcher := NewModeSwitcher(activeModeRepositoryStub{
		updateActiveModeByTokenHash: func(context.Context, string, ActiveMode) (SessionRecord, error) {
			t.Fatal("UpdateActiveModeByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})

	if err := switcher.SwitchActiveMode(context.Background(), "raw-session-token", ActiveMode("invalid")); !errors.Is(err, ErrInvalidActiveMode) {
		t.Fatalf("SwitchActiveMode() error got %v want %v", err, ErrInvalidActiveMode)
	}
}

func TestModeSwitcherRejectsBlankToken(t *testing.T) {
	t.Parallel()

	switcher := NewModeSwitcher(activeModeRepositoryStub{
		updateActiveModeByTokenHash: func(context.Context, string, ActiveMode) (SessionRecord, error) {
			t.Fatal("UpdateActiveModeByTokenHash() should not be called")
			return SessionRecord{}, nil
		},
	})

	if err := switcher.SwitchActiveMode(context.Background(), "   ", ActiveModeFan); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("SwitchActiveMode() error got %v want %v", err, ErrSessionNotFound)
	}
}

func TestModeSwitcherRejectsUninitializedSwitcher(t *testing.T) {
	t.Parallel()

	var switcher *ModeSwitcher

	if err := switcher.SwitchActiveMode(context.Background(), "raw-session-token", ActiveModeFan); err == nil {
		t.Fatal("SwitchActiveMode() error = nil, want initialization error")
	}
}

func TestModeSwitcherWrapsRepositoryError(t *testing.T) {
	t.Parallel()

	repositoryErr := errors.New("update failed")
	switcher := NewModeSwitcher(activeModeRepositoryStub{
		updateActiveModeByTokenHash: func(context.Context, string, ActiveMode) (SessionRecord, error) {
			return SessionRecord{}, repositoryErr
		},
	})

	err := switcher.SwitchActiveMode(context.Background(), "raw-session-token", ActiveModeFan)
	if !errors.Is(err, repositoryErr) {
		t.Fatalf("SwitchActiveMode() error got %v want wrapped %v", err, repositoryErr)
	}
	if strings.Contains(err.Error(), "raw-session-token") {
		t.Fatalf("SwitchActiveMode() error got %q want redacted token", err)
	}
	if !strings.Contains(err.Error(), HashSessionToken("raw-session-token")) {
		t.Fatalf("SwitchActiveMode() error got %q want token hash", err)
	}
}
