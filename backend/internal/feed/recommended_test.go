package feed

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type repositoryStub struct {
	listRecommendedFunc func(context.Context, *recommendedCursor, int32) ([]recommendedRecord, error)
}

func (s repositoryStub) listRecommended(ctx context.Context, cursor *recommendedCursor, limit int32) ([]recommendedRecord, error) {
	if s.listRecommendedFunc == nil {
		return nil, nil
	}

	return s.listRecommendedFunc(ctx, cursor, limit)
}

func TestListRecommendedReturnsMappedItemsAndPageInfo(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	nextShortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	creatorID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	mainID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	shortMediaID := uuid.MustParse("55555555-5555-5555-5555-555555555555")

	service := newRecommendedService(repositoryStub{
		listRecommendedFunc: func(_ context.Context, cursor *recommendedCursor, limit int32) ([]recommendedRecord, error) {
			if cursor != nil {
				t.Fatalf("cursor got %#v want nil", cursor)
			}
			if limit != 2 {
				t.Fatalf("limit got %d want 2", limit)
			}

			return []recommendedRecord{
				{
					ShortID:              shortID,
					CanonicalMainID:      mainID,
					CreatorUserID:        creatorID,
					ShortTitle:           "quiet rooftop preview",
					ShortCaption:         "quiet rooftop preview。",
					ShortMediaAssetID:    shortMediaID,
					ShortPublishedAt:     now,
					ShortPlaybackURL:     "https://cdn.example.com/short.mp4",
					ShortDurationSeconds: 16,
					CreatorDisplayName:   "Mina Rei",
					CreatorHandle:        "@minarei",
					CreatorBio:           "quiet rooftop と hotel light の preview を軸に投稿。",
					MainID:               mainID,
					MainPriceJPY:         1800,
					MainDurationSeconds:  480,
				},
				{
					ShortID:              nextShortID,
					CanonicalMainID:      mainID,
					CreatorUserID:        creatorID,
					ShortTitle:           "hotel light preview",
					ShortCaption:         "hotel light preview。",
					ShortMediaAssetID:    shortMediaID,
					ShortPublishedAt:     now.Add(-time.Minute),
					ShortPlaybackURL:     "https://cdn.example.com/short-2.mp4",
					ShortDurationSeconds: 14,
					CreatorDisplayName:   "Mina Rei",
					CreatorHandle:        "@minarei",
					CreatorBio:           "quiet rooftop と hotel light の preview を軸に投稿。",
					MainID:               mainID,
					MainPriceJPY:         1800,
					MainDurationSeconds:  480,
				},
			}, nil
		},
	})

	got, err := service.ListRecommended(context.Background(), ListRecommendedInput{Limit: 1})
	if err != nil {
		t.Fatalf("ListRecommended() error = %v, want nil", err)
	}

	if got.Tab != "recommended" {
		t.Fatalf("ListRecommended() tab got %q want %q", got.Tab, "recommended")
	}
	if len(got.Items) != 1 {
		t.Fatalf("ListRecommended() items len got %d want %d", len(got.Items), 1)
	}
	if got.Items[0].Short.ID != shortID.String() {
		t.Fatalf("ListRecommended() short id got %q want %q", got.Items[0].Short.ID, shortID.String())
	}
	if got.Items[0].UnlockCta.State != "unlock_available" {
		t.Fatalf("ListRecommended() unlock state got %q want %q", got.Items[0].UnlockCta.State, "unlock_available")
	}
	if got.NextCursor == nil {
		t.Fatal("ListRecommended() next cursor = nil, want non-nil")
	}
	if !got.HasNext {
		t.Fatal("ListRecommended() hasNext = false, want true")
	}
	if got.Items[0].Creator.Avatar.URL == "" {
		t.Fatal("ListRecommended() avatar url = empty, want non-empty")
	}
}

func TestListRecommendedDecodesCursor(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	wantCursor := &recommendedCursor{
		PublishedAt: now,
		ShortID:     shortID,
	}

	service := newRecommendedService(repositoryStub{
		listRecommendedFunc: func(_ context.Context, cursor *recommendedCursor, _ int32) ([]recommendedRecord, error) {
			if cursor == nil {
				t.Fatal("cursor = nil, want non-nil")
			}
			if !cursor.PublishedAt.Equal(wantCursor.PublishedAt) || cursor.ShortID != wantCursor.ShortID {
				t.Fatalf("cursor got %#v want %#v", cursor, wantCursor)
			}

			return nil, nil
		},
	})

	if _, err := service.ListRecommended(context.Background(), ListRecommendedInput{
		Cursor: encodeRecommendedCursor(*wantCursor),
	}); err != nil {
		t.Fatalf("ListRecommended() error = %v, want nil", err)
	}
}

func TestListRecommendedRejectsMalformedCursor(t *testing.T) {
	t.Parallel()

	service := newRecommendedService(repositoryStub{})

	if _, err := service.ListRecommended(context.Background(), ListRecommendedInput{
		Cursor: "not-a-valid-cursor",
	}); err == nil {
		t.Fatal("ListRecommended() error = nil, want malformed cursor error")
	}
}

func TestBuildAvatarURLFallsBackToGeneratedDataURL(t *testing.T) {
	t.Parallel()

	got := buildAvatarURL(uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), nil)
	if !strings.HasPrefix(got, "data:image/svg+xml;base64,") {
		t.Fatalf("buildAvatarURL() got %q want data URL prefix", got)
	}
}
