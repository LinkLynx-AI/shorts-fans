package httpserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
)

const creatorSearchPageSize = 20

type creatorSearchCursorPayload struct {
	Handle      string `json:"handle"`
	PublishedAt string `json:"publishedAt"`
}

type responseEnvelope[T any] struct {
	Data  *T             `json:"data"`
	Meta  responseMeta   `json:"meta"`
	Error *responseError `json:"error"`
}

type responseMeta struct {
	Page      *cursorPageInfo `json:"page"`
	RequestID string          `json:"requestId"`
}

type cursorPageInfo struct {
	HasNext    bool    `json:"hasNext"`
	NextCursor *string `json:"nextCursor"`
}

type responseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type creatorSearchResponseData struct {
	Items []creatorSearchResult `json:"items"`
	Query string                `json:"query"`
}

type creatorSearchResult struct {
	Creator creatorSummary `json:"creator"`
}

type creatorSummary struct {
	Avatar      *mediaAsset `json:"avatar"`
	Bio         string      `json:"bio"`
	DisplayName string      `json:"displayName"`
	Handle      string      `json:"handle"`
	ID          string      `json:"id"`
}

type mediaAsset struct {
	DurationSeconds *int64  `json:"durationSeconds"`
	ID              string  `json:"id"`
	Kind            string  `json:"kind"`
	PosterURL       *string `json:"posterUrl"`
	URL             string  `json:"url"`
}

// registerCreatorSearchRoutes は creator search API を router に登録します。
func registerCreatorSearchRoutes(router gin.IRouter, reader CreatorSearchReader) {
	if reader == nil {
		return
	}

	router.GET("/api/fan/creators/search", func(c *gin.Context) {
		handleCreatorSearch(c, reader)
	})
}

// handleCreatorSearch は query の有無に応じて recent 一覧または filtered search を返します。
func handleCreatorSearch(c *gin.Context, reader CreatorSearchReader) {
	query := strings.TrimSpace(c.Query("q"))
	cursor := decodeCreatorSearchCursor(strings.TrimSpace(c.Query("cursor")))

	var (
		items      []creator.Profile
		nextCursor *creator.PublicProfileCursor
		err        error
	)
	if query == "" {
		items, nextCursor, err = reader.ListRecentPublicProfiles(c.Request.Context(), cursor, creatorSearchPageSize)
	} else {
		items, nextCursor, err = reader.SearchPublicProfiles(c.Request.Context(), query, cursor, creatorSearchPageSize)
	}
	if err != nil {
		writeInternalServerError(c, "creator_search")
		return
	}

	results := make([]creatorSearchResult, 0, len(items))
	for _, item := range items {
		result, buildErr := buildCreatorSearchResult(item)
		if buildErr != nil {
			writeInternalServerError(c, "creator_search")
			return
		}
		results = append(results, result)
	}

	encodedCursor := encodeCreatorSearchCursor(nextCursor)
	c.JSON(http.StatusOK, responseEnvelope[creatorSearchResponseData]{
		Data: &creatorSearchResponseData{
			Items: results,
			Query: query,
		},
		Meta: responseMeta{
			RequestID: newRequestID("creator_search"),
			Page: &cursorPageInfo{
				HasNext:    nextCursor != nil,
				NextCursor: encodedCursor,
			},
		},
		Error: nil,
	})
}

// buildCreatorSearchResult は公開 creator profile を transport response に変換します。
func buildCreatorSearchResult(profile creator.Profile) (creatorSearchResult, error) {
	creatorSummary, err := buildCreatorSummary(profile)
	if err != nil {
		return creatorSearchResult{}, err
	}

	return creatorSearchResult{
		Creator: creatorSummary,
	}, nil
}

func buildCreatorSummary(profile creator.Profile) (creatorSummary, error) {
	if profile.DisplayName == nil || profile.Handle == nil {
		return creatorSummary{}, fmt.Errorf("creator summary に必要な display name または handle がありません")
	}

	return buildCreatorSummaryFields(profile.UserID, *profile.DisplayName, *profile.Handle, profile.AvatarURL, profile.Bio)
}

func buildCreatorSummaryFields(userID uuid.UUID, displayName string, handle string, avatarURL *string, bio string) (creatorSummary, error) {
	trimmedDisplayName := strings.TrimSpace(displayName)
	trimmedHandle := strings.TrimSpace(handle)
	if trimmedDisplayName == "" || trimmedHandle == "" {
		return creatorSummary{}, fmt.Errorf("creator summary に必要な display name または handle がありません")
	}

	var avatar *mediaAsset
	if avatarURL != nil && strings.TrimSpace(*avatarURL) != "" {
		avatar = &mediaAsset{
			DurationSeconds: nil,
			ID:              creatorAvatarAssetID(userID),
			Kind:            "image",
			PosterURL:       nil,
			URL:             strings.TrimSpace(*avatarURL),
		}
	}

	return creatorSummary{
		Avatar:      avatar,
		Bio:         bio,
		DisplayName: trimmedDisplayName,
		Handle:      formatHandle(trimmedHandle),
		ID:          creator.FormatPublicID(userID),
	}, nil
}

// decodeCreatorSearchCursor は base64url 文字列から keyset cursor を復元します。
func decodeCreatorSearchCursor(encoded string) *creator.PublicProfileCursor {
	if encoded == "" {
		return nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil
	}

	var payload creatorSearchCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil
	}

	publishedAt, err := time.Parse(time.RFC3339Nano, payload.PublishedAt)
	if err != nil || strings.TrimSpace(payload.Handle) == "" {
		return nil
	}

	return &creator.PublicProfileCursor{
		PublishedAt: publishedAt,
		Handle:      strings.TrimSpace(payload.Handle),
	}
}

// encodeCreatorSearchCursor は keyset cursor を response 用の base64url 文字列に変換します。
func encodeCreatorSearchCursor(cursor *creator.PublicProfileCursor) *string {
	if cursor == nil {
		return nil
	}

	payload, err := json.Marshal(creatorSearchCursorPayload{
		Handle:      cursor.Handle,
		PublishedAt: cursor.PublishedAt.Format(time.RFC3339Nano),
	})
	if err != nil {
		return nil
	}

	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return &encoded
}

// newRequestID は endpoint scope 付きの request identifier を生成します。
func newRequestID(prefix string) string {
	return fmt.Sprintf("req_%s_%s", prefix, strings.ReplaceAll(uuid.NewString(), "-", ""))
}

// writeInternalServerError は fan read surface 共通の internal error envelope を返します。
func writeInternalServerError(c *gin.Context, requestScope string) {
	c.JSON(http.StatusInternalServerError, responseEnvelope[struct{}]{
		Data: nil,
		Meta: responseMeta{
			RequestID: newRequestID(requestScope),
			Page:      nil,
		},
		Error: &responseError{
			Code:    "internal_error",
			Message: "internal server error",
		},
	})
}

// writeNotFoundError は fan read surface 共通の not found envelope を返します。
func writeNotFoundError(c *gin.Context, requestScope string, message string) {
	c.JSON(http.StatusNotFound, responseEnvelope[struct{}]{
		Data: nil,
		Meta: responseMeta{
			RequestID: newRequestID(requestScope),
			Page:      nil,
		},
		Error: &responseError{
			Code:    "not_found",
			Message: message,
		},
	})
}

// formatHandle は保存済み handle を API 向けの `@` 付き表現に揃えます。
func formatHandle(handle string) string {
	return "@" + strings.TrimPrefix(handle, "@")
}

// creatorAvatarAssetID は user ID から stable な avatar asset identifier を生成します。
func creatorAvatarAssetID(userID uuid.UUID) string {
	return fmt.Sprintf("asset_creator_%s_avatar", strings.ReplaceAll(userID.String(), "-", ""))
}
