package httpserver

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shortcomment"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	fanShortCommentCreateAuthRequiredMessage = "short comment requires authentication"
	fanShortCommentCreateRequestScope        = "fan_short_comment_create"
	fanShortCommentListRequestScope          = "fan_short_comment_list"
	maxFanShortCommentCursorLength           = 512
)

type shortCommentListResponseData struct {
	Items []shortCommentPayload `json:"items"`
}

type shortCommentCreateResponseData struct {
	Comment shortCommentPayload `json:"comment"`
}

type shortCommentPayload struct {
	Author    shortCommentAuthorPayload `json:"author"`
	Body      string                    `json:"body"`
	CreatedAt string                    `json:"createdAt"`
	ID        string                    `json:"id"`
	ShortID   string                    `json:"shortId"`
}

type shortCommentAuthorPayload struct {
	Avatar      *mediaAsset `json:"avatar"`
	DisplayName string      `json:"displayName"`
	Handle      string      `json:"handle"`
}

type shortCommentCreateRequest struct {
	Body string `json:"body"`
}

type shortCommentCursorPayload struct {
	CommentID string `json:"commentId"`
	CreatedAt string `json:"createdAt"`
	ShortID   string `json:"shortId"`
}

func registerFanShortCommentRoutes(
	router gin.IRouter,
	reader FanShortCommentReader,
	writer FanShortCommentWriter,
	viewerBootstrap ViewerBootstrapReader,
) {
	if router == nil {
		return
	}

	if reader != nil {
		router.GET("/api/fan/shorts/:shortId/comments", func(c *gin.Context) {
			handleFanShortCommentList(c, reader)
		})
	}

	if writer != nil && viewerBootstrap != nil {
		router.POST(
			"/api/fan/shorts/:shortId/comments",
			buildProtectedFanAuthGuard(viewerBootstrap, fanShortCommentCreateRequestScope, fanShortCommentCreateAuthRequiredMessage),
			func(c *gin.Context) {
				handleFanShortCommentCreate(c, writer)
			},
		)
	}
}

func handleFanShortCommentList(c *gin.Context, reader FanShortCommentReader) {
	shortID, err := shorts.ParsePublicShortID(c.Param("shortId"))
	if err != nil {
		writeFanShortCommentError(c, http.StatusNotFound, "not_found", "short was not found", fanShortCommentListRequestScope)
		return
	}

	cursor, ok := decodeShortCommentCursor(strings.TrimSpace(c.Query("cursor")), shortID)
	if !ok {
		writeFanShortCommentError(c, http.StatusBadRequest, "invalid_request", "short comment cursor was invalid", fanShortCommentListRequestScope)
		return
	}

	comments, nextCursor, err := reader.ListComments(c.Request.Context(), shortID, cursor, shortcomment.DefaultPageSize)
	if err != nil {
		if errors.Is(err, shortcomment.ErrShortNotFound) {
			writeFanShortCommentError(c, http.StatusNotFound, "not_found", "short was not found", fanShortCommentListRequestScope)
			return
		}

		writeInternalServerError(c, fanShortCommentListRequestScope)
		return
	}

	items, err := buildShortCommentPayloads(comments)
	if err != nil {
		writeInternalServerError(c, fanShortCommentListRequestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[shortCommentListResponseData]{
		Data: &shortCommentListResponseData{
			Items: items,
		},
		Meta: responseMeta{
			RequestID: newRequestID(fanShortCommentListRequestScope),
			Page: &cursorPageInfo{
				HasNext:    nextCursor != nil,
				NextCursor: encodeShortCommentCursor(shortID, nextCursor),
			},
		},
		Error: nil,
	})
}

func handleFanShortCommentCreate(c *gin.Context, writer FanShortCommentWriter) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeInternalServerError(c, fanShortCommentCreateRequestScope)
		return
	}

	shortID, err := shorts.ParsePublicShortID(c.Param("shortId"))
	if err != nil {
		writeFanShortCommentError(c, http.StatusNotFound, "not_found", "short was not found", fanShortCommentCreateRequestScope)
		return
	}

	var request shortCommentCreateRequest
	if err := decodeLimitedJSONBody(c, &request, true); err != nil {
		if isShortCommentBodyTypeError(err) {
			writeFanShortCommentError(c, http.StatusBadRequest, "validation_error", "comment body is invalid", fanShortCommentCreateRequestScope)
			return
		}
		writeFanShortCommentError(c, http.StatusBadRequest, "invalid_request", "short comment request was invalid", fanShortCommentCreateRequestScope)
		return
	}

	comment, err := writer.CreateComment(c.Request.Context(), shortcomment.CreateInput{
		AuthorUserID: viewerUserID,
		Body:         request.Body,
		ShortID:      shortID,
	})
	if err != nil {
		switch {
		case errors.Is(err, shortcomment.ErrInvalidCommentBody):
			writeFanShortCommentError(c, http.StatusBadRequest, "validation_error", "comment body is invalid", fanShortCommentCreateRequestScope)
			return
		case errors.Is(err, shortcomment.ErrShortNotFound):
			writeFanShortCommentError(c, http.StatusNotFound, "not_found", "short was not found", fanShortCommentCreateRequestScope)
			return
		default:
			writeInternalServerError(c, fanShortCommentCreateRequestScope)
			return
		}
	}

	payload, err := buildShortCommentPayload(comment)
	if err != nil {
		writeInternalServerError(c, fanShortCommentCreateRequestScope)
		return
	}

	c.JSON(http.StatusCreated, responseEnvelope[shortCommentCreateResponseData]{
		Data: &shortCommentCreateResponseData{
			Comment: payload,
		},
		Meta: responseMeta{
			RequestID: newRequestID(fanShortCommentCreateRequestScope),
			Page:      nil,
		},
		Error: nil,
	})
}

func buildShortCommentPayloads(comments []shortcomment.Comment) ([]shortCommentPayload, error) {
	items := make([]shortCommentPayload, 0, len(comments))
	for _, comment := range comments {
		item, err := buildShortCommentPayload(comment)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func buildShortCommentPayload(comment shortcomment.Comment) (shortCommentPayload, error) {
	author, err := buildShortCommentAuthorPayload(comment.Author)
	if err != nil {
		return shortCommentPayload{}, err
	}

	return shortCommentPayload{
		Author:    author,
		Body:      comment.Body,
		CreatedAt: comment.CreatedAt.UTC().Format(time.RFC3339Nano),
		ID:        shortCommentPublicID(comment.ID),
		ShortID:   shortPublicID(comment.ShortID),
	}, nil
}

func buildShortCommentAuthorPayload(author shortcomment.Author) (shortCommentAuthorPayload, error) {
	displayName := strings.TrimSpace(author.DisplayName)
	handle := strings.TrimSpace(author.Handle)
	if displayName == "" || handle == "" {
		return shortCommentAuthorPayload{}, fmt.Errorf("short comment author display name または handle がありません")
	}

	var avatar *mediaAsset
	if author.AvatarURL != nil && strings.TrimSpace(*author.AvatarURL) != "" {
		avatarURL := strings.TrimSpace(*author.AvatarURL)
		avatar = &mediaAsset{
			DurationSeconds: nil,
			ID:              shortCommentAuthorAvatarAssetID(avatarURL),
			Kind:            "image",
			PosterURL:       nil,
			URL:             avatarURL,
		}
	}

	return shortCommentAuthorPayload{
		Avatar:      avatar,
		DisplayName: displayName,
		Handle:      formatHandle(handle),
	}, nil
}

func decodeShortCommentCursor(encoded string, routeShortID uuid.UUID) (*shortcomment.Cursor, bool) {
	if encoded == "" {
		return nil, true
	}
	if len(encoded) > maxFanShortCommentCursorLength {
		return nil, false
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, false
	}

	var payload shortCommentCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, false
	}

	shortID, err := uuid.Parse(strings.TrimSpace(payload.ShortID))
	if err != nil || shortID != routeShortID {
		return nil, false
	}

	commentID, err := uuid.Parse(strings.TrimSpace(payload.CommentID))
	if err != nil {
		return nil, false
	}
	createdAt, err := time.Parse(time.RFC3339Nano, payload.CreatedAt)
	if err != nil {
		return nil, false
	}

	return &shortcomment.Cursor{
		CommentID: commentID,
		CreatedAt: createdAt,
	}, true
}

func encodeShortCommentCursor(shortID uuid.UUID, cursor *shortcomment.Cursor) *string {
	if cursor == nil {
		return nil
	}

	payload, err := json.Marshal(shortCommentCursorPayload{
		CommentID: cursor.CommentID.String(),
		CreatedAt: cursor.CreatedAt.UTC().Format(time.RFC3339Nano),
		ShortID:   shortID.String(),
	})
	if err != nil {
		return nil
	}

	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return &encoded
}

func shortCommentPublicID(commentID uuid.UUID) string {
	return fmt.Sprintf("comment_%s", strings.ReplaceAll(commentID.String(), "-", ""))
}

func shortCommentAuthorAvatarAssetID(avatarURL string) string {
	digest := sha256.Sum256([]byte(avatarURL))
	return fmt.Sprintf("asset_short_comment_author_avatar_%s", hex.EncodeToString(digest[:8]))
}

func isShortCommentBodyTypeError(err error) bool {
	var typeErr *json.UnmarshalTypeError
	if !errors.As(err, &typeErr) {
		return false
	}

	return typeErr.Field == "body" || typeErr.Field == "Body"
}

func writeFanShortCommentError(c *gin.Context, status int, code string, message string, requestScope string) {
	c.JSON(status, responseEnvelope[struct{}]{
		Data: nil,
		Meta: responseMeta{
			RequestID: newRequestID(requestScope),
			Page:      nil,
		},
		Error: &responseError{
			Code:    code,
			Message: message,
		},
	})
}
