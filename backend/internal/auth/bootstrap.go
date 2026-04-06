package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ActiveMode は現在前面に出す workspace を表します。
type ActiveMode string

const (
	// ActiveModeFan は fan mode を表します。
	ActiveModeFan ActiveMode = "fan"
	// ActiveModeCreator は creator mode を表します。
	ActiveModeCreator ActiveMode = "creator"
)

// CurrentViewer は app bootstrap で返す viewer の最小 state です。
type CurrentViewer struct {
	ID                   uuid.UUID
	ActiveMode           ActiveMode
	CanAccessCreatorMode bool
}

// Bootstrap は app bootstrap で返す current viewer state を表します。
type Bootstrap struct {
	CurrentViewer *CurrentViewer
}

type bootstrapRepository interface {
	GetCurrentViewerBySessionTokenHash(ctx context.Context, sessionTokenHash string) (CurrentViewer, error)
}

// Reader は session token から current viewer bootstrap を解決します。
type Reader struct {
	repository bootstrapRepository
}

// NewReader は bootstrap reader を構築します。
func NewReader(repository bootstrapRepository) *Reader {
	return &Reader{repository: repository}
}

// ReadCurrentViewer は raw session token から current viewer state を返します。
func (r *Reader) ReadCurrentViewer(ctx context.Context, rawSessionToken string) (Bootstrap, error) {
	if r == nil || r.repository == nil {
		return Bootstrap{}, fmt.Errorf("bootstrap reader が初期化されていません")
	}

	trimmedToken := strings.TrimSpace(rawSessionToken)
	if trimmedToken == "" {
		return Bootstrap{}, nil
	}

	viewer, err := r.repository.GetCurrentViewerBySessionTokenHash(ctx, HashSessionToken(trimmedToken))
	if err != nil {
		if errors.Is(err, ErrCurrentViewerNotFound) {
			return Bootstrap{}, nil
		}

		return Bootstrap{}, fmt.Errorf("current viewer bootstrap 読み取り: %w", err)
	}

	normalizedViewer := normalizeCurrentViewer(viewer)

	return Bootstrap{
		CurrentViewer: &normalizedViewer,
	}, nil
}

func normalizeCurrentViewer(viewer CurrentViewer) CurrentViewer {
	if viewer.ActiveMode == ActiveModeCreator && !viewer.CanAccessCreatorMode {
		viewer.ActiveMode = ActiveModeFan
	}

	return viewer
}
