package httpserver

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/fanprofile"
)

type fanProfileFollowingResponseData struct {
	Items []fanProfileFollowingItem `json:"items"`
}

type fanProfileFollowingItem struct {
	Creator creatorSummary          `json:"creator"`
	Viewer  fanProfileFollowingView `json:"viewer"`
}

type fanProfileFollowingView struct {
	IsFollowing bool `json:"isFollowing"`
}

type fanProfileFollowingCursorPayload struct {
	CreatorUserID string `json:"creatorUserId"`
	FollowedAt    string `json:"followedAt"`
}

// registerFanProfileRoutes は fan profile private hub API を router に登録します。
func registerFanProfileRoutes(
	router gin.IRouter,
	viewerBootstrapReader ViewerBootstrapReader,
	followingReader FanProfileFollowingReader,
) {
	if router == nil || viewerBootstrapReader == nil || followingReader == nil {
		return
	}

	protected := router.Group("/api/fan/profile")
	protected.Use(buildProtectedFanAuthGuard(viewerBootstrapReader, "fan_profile", "fan profile requires authentication"))
	protected.GET("/following", func(c *gin.Context) {
		handleFanProfileFollowing(c, followingReader)
	})
}

func handleFanProfileFollowing(c *gin.Context, reader FanProfileFollowingReader) {
	viewer, ok := authenticatedViewerFromContext(c)
	if !ok {
		writeInternalServerError(c, "fan_profile_following")
		return
	}

	cursor := decodeFanProfileFollowingCursor(strings.TrimSpace(c.Query("cursor")))
	items, nextCursor, err := reader.ListFollowing(c.Request.Context(), viewer.ID, cursor, fanprofile.DefaultFollowingPageSize)
	if err != nil {
		writeInternalServerError(c, "fan_profile_following")
		return
	}

	responseItems := make([]fanProfileFollowingItem, 0, len(items))
	for _, item := range items {
		responseItem, buildErr := buildFanProfileFollowingItem(item)
		if buildErr != nil {
			writeInternalServerError(c, "fan_profile_following")
			return
		}

		responseItems = append(responseItems, responseItem)
	}

	c.JSON(http.StatusOK, responseEnvelope[fanProfileFollowingResponseData]{
		Data: &fanProfileFollowingResponseData{
			Items: responseItems,
		},
		Meta: responseMeta{
			RequestID: newRequestID("fan_profile_following"),
			Page: &cursorPageInfo{
				HasNext:    nextCursor != nil,
				NextCursor: encodeFanProfileFollowingCursor(nextCursor),
			},
		},
		Error: nil,
	})
}

func buildFanProfileFollowingItem(item fanprofile.FollowingItem) (fanProfileFollowingItem, error) {
	creator, err := buildCreatorSummaryFields(item.CreatorUserID, item.DisplayName, item.Handle, item.AvatarURL, item.Bio)
	if err != nil {
		return fanProfileFollowingItem{}, err
	}

	return fanProfileFollowingItem{
		Creator: creator,
		Viewer: fanProfileFollowingView{
			IsFollowing: true,
		},
	}, nil
}

func decodeFanProfileFollowingCursor(encoded string) *fanprofile.FollowingCursor {
	if encoded == "" {
		return nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil
	}

	var payload fanProfileFollowingCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil
	}

	followedAt, err := time.Parse(time.RFC3339Nano, payload.FollowedAt)
	if err != nil {
		return nil
	}

	creatorUserID, err := uuid.Parse(strings.TrimSpace(payload.CreatorUserID))
	if err != nil {
		return nil
	}

	return &fanprofile.FollowingCursor{
		CreatorUserID: creatorUserID,
		FollowedAt:    followedAt,
	}
}

func encodeFanProfileFollowingCursor(cursor *fanprofile.FollowingCursor) *string {
	if cursor == nil {
		return nil
	}

	payload, err := json.Marshal(fanProfileFollowingCursorPayload{
		CreatorUserID: cursor.CreatorUserID.String(),
		FollowedAt:    cursor.FollowedAt.Format(time.RFC3339Nano),
	})
	if err != nil {
		return nil
	}

	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return &encoded
}
