package httpserver

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/fanprofile"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const fanProfilePinnedShortsRequestScope = "fan_profile_pinned_shorts"

type fanProfilePinnedShortsResponseData struct {
	Items []fanProfilePinnedShortItem `json:"items"`
}

type fanProfilePinnedShortItem struct {
	Creator creatorSummary   `json:"creator"`
	Short   feedShortSummary `json:"short"`
}

type fanProfilePinnedShortCursorPayload struct {
	PinnedAt string `json:"pinnedAt"`
	ShortID  string `json:"shortId"`
}

func handleFanProfilePinnedShorts(
	c *gin.Context,
	reader FanProfilePinnedShortsReader,
	shortDisplayAssets ShortDisplayAssetResolver,
) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		writeInternalServerError(c, fanProfilePinnedShortsRequestScope)
		return
	}

	cursor := decodeFanProfilePinnedShortsCursor(strings.TrimSpace(c.Query("cursor")))
	items, nextCursor, err := reader.ListPinnedShorts(c.Request.Context(), viewer.ID, cursor, fanprofile.DefaultPinnedShortsPageSize)
	if err != nil {
		writeInternalServerError(c, fanProfilePinnedShortsRequestScope)
		return
	}

	responseItems := make([]fanProfilePinnedShortItem, 0, len(items))
	for _, item := range items {
		responseItem, buildErr := buildFanProfilePinnedShortItem(item, shortDisplayAssets)
		if buildErr != nil {
			writeInternalServerError(c, fanProfilePinnedShortsRequestScope)
			return
		}

		responseItems = append(responseItems, responseItem)
	}

	c.JSON(http.StatusOK, responseEnvelope[fanProfilePinnedShortsResponseData]{
		Data: &fanProfilePinnedShortsResponseData{
			Items: responseItems,
		},
		Meta: responseMeta{
			RequestID: newRequestID(fanProfilePinnedShortsRequestScope),
			Page: &cursorPageInfo{
				HasNext:    nextCursor != nil,
				NextCursor: encodeFanProfilePinnedShortsCursor(nextCursor),
			},
		},
		Error: nil,
	})
}

func buildFanProfilePinnedShortItem(
	item fanprofile.PinnedShortItem,
	shortDisplayAssets ShortDisplayAssetResolver,
) (fanProfilePinnedShortItem, error) {
	if shortDisplayAssets == nil {
		return fanProfilePinnedShortItem{}, errors.New("short display asset resolver is required")
	}

	creator, err := buildCreatorSummaryFields(
		item.CreatorUserID,
		item.CreatorDisplayName,
		item.CreatorHandle,
		item.CreatorAvatarURL,
		item.CreatorBio,
	)
	if err != nil {
		return fanProfilePinnedShortItem{}, err
	}

	displayAsset, err := shortDisplayAssets.ResolveShortDisplayAsset(media.ShortDisplaySource{
		AssetID:    item.ShortMediaAssetID,
		ShortID:    item.ShortID,
		DurationMS: item.ShortPreviewDurationSeconds * 1000,
	}, media.AccessBoundaryPublic)
	if err != nil {
		return fanProfilePinnedShortItem{}, err
	}

	return fanProfilePinnedShortItem{
		Creator: creator,
		Short: feedShortSummary{
			Caption:                item.ShortCaption,
			CanonicalMainID:        shorts.FormatPublicMainID(item.ShortCanonicalMainID),
			CreatorID:              creator.ID,
			ID:                     shorts.FormatPublicShortID(item.ShortID),
			Media:                  buildVideoMediaAsset(displayAsset),
			PreviewDurationSeconds: item.ShortPreviewDurationSeconds,
		},
	}, nil
}

func decodeFanProfilePinnedShortsCursor(encoded string) *fanprofile.PinnedShortCursor {
	if encoded == "" {
		return nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil
	}

	var payload fanProfilePinnedShortCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil
	}

	pinnedAt, err := time.Parse(time.RFC3339Nano, payload.PinnedAt)
	if err != nil {
		return nil
	}

	shortID, err := uuid.Parse(strings.TrimSpace(payload.ShortID))
	if err != nil {
		return nil
	}

	return &fanprofile.PinnedShortCursor{
		PinnedAt: pinnedAt,
		ShortID:  shortID,
	}
}

func encodeFanProfilePinnedShortsCursor(cursor *fanprofile.PinnedShortCursor) *string {
	if cursor == nil {
		return nil
	}

	payload, err := json.Marshal(fanProfilePinnedShortCursorPayload{
		PinnedAt: cursor.PinnedAt.Format(time.RFC3339Nano),
		ShortID:  cursor.ShortID.String(),
	})
	if err != nil {
		return nil
	}

	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return &encoded
}
