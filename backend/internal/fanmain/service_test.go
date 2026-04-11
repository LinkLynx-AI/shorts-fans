package fanmain

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/unlock"
	"github.com/google/uuid"
)

type stubFeedReader struct {
	getDetail func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error)
}

func (s stubFeedReader) GetDetail(ctx context.Context, shortID uuid.UUID, viewerUserID *uuid.UUID) (feed.Detail, error) {
	return s.getDetail(ctx, shortID, viewerUserID)
}

type stubMainReader struct {
	getUnlockableMain func(context.Context, uuid.UUID) (shorts.Main, error)
}

func (s stubMainReader) GetUnlockableMain(ctx context.Context, id uuid.UUID) (shorts.Main, error) {
	return s.getUnlockableMain(ctx, id)
}

type stubUnlockRecorder struct {
	ensureMainUnlock func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error)
}

func (s stubUnlockRecorder) EnsureMainUnlock(ctx context.Context, input unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
	if s.ensureMainUnlock == nil {
		return unlock.MainUnlock{}, nil
	}

	return s.ensureMainUnlock(ctx, input)
}

func TestServiceUnlockEntryPlaybackFlow(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	mainAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	now := time.Unix(1_710_000_000, 0).UTC()
	recordedAt := now.Add(30 * time.Second)
	recordCallCount := 0

	service := NewService(
		stubFeedReader{
			getDetail: func(_ context.Context, gotShortID uuid.UUID, gotViewerID *uuid.UUID) (feed.Detail, error) {
				if gotShortID != shortID {
					t.Fatalf("GetDetail() shortID got %s want %s", gotShortID, shortID)
				}
				if gotViewerID == nil || *gotViewerID != viewerID {
					t.Fatalf("GetDetail() viewerUserID got %v want %s", gotViewerID, viewerID)
				}

				return feed.Detail{
					Item: feed.Item{
						Creator: feed.CreatorSummary{
							Bio:         "quiet rooftop specialist",
							DisplayName: "Mina Rei",
							Handle:      "minarei",
							ID:          viewerID,
						},
						Short: feed.ShortSummary{
							Caption:                "quiet rooftop preview",
							CanonicalMainID:        mainID,
							CreatorUserID:          viewerID,
							ID:                     shortID,
							MediaAssetID:           shortAssetID,
							PreviewDurationSeconds: 16,
						},
						Unlock: feed.UnlockPreview{
							IsOwner:             false,
							IsUnlocked:          false,
							MainDurationSeconds: 480,
							PriceJPY:            1800,
						},
					},
				}, nil
			},
		},
		stubMainReader{
			getUnlockableMain: func(_ context.Context, gotMainID uuid.UUID) (shorts.Main, error) {
				if gotMainID != mainID {
					t.Fatalf("GetUnlockableMain() id got %s want %s", gotMainID, mainID)
				}

				return shorts.Main{
					CreatedAt:     now,
					CreatorUserID: viewerID,
					ID:            mainID,
					MediaAssetID:  mainAssetID,
					PriceMinor:    1800,
				}, nil
			},
		},
		stubUnlockRecorder{
			ensureMainUnlock: func(_ context.Context, input unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
				recordCallCount++
				if input.UserID != viewerID || input.MainID != mainID {
					t.Fatalf("EnsureMainUnlock() input got %+v want viewer=%s main=%s", input, viewerID, mainID)
				}

				return unlock.MainUnlock{
					UserID:      viewerID,
					MainID:      mainID,
					PurchasedAt: recordedAt,
					CreatedAt:   recordedAt,
				}, nil
			},
		},
	)
	service.now = func() time.Time { return now }

	const sessionBinding = "session-hash"

	unlock, err := service.GetUnlockSurface(context.Background(), viewerID, sessionBinding, shortID)
	if err != nil {
		t.Fatalf("GetUnlockSurface() error = %v, want nil", err)
	}
	if unlock.UnlockCta.State != "unlock_available" {
		t.Fatalf("GetUnlockSurface() state got %q want %q", unlock.UnlockCta.State, "unlock_available")
	}
	if unlock.Short.Caption != "quiet rooftop preview" {
		t.Fatalf("GetUnlockSurface() short caption got %q want %q", unlock.Short.Caption, "quiet rooftop preview")
	}
	if unlock.MainAccessToken == "" {
		t.Fatal("GetUnlockSurface() main access token = empty, want value")
	}

	issued, err := service.IssueAccessEntry(context.Background(), sessionBinding, AccessEntryInput{
		AcceptedAge:   true,
		AcceptedTerms: true,
		EntryToken:    unlock.MainAccessToken,
		FromShortID:   shortID,
		MainID:        mainID,
		ViewerID:      viewerID,
	})
	if err != nil {
		t.Fatalf("IssueAccessEntry() error = %v, want nil", err)
	}
	if issued.GrantKind != MainPlaybackGrantKindUnlocked {
		t.Fatalf("IssueAccessEntry() grant kind got %q want %q", issued.GrantKind, MainPlaybackGrantKindUnlocked)
	}
	if recordCallCount != 1 {
		t.Fatalf("IssueAccessEntry() record call count got %d want %d", recordCallCount, 1)
	}

	playback, err := service.GetPlaybackSurface(context.Background(), viewerID, sessionBinding, mainID, shortID, issued.GrantToken)
	if err != nil {
		t.Fatalf("GetPlaybackSurface() error = %v, want nil", err)
	}
	if playback.Access.Status != "unlocked" {
		t.Fatalf("GetPlaybackSurface() access status got %q want %q", playback.Access.Status, "unlocked")
	}
	if playback.EntryShort.Caption != "quiet rooftop preview" {
		t.Fatalf("GetPlaybackSurface() entry short caption got %q want %q", playback.EntryShort.Caption, "quiet rooftop preview")
	}
}

func TestServiceIssueAccessEntryRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	now := time.Unix(1_710_000_000, 0).UTC()

	service := NewService(
		stubFeedReader{
			getDetail: func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error) {
				return feed.Detail{
					Item: feed.Item{
						Creator: feed.CreatorSummary{
							DisplayName: "Mina Rei",
							Handle:      "minarei",
							ID:          viewerID,
						},
						Short: feed.ShortSummary{
							CanonicalMainID:        mainID,
							CreatorUserID:          viewerID,
							ID:                     shortID,
							PreviewDurationSeconds: 16,
						},
						Unlock: feed.UnlockPreview{
							MainDurationSeconds: 480,
							PriceJPY:            1800,
						},
					},
				}, nil
			},
		},
		stubMainReader{
			getUnlockableMain: func(context.Context, uuid.UUID) (shorts.Main, error) {
				return shorts.Main{CreatedAt: now, ID: mainID, PriceMinor: 1800}, nil
			},
		},
		stubUnlockRecorder{},
	)
	service.now = func() time.Time { return now }

	_, err := service.IssueAccessEntry(context.Background(), "session-hash", AccessEntryInput{
		EntryToken:  "invalid-token",
		FromShortID: shortID,
		MainID:      mainID,
		ViewerID:    viewerID,
	})
	if !errors.Is(err, ErrMainLocked) {
		t.Fatalf("IssueAccessEntry() error got %v want %v", err, ErrMainLocked)
	}
}

func TestServiceIssueAccessEntryAllowsExistingUnlockRecord(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	now := time.Unix(1_710_000_000, 0).UTC()
	recordCallCount := 0

	service := NewService(
		stubFeedReader{
			getDetail: func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error) {
				return feed.Detail{
					Item: feed.Item{
						Creator: feed.CreatorSummary{
							DisplayName: "Mina Rei",
							Handle:      "minarei",
							ID:          viewerID,
						},
						Short: feed.ShortSummary{
							CanonicalMainID:        mainID,
							CreatorUserID:          viewerID,
							ID:                     shortID,
							PreviewDurationSeconds: 16,
						},
						Unlock: feed.UnlockPreview{
							IsUnlocked:          true,
							MainDurationSeconds: 480,
							PriceJPY:            1800,
						},
					},
				}, nil
			},
		},
		stubMainReader{
			getUnlockableMain: func(context.Context, uuid.UUID) (shorts.Main, error) {
				return shorts.Main{CreatedAt: now, ID: mainID, PriceMinor: 1800}, nil
			},
		},
		stubUnlockRecorder{
			ensureMainUnlock: func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
				recordCallCount++
				return unlock.MainUnlock{
					UserID:      viewerID,
					MainID:      mainID,
					PurchasedAt: now,
					CreatedAt:   now,
				}, nil
			},
		},
	)
	service.now = func() time.Time { return now }

	unlockSurface, err := service.GetUnlockSurface(context.Background(), viewerID, "session-hash", shortID)
	if err != nil {
		t.Fatalf("GetUnlockSurface() error = %v, want nil", err)
	}

	issued, err := service.IssueAccessEntry(context.Background(), "session-hash", AccessEntryInput{
		EntryToken:  unlockSurface.MainAccessToken,
		FromShortID: shortID,
		MainID:      mainID,
		ViewerID:    viewerID,
	})
	if err != nil {
		t.Fatalf("IssueAccessEntry() error = %v, want nil", err)
	}
	if issued.GrantKind != MainPlaybackGrantKindUnlocked {
		t.Fatalf("IssueAccessEntry() grant kind got %q want %q", issued.GrantKind, MainPlaybackGrantKindUnlocked)
	}
	if recordCallCount != 1 {
		t.Fatalf("IssueAccessEntry() record call count got %d want %d", recordCallCount, 1)
	}
}

func TestServiceIssueAccessEntrySkipsOwnerPersistence(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	now := time.Unix(1_710_000_000, 0).UTC()
	recordCallCount := 0

	service := NewService(
		stubFeedReader{
			getDetail: func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error) {
				return feed.Detail{
					Item: feed.Item{
						Creator: feed.CreatorSummary{
							DisplayName: "Aoi N",
							Handle:      "aoina",
							ID:          viewerID,
						},
						Short: feed.ShortSummary{
							CanonicalMainID:        mainID,
							CreatorUserID:          viewerID,
							ID:                     shortID,
							PreviewDurationSeconds: 16,
						},
						Unlock: feed.UnlockPreview{
							IsOwner: true,
						},
					},
				}, nil
			},
		},
		stubMainReader{
			getUnlockableMain: func(context.Context, uuid.UUID) (shorts.Main, error) {
				return shorts.Main{CreatedAt: now, ID: mainID, PriceMinor: 1800}, nil
			},
		},
		stubUnlockRecorder{
			ensureMainUnlock: func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
				recordCallCount++
				return unlock.MainUnlock{}, nil
			},
		},
	)
	service.now = func() time.Time { return now }

	unlockSurface, err := service.GetUnlockSurface(context.Background(), viewerID, "session-hash", shortID)
	if err != nil {
		t.Fatalf("GetUnlockSurface() error = %v, want nil", err)
	}

	issued, err := service.IssueAccessEntry(context.Background(), "session-hash", AccessEntryInput{
		EntryToken:  unlockSurface.MainAccessToken,
		FromShortID: shortID,
		MainID:      mainID,
		ViewerID:    viewerID,
	})
	if err != nil {
		t.Fatalf("IssueAccessEntry() error = %v, want nil", err)
	}
	if issued.GrantKind != MainPlaybackGrantKindOwner {
		t.Fatalf("IssueAccessEntry() grant kind got %q want %q", issued.GrantKind, MainPlaybackGrantKindOwner)
	}
	if recordCallCount != 0 {
		t.Fatalf("IssueAccessEntry() record call count got %d want %d", recordCallCount, 0)
	}
}

func TestServiceIssueAccessEntryFailsWhenPersistenceFails(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	now := time.Unix(1_710_000_000, 0).UTC()
	recordErr := errors.New("write failed")

	service := NewService(
		stubFeedReader{
			getDetail: func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error) {
				return feed.Detail{
					Item: feed.Item{
						Creator: feed.CreatorSummary{
							DisplayName: "Mina Rei",
							Handle:      "minarei",
							ID:          viewerID,
						},
						Short: feed.ShortSummary{
							CanonicalMainID:        mainID,
							CreatorUserID:          viewerID,
							ID:                     shortID,
							PreviewDurationSeconds: 16,
						},
						Unlock: feed.UnlockPreview{
							MainDurationSeconds: 480,
							PriceJPY:            1800,
						},
					},
				}, nil
			},
		},
		stubMainReader{
			getUnlockableMain: func(context.Context, uuid.UUID) (shorts.Main, error) {
				return shorts.Main{CreatedAt: now, ID: mainID, PriceMinor: 1800}, nil
			},
		},
		stubUnlockRecorder{
			ensureMainUnlock: func(context.Context, unlock.RecordMainUnlockInput) (unlock.MainUnlock, error) {
				return unlock.MainUnlock{}, recordErr
			},
		},
	)
	service.now = func() time.Time { return now }

	unlockSurface, err := service.GetUnlockSurface(context.Background(), viewerID, "session-hash", shortID)
	if err != nil {
		t.Fatalf("GetUnlockSurface() error = %v, want nil", err)
	}

	_, err = service.IssueAccessEntry(context.Background(), "session-hash", AccessEntryInput{
		EntryToken:  unlockSurface.MainAccessToken,
		FromShortID: shortID,
		MainID:      mainID,
		ViewerID:    viewerID,
	})
	if !errors.Is(err, recordErr) {
		t.Fatalf("IssueAccessEntry() error got %v want wrapped %v", err, recordErr)
	}
}

func TestServiceMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	t.Run("unlock surface", func(t *testing.T) {
		t.Parallel()

		service := NewService(
			stubFeedReader{
				getDetail: func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error) {
					return feed.Detail{}, feed.ErrPublicShortNotFound
				},
			},
			stubMainReader{
				getUnlockableMain: func(context.Context, uuid.UUID) (shorts.Main, error) {
					t.Fatal("GetUnlockableMain() should not be called")
					return shorts.Main{}, nil
				},
			},
			stubUnlockRecorder{},
		)

		_, err := service.GetUnlockSurface(context.Background(), viewerID, "session-hash", shortID)
		if !errors.Is(err, ErrShortUnlockNotFound) {
			t.Fatalf("GetUnlockSurface() error got %v want %v", err, ErrShortUnlockNotFound)
		}
	})

	t.Run("access entry", func(t *testing.T) {
		t.Parallel()

		service := NewService(
			stubFeedReader{
				getDetail: func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error) {
					return feed.Detail{}, feed.ErrPublicShortNotFound
				},
			},
			stubMainReader{
				getUnlockableMain: func(context.Context, uuid.UUID) (shorts.Main, error) {
					t.Fatal("GetUnlockableMain() should not be called")
					return shorts.Main{}, nil
				},
			},
			stubUnlockRecorder{},
		)

		_, err := service.IssueAccessEntry(context.Background(), "session-hash", AccessEntryInput{
			EntryToken:  "ignored",
			FromShortID: shortID,
			MainID:      mainID,
			ViewerID:    viewerID,
		})
		if !errors.Is(err, ErrAccessEntryNotFound) {
			t.Fatalf("IssueAccessEntry() error got %v want %v", err, ErrAccessEntryNotFound)
		}
	})

	t.Run("playback", func(t *testing.T) {
		t.Parallel()

		service := NewService(
			stubFeedReader{
				getDetail: func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error) {
					return feed.Detail{}, feed.ErrPublicShortNotFound
				},
			},
			stubMainReader{
				getUnlockableMain: func(context.Context, uuid.UUID) (shorts.Main, error) {
					t.Fatal("GetUnlockableMain() should not be called")
					return shorts.Main{}, nil
				},
			},
			stubUnlockRecorder{},
		)

		_, err := service.GetPlaybackSurface(context.Background(), viewerID, "session-hash", mainID, shortID, "ignored")
		if !errors.Is(err, ErrPlaybackNotFound) {
			t.Fatalf("GetPlaybackSurface() error got %v want %v", err, ErrPlaybackNotFound)
		}
	})
}

func TestServiceRejectsMismatchedIdentifiers(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	otherMainID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	now := time.Unix(1_710_000_000, 0).UTC()

	service := NewService(
		stubFeedReader{
			getDetail: func(context.Context, uuid.UUID, *uuid.UUID) (feed.Detail, error) {
				return feed.Detail{
					Item: feed.Item{
						Creator: feed.CreatorSummary{
							DisplayName: "Mina Rei",
							Handle:      "minarei",
							ID:          viewerID,
						},
						Short: feed.ShortSummary{
							CanonicalMainID:        mainID,
							CreatorUserID:          viewerID,
							ID:                     shortID,
							PreviewDurationSeconds: 16,
						},
						Unlock: feed.UnlockPreview{
							MainDurationSeconds: 480,
							PriceJPY:            1800,
						},
					},
				}, nil
			},
		},
		stubMainReader{
			getUnlockableMain: func(context.Context, uuid.UUID) (shorts.Main, error) {
				return shorts.Main{CreatedAt: now, ID: mainID, PriceMinor: 1800}, nil
			},
		},
		stubUnlockRecorder{},
	)
	service.now = func() time.Time { return now }

	unlock, err := service.GetUnlockSurface(context.Background(), viewerID, "session-hash", shortID)
	if err != nil {
		t.Fatalf("GetUnlockSurface() error = %v, want nil", err)
	}

	_, err = service.IssueAccessEntry(context.Background(), "session-hash", AccessEntryInput{
		EntryToken:  unlock.MainAccessToken,
		FromShortID: shortID,
		MainID:      otherMainID,
		ViewerID:    viewerID,
	})
	if !errors.Is(err, ErrAccessEntryNotFound) {
		t.Fatalf("IssueAccessEntry() error got %v want %v", err, ErrAccessEntryNotFound)
	}

	issued, err := service.IssueAccessEntry(context.Background(), "session-hash", AccessEntryInput{
		EntryToken:  unlock.MainAccessToken,
		FromShortID: shortID,
		MainID:      mainID,
		ViewerID:    viewerID,
	})
	if err != nil {
		t.Fatalf("IssueAccessEntry() error = %v, want nil", err)
	}

	_, err = service.GetPlaybackSurface(context.Background(), viewerID, "session-hash", otherMainID, shortID, issued.GrantToken)
	if !errors.Is(err, ErrPlaybackNotFound) {
		t.Fatalf("GetPlaybackSurface() error got %v want %v", err, ErrPlaybackNotFound)
	}
}

func TestServiceHelpers(t *testing.T) {
	t.Parallel()

	mainID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	if got := buildGrantedAccessState(mainID, MainPlaybackGrantKindOwner); got.Status != "owner" || got.Reason != "owner_preview" {
		t.Fatalf("buildGrantedAccessState(owner) got %+v", got)
	}
	if got := buildGrantedAccessState(mainID, MainPlaybackGrantKindUnlocked); got.Status != "unlocked" || got.Reason != "session_unlocked" {
		t.Fatalf("buildGrantedAccessState(unlocked) got %+v", got)
	}

	if got := buildMainAccessState(feed.UnlockPreview{IsOwner: true}, mainID); got.Status != "owner" {
		t.Fatalf("buildMainAccessState(owner) got %+v", got)
	}
	if got := buildMainAccessState(feed.UnlockPreview{IsUnlocked: true}, mainID); got.Status != "unlocked" {
		t.Fatalf("buildMainAccessState(unlocked) got %+v", got)
	}
	if got := buildMainAccessState(feed.UnlockPreview{}, mainID); got.Status != "locked" {
		t.Fatalf("buildMainAccessState(locked) got %+v", got)
	}

	if got := buildUnlockCtaState(feed.UnlockPreview{IsOwner: true}); got.State != "owner_preview" {
		t.Fatalf("buildUnlockCtaState(owner) got %+v", got)
	}
	if got := buildUnlockCtaState(feed.UnlockPreview{IsUnlocked: true}); got.State != "continue_main" {
		t.Fatalf("buildUnlockCtaState(unlocked) got %+v", got)
	}
	if got := buildUnlockCtaState(feed.UnlockPreview{MainDurationSeconds: 480, PriceJPY: 1800}); got.State != "unlock_available" || got.MainDurationSeconds == nil || got.PriceJPY == nil {
		t.Fatalf("buildUnlockCtaState(available) got %+v", got)
	}

	if got := resolveGrantKind(feed.UnlockPreview{IsOwner: true}); got != MainPlaybackGrantKindOwner {
		t.Fatalf("resolveGrantKind(owner) got %q", got)
	}
	if got := resolveGrantKind(feed.UnlockPreview{IsUnlocked: true}); got != MainPlaybackGrantKindUnlocked {
		t.Fatalf("resolveGrantKind(unlocked) got %q", got)
	}
	if got := resolveGrantKind(feed.UnlockPreview{}); got != MainPlaybackGrantKindUnlocked {
		t.Fatalf("resolveGrantKind(default) got %q", got)
	}
}

func TestSignedTokenValidation(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_710_000_000, 0).UTC()
	payload := signedTokenPayload{
		Kind:        entryTokenKind,
		MainID:      uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		FromShortID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		ViewerID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	}

	if _, err := issueSignedToken("", now, time.Minute, payload); err == nil {
		t.Fatal("issueSignedToken() error = nil, want value")
	}
	if _, err := issueSignedToken("session-hash", now, 0, payload); err == nil {
		t.Fatal("issueSignedToken() ttl error = nil, want value")
	}

	token, err := issueSignedToken("session-hash", now, time.Minute, payload)
	if err != nil {
		t.Fatalf("issueSignedToken() error = %v, want nil", err)
	}

	if _, err := readSignedToken("session-hash", now, "invalid"); err == nil {
		t.Fatal("readSignedToken() format error = nil, want value")
	}
	if _, err := readSignedToken("session-hash", now, "%%%."+token[strings.Index(token, ".")+1:]); err == nil {
		t.Fatal("readSignedToken() decode payload error = nil, want value")
	}
	if _, err := readSignedToken("session-hash", now, token[:strings.LastIndex(token, ".")+1]+"zz"); err == nil {
		t.Fatal("readSignedToken() decode signature error = nil, want value")
	}
	if _, err := readSignedToken("other-session", now, token); err == nil {
		t.Fatal("readSignedToken() signature mismatch error = nil, want value")
	}
	if _, err := readSignedToken("session-hash", now.Add(2*time.Minute), token); err == nil {
		t.Fatal("readSignedToken() expiry error = nil, want value")
	}
}

func TestLoadLinkedSurfaceRequiresReaders(t *testing.T) {
	t.Parallel()

	var service *Service
	_, _, err := service.loadLinkedSurface(context.Background(), uuid.Nil, uuid.Nil)
	if err == nil {
		t.Fatal("loadLinkedSurface() error = nil, want value")
	}
}
