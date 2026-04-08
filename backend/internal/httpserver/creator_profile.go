package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	creatorFollowDeleteAuthRequiredMessage = "creator unfollow requires authentication"
	creatorFollowDeleteRequestScope        = "creator_follow_delete"
	creatorFollowPutAuthRequiredMessage    = "creator follow requires authentication"
	creatorFollowPutRequestScope           = "creator_follow_put"
)

type creatorProfileResponseData struct {
	Profile creatorProfilePayload `json:"profile"`
}

type creatorProfilePayload struct {
	Creator creatorSummary       `json:"creator"`
	Stats   creatorProfileStats  `json:"stats"`
	Viewer  creatorProfileViewer `json:"viewer"`
}

type creatorProfileStats struct {
	FanCount   int64 `json:"fanCount"`
	ShortCount int64 `json:"shortCount"`
}

type creatorProfileViewer struct {
	IsFollowing bool `json:"isFollowing"`
}

type creatorProfileShortGridResponseData struct {
	Items []creatorProfileShortGridItem `json:"items"`
}

type creatorFollowResponseData struct {
	Stats  creatorFollowStats   `json:"stats"`
	Viewer creatorProfileViewer `json:"viewer"`
}

type creatorFollowStats struct {
	FanCount int64 `json:"fanCount"`
}

type creatorProfileShortGridItem struct {
	CanonicalMainID        string     `json:"canonicalMainId"`
	CreatorID              string     `json:"creatorId"`
	ID                     string     `json:"id"`
	Media                  mediaAsset `json:"media"`
	PreviewDurationSeconds int64      `json:"previewDurationSeconds"`
}

type creatorProfileShortGridCursorPayload struct {
	PublishedAt string `json:"publishedAt"`
	ShortID     string `json:"shortId"`
}

// registerCreatorProfileRoutes は creator profile API を router に登録します。
func registerCreatorProfileRoutes(
	router gin.IRouter,
	profileReader CreatorProfileReader,
	shortsReader CreatorProfileShortsReader,
	followWriter CreatorFollowWriter,
	viewerBootstrap ViewerBootstrapReader,
) {
	if profileReader != nil {
		router.GET("/api/fan/creators/:creatorId", func(c *gin.Context) {
			handleCreatorProfile(c, profileReader, viewerBootstrap)
		})
	}
	if shortsReader != nil {
		router.GET("/api/fan/creators/:creatorId/shorts", func(c *gin.Context) {
			handleCreatorProfileShorts(c, shortsReader)
		})
	}
	if followWriter != nil && viewerBootstrap != nil {
		router.PUT(
			"/api/fan/creators/:creatorId/follow",
			buildProtectedFanAuthGuard(viewerBootstrap, creatorFollowPutRequestScope, creatorFollowPutAuthRequiredMessage),
			func(c *gin.Context) {
				handleCreatorFollowPut(c, followWriter)
			},
		)
		router.DELETE(
			"/api/fan/creators/:creatorId/follow",
			buildProtectedFanAuthGuard(viewerBootstrap, creatorFollowDeleteRequestScope, creatorFollowDeleteAuthRequiredMessage),
			func(c *gin.Context) {
				handleCreatorFollowDelete(c, followWriter)
			},
		)
	}
}

// handleCreatorProfile は creator profile header を返します。
func handleCreatorProfile(c *gin.Context, reader CreatorProfileReader, viewerBootstrap ViewerBootstrapReader) {
	creatorID := strings.TrimSpace(c.Param("creatorId"))
	viewerUserID, err := resolveOptionalViewerUserID(c, viewerBootstrap)
	if err != nil {
		writeInternalServerError(c, "creator_profile")
		return
	}
	profileHeader, err := reader.GetPublicProfileHeader(c.Request.Context(), creatorID, viewerUserID)
	if err != nil {
		if errors.Is(err, creator.ErrProfileNotFound) {
			writeNotFoundError(c, "creator_profile", "creator was not found")
			return
		}

		writeInternalServerError(c, "creator_profile")
		return
	}

	creatorSummary, err := buildCreatorSummary(profileHeader.Profile)
	if err != nil {
		writeInternalServerError(c, "creator_profile")
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[creatorProfileResponseData]{
		Data: &creatorProfileResponseData{
			Profile: creatorProfilePayload{
				Creator: creatorSummary,
				Stats: creatorProfileStats{
					FanCount:   profileHeader.FanCount,
					ShortCount: profileHeader.ShortCount,
				},
				Viewer: creatorProfileViewer{
					IsFollowing: profileHeader.IsFollowing,
				},
			},
		},
		Meta: responseMeta{
			RequestID: newRequestID("creator_profile"),
			Page:      nil,
		},
		Error: nil,
	})
}

func handleCreatorFollowPut(c *gin.Context, writer CreatorFollowWriter) {
	handleCreatorFollowMutation(c, writer.FollowPublicCreator, creatorFollowPutRequestScope)
}

func handleCreatorFollowDelete(c *gin.Context, writer CreatorFollowWriter) {
	handleCreatorFollowMutation(c, writer.UnfollowPublicCreator, creatorFollowDeleteRequestScope)
}

func handleCreatorFollowMutation(
	c *gin.Context,
	mutate func(context.Context, uuid.UUID, string) (creator.FollowMutationResult, error),
	requestScope string,
) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeInternalServerError(c, requestScope)
		return
	}

	creatorID := strings.TrimSpace(c.Param("creatorId"))
	result, err := mutate(c.Request.Context(), viewerUserID, creatorID)
	if err != nil {
		if errors.Is(err, creator.ErrProfileNotFound) {
			writeNotFoundError(c, requestScope, "creator was not found")
			return
		}

		writeInternalServerError(c, requestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[creatorFollowResponseData]{
		Data: &creatorFollowResponseData{
			Stats: creatorFollowStats{
				FanCount: result.FanCount,
			},
			Viewer: creatorProfileViewer{
				IsFollowing: result.IsFollowing,
			},
		},
		Meta: responseMeta{
			RequestID: newRequestID(requestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

// handleCreatorProfileShorts は creator profile short grid を返します。
func handleCreatorProfileShorts(c *gin.Context, reader CreatorProfileShortsReader) {
	creatorID := strings.TrimSpace(c.Param("creatorId"))
	cursor := decodeCreatorProfileShortGridCursor(strings.TrimSpace(c.Query("cursor")))

	items, nextCursor, err := reader.ListPublicProfileShorts(c.Request.Context(), creatorID, cursor, creator.DefaultPublicProfileShortGridPageSize)
	if err != nil {
		if errors.Is(err, creator.ErrProfileNotFound) {
			writeNotFoundError(c, "creator_profile_shorts", "creator was not found")
			return
		}

		writeInternalServerError(c, "creator_profile_shorts")
		return
	}

	responseItems := make([]creatorProfileShortGridItem, 0, len(items))
	for _, item := range items {
		responseItem, buildErr := buildCreatorProfileShortGridItem(item)
		if buildErr != nil {
			writeInternalServerError(c, "creator_profile_shorts")
			return
		}
		responseItems = append(responseItems, responseItem)
	}

	c.JSON(http.StatusOK, responseEnvelope[creatorProfileShortGridResponseData]{
		Data: &creatorProfileShortGridResponseData{
			Items: responseItems,
		},
		Meta: responseMeta{
			RequestID: newRequestID("creator_profile_shorts"),
			Page: &cursorPageInfo{
				HasNext:    nextCursor != nil,
				NextCursor: encodeCreatorProfileShortGridCursor(nextCursor),
			},
		},
		Error: nil,
	})
}

func buildCreatorProfileShortGridItem(item creator.PublicProfileShort) (creatorProfileShortGridItem, error) {
	if strings.TrimSpace(item.MediaURL) == "" {
		return creatorProfileShortGridItem{}, fmt.Errorf("creator profile short grid item に必要な media url がありません")
	}

	durationSeconds := item.PreviewDurationSeconds

	return creatorProfileShortGridItem{
		CanonicalMainID: mainPublicID(item.CanonicalMainID),
		CreatorID:       creator.FormatPublicID(item.CreatorUserID),
		ID:              shortPublicID(item.ID),
		Media: mediaAsset{
			DurationSeconds: &durationSeconds,
			ID:              mediaAssetPublicID(item.MediaAssetID),
			Kind:            "video",
			PosterURL:       nil,
			URL:             item.MediaURL,
		},
		PreviewDurationSeconds: item.PreviewDurationSeconds,
	}, nil
}

func decodeCreatorProfileShortGridCursor(encoded string) *creator.PublicProfileShortCursor {
	if encoded == "" {
		return nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil
	}

	var payload creatorProfileShortGridCursorPayload
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

	return &creator.PublicProfileShortCursor{
		PublishedAt: publishedAt,
		ShortID:     shortID,
	}
}

func encodeCreatorProfileShortGridCursor(cursor *creator.PublicProfileShortCursor) *string {
	if cursor == nil {
		return nil
	}

	payload, err := json.Marshal(creatorProfileShortGridCursorPayload{
		PublishedAt: cursor.PublishedAt.Format(time.RFC3339Nano),
		ShortID:     cursor.ShortID.String(),
	})
	if err != nil {
		// Strings-only payload should always marshal; nil omits nextCursor if something unexpected happens.
		return nil
	}

	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return &encoded
}

func shortPublicID(shortID uuid.UUID) string {
	return fmt.Sprintf("short_%s", strings.ReplaceAll(shortID.String(), "-", ""))
}

func mainPublicID(mainID uuid.UUID) string {
	return fmt.Sprintf("main_%s", strings.ReplaceAll(mainID.String(), "-", ""))
}

func mediaAssetPublicID(mediaAssetID uuid.UUID) string {
	return fmt.Sprintf("asset_%s", strings.ReplaceAll(mediaAssetID.String(), "-", ""))
}

func resolveOptionalViewerUserID(c *gin.Context, viewerBootstrap ViewerBootstrapReader) (*uuid.UUID, error) {
	if viewerBootstrap == nil {
		return nil, nil
	}

	rawSessionToken, err := c.Cookie(auth.SessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, nil
		}

		return nil, err
	}

	bootstrap, err := viewerBootstrap.ReadCurrentViewer(c.Request.Context(), rawSessionToken)
	if err != nil {
		return nil, err
	}
	if bootstrap.CurrentViewer == nil {
		return nil, nil
	}

	viewerUserID := bootstrap.CurrentViewer.ID
	return &viewerUserID, nil
}
