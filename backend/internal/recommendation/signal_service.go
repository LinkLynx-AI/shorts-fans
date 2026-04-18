package recommendation

import (
	"context"
	"errors"
	"fmt"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
	"github.com/google/uuid"
)

// ErrSignalTargetNotFound は recommendation signal の対象 short / creator が見つからないことを表します。
var ErrSignalTargetNotFound = errors.New("recommendation signal target が見つかりません")

type signalShortDetailReader interface {
	GetDetail(ctx context.Context, shortID uuid.UUID, viewerUserID *uuid.UUID) (feed.Detail, error)
}

type signalCreatorProfileReader interface {
	GetPublicProfile(ctx context.Context, userID uuid.UUID) (creator.Profile, error)
}

type signalEventRecorder interface {
	RecordEvent(ctx context.Context, input RecordEventInput) (RecordEventResult, error)
}

// SignalService は recommendation signal の target identity 解決と event 記録を扱います。
type SignalService struct {
	shortDetailReader     signalShortDetailReader
	creatorProfileReader  signalCreatorProfileReader
	recommendationEventDB signalEventRecorder
}

// NewSignalService は recommendation signal service を構築します。
func NewSignalService(
	shortDetailReader signalShortDetailReader,
	creatorProfileReader signalCreatorProfileReader,
	recommendationEventDB signalEventRecorder,
) *SignalService {
	return &SignalService{
		shortDetailReader:     shortDetailReader,
		creatorProfileReader:  creatorProfileReader,
		recommendationEventDB: recommendationEventDB,
	}
}

// RecordShortSignal は public short 起点の recommendation signal を記録します。
func (s *SignalService) RecordShortSignal(
	ctx context.Context,
	viewerID uuid.UUID,
	shortID uuid.UUID,
	eventKind EventKind,
	idempotencyKey string,
) (RecordEventResult, error) {
	switch eventKind {
	case EventKindImpression, EventKindViewStart, EventKindViewCompletion, EventKindRewatchLoop, EventKindMainClick:
	default:
		return RecordEventResult{}, fmt.Errorf("recommendation short signal 記録: %w", ErrEventKindInvalid)
	}

	detail, err := s.resolveShortDetail(ctx, viewerID, shortID)
	if err != nil {
		return RecordEventResult{}, err
	}

	return s.recordEvent(ctx, RecordEventInput{
		ViewerUserID:    viewerID,
		EventKind:       eventKind,
		CreatorUserID:   signalUUIDPtr(detail.Item.Short.CreatorUserID),
		CanonicalMainID: signalUUIDPtr(detail.Item.Short.CanonicalMainID),
		ShortID:         signalUUIDPtr(detail.Item.Short.ID),
		IdempotencyKey:  idempotencyKey,
	})
}

// RecordProfileClick は creator profile click signal を記録します。
func (s *SignalService) RecordProfileClick(
	ctx context.Context,
	viewerID uuid.UUID,
	creatorUserID uuid.UUID,
	idempotencyKey string,
) (RecordEventResult, error) {
	if s == nil || s.creatorProfileReader == nil {
		return RecordEventResult{}, fmt.Errorf("recommendation signal service の creator profile reader が初期化されていません")
	}

	if _, err := s.creatorProfileReader.GetPublicProfile(ctx, creatorUserID); err != nil {
		if errors.Is(err, creator.ErrProfileNotFound) {
			return RecordEventResult{}, fmt.Errorf("recommendation creator signal 対象取得 creator=%s: %w", creatorUserID, ErrSignalTargetNotFound)
		}

		return RecordEventResult{}, fmt.Errorf("recommendation creator signal 対象取得 creator=%s: %w", creatorUserID, err)
	}

	return s.recordEvent(ctx, RecordEventInput{
		ViewerUserID:   viewerID,
		EventKind:      EventKindProfileClick,
		CreatorUserID:  signalUUIDPtr(creatorUserID),
		IdempotencyKey: idempotencyKey,
	})
}

// RecordUnlockConversion は access-entry 成功に紐づく unlock conversion signal を記録します。
func (s *SignalService) RecordUnlockConversion(
	ctx context.Context,
	viewerID uuid.UUID,
	detail feed.Detail,
	idempotencyKey string,
) error {
	if s == nil {
		return fmt.Errorf("recommendation signal service が初期化されていません")
	}

	_, err := s.recordEvent(ctx, RecordEventInput{
		ViewerUserID:    viewerID,
		EventKind:       EventKindUnlockConversion,
		CreatorUserID:   signalUUIDPtr(detail.Item.Short.CreatorUserID),
		CanonicalMainID: signalUUIDPtr(detail.Item.Short.CanonicalMainID),
		ShortID:         signalUUIDPtr(detail.Item.Short.ID),
		IdempotencyKey:  idempotencyKey,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *SignalService) resolveShortDetail(ctx context.Context, viewerID uuid.UUID, shortID uuid.UUID) (feed.Detail, error) {
	if s == nil || s.shortDetailReader == nil {
		return feed.Detail{}, fmt.Errorf("recommendation signal service の short detail reader が初期化されていません")
	}

	detail, err := s.shortDetailReader.GetDetail(ctx, shortID, &viewerID)
	if err != nil {
		if errors.Is(err, feed.ErrPublicShortNotFound) {
			return feed.Detail{}, fmt.Errorf("recommendation short signal 対象取得 short=%s: %w", shortID, ErrSignalTargetNotFound)
		}

		return feed.Detail{}, fmt.Errorf("recommendation short signal 対象取得 short=%s: %w", shortID, err)
	}

	return detail, nil
}

func (s *SignalService) recordEvent(ctx context.Context, input RecordEventInput) (RecordEventResult, error) {
	if s == nil || s.recommendationEventDB == nil {
		return RecordEventResult{}, fmt.Errorf("recommendation signal service の event recorder が初期化されていません")
	}

	result, err := s.recommendationEventDB.RecordEvent(ctx, input)
	if err != nil {
		return RecordEventResult{}, fmt.Errorf("recommendation signal 記録 event=%s viewer=%s: %w", input.EventKind, input.ViewerUserID, err)
	}

	return result, nil
}

func signalUUIDPtr(value uuid.UUID) *uuid.UUID {
	return &value
}
