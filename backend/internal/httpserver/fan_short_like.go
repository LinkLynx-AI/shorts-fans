package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	fanShortLikeAuthRequiredMessage = "short like requires authentication"
	fanShortLikeDeleteRequestScope  = "fan_short_like_delete"
	fanShortLikePutRequestScope     = "fan_short_like_put"
)

type fanShortLikeResponseData struct {
	Viewer     fanShortLikeViewerPayload `json:"viewer"`
	Engagement shortEngagementPayload    `json:"engagement"`
}

type fanShortLikeViewerPayload struct {
	HasLiked bool `json:"hasLiked"`
}

func registerFanShortLikeRoutes(
	router gin.IRouter,
	writer FanShortLikeWriter,
	viewerBootstrap ViewerBootstrapReader,
) {
	if router == nil || writer == nil || viewerBootstrap == nil {
		return
	}

	router.PUT(
		"/api/fan/shorts/:shortId/like",
		buildProtectedFanAuthGuard(viewerBootstrap, fanShortLikePutRequestScope, fanShortLikeAuthRequiredMessage),
		func(c *gin.Context) {
			handleFanShortLikePut(c, writer)
		},
	)
	router.DELETE(
		"/api/fan/shorts/:shortId/like",
		buildProtectedFanAuthGuard(viewerBootstrap, fanShortLikeDeleteRequestScope, fanShortLikeAuthRequiredMessage),
		func(c *gin.Context) {
			handleFanShortLikeDelete(c, writer)
		},
	)
}

func handleFanShortLikePut(c *gin.Context, writer FanShortLikeWriter) {
	handleFanShortLikeMutation(c, writer.LikePublicShort, fanShortLikePutRequestScope)
}

func handleFanShortLikeDelete(c *gin.Context, writer FanShortLikeWriter) {
	handleFanShortLikeMutation(c, writer.UnlikePublicShort, fanShortLikeDeleteRequestScope)
}

func handleFanShortLikeMutation(
	c *gin.Context,
	mutate func(context.Context, uuid.UUID, uuid.UUID) (shorts.LikeMutationResult, error),
	requestScope string,
) {
	viewerUserID, ok := authenticatedViewerIDFromContext(c)
	if !ok {
		writeInternalServerError(c, requestScope)
		return
	}

	shortID, err := shorts.ParsePublicShortID(c.Param("shortId"))
	if err != nil {
		writeNotFoundError(c, requestScope, "short was not found")
		return
	}

	result, err := mutate(c.Request.Context(), viewerUserID, shortID)
	if err != nil {
		if errors.Is(err, shorts.ErrShortNotFound) {
			writeNotFoundError(c, requestScope, "short was not found")
			return
		}

		writeInternalServerError(c, requestScope)
		return
	}

	c.JSON(http.StatusOK, responseEnvelope[fanShortLikeResponseData]{
		Data: &fanShortLikeResponseData{
			Viewer: fanShortLikeViewerPayload{
				HasLiked: result.HasLiked,
			},
			Engagement: shortEngagementPayload{
				LikeCount: result.LikeCount,
			},
		},
		Meta: responseMeta{
			RequestID: newRequestID(requestScope),
			Page:      nil,
		},
		Error: nil,
	})
}
