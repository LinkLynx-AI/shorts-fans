package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	fanFeedRequestScope                 = "fan_feed"
	fanShortDetailRequestScope          = "fan_short_detail"
	fanFeedFollowingAuthRequiredMessage = "following feed requires authentication"
)

type fanFeedResponseData struct {
	Items []fanFeedItemPayload `json:"items"`
	Tab   string               `json:"tab"`
}

type fanFeedItemPayload struct {
	Creator   creatorSummary        `json:"creator"`
	Short     feedShortSummary      `json:"short"`
	UnlockCta unlockCtaStatePayload `json:"unlockCta"`
	Viewer    feedViewerPayload     `json:"viewer"`
}

type shortDetailResponseData struct {
	Detail shortDetailPayload `json:"detail"`
}

type shortDetailPayload struct {
	Creator   creatorSummary        `json:"creator"`
	Short     feedShortSummary      `json:"short"`
	UnlockCta unlockCtaStatePayload `json:"unlockCta"`
	Viewer    shortDetailViewer     `json:"viewer"`
}

type feedShortSummary struct {
	Caption                string     `json:"caption"`
	CanonicalMainID        string     `json:"canonicalMainId"`
	CreatorID              string     `json:"creatorId"`
	ID                     string     `json:"id"`
	Media                  mediaAsset `json:"media"`
	PreviewDurationSeconds int64      `json:"previewDurationSeconds"`
}

type feedViewerPayload struct {
	IsFollowingCreator bool `json:"isFollowingCreator"`
	IsPinned           bool `json:"isPinned"`
}

type shortDetailViewer struct {
	IsFollowingCreator bool `json:"isFollowingCreator"`
	IsPinned           bool `json:"isPinned"`
}

type unlockCtaStatePayload struct {
	MainDurationSeconds   *int64 `json:"mainDurationSeconds"`
	PriceJPY              *int64 `json:"priceJpy"`
	ResumePositionSeconds *int64 `json:"resumePositionSeconds"`
	State                 string `json:"state"`
}

type fanFeedCursorPayload struct {
	PublishedAt string `json:"publishedAt"`
	ShortID     string `json:"shortId"`
}

func registerFanFeedRoutes(
	router gin.IRouter,
	reader FanFeedReader,
	shortDisplayAssets ShortDisplayAssetResolver,
	recommendationSignalExposure RecommendationSignalExposureStore,
	viewerBootstrap ViewerBootstrapReader,
) {
	if router == nil || reader == nil || shortDisplayAssets == nil {
		return
	}

	router.GET("/api/fan/feed", func(c *gin.Context) {
		handleFanFeed(c, reader, shortDisplayAssets, recommendationSignalExposure, viewerBootstrap)
	})
	router.GET("/api/fan/shorts/:shortId", func(c *gin.Context) {
		handleFanShortDetail(c, reader, shortDisplayAssets, recommendationSignalExposure, viewerBootstrap)
	})
}

func handleFanFeed(
	c *gin.Context,
	reader FanFeedReader,
	shortDisplayAssets ShortDisplayAssetResolver,
	recommendationSignalExposure RecommendationSignalExposureStore,
	viewerBootstrap ViewerBootstrapReader,
) {
	tab := normalizeFanFeedTab(c.Query("tab"))
	cursor := decodeFanFeedCursor(strings.TrimSpace(c.Query("cursor")))

	viewerUserID, err := resolveOptionalViewerUserID(c, viewerBootstrap)
	if err != nil {
		writeInternalServerError(c, fanFeedRequestScope)
		return
	}

	var (
		items      []feed.Item
		nextCursor *feed.Cursor
	)
	switch tab {
	case "following":
		if viewerUserID == nil {
			writeAuthRequiredError(c, fanFeedRequestScope, fanFeedFollowingAuthRequiredMessage)
			return
		}

		items, nextCursor, err = reader.ListFollowing(c.Request.Context(), *viewerUserID, cursor, feed.DefaultPageSize)
	default:
		items, nextCursor, err = reader.ListRecommended(c.Request.Context(), viewerUserID, cursor, feed.DefaultPageSize)
	}
	if err != nil {
		writeInternalServerError(c, fanFeedRequestScope)
		return
	}

	responseItems := make([]fanFeedItemPayload, 0, len(items))
	for _, item := range items {
		responseItem, buildErr := buildFanFeedItemPayload(item, shortDisplayAssets)
		if buildErr != nil {
			writeInternalServerError(c, fanFeedRequestScope)
			return
		}

		responseItems = append(responseItems, responseItem)
	}

	shortIDs := collectRecommendationFeedShortIDs(items)
	creatorIDs := collectRecommendationFeedCreatorIDs(items)
	rememberRecommendationFeedExposure(c.Request.Context(), recommendationSignalExposure, viewerUserID, shortIDs, creatorIDs)

	c.JSON(http.StatusOK, responseEnvelope[fanFeedResponseData]{
		Data: &fanFeedResponseData{
			Items: responseItems,
			Tab:   tab,
		},
		Meta: responseMeta{
			RequestID: newRequestID(fanFeedRequestScope),
			Page: &cursorPageInfo{
				HasNext:    nextCursor != nil,
				NextCursor: encodeFanFeedCursor(nextCursor),
			},
		},
		Error: nil,
	})
}

func handleFanShortDetail(
	c *gin.Context,
	reader FanFeedReader,
	shortDisplayAssets ShortDisplayAssetResolver,
	recommendationSignalExposure RecommendationSignalExposureStore,
	viewerBootstrap ViewerBootstrapReader,
) {
	shortID, err := shorts.ParsePublicShortID(c.Param("shortId"))
	if err != nil {
		writeNotFoundError(c, fanShortDetailRequestScope, "short was not found")
		return
	}

	viewerUserID, err := resolveOptionalViewerUserID(c, viewerBootstrap)
	if err != nil {
		writeInternalServerError(c, fanShortDetailRequestScope)
		return
	}

	detail, err := reader.GetDetail(c.Request.Context(), shortID, viewerUserID)
	if err != nil {
		if errors.Is(err, feed.ErrPublicShortNotFound) {
			writeNotFoundError(c, fanShortDetailRequestScope, "short was not found")
			return
		}

		writeInternalServerError(c, fanShortDetailRequestScope)
		return
	}

	responseDetail, err := buildShortDetailPayload(detail, shortDisplayAssets)
	if err != nil {
		writeInternalServerError(c, fanShortDetailRequestScope)
		return
	}

	rememberRecommendationShortDetailExposure(c.Request.Context(), recommendationSignalExposure, viewerUserID, detail)

	c.JSON(http.StatusOK, responseEnvelope[shortDetailResponseData]{
		Data: &shortDetailResponseData{
			Detail: responseDetail,
		},
		Meta: responseMeta{
			RequestID: newRequestID(fanShortDetailRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func buildFanFeedItemPayload(item feed.Item, shortDisplayAssets ShortDisplayAssetResolver) (fanFeedItemPayload, error) {
	creator, err := buildFeedCreatorSummary(item)
	if err != nil {
		return fanFeedItemPayload{}, err
	}
	short, err := buildFeedShortSummary(item, shortDisplayAssets)
	if err != nil {
		return fanFeedItemPayload{}, err
	}

	return fanFeedItemPayload{
		Creator:   creator,
		Short:     short,
		UnlockCta: buildUnlockCtaStatePayload(item),
		Viewer: feedViewerPayload{
			IsFollowingCreator: item.Viewer.IsFollowingCreator,
			IsPinned:           item.Viewer.IsPinned,
		},
	}, nil
}

func buildShortDetailPayload(detail feed.Detail, shortDisplayAssets ShortDisplayAssetResolver) (shortDetailPayload, error) {
	creator, err := buildFeedCreatorSummary(detail.Item)
	if err != nil {
		return shortDetailPayload{}, err
	}
	short, err := buildFeedShortSummary(detail.Item, shortDisplayAssets)
	if err != nil {
		return shortDetailPayload{}, err
	}

	return shortDetailPayload{
		Creator:   creator,
		Short:     short,
		UnlockCta: buildUnlockCtaStatePayload(detail.Item),
		Viewer: shortDetailViewer{
			IsFollowingCreator: detail.Viewer.IsFollowingCreator,
			IsPinned:           detail.Item.Viewer.IsPinned,
		},
	}, nil
}

func buildFeedCreatorSummary(item feed.Item) (creatorSummary, error) {
	return buildCreatorSummaryFields(
		item.Creator.ID,
		item.Creator.DisplayName,
		item.Creator.Handle,
		item.Creator.AvatarURL,
		item.Creator.Bio,
	)
}

func buildFeedShortSummary(item feed.Item, shortDisplayAssets ShortDisplayAssetResolver) (feedShortSummary, error) {
	return buildPublicShortSummaryFields(
		item.Short.ID,
		item.Short.CanonicalMainID,
		item.Short.CreatorUserID,
		item.Short.Caption,
		item.Short.MediaAssetID,
		item.Short.PreviewDurationSeconds,
		shortDisplayAssets,
	)
}

func buildVideoMediaAsset(asset media.VideoDisplayAsset) mediaAsset {
	durationSeconds := asset.DurationSeconds
	posterURL := asset.PosterURL

	return mediaAsset{
		DurationSeconds: &durationSeconds,
		ID:              mediaAssetPublicID(asset.ID),
		Kind:            asset.Kind,
		PosterURL:       &posterURL,
		URL:             asset.URL,
	}
}

func rememberRecommendationFeedExposure(
	ctx context.Context,
	recommendationSignalExposure RecommendationSignalExposureStore,
	viewerUserID *uuid.UUID,
	shortIDs []uuid.UUID,
	creatorIDs []uuid.UUID,
) {
	if recommendationSignalExposure == nil || viewerUserID == nil {
		return
	}

	_ = recommendationSignalExposure.RememberShortExposures(ctx, *viewerUserID, shortIDs)
	_ = recommendationSignalExposure.RememberCreatorExposures(ctx, *viewerUserID, creatorIDs)
}

func rememberRecommendationShortDetailExposure(
	ctx context.Context,
	recommendationSignalExposure RecommendationSignalExposureStore,
	viewerUserID *uuid.UUID,
	detail feed.Detail,
) {
	if recommendationSignalExposure == nil || viewerUserID == nil {
		return
	}

	_ = recommendationSignalExposure.RememberShortExposure(ctx, *viewerUserID, detail.Item.Short.ID)
	_ = recommendationSignalExposure.RememberCreatorExposure(ctx, *viewerUserID, detail.Item.Creator.ID)
}

func collectRecommendationFeedShortIDs(items []feed.Item) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.Short.ID)
	}

	return ids
}

func collectRecommendationFeedCreatorIDs(items []feed.Item) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.Creator.ID)
	}

	return ids
}

func buildUnlockCtaStatePayload(item feed.Item) unlockCtaStatePayload {
	switch {
	case item.Unlock.IsOwner:
		return unlockCtaStatePayload{
			State: "owner_preview",
		}
	case item.Unlock.IsUnlocked:
		return unlockCtaStatePayload{
			State: "continue_main",
		}
	default:
		mainDurationSeconds := item.Unlock.MainDurationSeconds
		priceJPY := item.Unlock.PriceJPY

		return unlockCtaStatePayload{
			MainDurationSeconds: &mainDurationSeconds,
			PriceJPY:            &priceJPY,
			State:               "unlock_available",
		}
	}
}

func normalizeFanFeedTab(value string) string {
	if strings.TrimSpace(value) == "following" {
		return "following"
	}

	return "recommended"
}

func decodeFanFeedCursor(encoded string) *feed.Cursor {
	if encoded == "" {
		return nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil
	}

	var payload fanFeedCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil
	}

	publishedAt, err := time.Parse(time.RFC3339Nano, payload.PublishedAt)
	if err != nil {
		return nil
	}

	shortID, err := uuid.Parse(strings.TrimSpace(payload.ShortID))
	if err != nil {
		return nil
	}

	return &feed.Cursor{
		PublishedAt: publishedAt,
		ShortID:     shortID,
	}
}

func encodeFanFeedCursor(cursor *feed.Cursor) *string {
	if cursor == nil {
		return nil
	}

	payload, err := json.Marshal(fanFeedCursorPayload{
		PublishedAt: cursor.PublishedAt.Format(time.RFC3339Nano),
		ShortID:     cursor.ShortID.String(),
	})
	if err != nil {
		return nil
	}

	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return &encoded
}
