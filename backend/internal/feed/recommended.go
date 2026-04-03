package feed

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const defaultRecommendedPageSize = 20

// ErrInvalidCursor は recommended feed cursor が不正なことを表します。
var ErrInvalidCursor = errors.New("recommended feed cursor が不正です")

type recommendedRepository interface {
	listRecommended(ctx context.Context, cursor *recommendedCursor, limit int32) ([]recommendedRecord, error)
}

// MediaAsset は fan public surface で返す media asset を表します。
type MediaAsset struct {
	ID              string  `json:"id"`
	Kind            string  `json:"kind"`
	URL             string  `json:"url"`
	PosterURL       *string `json:"posterUrl"`
	DurationSeconds *int    `json:"durationSeconds"`
}

// CreatorSummary は feed item 用の creator summary を表します。
type CreatorSummary struct {
	ID          string     `json:"id"`
	DisplayName string     `json:"displayName"`
	Handle      string     `json:"handle"`
	Avatar      MediaAsset `json:"avatar"`
	Bio         string     `json:"bio"`
}

// ShortSummary は feed item 用の short summary を表します。
type ShortSummary struct {
	ID                     string     `json:"id"`
	CanonicalMainID        string     `json:"canonicalMainId"`
	CreatorID              string     `json:"creatorId"`
	Title                  string     `json:"title"`
	Caption                string     `json:"caption"`
	Media                  MediaAsset `json:"media"`
	PreviewDurationSeconds int        `json:"previewDurationSeconds"`
}

// FeedViewerState は recommended feed 上の viewer relation state を表します。
type FeedViewerState struct {
	IsPinned bool `json:"isPinned"`
}

// UnlockCtaState は feed item の CTA state を表します。
type UnlockCtaState struct {
	State                 string `json:"state"`
	PriceJPY              *int64 `json:"priceJpy"`
	MainDurationSeconds   *int   `json:"mainDurationSeconds"`
	ResumePositionSeconds *int   `json:"resumePositionSeconds"`
}

// FeedItem は recommended feed の item を表します。
type FeedItem struct {
	Short     ShortSummary     `json:"short"`
	Creator   CreatorSummary   `json:"creator"`
	Viewer    FeedViewerState  `json:"viewer"`
	UnlockCta UnlockCtaState   `json:"unlockCta"`
}

// RecommendedFeed は `GET /api/fan/feed` の recommended data を表します。
type RecommendedFeed struct {
	Tab        string     `json:"tab"`
	Items      []FeedItem `json:"items"`
	NextCursor *string    `json:"-"`
	HasNext    bool       `json:"-"`
}

// ListRecommendedInput は recommended feed 一覧取得の入力です。
type ListRecommendedInput struct {
	Cursor string
	Limit  int
}

// RecommendedService は recommended feed の read use case を提供します。
type RecommendedService struct {
	repo recommendedRepository
}

// NewRecommendedService は recommended feed service を構築します。
func NewRecommendedService(repo *Repository) *RecommendedService {
	return &RecommendedService{repo: repo}
}

func newRecommendedService(repo recommendedRepository) *RecommendedService {
	return &RecommendedService{repo: repo}
}

// ListRecommended は recommended feed を返します。
func (s *RecommendedService) ListRecommended(ctx context.Context, input ListRecommendedInput) (RecommendedFeed, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = defaultRecommendedPageSize
	}

	cursor, err := decodeRecommendedCursor(strings.TrimSpace(input.Cursor))
	if err != nil {
		return RecommendedFeed{}, fmt.Errorf("recommended feed cursor 解釈: %w", err)
	}

	records, err := s.repo.listRecommended(ctx, cursor, int32(limit+1))
	if err != nil {
		return RecommendedFeed{}, err
	}

	hasNext := len(records) > limit
	if hasNext {
		records = records[:limit]
	}

	items := make([]FeedItem, 0, len(records))
	for _, record := range records {
		items = append(items, buildFeedItem(record))
	}

	var nextCursor *string
	if hasNext && len(records) > 0 {
		cursorValue := encodeRecommendedCursor(recommendedCursor{
			PublishedAt: records[len(records)-1].ShortPublishedAt,
			ShortID:     records[len(records)-1].ShortID,
		})
		nextCursor = &cursorValue
	}

	return RecommendedFeed{
		Tab:        "recommended",
		Items:      items,
		NextCursor: nextCursor,
		HasNext:    hasNext,
	}, nil
}

func buildFeedItem(record recommendedRecord) FeedItem {
	shortDuration := record.ShortDurationSeconds
	mainDuration := record.MainDurationSeconds
	price := record.MainPriceJPY

	return FeedItem{
		Short: ShortSummary{
			ID:              record.ShortID.String(),
			CanonicalMainID: record.CanonicalMainID.String(),
			CreatorID:       record.CreatorUserID.String(),
			Title:           record.ShortTitle,
			Caption:         record.ShortCaption,
			Media: MediaAsset{
				ID:              record.ShortMediaAssetID.String(),
				Kind:            "video",
				URL:             record.ShortPlaybackURL,
				PosterURL:       nil,
				DurationSeconds: &shortDuration,
			},
			PreviewDurationSeconds: record.ShortDurationSeconds,
		},
		Creator: CreatorSummary{
			ID:          record.CreatorUserID.String(),
			DisplayName: record.CreatorDisplayName,
			Handle:      normalizeCreatorHandle(record.CreatorHandle),
			Avatar: MediaAsset{
				ID:              buildAvatarAssetID(record.CreatorUserID),
				Kind:            "image",
				URL:             buildAvatarURL(record.CreatorUserID, record.CreatorAvatarURL),
				PosterURL:       nil,
				DurationSeconds: nil,
			},
			Bio: record.CreatorBio,
		},
		Viewer: FeedViewerState{
			IsPinned: false,
		},
		UnlockCta: UnlockCtaState{
			State:                 "unlock_available",
			PriceJPY:              &price,
			MainDurationSeconds:   &mainDuration,
			ResumePositionSeconds: nil,
		},
	}
}

func encodeRecommendedCursor(cursor recommendedCursor) string {
	payload := cursor.PublishedAt.UTC().Format(time.RFC3339Nano) + "|" + cursor.ShortID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func decodeRecommendedCursor(raw string) (*recommendedCursor, error) {
	if raw == "" {
		return nil, nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: cursor decode: %v", ErrInvalidCursor, err)
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w: cursor format が不正です", ErrInvalidCursor)
	}

	publishedAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: cursor published_at 解釈: %v", ErrInvalidCursor, err)
	}

	shortID, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: cursor short id 解釈: %v", ErrInvalidCursor, err)
	}

	return &recommendedCursor{
		PublishedAt: publishedAt,
		ShortID:     shortID,
	}, nil
}

func normalizeCreatorHandle(handle string) string {
	if strings.HasPrefix(handle, "@") {
		return handle
	}

	return "@" + handle
}

func buildAvatarAssetID(creatorID uuid.UUID) string {
	return "asset_creator_" + strings.ReplaceAll(creatorID.String(), "-", "") + "_avatar"
}

func buildAvatarURL(creatorID uuid.UUID, avatarURL *string) string {
	if avatarURL != nil && strings.TrimSpace(*avatarURL) != "" {
		return *avatarURL
	}

	palette := [][]string{
		{"#d6f5ff", "#65bae0", "#1c4e6f"},
		{"#edf7ff", "#7bcbe6", "#315f8d"},
		{"#fff4dc", "#79c8ef", "#264f70"},
		{"#dff9ff", "#70b0d1", "#233e57"},
	}
	index := int(creatorID[0]) % len(palette)
	colors := palette[index]
	svg := fmt.Sprintf(
		"<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 80 80' fill='none'><defs><linearGradient id='avatar-gradient' x1='40' y1='0' x2='40' y2='80' gradientUnits='userSpaceOnUse'><stop stop-color='%s' /><stop offset='0.56' stop-color='%s' /><stop offset='1' stop-color='%s' /></linearGradient></defs><rect width='80' height='80' rx='40' fill='url(#avatar-gradient)' /></svg>",
		colors[0],
		colors[1],
		colors[2],
	)
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}
