package recommendation

import (
	"context"
	"errors"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
	"github.com/google/uuid"
)

type stubSignalShortDetailReader struct {
	getDetail func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error)
}

func (s stubSignalShortDetailReader) GetDetail(ctx context.Context, shortID uuid.UUID, viewerUserID *uuid.UUID) (feed.Detail, error) {
	return s.getDetail(ctx, shortID, viewerUserID)
}

type stubSignalCreatorProfileReader struct {
	getPublicProfile func(context.Context, uuid.UUID) (creator.Profile, error)
}

func (s stubSignalCreatorProfileReader) GetPublicProfile(ctx context.Context, userID uuid.UUID) (creator.Profile, error) {
	return s.getPublicProfile(ctx, userID)
}

type stubSignalEventRecorder struct {
	recordEvent func(context.Context, RecordEventInput) (RecordEventResult, error)
}

func (s stubSignalEventRecorder) RecordEvent(ctx context.Context, input RecordEventInput) (RecordEventResult, error) {
	return s.recordEvent(ctx, input)
}

func TestSignalServiceRecordShortSignalResolvesFeedIdentity(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorUserID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	service := NewSignalService(
		stubSignalShortDetailReader{
			getDetail: func(_ context.Context, gotShortID uuid.UUID, gotViewerUserID *uuid.UUID) (feed.Detail, error) {
				if gotShortID != shortID {
					t.Fatalf("GetDetail() shortID got %s want %s", gotShortID, shortID)
				}
				if gotViewerUserID == nil || *gotViewerUserID != viewerID {
					t.Fatalf("GetDetail() viewerUserID got %v want %s", gotViewerUserID, viewerID)
				}

				return feed.Detail{
					Item: feed.Item{
						Short: feed.ShortSummary{
							CanonicalMainID: mainID,
							CreatorUserID:   creatorUserID,
							ID:              shortID,
						},
					},
				}, nil
			},
		},
		nil,
		stubSignalEventRecorder{
			recordEvent: func(_ context.Context, input RecordEventInput) (RecordEventResult, error) {
				if input.ViewerUserID != viewerID {
					t.Fatalf("RecordEvent() viewerUserID got %s want %s", input.ViewerUserID, viewerID)
				}
				if input.EventKind != EventKindMainClick {
					t.Fatalf("RecordEvent() eventKind got %q want %q", input.EventKind, EventKindMainClick)
				}
				if input.CreatorUserID == nil || *input.CreatorUserID != creatorUserID {
					t.Fatalf("RecordEvent() creatorUserID got %v want %s", input.CreatorUserID, creatorUserID)
				}
				if input.CanonicalMainID == nil || *input.CanonicalMainID != mainID {
					t.Fatalf("RecordEvent() canonicalMainID got %v want %s", input.CanonicalMainID, mainID)
				}
				if input.ShortID == nil || *input.ShortID != shortID {
					t.Fatalf("RecordEvent() shortID got %v want %s", input.ShortID, shortID)
				}
				if input.IdempotencyKey != "main-click:session-1" {
					t.Fatalf("RecordEvent() idempotencyKey got %q want %q", input.IdempotencyKey, "main-click:session-1")
				}

				return RecordEventResult{Recorded: true}, nil
			},
		},
	)

	result, err := service.RecordShortSignal(context.Background(), viewerID, shortID, EventKindMainClick, "main-click:session-1")
	if err != nil {
		t.Fatalf("RecordShortSignal() error = %v, want nil", err)
	}
	if !result.Recorded {
		t.Fatal("RecordShortSignal() recorded = false, want true")
	}
}

func TestSignalServiceRecordShortSignalMapsNotFound(t *testing.T) {
	t.Parallel()

	service := NewSignalService(
		stubSignalShortDetailReader{
			getDetail: func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error) {
				return feed.Detail{}, feed.ErrPublicShortNotFound
			},
		},
		nil,
		stubSignalEventRecorder{
			recordEvent: func(context.Context, RecordEventInput) (RecordEventResult, error) {
				t.Fatal("RecordEvent() should not be called")
				return RecordEventResult{}, nil
			},
		},
	)

	_, err := service.RecordShortSignal(context.Background(), uuid.New(), uuid.New(), EventKindImpression, "signal:1")
	if !errors.Is(err, ErrSignalTargetNotFound) {
		t.Fatalf("RecordShortSignal() error got %v want %v", err, ErrSignalTargetNotFound)
	}
}

func TestSignalServiceRecordProfileClickChecksCreatorExistence(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorUserID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	service := NewSignalService(
		nil,
		stubSignalCreatorProfileReader{
			getPublicProfile: func(_ context.Context, gotCreatorUserID uuid.UUID) (creator.Profile, error) {
				if gotCreatorUserID != creatorUserID {
					t.Fatalf("GetPublicProfile() userID got %s want %s", gotCreatorUserID, creatorUserID)
				}

				return creator.Profile{
					UserID: creatorUserID,
				}, nil
			},
		},
		stubSignalEventRecorder{
			recordEvent: func(_ context.Context, input RecordEventInput) (RecordEventResult, error) {
				if input.EventKind != EventKindProfileClick {
					t.Fatalf("RecordEvent() eventKind got %q want %q", input.EventKind, EventKindProfileClick)
				}
				if input.CanonicalMainID != nil || input.ShortID != nil {
					t.Fatalf("RecordEvent() main/short ids got main=%v short=%v want nil", input.CanonicalMainID, input.ShortID)
				}
				if input.CreatorUserID == nil || *input.CreatorUserID != creatorUserID {
					t.Fatalf("RecordEvent() creatorUserID got %v want %s", input.CreatorUserID, creatorUserID)
				}
				if input.ViewerUserID != viewerID {
					t.Fatalf("RecordEvent() viewerUserID got %s want %s", input.ViewerUserID, viewerID)
				}

				return RecordEventResult{Recorded: true}, nil
			},
		},
	)

	result, err := service.RecordProfileClick(context.Background(), viewerID, creatorUserID, "profile-click:1")
	if err != nil {
		t.Fatalf("RecordProfileClick() error = %v, want nil", err)
	}
	if !result.Recorded {
		t.Fatal("RecordProfileClick() recorded = false, want true")
	}
}

func TestSignalServiceRecordUnlockConversionUsesResolvedIdentity(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorUserID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	service := NewSignalService(
		nil,
		nil,
		stubSignalEventRecorder{
			recordEvent: func(_ context.Context, input RecordEventInput) (RecordEventResult, error) {
				if input.ViewerUserID != viewerID {
					t.Fatalf("RecordEvent() viewerUserID got %s want %s", input.ViewerUserID, viewerID)
				}
				if input.EventKind != EventKindUnlockConversion {
					t.Fatalf("RecordEvent() eventKind got %q want %q", input.EventKind, EventKindUnlockConversion)
				}
				if input.CreatorUserID == nil || *input.CreatorUserID != creatorUserID {
					t.Fatalf("RecordEvent() creatorUserID got %v want %s", input.CreatorUserID, creatorUserID)
				}
				if input.CanonicalMainID == nil || *input.CanonicalMainID != mainID {
					t.Fatalf("RecordEvent() canonicalMainID got %v want %s", input.CanonicalMainID, mainID)
				}
				if input.ShortID == nil || *input.ShortID != shortID {
					t.Fatalf("RecordEvent() shortID got %v want %s", input.ShortID, shortID)
				}

				return RecordEventResult{Recorded: true}, nil
			},
		},
	)

	err := service.RecordUnlockConversion(context.Background(), viewerID, feed.Detail{
		Item: feed.Item{
			Short: feed.ShortSummary{
				CanonicalMainID: mainID,
				CreatorUserID:   creatorUserID,
				ID:              shortID,
			},
		},
	}, "access-entry:token")
	if err != nil {
		t.Fatalf("RecordUnlockConversion() error = %v, want nil", err)
	}
}
