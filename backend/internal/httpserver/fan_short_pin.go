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
	fanShortPinAuthRequiredMessage = "short pin requires authentication"
	fanShortPinDeleteRequestScope  = "fan_short_pin_delete"
	fanShortPinPutRequestScope     = "fan_short_pin_put"
)

type fanShortPinResponseData struct {
	Viewer feedViewerPayload `json:"viewer"`
}

func registerFanShortPinRoutes(
	router gin.IRouter,
	writer FanShortPinWriter,
	viewerBootstrap ViewerBootstrapReader,
) {
	if router == nil || writer == nil || viewerBootstrap == nil {
		return
	}

	router.PUT(
		"/api/fan/shorts/:shortId/pin",
		buildProtectedFanAuthGuard(viewerBootstrap, fanShortPinPutRequestScope, fanShortPinAuthRequiredMessage),
		func(c *gin.Context) {
			handleFanShortPinPut(c, writer)
		},
	)
	router.DELETE(
		"/api/fan/shorts/:shortId/pin",
		buildProtectedFanAuthGuard(viewerBootstrap, fanShortPinDeleteRequestScope, fanShortPinAuthRequiredMessage),
		func(c *gin.Context) {
			handleFanShortPinDelete(c, writer)
		},
	)
}

func handleFanShortPinPut(c *gin.Context, writer FanShortPinWriter) {
	handleFanShortPinMutation(c, writer.PinPublicShort, fanShortPinPutRequestScope)
}

func handleFanShortPinDelete(c *gin.Context, writer FanShortPinWriter) {
	handleFanShortPinMutation(c, writer.UnpinPublicShort, fanShortPinDeleteRequestScope)
}

func handleFanShortPinMutation(
	c *gin.Context,
	mutate func(context.Context, uuid.UUID, uuid.UUID) (shorts.PinMutationResult, error),
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

	c.JSON(http.StatusOK, responseEnvelope[fanShortPinResponseData]{
		Data: &fanShortPinResponseData{
			Viewer: feedViewerPayload{
				IsPinned: result.IsPinned,
			},
		},
		Meta: responseMeta{
			RequestID: newRequestID(requestScope),
			Page:      nil,
		},
		Error: nil,
	})
}
